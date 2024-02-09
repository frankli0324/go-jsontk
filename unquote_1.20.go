//go:build go1.20

package jsontk

import "unsafe"

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = unsafe.String(unsafe.SliceData(s), len(s))
	return
}
