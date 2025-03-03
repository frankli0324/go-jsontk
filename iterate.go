package jsontk

import (
	"fmt"
)

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

func eos(s []byte, i int) int {
	isEscaped := false
	for ; i < len(s) && (s[i] != '"' || isEscaped); i++ {
		if s[i] == '\\' && !isEscaped {
			isEscaped = true
		} else {
			isEscaped = false
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
		j := i + 1
		for ; j < len(s) && s[j] != '"'; j++ {
			if s[j] == '\\' {
				j = eos(s, j)
				break
			}
		}
		if j == len(s) {
			return INVALID, 0, fmt.Errorf("%w, expected end of string", ErrEarlyEOF)
		}
		return STRING, j - i + 1, nil
	case '{':
		return BEGIN_OBJECT, 1, nil
	case '[':
		return BEGIN_ARRAY, 1, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		j := i + 1
		for ; j < len(s); j++ {
			if s[j] >= '0' && s[j] <= '9' {
				continue
			}
			if s[j] == '.' || s[j] == 'e' || s[j] == 'E' || s[j] == '+' || s[j] == '-' {
				continue
			}
			break
		}
		return NUMBER, j - i, nil
	case 't':
		if len(s)-i < 4 || s[i+1] != 'r' || s[i+2] != 'u' || s[i+3] != 'e' {
			return INVALID, 0, fmt.Errorf("%w, expected boolean", ErrUnexpectedToken)
		}
		return BOOLEAN, 4, nil
	case 'f':
		if len(s)-i < 5 || s[i+1] != 'a' || s[i+2] != 'l' || s[i+3] != 's' || s[i+4] != 'e' {
			return INVALID, 0, fmt.Errorf("%w, expected boolean", ErrUnexpectedToken)
		}
		return BOOLEAN, 5, nil
	case 'n':
		if len(s)-i < 4 || s[i+1] != 'u' || s[i+2] != 'l' || s[i+3] != 'l' {
			return INVALID, 0, fmt.Errorf("%w, expected null", ErrUnexpectedToken)
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

func Iterate(s []byte, cb func(typ TokenType, idx, len int)) error {
	hadComma, wantComma := false, false

	for i := 0; i < len(s); {
		i = skip(s, i)

		currentType, length, errOnce := next(s, i)

		start := i
		// prepare for lookahead, consume until next char is valid
		i = skip(s, i+length)
		if i < len(s) && s[i] == ':' {
			if currentType == STRING {
				currentType = KEY
				i++
			} else {
				return fmt.Errorf("%w at %d, expected string key", ErrUnexpectedToken, i)
			}
		}

		if currentType == END_ARRAY || currentType == END_OBJECT {
			// intensionally don't check for previous comma
			// if StrictComma && hadComma {
			// 	return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start-1)
			// }
		} else if wantComma && !hadComma {
			return fmt.Errorf("%w at %d, expected comma", ErrUnexpectedSep, start)
		} else if !wantComma && hadComma {
			return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start-1)
		}
		wantComma = commaAfterToken[currentType]
		hadComma = i < len(s) && s[i] == ','
		if hadComma {
			i++
		}

		cb(currentType, start, length)
		if errOnce != nil {
			return fmt.Errorf("%w at %d", errOnce, start)
		}
	}
	// if StrictComma && hadComma {
	// 	return fmt.Errorf("%w at end, unexpected comma", ErrUnexpectedSep)
	// }
	return nil
}
