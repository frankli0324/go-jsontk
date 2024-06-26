package jsontk

import (
	"fmt"
	"sync"
)

func skip(s []byte, i int) int {
	for i < len(s) {
		if s[i] > 0x20 && s[i] != '/' {
			return i
		}
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

func next(s []byte) (TokenType, int, error) {
	if len(s) == 0 {
		return INVALID, 0, ErrEarlyEOF
	}
	switch s[0] {
	case '"':
		i := 1
		for ; i < len(s) && s[i] != '"'; i++ {
			if s[i] == '\\' {
				i = eos(s, i)
				break
			}
		}
		if i == len(s) {
			return INVALID, i, fmt.Errorf("%w, expected end of string", ErrEarlyEOF)
		}
		return STRING, i + 1, nil
	case '{':
		return BEGIN_OBJECT, 1, nil
	case '[':
		return BEGIN_ARRAY, 1, nil
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
	case '}':
		return END_OBJECT, 1, nil
	case ']':
		return END_ARRAY, 1, nil
	default:
		return INVALID, 1, ErrUnexpectedToken
	}
}

type JSON struct {
	store []Token
}

func (c *JSON) Close() {
	c.store = c.store[:0]
	pool.Put(c)
}

func Iterate(s []byte, cb func(typ TokenType, idx, len int)) error {
	hadComma, wantComma := false, false

	for i := 0; i < len(s); {
		i = skip(s, i)

		currentType, length, errOnce := next(s[i:])

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

		if currentType != END_ARRAY && currentType != END_OBJECT && wantComma && !hadComma {
			return fmt.Errorf("%w at %d, expected comma", ErrUnexpectedSep, start)
		}
		if !wantComma && hadComma {
			return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start)
		}
		wantComma = currentType >= STRING && currentType <= NULL || currentType == END_ARRAY || currentType == END_OBJECT
		hadComma = i < len(s) && s[i] == ','
		if hadComma {
			i++
		}

		cb(currentType, start, length)
		if errOnce != nil {
			return fmt.Errorf("%w at %d", errOnce, start)
		}
	}
	return nil
}

// Tokenize reads binary data into an array of JSON tokens.
// The data passed in tokenize should never be changed even after calling [Tokenize]
// or unexpected data could be read from the result.
func (c *JSON) Tokenize(s []byte) error {
	c.store = c.store[:0]
	return Iterate(s, func(typ TokenType, idx, len int) {
		c.store = append(c.store, Token{Type: typ, Value: s[idx : idx+len]})
	})
}

var pool = sync.Pool{
	New: func() any {
		return new(JSON)
	},
}

func Tokenize(s []byte) (result *JSON, err error) {
	result = pool.Get().(*JSON)
	return result, result.Tokenize(s)
}
