package jsontk

import (
	"fmt"
)

var expInitial = func(m map[TokenType]bool) (r [cntTokenType]bool) {
	for k := range m {
		r[k] = true
	}
	return
}(map[TokenType]bool{
	BEGIN_OBJECT: true, BEGIN_ARRAY: true,
	NUMBER: true, STRING: true, BOOLEAN: true, NULL: true,
})

// within an object, expected next token when encountered n
var expObj = func(m map[TokenType]map[TokenType]bool) (r [cntTokenType][cntTokenType]bool) {
	for k := range m {
		for l := range m[k] {
			r[k][l] = true
		}
	}
	return
}(map[TokenType]map[TokenType]bool{
	BEGIN_OBJECT: {KEY: true, END_OBJECT: true},
	END_OBJECT:   {KEY: true, END_OBJECT: true},
	END_ARRAY:    {KEY: true, END_OBJECT: true},
	NUMBER:       {KEY: true, END_OBJECT: true},
	STRING:       {KEY: true, END_OBJECT: true},
	BOOLEAN:      {KEY: true, END_OBJECT: true},
	NULL:         {KEY: true, END_OBJECT: true},
	KEY: {
		NUMBER: true, STRING: true, BOOLEAN: true, NULL: true,
		BEGIN_ARRAY: true, BEGIN_OBJECT: true,
	},
})

// within an array, expected next token when encountered n
var expArr = func(m map[TokenType]map[TokenType]bool) (r [cntTokenType][cntTokenType]bool) {
	for k := range m {
		for l := range m[k] {
			r[k][l] = true
		}
	}
	return
}(map[TokenType]map[TokenType]bool{
	NUMBER:    {NUMBER: true, STRING: true, BOOLEAN: true, NULL: true, BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true},
	STRING:    {NUMBER: true, STRING: true, BOOLEAN: true, NULL: true, BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true},
	BOOLEAN:   {NUMBER: true, STRING: true, BOOLEAN: true, NULL: true, BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true},
	NULL:      {NUMBER: true, STRING: true, BOOLEAN: true, NULL: true, BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true},
	END_ARRAY: {NUMBER: true, STRING: true, BOOLEAN: true, NULL: true, BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true},
	END_OBJECT: {
		NUMBER: true, STRING: true, BOOLEAN: true, NULL: true,
		END_ARRAY: true, BEGIN_OBJECT: true, BEGIN_ARRAY: true,
	},
	BEGIN_ARRAY: {
		NUMBER: true, STRING: true, BOOLEAN: true, NULL: true,
		BEGIN_ARRAY: true, BEGIN_OBJECT: true, END_ARRAY: true,
	},
})

var expNone = [cntTokenType]bool{}

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

type JSON []Token

