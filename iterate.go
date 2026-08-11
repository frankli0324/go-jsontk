package jsontk

import "bytes"

var numByte = [256]bool{
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
	'.': true, 'e': true, 'E': true, '+': true, '-': true,
}

func skip(s []byte, i int) int {
	for i < len(s) {
		switch s[i] {
		case ' ', '\n', '\t', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

func next(s []byte, i int) (typ TokenType, length int, err error) {
	if len(s) <= i {
		return INVALID, 0, ErrEarlyEOF
	}
	switch s[i] {
	case '"':
		for j := i + 1; ; j++ {
			q := bytes.IndexByte(s[j:], '"')
			if q < 0 {
				return INVALID, 0, ErrEarlyEOF.at(i, "expected end of string")
			}
			j += q
			for q = j - 1; q > i && s[q] == '\\'; q-- {
			}
			if (j-q)&1 == 1 { // even backslash run: quote closes the string
				return STRING, j - i + 1, nil
			}
		}
	case '{':
		return BEGIN_OBJECT, 1, nil
	case '[':
		return BEGIN_ARRAY, 1, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		j := i + 1
		for ; j < len(s) && numByte[s[j]]; j++ {
		}
		return NUMBER, j - i, nil
	case 't':
		if len(s)-i < 4 || string(s[i:i+4]) != "true" {
			return INVALID, 0, ErrUnexpectedToken.at(i, "expected boolean")
		}
		return BOOLEAN, 4, nil
	case 'f':
		if len(s)-i < 5 || string(s[i+1:i+5]) != "alse" {
			return INVALID, 0, ErrUnexpectedToken.at(i, "expected boolean")
		}
		return BOOLEAN, 5, nil
	case 'n':
		if len(s)-i < 4 || string(s[i:i+4]) != "null" {
			return INVALID, 0, ErrUnexpectedToken.at(i, "expected null")
		}
		return NULL, 4, nil
	case '}':
		return END_OBJECT, 1, nil
	case ']':
		return END_ARRAY, 1, nil
	default:
		return INVALID, 0, ErrUnexpectedToken
	}
}
