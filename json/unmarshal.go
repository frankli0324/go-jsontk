package json

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/frankli0324/go-jsontk"
)

var iterPool = sync.Pool{New: func() any { return &jsontk.Iterator{} }}

// Unmarshal decodes JSON-encoded data and stores the result
// just like json.Unmarshal from the standard library.
// plain interface is currently not supported as a target type.
func Unmarshal(data []byte, into interface{}) error {
	iter := iterPool.Get().(*jsontk.Iterator)
	defer iterPool.Put(iter)
	iter.Reset(data)
	v := reflect.ValueOf(into)
	if v.Kind() != reflect.Pointer {
		return fmt.Errorf("must Unmarshal into a pointer")
	}
	return writeVal(iter, v.Elem())
}

func writeStruct(iter *jsontk.Iterator, v reflect.Value) error {
	sc := cachedStructIndex(v.Type())
	return iter.NextObject(func(key *jsontk.Token) bool {
		fn := key.String()
		field, ok := sc[fn]
		if !ok {
			iter.Skip()
			return true
		}
		f := v.FieldByIndex(field)
		if f.Kind() == reflect.Invalid {
			iter.Error = fmt.Errorf("invalid field %s", fn)
		}
		if err := writeVal(iter, f); err != nil {
			iter.Error = fmt.Errorf("%w for field %s", err, fn)
			return false
		}
		return true
	})
}

func writeMap(iter *jsontk.Iterator, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	keyType := v.Type().Key()
	valType := v.Type().Elem()
	mkey := reflect.New(keyType).Elem()
	return iter.NextObject(func(key *jsontk.Token) bool {
		mkey.SetString(key.String())
		val := reflect.New(valType).Elem()
		if err := writeVal(iter, val); err != nil {
			iter.Error = fmt.Errorf("%w for key %s", err, key.String())
			return false
		}
		v.SetMapIndex(mkey, val)
		return true
	})
}

func writeSlice(iter *jsontk.Iterator, v reflect.Value) error {
	vtyp := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeSlice(vtyp, 0, 4))
	}
	return iter.NextArray(func(idx int) bool {
		if idx >= v.Cap() {
			newcap := v.Cap() + v.Cap()/2
			newv := reflect.MakeSlice(vtyp, v.Len(), newcap)
			reflect.Copy(newv, v)
			v.Set(newv)
		}
		v.SetLen(idx + 1)
		if err := writeVal(iter, v.Index(idx)); err != nil {
			iter.Error = fmt.Errorf("%w at index %d", err, idx)
			return false
		}
		return true
	})
}

func writeArray(iter *jsontk.Iterator, v reflect.Value) error {
	length := v.Len()
	return iter.NextArray(func(idx int) bool {
		// If JSON array has more elements than Go array capacity — skip extras
		if idx >= length {
			// skip remaining elements but keep consuming tokens
			iter.Skip()
			return true
		}
		if err := writeVal(iter, v.Index(idx)); err != nil {
			iter.Error = err
			return false
		}
		return true
	})
}

var jsonUnmarshaller = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

func writeVal(iter *jsontk.Iterator, f reflect.Value) error {
	nxt, fkind, ftyp := iter.Peek(), f.Kind(), f.Type()
	if f.CanInterface() && ftyp.Implements(jsonUnmarshaller) {
		// TODO: json.Unmarshaler
	}
	if nxt == jsontk.NULL {
		if t, _, _ := iter.Next(); t != jsontk.NULL {
			if iter.Error != nil {
				return iter.Error
			}
			return fmt.Errorf("invalid jsontk internal state: expected null but got %s", t.String())
		}
		switch fkind {
		case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
		default:
			// this is intensional, std lib explicitly behaves like this!
			return nil
		}
		if f.IsNil() {
			return nil
		}
		if !f.CanSet() {
			return fmt.Errorf("can't assign %s to %s", nxt.String(), fkind.String())
		}
		f.Set(reflect.Zero(ftyp)) // f.SetZero is added in go1.20
		return nil
	}
	if !canUnmarshal[nxt][fkind] {
		return fmt.Errorf("can't assign %s to %s: type mismatch", nxt.String(), fkind.String())
	}
	var tk jsontk.Token
	switch fkind {
	case reflect.Interface:
		if f.IsNil() {
			if !f.CanSet() {
				return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
			}
			v := reflect.New(createInterface[nxt]).Elem()
			if err := writeVal(iter, v); err != nil {
				return err
			}
			f.Set(v)
		} else {
			if err := writeVal(iter, f.Elem()); err != nil {
				return err
			}
		}
	case reflect.Pointer:
		if f.IsNil() {
			if !f.CanSet() {
				return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
			}
			f.Set(reflect.New(ftyp.Elem()))
		}
		if err := writeVal(iter, f.Elem()); err != nil {
			return err
		}
	case reflect.String:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid string: %w", iter.Error)
		}
		s, ok := tk.UnquoteBytes()
		if !ok {
			return fmt.Errorf("invalid string: unquote failed")
		}
		if !f.CanSet() {
			return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
		}
		f.SetString(string(s))
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid number: %w", iter.Error)
		}
		if num, err := tk.Number().Int64(); err == nil {
			if !f.CanSet() {
				return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
			}
			f.SetInt(num)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid number: %w", iter.Error)
		}
		if num, err := tk.Number().Float64(); err == nil {
			if !f.CanSet() {
				return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
			}
			f.SetFloat(num)
		} else {
			return err
		}
	case reflect.Slice:
		if !f.CanSet() {
			return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
		}
		if err := writeSlice(iter, f); err != nil {
			return err
		}
	case reflect.Array:
		return writeArray(iter, f)
	case reflect.Map:
		if !f.CanSet() {
			return fmt.Errorf("unable to assign %s to %s", nxt.String(), fkind.String())
		}
		if err := writeMap(iter, f); err != nil {
			return err
		}
	case reflect.Struct:
		return writeStruct(iter, f)
	default:
		iter.Skip()
	}
	return nil
}

var canUnmarshal = func() [10][26]bool {
	rev := [10][26]bool{}
	for t, k := range map[jsontk.TokenType][]reflect.Kind{
		jsontk.STRING:       {reflect.String},
		jsontk.BEGIN_ARRAY:  {reflect.Array, reflect.Slice},
		jsontk.BEGIN_OBJECT: {reflect.Struct, reflect.Map},
		jsontk.BOOLEAN:      {reflect.Bool},
		jsontk.NUMBER: {
			reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64,
		},
	} {
		for _, k := range k {
			rev[t][k] = true
		}
	}
	for t, f := range createInterface {
		if f != nil {
			rev[t][reflect.Interface] = true
		}
	}
	for i := 0; i < 10; i++ {
		rev[i][reflect.Pointer] = true
	}
	return rev
}()

var createInterface = [10]reflect.Type{
	jsontk.STRING:       reflect.TypeOf(""),
	jsontk.BEGIN_ARRAY:  reflect.TypeOf([]interface{}{}),
	jsontk.BEGIN_OBJECT: reflect.TypeOf(map[string]interface{}{}),
	jsontk.BOOLEAN:      reflect.TypeOf(false),
	jsontk.NUMBER:       reflect.TypeOf(float64(0)),
}
