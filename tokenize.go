package jsontk

import (
	"fmt"
)

func skip(s []byte) (i int) {
	for i < len(s) {
		switch s[i] {
		case ' ', '\n', '\t', '\r':
			i++
		case '/':
			if i+1 >= len(s) {
				return i
			}
			switch s[i+1] {
			case '/':
				for i < len(s) && s[i] != '\r' && s[i] != '\n' {
					i++
				}
				continue
			case '*':
				for i+1 < len(s) && (s[i] != '*' || s[i+1] != '/') {
					i++
				}
				if i+2 == len(s) {
					return i + 2
				} else {
					i += 2
					continue
				}
			default:
				return i
			}
		default:
			return i
		}
	}
	return i
}

func next(s []byte) (TokenType, int, error) {
	if len(s) == 0 {
		return INVALID, 0, ErrEarlyEOF
	}
	switch s[0] {
	case '{':
		return BEGIN_OBJECT, 1, nil
	case '}':
		return END_OBJECT, 1, nil
	case '[':
		return BEGIN_ARRAY, 1, nil
	case ']':
		return END_ARRAY, 1, nil
	case '"':
		i := 1
		isEscaped := false
		for ; i < len(s) && (s[i] != '"' || isEscaped); i++ {
			if s[i] == '\\' && !isEscaped {
				isEscaped = true
			} else {
				isEscaped = false
			}
		}
		if i == len(s) {
			return INVALID, i, fmt.Errorf("%w, expected end of string", ErrEarlyEOF)
		}
		return STRING, i + 1, nil
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		i := 1
		for ; i < len(s); i++ {
			if s[i] >= '0' && s[i] <= '9' {
				continue
			}
			if s[i] == '.' || s[i] == 'e' || s[i] == 'E' || s[i] == '+' || s[i] == '-' {
				continue
			}
			break
		}
		return NUMBER, i, nil
	case 't':
		if len(s) < 4 || s[1] != 'r' || s[2] != 'u' || s[3] != 'e' {
			return INVALID, len(s), fmt.Errorf("%w, expected boolean", ErrUnexpectedToken)
		}
		return BOOLEAN, 4, nil
	case 'f':
		if len(s) < 5 || s[1] != 'a' || s[2] != 'l' || s[3] != 's' || s[4] != 'e' {
			return INVALID, len(s), fmt.Errorf("%w, expected boolean", ErrUnexpectedToken)
		}
		return BOOLEAN, 5, nil
	case 'n':
		if len(s) < 4 || s[1] != 'u' || s[2] != 'l' || s[3] != 'l' {
			return INVALID, len(s), fmt.Errorf("%w, expected null", ErrUnexpectedToken)
		}
		return NULL, 4, nil
	default:
		return INVALID, 1, ErrUnexpectedToken
	}
}

type JSON []Token

func TokenizeInto(s []byte, result JSON) (JSON, error) {
	hadComma, wantComma := false, false

	for i := 0; i < len(s); {
		if i < len(s) {
			i += skip(s[i:])
		}
		currentType, length, errOnce := next(s[i:])
		start := i

		i += length
		// prepare for lookahead, consume until next char is valid
		if i < len(s) {
			i += skip(s[i:])
		}
		if i < len(s) && s[i] == ':' {
			if currentType == STRING {
				currentType = KEY
				i++
			} else {
				return result, fmt.Errorf("%w at %d, expected string key", ErrUnexpectedToken, i)
			}
		}

		if currentType != END_ARRAY && currentType != END_OBJECT && wantComma && !hadComma {
			return result, fmt.Errorf("%w at %d, expected comma", ErrUnexpectedSep, start)
		}
		if !wantComma && hadComma {
			return result, fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start)
		}
		wantComma = currentType >= STRING && currentType <= NULL || currentType == END_ARRAY || currentType == END_OBJECT
		hadComma = i < len(s) && s[i] == ','
		if hadComma {
			i++
		}

		result = append(result, Token{
			Type: currentType, Value: s[start : start+length],
		})
		if errOnce != nil {
			return result, fmt.Errorf("%w at %d", errOnce, start)
		}
	}
	return result, nil
}

func Tokenize(s []byte) (result JSON, err error) {
	return TokenizeInto(s, make(JSON, 0, 3))
}
