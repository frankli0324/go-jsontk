package tests

import (
	"fmt"
	"strings"
	"testing"

	_ "unsafe"

	"github.com/frankli0324/go-jsontk"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fastjson"
)

var benchPool fastjson.ParserPool

func benchmarkObjectGet(b *testing.B, itemsCount int) {
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
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			p := benchPool.Get()
			for pb.Next() {
				v, err := p.Parse(s)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				o := v.GetObject()
				sb := o.Get(key).GetStringBytes()
				if string(sb) != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", sb, expectedValue))
				}
			}
			benchPool.Put(p)
		})
	})
	b.Run("jsontk", func(b *testing.B) {
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				tks, err := jsontk.Tokenize(bs)
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
	b.Run("jsoniter", func(b *testing.B) {
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if v := jsoniter.Get(bs, key).ToString(); v != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", v, expectedValue))
				}
			}
		})
	})
	b.Run("jsontk-iterate", func(b *testing.B) {
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			iter := jsontk.Iterator{}
			for pb.Next() {
				iter.Reset(bs)
				iter.NextObject(func(k *jsontk.Token) bool {
					if k.EqualString(key) {
						iter.NextToken(k)
						if !k.EqualString(expectedValue) {
							b.Error(fmt.Errorf("unexpected value; got %q; want %q", k.Value, expectedValue))
						}
						return false
					}
					iter.Skip()
					return true
				})
			}
		})
	})
}

func BenchmarkObjectGet(b *testing.B) {
	for _, itemsCount := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", itemsCount), func(b *testing.B) {
			benchmarkObjectGet(b, itemsCount)
		})
	}
}
