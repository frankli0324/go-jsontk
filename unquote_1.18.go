//go:build go1.18 && !go1.20

package jsontk

import (
	"bytes"
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
