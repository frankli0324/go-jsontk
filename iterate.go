package jsontk

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
			return INVALID, 0, ErrEarlyEOF.at(i, "expected end of string")
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
