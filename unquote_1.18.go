//go:build go1.18 && !go1.20

package jsontk

import (
	"reflect"
	"unsafe"
)

func unquote(s []byte) (t string, ok bool) {
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
