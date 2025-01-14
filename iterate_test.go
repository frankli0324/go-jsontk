package jsontk

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"
)

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
	valList := [][]byte{nil, b(`"test"`), b("1"), b(`"xx"`), b("true"), b(`"vv"`), b("false"), nil}
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
