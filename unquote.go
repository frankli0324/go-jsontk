package jsontk

import (
	"bytes"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var escapeChars = [256]byte{
	'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t',
	'"': '"', '/': '/', '\'': '\'', '\\': '\\', 'u': 0xff,
}

var u4map = func() (r [256]rune) {
	for i := 0; i < 256; i++ {
		r[i] = -1
	}
	for i := '0'; i <= '9'; i++ {
		r[i] = i - '0'
	}
	for i := 'A'; i <= 'F'; i++ {
		r[i] = i - 'A' + 10
	}
	for i := 'a'; i <= 'f'; i++ {
		r[i] = i - 'a' + 10
	}
	return
}()

// This file was taken and modified from encoding/json library.
// (c) Golang: encoding/json/decode.go

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) (r rune) {
	for _, c := range s {
		r = (r << 4) | u4map[c]
	}
	return
}

// unquoteBytes unquotes json strings
// it assumes that quote escape is always correctly handled, so it won't
// complain about unescaped quotes ("te"st" -> te"st)
func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]
	r := bytes.IndexByte(s, '\\')
	if r == -1 {
		return s, true
	}

	b := make([]byte, len(s))
	w := 0
	for r != -1 {
		w += copy(b[w:], s[:r])
		r++
		if r >= len(s) {
			return
		}
		switch c := escapeChars[s[r]]; c {
		default:
			b[w] = c
			r++
			w++
		case 0:
			return
		case 0xff:
			if r+5 > len(s) {
				return
			}
			rr := getu4(s[r+1 : r+5])
			if rr < 0 {
				return
			}
			r += 5
			if utf16.IsSurrogate(rr) {
				if r+6 > len(s) || s[r] != '\\' || s[r+1] != 'u' {
					rr = unicode.ReplacementChar
				} else {
					rr = utf16.DecodeRune(rr, getu4(s[r+2:r+6]))
					r += 6
				}
			}
			w += utf8.EncodeRune(b[w:], rr)
		}
		if r == len(s) {
			return b[:w], true
		}
		s = s[r:]
		if s[0] == '\\' {
			r = 0
			continue
		}
		r = bytes.IndexByte(s, '\\')
	}
	w += copy(b[w:], s)
	return b[:w], true
}

// unquotedEqual returns unquoteBytes(s) == d without heap allocations
// it assumes that quote escape is always correctly handled, so it won't
// complain about unescaped quotes ("te"st" -> te"st)
func unquotedEqual(s, d []byte) (equal bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	s = s[1 : len(s)-1]

	r := bytes.IndexByte(s, '\\')
	for r != -1 {
		if r >= len(d) || r+1 == len(s) {
			return
		}
		if !bytes.Equal(s[:r], d[:r]) {
			return
		}
		switch esc := escapeChars[s[r+1]]; esc {
		default:
			if d[r] != esc {
				return
			}
			d = d[r+1:]
			r += 2
		case 0:
			return
		case 0xff:
			drune, sz := utf8.DecodeRune(d[r:])
			d = d[r+sz:]
			if drune == utf8.RuneError || r+6 > len(s) {
				return
			}
			srune := getu4(s[r+2 : r+6])
			if srune < 0 {
				return
			}
			r += 6
			if utf16.IsSurrogate(srune) {
				if r+6 > len(s) || s[r] != '\\' || s[r+1] != 'u' {
					return
				}
				srune = utf16.DecodeRune(srune, getu4(s[r+2:r+6]))
				if srune == unicode.ReplacementChar {
					return
				}
				r += 6
			}
			if srune != drune {
				return
			}
		}
		if r == len(s) {
			return len(d) == 0
		}
		s = s[r:]
		if s[0] == '\\' {
			r = 0
			continue
		}
		r = bytes.IndexByte(s, '\\')
	}
	return bytes.Equal(s, d)
}
