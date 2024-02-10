package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/frankli0324/go-jsontk"
	"github.com/valyala/fastjson"
)

var benchPool fastjson.ParserPool

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

	b.Run("fastjson", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			p := benchPool.Get()
			for pb.Next() {
				v, err := p.Parse(s)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				o := v.GetObject()
				for i := 0; i < lookupsCount; i++ {
					sb := o.Get(key).GetStringBytes()
					if string(sb) != expectedValue {
						panic(fmt.Errorf("unexpected value; got %q; want %q", sb, expectedValue))
					}
				}
			}
			benchPool.Put(p)
		})
	})
	b.Run("jsontk", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			tks, err := jsontk.Tokenize([]byte(s))
			if err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
			for pb.Next() {
				v, _ := tks.Get(key).String()
				if v != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", v, expectedValue))
				}
			}
		})
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
