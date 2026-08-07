package jsontk

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"
)

func b(s ...string) [][]byte {
	ret := make([][]byte, len(s))
	for i, s := range s {
		if s == "" {
			ret[i] = nil
		} else {
			ret[i] = []byte(s)
		}
	}
	return ret
}

func Iterate(s []byte, cb func(typ TokenType, idx, len int)) error {
	hadComma, wantComma := false, false

	for i := 0; i < len(s); {
		i = skip(s, i)

		currentType, length, errOnce := next(s, i)

		start := i
		// prepare for lookahead, consume until next char is valid
		i = skip(s, i+length)
		if i < len(s) && s[i] == ':' {
			if currentType == STRING {
				currentType = KEY
				i++
			} else {
				return ErrUnexpectedToken
			}
		}

		if currentType == END_ARRAY || currentType == END_OBJECT {
			// intensionally don't check for previous comma
			// if StrictComma && hadComma {
			// 	return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start-1)
			// }
		} else if wantComma && !hadComma {
			return ErrUnexpectedSep
		} else if !wantComma && hadComma {
			return ErrUnexpectedSep
		}
		wantComma = commaAfterToken[currentType]
		hadComma = i < len(s) && s[i] == ','
		if hadComma {
			i++
		}

		cb(currentType, start, length)
		if errOnce != nil {
			return fmt.Errorf("%w at %d", errOnce, start)
		}
	}
	// if StrictComma && hadComma {
	// 	return fmt.Errorf("%w at end, unexpected comma", ErrUnexpectedSep)
	// }
	return nil
}

func Tokenize(s []byte) ([]Token, error) {
	store := []Token{}
	return store, Iterate(s, func(typ TokenType, idx, len int) {
		switch typ {
		case BEGIN_ARRAY, END_ARRAY, BEGIN_OBJECT, END_OBJECT, NULL:
			store = append(store, Token{Type: typ})
		default:
			store = append(store, Token{Type: typ, Value: s[idx : idx+len]})
		}
	})
}

func TestIterate(t *testing.T) {
	res, err := Tokenize([]byte(`{"test":1,"xx":true,
	"vv"
	: false}`))
	typList := []TokenType{BEGIN_OBJECT, KEY, NUMBER, KEY, BOOLEAN, KEY, BOOLEAN, END_OBJECT}
	valList := b("", `"test"`, "1", `"xx"`, "true", `"vv"`, "false", "")
	if err != nil {
		t.Error(err)
	}
	for i, tk := range res {
		if typList[i] != tk.Type {
			t.Errorf("invalid type at idx %d", i)
		}
		if !bytes.Equal(valList[i], tk.Value) {
			t.Errorf("invalid value at idx %d", i)
		}
	}
}

// test cases taken from https://github.com/valyala/fastjson
func TestJSONDatasets(t *testing.T) {
	entries, _ := os.ReadDir("./testdata")
	for _, ent := range entries {
		if !ent.IsDir() {
			file, _ := os.ReadFile(path.Join("./testdata", ent.Name()))
			fmt.Println(path.Join("./testdata", ent.Name()), len(file))
			_, err := Tokenize(file)
			if err != nil {
				t.Error(err)
			}
		}
	}
	// for _, tk := range res {
	// 	fmt.Printf("%s->%s\n", tk.Type.String(), string(tk.Value))
	// }
}
