package jsontk

import (
	"strconv"
	"unsafe"
)

func unquoteBytes(b []byte) (string, error) {
	return strconv.Unquote(*(*string)(unsafe.Pointer(&b)))
}
