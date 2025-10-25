package json

import (
	"fmt"
	"reflect"

	"github.com/frankli0324/go-jsontk"
)

// Unmarshal decodes JSON-encoded data and stores the result
// just like json.Unmarshal from the standard library.
// plain interface is currently not supported as a target type.
func Unmarshal(data []byte, into interface{}) error {
	var iter jsontk.Iterator
	iter.Reset(data)
	return writeVal(&iter, reflect.ValueOf(into))
}

var mapping = func() [10][26]bool {
	f := map[jsontk.TokenType][]reflect.Kind{
		jsontk.NULL:         {reflect.Slice, reflect.Map, reflect.Pointer},
		jsontk.STRING:       {reflect.String},
		jsontk.BEGIN_ARRAY:  {reflect.Array, reflect.Slice},
		jsontk.BEGIN_OBJECT: {reflect.Struct, reflect.Map},
		jsontk.BOOLEAN:      {reflect.Bool},
		jsontk.NUMBER: {
			reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64,
		},
	}
	rev := [10][26]bool{}
	for t, k := range f {
		for _, k := range k {
			rev[t][k] = true
		}
	}
	return rev
}()

func writeStruct(iter *jsontk.Iterator, v reflect.Value) error {
	sc := cachedStructIndex(v.Type())
	return iter.NextObject(func(key *jsontk.Token) bool {
		fn := key.String()
		f := v.FieldByIndex(sc[fn])
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

func writeMap(iter *jsontk.Iterator, v reflect.Value) (reflect.Value, error) {
	if v.IsNil() {
		v = reflect.MakeMap(v.Type())
	}
	keyType := v.Type().Key()
	valType := v.Type().Elem()
	err := iter.NextObject(func(key *jsontk.Token) bool {
		k := reflect.New(keyType).Elem()
		k.SetString(key.String())
		val := reflect.New(valType).Elem()
		if e := writeVal(iter, val); e != nil {
			iter.Error = e
			return false
		}
		v.SetMapIndex(k, val)
		return true
	})
	return v, err
}

func writeSlice(iter *jsontk.Iterator, v reflect.Value) (reflect.Value, error) {
	elem := v.Type().Elem()
	v = reflect.MakeSlice(v.Type(), 0, 3)
	return v, iter.NextArray(func(idx int) bool {
		val := reflect.New(elem).Elem()
		if err := writeVal(iter, val); err != nil {
			iter.Error = err
			return false
		}
		v = reflect.Append(v, val)
		return true
	})
}

func writeArray(iter *jsontk.Iterator, v reflect.Value) error {
	elemType := v.Type().Elem()
	length := v.Len()
	i := 0

	return iter.NextArray(func(idx int) bool {
		// If JSON array has more elements than Go array capacity — skip extras
		if i >= length {
			// skip remaining elements but keep consuming tokens
			iter.Skip()
			return true
		}

		elem := reflect.New(elemType).Elem()
		if err := writeVal(iter, elem); err != nil {
			iter.Error = err
			return false
		}
		v.Index(i).Set(elem)
		i++
		return true
	})
}

func writeVal(iter *jsontk.Iterator, f reflect.Value) error {
	if iter.Peek() == jsontk.NULL {
		iter.Next()
		return nil
	}
	if f.Kind() == reflect.Pointer {
		if f.IsNil() {
			f.Set(reflect.New(f.Type().Elem()))
		}
		return writeVal(iter, f.Elem())
	}
	if !mapping[iter.Peek()][f.Kind()] {
		return fmt.Errorf("can't assign %s to %s", iter.Peek().String(), f.Kind().String())
	}
	var tk jsontk.Token
	switch f.Kind() {
	case reflect.String:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid string")
		}
		f.SetString(tk.String())
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid number")
		}
		if num, err := tk.Number().Int64(); err == nil {
			f.SetInt(num)
		} else {
			return err
		}
	case reflect.Float32, reflect.Float64:
		iter.NextToken(&tk)
		if tk.Type == jsontk.INVALID {
			return fmt.Errorf("invalid number")
		}
		if num, err := tk.Number().Float64(); err == nil {
			f.SetFloat(num)
		} else {
			return err
		}
	case reflect.Slice:
		res, err := writeSlice(iter, f)
		if err != nil {
			return err
		}
		f.Set(res)
	case reflect.Array:
		return writeArray(iter, f)
	case reflect.Map:
		res, err := writeMap(iter, f)
		if err != nil {
			return err
		}
		f.Set(res)
	case reflect.Struct:
		return writeStruct(iter, f)
	default:
		iter.Skip()
	}
	return nil
}
