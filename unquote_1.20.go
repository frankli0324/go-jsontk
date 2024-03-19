//go:build go1.20

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
		return unsafe.String(unsafe.SliceData(s[1:len(s)-1]), len(s)-2), true
	}
	s, ok = unquoteBytes(s)
	t = unsafe.String(unsafe.SliceData(s), len(s))
	return
}
