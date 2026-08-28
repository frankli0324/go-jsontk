package jsontk

func (iter *Iterator) Validate() error {
	if err := walk(iter); err != nil {
		return err
	}
	if iter.Peek() != INVALID {
		return ErrUnexpectedToken
	}
	return nil
}

func walk(iter *Iterator) (err error) {
	switch iter.Peek() {
	case END_OBJECT, END_ARRAY:
		return ErrStandardViolation
	case BEGIN_OBJECT:
		return iter.NextObject(func(key *Token) bool {
			if !validString(key.Value) {
				err = ErrStandardViolation
				return false
			}
			iter.Error = walk(iter)
			return iter.Error == nil
		})
	case BEGIN_ARRAY:
		return iter.NextArray(func(idx int) bool {
			iter.Error = walk(iter)
			return iter.Error == nil
		})
	case STRING:
		var tk Token
		iter.NextToken(&tk)
		if !validString(tk.Value) {
			return ErrStandardViolation
		}
	case NUMBER:
		var tk Token
		iter.NextToken(&tk)
		if !validNumber(tk.Value) {
			return ErrStandardViolation
		}
	case INVALID:
		_, _, err := next(iter.data, iter.head)
		return err
	default:
		iter.Skip()
	}
	return
}

func validNumber(s []byte) bool {
	if len(s) == 0 {
		return false
	}
	i := 0
	if s[i] == '-' {
		i++
	}
	if i == len(s) {
		return false
	}
	// integer part
	if s[i] == '0' {
		i++
	} else if s[i] >= '1' && s[i] <= '9' {
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	} else {
		return false
	}
	// fraction
	if i < len(s) && s[i] == '.' {
		i++
		if i == len(s) || s[i] < '0' || s[i] > '9' {
			return false
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	// exponent
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if i == len(s) || s[i] < '0' || s[i] > '9' {
			return false
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	return i == len(s)
}

// validString validates a JSON string without allocating.
func validString(s []byte) bool {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return false
	}
	for i := 1; i < len(s)-1; i++ {
		if s[i] == '\\' {
			i++
			if i >= len(s)-1 {
				return false
			}
			switch escapeChars[s[i]] {
			case 0:
				return false
			case 0xff: // \uXXXX
				if i+5 > len(s)-1 {
					return false
				}
				for j := i + 1; j < i+5; j++ {
					if u4map[s[j]] < 0 {
						return false
					}
				}
				i += 4
			}
		}
	}
	return true
}
