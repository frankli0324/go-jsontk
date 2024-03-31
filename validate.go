package jsontk

import (
	"fmt"
)

func validateValue(t []Token) (int, error) {
	switch t[0].Type {
	case BEGIN_ARRAY:
		return validateArray(t)
	case BEGIN_OBJECT:
		return validateObject(t)
	case END_ARRAY, END_OBJECT:
		return 0, ErrUnexpectedToken
	case STRING:
		for _, c := range t[0].Value {
			if c < 0x20 {
				return 0, ErrStandardViolation
			}
		}
	}
	return 1, nil
}

func validateArray(t []Token) (int, error) {
	for i := 1; i < len(t); i++ {
		switch t[i].Type {
		case END_ARRAY:
			return i + 1, nil
		case END_OBJECT:
			return i, ErrInvalidParentheses
		case KEY:
			return i, ErrUnexpectedSep
		default:
			if x, err := validateValue(t[i:]); err != nil {
				return i + x, err
			} else {
				i += x - 1
			}
		}
	}
	return len(t), ErrEarlyEOF
}

func validateObject(t []Token) (int, error) {
	wantKey := true
	for i := 1; i < len(t); i++ {
		switch t[i].Type {
		case END_OBJECT:
			return i + 1, nil
		case END_ARRAY:
			return i, ErrInvalidParentheses
		case KEY:
			if wantKey {
				wantKey = false
			} else {
				return i, ErrUnexpectedToken
			}
		default:
			if !wantKey {
				wantKey = true
			} else {
				return i, ErrUnexpectedToken
			}
			if x, err := validateValue(t[i:]); err != nil {
				return i + x, err
			} else {
				i += x - 1
			}
		}
	}
	return len(t), ErrEarlyEOF
}

func (j *JSON) Validate() error {
	if len(j.store) == 0 {
		return ErrEarlyEOF
	}
	idx, err := validateValue(j.store)
	// return nil
	if err != nil {
		return fmt.Errorf("%w at %d", err, idx)
	}
	if idx != len(j.store) {
		return ErrStandardViolation
	}
	return nil
}
