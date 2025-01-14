package jsontk

import "encoding/json"

func (iter *Iterator) Validate() error {
	return walk(iter)
}

func walk(iter *Iterator) (err error) {
	switch iter.Peek() {
	case BEGIN_OBJECT:
		iter.NextObject(func(key *Token) bool {
			if !json.Valid(key.Value) {
				err = ErrUnexpectedToken
				return false
			}
			err = walk(iter)
			return err == nil
		})
	case BEGIN_ARRAY:
		return iter.NextArray(func(idx int) bool {
			err = walk(iter)
			return err == nil
		})
	case STRING:
		var tk Token
		iter.NextToken(&tk)
		if !json.Valid(tk.Value) {
			return ErrUnexpectedToken
		}
	case NUMBER:
		var tk Token
		iter.NextToken(&tk)
		_, err := tk.Number().Float64()
		return err
	case INVALID:
		_, _, err := next(iter.data[iter.head:])
		return err
	default:
		iter.Skip()
	}
	return
}
