package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/frankli0324/go-jsontk"
	"github.com/valyala/fastjson"
)

func BenchmarkValidate(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkValidate(b, smallFixture)
	})
	b.Run("medium", func(b *testing.B) {
		benchmarkValidate(b, mediumFixture)
	})
	b.Run("large", func(b *testing.B) {
		benchmarkValidate(b, largeFixture)
	})
	b.Run("canada", func(b *testing.B) {
		benchmarkValidate(b, canadaFixture)
	})
	b.Run("citm", func(b *testing.B) {
		benchmarkValidate(b, citmFixture)
	})
	b.Run("twitter", func(b *testing.B) {
		benchmarkValidate(b, twitterFixture)
	})
}

func benchmarkValidate(b *testing.B, s string) {
	b.Run("stdjson", func(b *testing.B) {
		benchmarkValidateStdJSON(b, s)
	})
	b.Run("fastjson", func(b *testing.B) {
		benchmarkValidateFastJSON(b, s)
	})
	b.Run("jsontk", func(b *testing.B) {
		benchmarkValidateJSONtk(b, s)
	})
}

func benchmarkValidateStdJSON(b *testing.B, s string) {
	b.StopTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	d := unsafe.StringData(s)
	bb := unsafe.Slice(d, len(s))
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !json.Valid(bb) {
				panic("json.Valid unexpectedly returned false")
			}
		}
	})
}

func benchmarkValidateFastJSON(b *testing.B, s string) {
	b.StopTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := fastjson.Validate(s); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}

func benchmarkValidateJSONtk(b *testing.B, s string) {
	b.StopTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	d := unsafe.StringData(s)
	bb := unsafe.Slice(d, len(s))
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		var iter jsontk.Iterator
		for pb.Next() {
			iter.Reset(bb)
			if err := iter.Validate(); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}
