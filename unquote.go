package jsontk

import (
	"bytes"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var escapeChars = [10]byte{
	('b' - 'b') >> 1: '\b', //lint:ignore SA4000 consistency
	('f' - 'b') >> 1: '\f',
	('n' - 'b') >> 1: '\n',
	('r' - 'b') >> 1: '\r',
	('t' - 'b') >> 1: '\t',
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

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b', 'f', 'n', 'r', 't':
				b[w] = escapeChars[(s[r]-'b')>>1]
				r++
				w++
			case 'u':
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
						if rr != unicode.ReplacementChar {
							r += 6
						}
					}
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// control characters are invalid.
		case c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// unquotedEqual returns unquoteBytes(s) == d without heap allocations
// it assumes that quote escape is always correctly handled, so it won't
// complain about unescaped quotes ("te"st" -> te"st)
func unquotedEqual(s, d []byte) (equal bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	s = s[1 : len(s)-1]

	ps := bytes.IndexByte(s, '\\')
	if ps == -1 {
		return bytes.Equal(s, d)
	}

	pd := ps
	if pd > len(d) || !bytes.Equal(s[:ps], d[:pd]) {
		return false
	}
	for ps < len(s) && pd < len(d) {
		switch c := s[ps]; {
		case c == '\\':
			ps++
			if ps == len(s) {
				return
			}
			switch s[ps] {
			default:
				return
			case '"', '\\', '/', '\'':
				if d[pd] != s[ps] {
					return false
				}
				ps++
				pd++
			case 'b', 'f', 'n', 'r', 't':
				if d[pd] != escapeChars[(s[ps]-'b')>>1] {
					return false
				}
				ps++
				pd++
			case 'u':
				if ps+5 > len(s) {
					return
				}
				rr := getu4(s[ps+1 : ps+5])
				if rr < 0 {
					return
				}
				ps += 5
				if utf16.IsSurrogate(rr) {
					if s[ps] != '\\' || s[ps+1] != 'u' || ps+6 > len(s) {
						return
					}
					rr1 := getu4(s[ps+2 : ps+6])
					rr = utf16.DecodeRune(rr, rr1)
					if rr != unicode.ReplacementChar {
						ps += 6
					}
				}
				r, sz := utf8.DecodeRune(d[pd:])
				if rr != r {
					// could be RuneError, but wouldn't be equal anyway
					return
				}
				pd += sz
			}
		// control characters are invalid.
		case c < ' ':
			return
		default:
			if s[ps] != d[pd] {
				return
			}
			ps, pd = ps+1, pd+1
		}
	}
	return ps == len(s) && pd == len(d)
}
