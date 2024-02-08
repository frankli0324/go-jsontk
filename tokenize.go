package jsontk

import "fmt"

var maybeNext = map[TokenType]byte{
	END_OBJECT: ',',
	END_ARRAY:  ',',
	STRING:     ',',
	NUMBER:     ',',
	BOOLEAN:    ',',
	NULL:       ',',
	KEY:        ':',
}

// within an object, expected next token when encountered n
var expObj = map[TokenType]map[TokenType]bool{
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
}

// within an array, expected next token when encountered n
var expArr = map[TokenType]map[TokenType]bool{
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
}

type JSON []Token

func Tokenize(s []byte) (result JSON, err error) {
	stack := make([]uint8, 0, 8)
	stack = append(stack, 0) // root
	// stack[-1] == 1 means currently in object, use expObj
	// stack[-1] == 2 means currently in array, use expArr

	var nowExpect = map[TokenType]bool{
		BEGIN_OBJECT: true, BEGIN_ARRAY: true,
		NUMBER: true, STRING: true, BOOLEAN: true, NULL: true,
	}
	for i := 0; i < len(s); i++ {
		for s[i] == ' ' || s[i] == '\n' || s[i] == '\r' || s[i] == '\t' {
			i++
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
			cntEsc := 0
			for i < len(s) && s[i] != '"' || cntEsc%2 == 1 {
				if s[i] == '\\' {
					cntEsc++
				} else {
					cntEsc = 0
				}
				i++
			}
			currentType = STRING
		case '-':
			i++
			fallthrough // negative NUMBER
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for i < len(s) && (s[i] >= '0' && s[i] <= '9') || s[i] == '.' {
				i++
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
		case '/':
			if i+1 < len(s) && s[i+1] == '/' {
				for i < len(s) && s[i] != '\r' && s[i] != '\n' {
					i++
				}
				continue
			}
		}
		if currentType == STRING && nowExpect[KEY] {
			currentType = KEY
		}
		if !nowExpect[currentType] {
			want := ""
			for k := range nowExpect {
				want += k.String() + ","
			}
			err = fmt.Errorf(
				"%w: got %s, want one of:%s",
				ErrUnexpectedToken, currentType.String(), want[:len(want)-1],
			)
			currentType = INVALID
		}
		switch stack[len(stack)-1] {
		case 1:
			if v, ok := expObj[currentType]; ok {
				nowExpect = v
			}
		case 2:
			if v, ok := expArr[currentType]; ok {
				nowExpect = v
			}
		default:
			nowExpect = nil
		}
		result = append(result, Token{
			Type: currentType, Value: s[start : i+1],
		})
		for i+1 < len(s) && (s[i+1] == ' ' || s[i+1] == '\n' || s[i+1] == '\r' || s[i+1] == '\t') {
			i++
		}
		if n, ok := maybeNext[currentType]; ok {
			if i+1 != len(s) && s[i+1] != n {
				if s[i+1] != '}' && s[i+1] != ']' {
					currentType = INVALID
					err = fmt.Errorf("%w: expected '%c', got %c", ErrUnexpectedSep, n, s[i+1])
					result = append(result, Token{
						Type: INVALID, Value: s[i+1:],
					})
				}
			} else {
				i++
			}
		}
		if currentType == INVALID {
			return
		}
	}
	return result, nil
}
