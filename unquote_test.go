package jsontk

import (
	"bytes"
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func BenchmarkUnquote(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pairs := [][2][]byte{
			// {[]byte("\"\x80\""), []byte("\x80")},
			{[]byte(`"test"`), []byte("test")},
			{[]byte(`"\test"`), []byte("\test")},
			{[]byte(`"\\test"`), []byte("\\test")},
			{[]byte(`"\u0911est"`), []byte("ऑest")},
			{[]byte(`"test\u0911est"`), []byte("testऑest")},
			{[]byte(`"阿斯顿发个好借口了"`), []byte("阿斯顿发个好借口了")},
		}
		for _, cs := range pairs {
			if res, ok := unquoteBytes(cs[0]); !ok {
				b.Errorf("invalid, want:%s, has:%s", cs[1], res)
			} else if !bytes.Equal(res, cs[1]) {
				b.Errorf("unequal, want:%v, has:%v", cs[1], res)
			}
		}
	}
}

func BenchmarkUnquotedEqual(b *testing.B) {
	pairs := [][2][]byte{
		{[]byte("\"\x80\""), []byte("\x80")},
		{[]byte(`"test"`), []byte("test")},
		{[]byte(`"\test"`), []byte("\test")},
		{[]byte(`"\\test"`), []byte("\\test")},
		{[]byte(`"\u0911est"`), []byte("ऑest")},
		{[]byte(`"test\u0911est"`), []byte("testऑest")},
		{[]byte(`"阿斯顿发个好借口了"`), []byte("阿斯顿发个好借口了")},
	}
	for i := 0; i < b.N; i++ {
		for _, cs := range pairs {
			if !unquotedEqualStr(cs[0], string(cs[1])) {
				b.Errorf("unequal unquoted string, want:%s, has:%s", cs[1], cs[0])
			}
		}
	}
}
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

func TestEqualStringNoEscape(t *testing.T) {
	cases := []struct {
		s    []byte
		d    []byte
		want bool
	}{
		{[]byte(`"test"`), []byte("test"), true},
		{[]byte(`"test"`), []byte("nest"), false},
		{[]byte(`""`), []byte(""), true},
		{[]byte(`"\test"`), []byte(`\test`), true},  // backslash kept literally
		{[]byte(`"\test"`), []byte("\test"), false}, // unescaped form does not match
		{[]byte(`"a\"b"`), []byte(`a\"b`), true},
		{[]byte(`"a\"b"`), []byte(`a"b`), false},
		{[]byte(`"阿斯顿"`), []byte("阿斯顿"), true},
		{[]byte(`123`), []byte("123"), false},  // not a string token
		{[]byte(`"abc`), []byte("abc"), false}, // unterminated
	}
	for _, c := range cases {
		tk := Token{Type: STRING, Value: c.s}
		if got := tk.EqualStringNoEscape(string(c.d)); got != c.want {
			t.Errorf("EqualStringNoEscape(%q, %q) = %v, want %v", c.s, c.d, got, c.want)
		}
	}
}

func BenchmarkEqualStringNoEscape(b *testing.B) {
	pairs := [][2][]byte{
		{[]byte(`"test"`), []byte("test")},
		{[]byte(`"key_5000"`), []byte("key_5000")},
		{[]byte(`"阿斯顿发个好借口了"`), []byte("阿斯顿发个好借口了")},
	}
	for i := 0; i < b.N; i++ {
		for _, cs := range pairs {
			tk := Token{Type: STRING, Value: cs[0]}
			if !tk.EqualStringNoEscape(string(cs[1])) {
				b.Errorf("unequal raw string, want:%s, has:%s", cs[1], cs[0])
			}
		}
	}
}

func FuzzUnquotedEqual(f *testing.F) {
	f.Add([]byte(`""`))
	f.Add([]byte(`"abc"`))
	f.Add([]byte(`"\n\t\r"`))
	f.Add([]byte(`"\\\""`))
	f.Add([]byte(`"\u0000"`))
	f.Add([]byte(`"\uD800"`))
	f.Add([]byte(`"\uD800\uDC00"`))
	f.Add([]byte(`"\uFFFF"`))
	f.Add([]byte(`"\uZZZZ"`))
	f.Add([]byte(`"\u12"`))
	f.Add([]byte(`"\\"`))
	f.Add([]byte(`"阿巴阿巴"`))

	f.Fuzz(func(t *testing.T, s []byte) {
		if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
			return
		}
		out, ok := unquoteBytes(s)
		if ok {
			_, _ = unquoteBytes(append([]byte{}, s...))
			if bytes.ContainsRune(out, utf8.RuneError) {
				return
			}
			if !unquotedEqual(s, out) {
				t.Fatalf("inconsistent:\ninput: %q\nunquoted: %q", s, out)
			}
		}
		var std string
		if json.Unmarshal(s, &std) == nil && (!ok || string(out) != std) {
			t.Fatalf("mismatch against std: %q", s)
		}
	})
}
