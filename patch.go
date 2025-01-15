package jsontk

import (
	"fmt"
)

// Patch API is currently unstable
func Patch(data []byte, path string, f func([]byte) []byte) ([]byte, int) {
	type replaceOp struct {
		start, length int
	}
	var replaces []replaceOp
	iter := Iterator{}
	iter.Reset(data)

	err := iter.Select(path, func(iter *Iterator) {
		var loc, length int
		switch iter.Peek() {
		case BEGIN_ARRAY, BEGIN_OBJECT:
			_, loc, length = iter.Skip()
		case INVALID, END_ARRAY, END_OBJECT:
			iter.Error = fmt.Errorf("%w at %d", ErrUnexpectedToken, iter.head)
			return
		default:
			_, loc, length = iter.Next()
		}
		replaces = append(replaces, replaceOp{loc, length})
	})
	if err == nil && iter.Error == ErrInterrupt {
		err = fmt.Errorf("%w at %d", ErrUnexpectedToken, iter.head)
	}
	if err != nil {
		// TODO: handle errors
		return data, 0
	}
	result := data
	for i := len(replaces) - 1; i >= 0; i-- {
		op := replaces[i]
		newVal := f(data[op.start : op.start+op.length])
		result = append(result[:op.start], append(newVal, result[op.start+op.length:]...)...)
	}
	return result, len(replaces)
}
