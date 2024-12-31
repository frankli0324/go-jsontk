package jsontk

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
)

func TestTokenize(t *testing.T) {
	res, err := Tokenize([]byte(`{"test":1,"xx":true,
	"vv"
	: false}`))
	if err != nil {
		t.Error(err)
	}
	for _, tk := range res.store {
		fmt.Printf("%s->%s\n", tk.Type.String(), string(tk.Value))
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

func TestArrayIndex(t *testing.T) {
	file, _ := os.ReadFile(path.Join("./testdata", "citm_catalog.json"))
	j, _ := Tokenize(file)
	i, _ := j.Get("events").Get("138586341").Get("subTopicIds").Index(1).Int64()
	if i != 337184283 {
		t.Fail()
	}
}

func TestString(t *testing.T) {
	j, _ := Tokenize([]byte(`{"test":"zxcv"}`))
	i, _ := j.Get("test").String()
	if i != "zxcv" {
		t.Fail()
	}
	j, _ = Tokenize([]byte(`{"test":"z\txcv"}`))
	i, _ = j.Get("test").String()
	if i != "z\txcv" {
		t.Fail()
	}
}

func TestValidate(t *testing.T) {
	v := []byte(`[null, 1, "1", {}]`)
	fmt.Println(string(v))
	j, err := Tokenize(v)
	fmt.Println(j.store, err)

	if err := j.Validate(); err != nil {
		t.Error(err)
	}

	v = []byte(`[null, 1, "1\t", {"a":1}],1234`)
	fmt.Println(string(v))
	j, _ = Tokenize(v)
	fmt.Println(j.store)

	if err := j.Validate(); err == nil {
		t.Error(errors.New("should have failed in Validation"))
	}
}
