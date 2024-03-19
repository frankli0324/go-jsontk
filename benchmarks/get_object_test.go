package tests

import (
	"encoding/json"
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
	bs := []byte(s)
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
			var tks = &jsontk.JSON{}
			var err error
			for pb.Next() {
				tks, err = jsontk.Tokenize(bs)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				v, _ := tks.Get(key).String()
				if v != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", v, expectedValue))
				}
				tks.Close()
			}
		})
	})
	b.Run("stdjson", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m := map[string]string{}
				err := json.Unmarshal(bs, &m)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				if m[key] != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", m[key], expectedValue))
				}
			}
		})
	})
}

func BenchmarkObjectGet(b *testing.B) {
	for _, itemsCount := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", itemsCount), func(b *testing.B) {
			for _, lookupsCount := range []int{64} {
				b.Run(fmt.Sprintf("lookups_%d", lookupsCount), func(b *testing.B) {
					benchmarkObjectGet(b, itemsCount, lookupsCount)
				})
			}
		})
	}
}
