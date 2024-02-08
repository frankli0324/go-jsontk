package jsontk

import (
	"fmt"
	"os"
	"path"
	"strings"
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

func benchmarkObjectGet(b *testing.B, itemsCount, lookupsCount int) {
	b.StopTimer()
	var ss []string
	for i := 0; i < itemsCount; i++ {
		s := fmt.Sprintf(`"key_%d": "value_%d"`, i, i)
		ss = append(ss, s)
	}
	s := "{" + strings.Join(ss, ",") + "}"
	key := fmt.Sprintf("key_%d", len(ss)/2)
	expectedValue := fmt.Sprintf("value_%d", len(ss)/2)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))

	b.RunParallel(func(pb *testing.PB) {
		tks, err := Tokenize([]byte(s))
		if err != nil {
			panic(fmt.Errorf("unexpected error: %s", err))
		}
		for pb.Next() {
			for i := range tks {
				k, _ := unquoteBytes(tks[i].Value)
				if tks[i].Type == KEY && k == key {
					v, _ := unquoteBytes(tks[i+1].Value)
					if v != expectedValue {
						panic(fmt.Errorf("unexpected value; got %q; want %q", string(tks[i+1].Value), expectedValue))
					}
				}
			}
		}
	})
}

func BenchmarkObjectGet(b *testing.B) {
	for _, itemsCount := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", itemsCount), func(b *testing.B) {
			for _, lookupsCount := range []int{0, 1, 2, 4, 8, 16, 32, 64} {
				b.Run(fmt.Sprintf("lookups_%d", lookupsCount), func(b *testing.B) {
					benchmarkObjectGet(b, itemsCount, lookupsCount)
				})
			}
		})
	}
}
