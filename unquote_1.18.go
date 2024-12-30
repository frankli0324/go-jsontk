//go:build go1.18 && !go1.20

package jsontk

import (
	"bytes"
	"reflect"
	"unsafe"
)

func unquote(s []byte) (t string, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	if bytes.IndexByte(s, '\\') < 0 {
		s = s[1 : len(s)-1]
		return *(*string)(unsafe.Pointer(&s)), true
	}
	s, ok = unquoteBytes(s)
	t = *(*string)(unsafe.Pointer(&s))
	return
}

func unquotedEqualStr(s []byte, t string) bool {
	const MaxInt32 = 1<<31 - 1
	d := (*[MaxInt32]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&t)).Data),
	)[:len(t):len(t)]
	return unquotedEqual(s, d)
}
