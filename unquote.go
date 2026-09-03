package jsontk

import (
	"bytes"
	"encoding/binary"
	"math/bits"
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
const (
	loBytes = 0x0101010101010101
	hiBytes = 0x8080808080808080
	slash8  = uint64('\\') * loBytes
)

func unquotedEqual(s, d []byte) bool {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	s = s[1 : len(s)-1]
loop:
	if len(d) > len(s) {
		return false
	}
	if len(s) >= 24 {
		r := bytes.IndexByte(s, '\\')
		if r < 0 {
			return bytes.Equal(s, d)
		}
		if r >= len(d) || r+1 >= len(s) {
			return false
		}
		if !bytes.Equal(s[:r], d[:r]) {
			return false
		}
		s = s[r:]
		d = d[r:]
		goto escape
	}
	for i := 0; i+8 <= len(d); i += 8 {
		vs := binary.LittleEndian.Uint64(s[i:])
		x := vs ^ slash8
		z := (x - loBytes) &^ x & hiBytes
		if z == 0 { // no backslash in this chunk: decoded == raw
			if vs != binary.LittleEndian.Uint64(d[i:]) {
				return false
			}
			continue
		}
		slash := bits.TrailingZeros64(z) >> 3
		if vd := binary.LittleEndian.Uint64(d[i:]); vs != vd {
			mismatch := bits.TrailingZeros64(vs^vd) >> 3
			if mismatch < slash {
				return false
			}
		}
		i += slash
		s = s[i:]
		d = d[i:]
		goto escape
	}
	for i := 0; i < len(d); i++ {
		if s[i] == '\\' {
			s = s[i:]
			d = d[i:]
			goto escape
		}
		if s[i] != d[i] {
			return false
		}
	}
	return len(s) == len(d)

escape:
	if len(s) < 2 || len(d) == 0 {
		return false
	}
	switch esc := escapeChars[s[1]]; esc {
	case 0:
		return false
	default:
		if d[0] != esc {
			return false
		}
		s = s[2:]
		d = d[1:]
	case 0xff:
		dr, sz := utf8.DecodeRune(d)
		if dr == utf8.RuneError || len(s) < 6 {
			return false
		}
		sr := getu4(s[2:6])
		if sr < 0 {
			return false
		}
		if utf16.IsSurrogate(sr) {
			if len(s) < 12 || s[6] != '\\' || s[7] != 'u' {
				return false
			}
			sr = utf16.DecodeRune(sr, getu4(s[8:12]))
			if sr == unicode.ReplacementChar {
				return false
			}
			s = s[12:]
		} else {
			s = s[6:]
		}
		if sr != dr {
			return false
		}
		d = d[sz:]
	}
	goto loop
}
