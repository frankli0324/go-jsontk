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
		value, loc, length := iter.Next()
		switch value {
		case BEGIN_ARRAY:
			iter.Skip()
		case BEGIN_OBJECT:
			iter.Skip()
		case INVALID:
			iter.Error = fmt.Errorf("%w at %d", ErrUnexpectedToken, iter.head)
		default:
			replaces = append(replaces, replaceOp{loc, length})
		}
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
