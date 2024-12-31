//go:build go1.20

package jsontk

import (
	"unsafe"
)

func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = unsafe.String(unsafe.SliceData(s), len(s))
	return
}

func unquotedEqualStr(s []byte, t string) bool {
	d := unsafe.StringData(t)
	b := unsafe.Slice(d, len(t))
	return unquotedEqual(s, b)
}