func Tokenize(s []byte) (result JSON, err error) {
	result = make(JSON, 0, 3)
	stack := make([]uint8, 1, 8)
	// stack[-1] == 1 means currently in object, use expObj
	// stack[-1] == 2 means currently in array, use expArr

	nowExpect := expInitial
	for i := 0; i < len(s); i++ {
		i += skip(s[i:])
		if i >= len(s) {
			break
		}
		start := i
		var currentType TokenType
		switch s[i] {
		case '{':
			stack = append(stack, 1)
			currentType = BEGIN_OBJECT
		case '}':
			currentType = END_OBJECT
			if stack[len(stack)-1] != 1 {
				err = fmt.Errorf("%w at char %d", ErrInvalidParentheses, i)
				currentType = INVALID
			} else {
				stack = stack[:len(stack)-1]
			}
		case '[':
			stack = append(stack, 2)
			currentType = BEGIN_ARRAY
		case ']':
			currentType = END_ARRAY
			if stack[len(stack)-1] != 2 {
				err = fmt.Errorf("%w at char %d", ErrInvalidParentheses, i)
				currentType = INVALID
			} else {
				stack = stack[:len(stack)-1]
			}
		case '"':
			i++
			isEscaped := false
			for ; i < len(s) && (s[i] != '"' || isEscaped); i++ {
				if s[i] == '\\' && !isEscaped {
					isEscaped = true
				} else {
					isEscaped = false
				}
			}
			currentType = STRING
			if i == len(s) {
				err = fmt.Errorf("%w, expected end of string", ErrEarlyEOF)
				currentType = INVALID
				i--
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for i++; i < len(s); i++ {
				if s[i] >= '0' && s[i] <= '9' {
					continue
				}
				if s[i] == '.' || s[i] == 'e' || s[i] == 'E' || s[i] == '+' || s[i] == '-' {
					continue
				}
				break
			}
			i--
			currentType = NUMBER
		case 't':
			if i+3 >= len(s) || s[i+1] != 'r' || s[i+2] != 'u' || s[i+3] != 'e' {
				err = fmt.Errorf("%w at char %d, expected boolean", ErrUnexpectedToken, i)
				currentType = INVALID
			}
			i += 3
			currentType = BOOLEAN
		case 'f':
			if i+4 >= len(s) || s[i+1] != 'a' || s[i+2] != 'l' || s[i+3] != 's' || s[i+4] != 'e' {
				err = fmt.Errorf("%w at char %d, expected boolean", ErrUnexpectedToken, i)
				currentType = INVALID
			}
			i += 4
			currentType = BOOLEAN
		case 'n':
			if i+3 >= len(s) || s[i+1] != 'u' || s[i+2] != 'l' || s[i+3] != 'l' {
				err = fmt.Errorf("%w at char %d, expected null", ErrUnexpectedToken, i)
				currentType = INVALID
			}
			i += 3
			currentType = NULL
		default:
			err = fmt.Errorf("%w at %d", ErrUnexpectedToken, i)
		}
		if currentType == STRING && nowExpect[KEY] {
			currentType = KEY
		}
		if !nowExpect[currentType] && currentType != INVALID {
			want := ""
			for t, k := range nowExpect {
				if k {
					want += "," + TokenType(t).String()
				}
			}
			if len(want) != 0 {
				want = want[1:]
			}
			err = fmt.Errorf("%w: got %s, want one of:[%s]", ErrUnexpectedToken, currentType.String(), want)
			currentType = INVALID
		}
		switch stack[len(stack)-1] {
		case 1:
			nowExpect = expObj[currentType]
		case 2:
			nowExpect = expArr[currentType]
		default:
			nowExpect = expNone
		}
		result = append(result, Token{
			Type: currentType, Value: s[start : i+1],
		})

		// prepare for lookahead, consume until *next* char is valid
		if i+1 < len(s) {
			i += skip(s[i+1:])
		}

		// lookahead
		switch currentType {
		case KEY: // key must be followed by a ':'
			if i+1 >= len(s) {
				currentType = INVALID
				err = fmt.Errorf("%w, expected ':' after object key", ErrEarlyEOF)
			} else if s[i+1] != ':' {
				currentType = INVALID
				err = fmt.Errorf("%w: expected ':' after object key at %d, got %c", ErrUnexpectedSep, i+1, s[i+1])
			} else {
				i++
			}
		case NUMBER, STRING, BOOLEAN, NULL, END_OBJECT, END_ARRAY:
			if i+1 >= len(s) {
				if len(stack) != 1 {
					currentType = INVALID
					err = fmt.Errorf("%w, expected ',' after value", ErrEarlyEOF)
				}
			} else if s[i+1] != ',' {
				if s[i+1] != '}' && s[i+1] != ']' && len(stack) != 1 {
					currentType = INVALID
					err = fmt.Errorf("%w: expected ',' at %d, got %c", ErrUnexpectedSep, i+1, s[i+1])
				}
			} else {
				i++
			}
		case INVALID:
			return // don't need to add new token if already invalid
		}

		if currentType == INVALID {
			if i+1 < len(s) {
				result = append(result, Token{Type: INVALID, Value: s[i+1:]})
			}
			return
		}
	}
	if len(stack) != 1 {
		return result, ErrEarlyEOF
	}
	return result, nil
}
