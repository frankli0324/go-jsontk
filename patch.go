package jsontk

import (
	"fmt"
	"strconv"
	"strings"
)

func walkSegments(iter *Iterator, segments []string, f func(iter *Iterator) bool) error {
	switch iter.Peek() {
	case BEGIN_OBJECT:
		return iter.NextObject(func(key *Token) bool {
			if segments[0] == "*" || key.EqualString(segments[0]) {
				if len(segments) == 1 {
					return f(iter)
				}
				return walkSegments(iter, segments[1:], f) == nil
			}
			iter.Skip()
			return true
		})
	case BEGIN_ARRAY:
		wantIdx, okIdx := strconv.Atoi(segments[0])
		return iter.NextArray(func(idx int) bool {
			if segments[0] == "*" || (okIdx == nil && idx == wantIdx) {
				if len(segments) == 1 {
					return f(iter)
				}
				return walkSegments(iter, segments[1:], f) == nil
			}
			iter.Skip()
			return true
		})
	default:
		iter.Skip()
		return ErrUnexpectedToken
	}
}

// Patch API is currently unstable
func Patch(data []byte, path string, f func([]byte) []byte) ([]byte, int) {
	type replaceOp struct {
		start, length int
	}
	var replaces []replaceOp

	if !strings.HasPrefix(path, "$.") {
		return nil, 0
	}

	segments := strings.Split(path[2:], ".")
	if len(segments) == 0 {
		return nil, 0
	}

	iter := Iterator{}
	iter.Reset(data)
	err := walkSegments(&iter, segments, func(iter *Iterator) bool {
		value, loc, length := iter.Next()
		switch value {
		case BEGIN_ARRAY:
			iter.Skip()
		case BEGIN_OBJECT:
			iter.Skip()
		case INVALID:
			return false
		default:
			replaces = append(replaces, replaceOp{loc, length})
		}
		return true
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
