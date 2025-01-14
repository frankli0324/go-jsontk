package jsontk

import (
	"bytes"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// This file was copied from encoding/json library.
// (c) Golang: encoding/json/decode.go

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

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
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
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
func unquotedEqual(s, d []byte) (equal bool) {
	s = s[1 : len(s)-1]

	ps := 0
	for ps < len(s) {
		c := s[ps]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			ps++
			continue
		}
		rr, size := utf8.DecodeRune(s[ps:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		ps += size
	}
	if ps == len(s) {
		return bytes.Equal(s, d)
	}

	pd := ps
	if pd > len(d) || pd > 0 && !bytes.Equal(s[:ps], d[:pd]) {
		return false
	}
	for ps < len(s) {
		switch c := s[ps]; {
		case c == '\\':
			ps++
			if ps >= len(s) || pd >= len(d) {
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
			case 'b':
				if d[pd] != '\b' {
					return false
				}
				ps++
				pd++
			case 'f':
				if d[pd] != '\f' {
					return false
				}
				ps++
				pd++
			case 'n':
				if d[pd] != '\n' {
					return false
				}
				ps++
				pd++
			case 'r':
				if d[pd] != '\r' {
					return false
				}
				ps++
				pd++
			case 't':
				if d[pd] != '\t' {
					return false
				}
				ps++
				pd++
			case 'u':
				rr := getu4(s[ps-1:])
				if rr < 0 {
					return
				}
				ps += 5
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[ps:])
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
		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return
		}
		psn := ps
		for ps < len(s) {
			c := s[ps]
			if c == '\\' || c == '"' || c < ' ' {
				break
			}

			if c < utf8.RuneSelf {
				ps++
				continue
			}
			rr, size := utf8.DecodeRune(s[ps:])
			ps += size
			if rr == utf8.RuneError && size == 1 {
				break
			}
		}
		if psn < ps && !bytes.Equal(s[psn:ps], d[pd:pd+ps-psn]) {
			return
		}
		pd += ps - psn
	}
	return true
}
