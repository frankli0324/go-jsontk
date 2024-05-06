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
			tks := &jsontk.JSON{}
			for pb.Next() {
				err := tks.Tokenize(bs)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				for i := 0; i < lookupsCount; i++ {
					v, _ := tks.Get(key).String()
					if v != expectedValue {
						panic(fmt.Errorf("unexpected value; got %q; want %q", v, expectedValue))
					}
				}
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
	if lookupsCount != 1 {
		b.Log("lookups Count not 1, not running jsontk-iterate benchmark since it's not meaningful")
		return
	}
	b.Run("jsontk-iterate", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				want := false
				if err := jsontk.Iterate(bs, func(typ jsontk.TokenType, idx, len int) {
					if want && (typ != jsontk.STRING || string(bs[idx:idx+len]) != expectedValue) {
						panic(fmt.Errorf("unexpected value; got %q; want %q", string(bs[idx:idx+len]), expectedValue))
					}
					if typ == jsontk.KEY && string(bs[idx:idx+len]) == key {
						want = true
					}
				}); err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
			}
		})
	})
}

func BenchmarkObjectGet(b *testing.B) {
	for _, itemsCount := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", itemsCount), func(b *testing.B) {
			benchmarkObjectGet(b, itemsCount, 1)
			benchmarkObjectGet(b, itemsCount, 10)
		})
	}
}
