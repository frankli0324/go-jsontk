package jsontk

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestTokenize(t *testing.T) {
	res, err := Tokenize([]byte(`{"test":1,"xx":true,
	"vv": // test
	false}`))
	fmt.Println(err)
	for _, tk := range res {
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
				t.Fail()
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
