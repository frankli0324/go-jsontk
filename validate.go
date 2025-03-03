package jsontk

import (
	"encoding/json"
)

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
			if !json.Valid(key.Value) {
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
		if !json.Valid(tk.Value) {
			return ErrStandardViolation
		}
	case NUMBER:
		var tk Token
		iter.NextToken(&tk)
		_, err := tk.Number().Float64()
		return err
	case INVALID:
		_, _, err := next(iter.data, iter.head)
		return err
	default:
		iter.Skip()
	}
	return
}
