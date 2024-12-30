package jsontk

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestUnquotedEqual(t *testing.T) {
	pairs := [][2][]byte{
		{[]byte("\"\x80\""), []byte("\x80")},
		{[]byte(`"test"`), []byte("test")},
		{[]byte(`"\test"`), []byte("\test")},
		{[]byte(`"\\test"`), []byte("\\test")},
		{[]byte(`"\u0911est"`), []byte("ऑest")},
		{[]byte(`"test\u0911est"`), []byte("testऑest")},
		{[]byte(`"阿斯顿发个好借口了"`), []byte("阿斯顿发个好借口了")},
	}
	for _, cs := range pairs {
		if !unquotedEqual(cs[0], cs[1]) {
			t.Errorf("unequal unquoted string, want:%s, has:%s", cs[1], cs[0])
		}
	}
}

func FuzzUnquotedEqual(f *testing.F) {
	f.Add([]byte(`"test"`))

	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) <= 2 || b[0] != '"' || b[len(b)-1] != '"' {
			return
		}
		actual, ok := unquoteBytes(b)
		if !ok || bytes.ContainsRune(actual, utf8.RuneError) {
			return
		}
		t.Log(b, actual)
		if !unquotedEqual(b, actual) {
			t.Fail()
		}
	})
}
