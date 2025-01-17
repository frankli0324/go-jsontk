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
	f.Add([]byte(`"\r\n test\t"`))

	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) <= 2 || b[0] != '"' || b[len(b)-1] != '"' {
			return
		}
		actual, ok := unquoteBytes(b)
		if !ok || bytes.ContainsRune(actual, utf8.RuneError) {
			return
		}
		t.Log(string(b), string(actual))
		if !bytes.Equal([]byte("test"), actual) {
			if unquotedEqual(b, []byte("test")) {
				t.Error("should not equal test")
				t.Fail()
			}
			if unquotedEqual([]byte(`"test"`), actual) {
				t.Error("test should not equal " + string(actual))
				t.Fail()
			}
		}
		if !unquotedEqual(b, actual) {
			t.Fail()
		}
	})
}
