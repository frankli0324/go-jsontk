package jsontk

import (
	"fmt"
)

var typMap = [256]TokenType{
	'-': NUMBER, '0': NUMBER, '1': NUMBER,
	'2': NUMBER, '3': NUMBER, '4': NUMBER,
	'5': NUMBER, '6': NUMBER, '7': NUMBER,
	'8': NUMBER, '9': NUMBER, '"': STRING,
	't': BOOLEAN, 'f': BOOLEAN, 'n': NULL,
	'[': BEGIN_ARRAY, '{': BEGIN_OBJECT,
	']': END_ARRAY, '}': END_OBJECT,
}

type Iterator struct {
	data  []byte
	head  int
	Error error
	key   Token // used for temporarily storing object keys to avoid alloc
}

func (iter *Iterator) Reset(data []byte) {
	iter.Error = nil
	iter.head = 0
	iter.data = data
}

func (iter *Iterator) Peek() TokenType {
	if iter.Error != nil {
		return INVALID
	}
	iter.head = skip(iter.data, iter.head)
	if iter.head >= len(iter.data) {
		return INVALID
	}
	return typMap[iter.data[iter.head]]
}

func (iter *Iterator) Next() (TokenType, int, int) {
	if iter.Error != nil {
		return INVALID, 0, 0
	}
	iter.head = skip(iter.data, iter.head)
	loc := iter.head
	typ, length, err := next(iter.data[iter.head:])
	iter.Error = err
	iter.head += length
	return typ, loc, length
}

func (iter *Iterator) NextToken(t *Token) *Token {
	if t == nil {
		t = new(Token)
	}
	typ, idx, l := iter.Next()
	t.Type = typ
	if typ < cntTokenType && assuredToken[typ] == "" {
		t.Value = iter.data[idx : idx+l]
	}
	return t
}

func (iter *Iterator) Skip() (TokenType, int, int) {
	if iter.Error != nil {
		return INVALID, 0, 0
	}
	iter.head = skip(iter.data, iter.head)
	loc := iter.head
	typ, length, err := next(iter.data[iter.head:])
	iter.head += length

	switch typ {
	case BEGIN_ARRAY:
		iter.NextArray(func(int) bool { iter.Skip(); return true })
	case BEGIN_OBJECT:
		iter.NextObject(func(*Token) bool { iter.Skip(); return true })
	}
	if err != nil {
		iter.Error = err
	}
	return typ, loc, iter.head - loc
}

// NextObject iterates over the next value as an object, assuming that it is one.
// One MUST be aware that the "key" callback parameter is only valid before next call to ANY method on [Iterator]
func (iter *Iterator) NextObject(cb func(key *Token) bool) error {
	if iter.Error != nil {
		return iter.Error
	}
	iter.head = skip(iter.data, iter.head)
	if iter.head >= len(iter.data) {
		iter.Error = fmt.Errorf("%w while reading object", ErrEarlyEOF)
		return iter.Error
	}
	if iter.data[iter.head] == '{' {
		iter.head = skip(iter.data, iter.head+1)
	}
	for {
		currentType, length, errOnce := next(iter.data[iter.head:])
		iter.Error = errOnce
		if currentType != STRING {
			if currentType == END_OBJECT {
				// intensionally don't check for previous comma
				// if StrictComma {
				// 	return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start-1)
				// }
				iter.head = skip(iter.data, iter.head+1)
				return nil
			}
			if iter.Error == nil {
				iter.Error = fmt.Errorf("%w at %d, expected string key", ErrUnexpectedToken, iter.head)
			}
			return iter.Error
		}
		iter.key = Token{Type: KEY, Value: iter.data[iter.head : iter.head+length]}
		iter.head = skip(iter.data, iter.head+length)
		if iter.head >= len(iter.data) || iter.data[iter.head] != ':' {
			iter.Error = fmt.Errorf("%w at %d, expected colon", ErrUnexpectedToken, iter.head)
			return iter.Error
		}
		iter.head++
		if !cb(&iter.key) {
			iter.Error = ErrInterrupt
			return nil
		}
		if iter.Error != nil {
			return iter.Error
		}

		iter.head = skip(iter.data, iter.head)
		if iter.head >= len(iter.data) || iter.data[iter.head] != ',' {
			break
		}
		iter.head = skip(iter.data, iter.head+1)
	}
	if iter.head >= len(iter.data) {
		iter.Error = fmt.Errorf("%w while reading object, expecting END_OBJECT", ErrEarlyEOF)
		return iter.Error
	}
	if iter.data[iter.head] != '}' {
		iter.Error = fmt.Errorf("%w at %d, expected END_OBJECT", ErrUnexpectedToken, iter.head)
		return iter.Error
	}
	iter.head++
	return nil
}

func (iter *Iterator) NextArray(cb func(idx int) bool) error {
	if iter.Error != nil {
		return iter.Error
	}
	iter.head = skip(iter.data, iter.head)
	if iter.head >= len(iter.data) {
		iter.Error = fmt.Errorf("%w while reading array", ErrEarlyEOF)
		return iter.Error
	}
	if iter.data[iter.head] == '[' {
		iter.head = skip(iter.data, iter.head+1)
	}
	idx := 0
	for {
		if iter.Peek() == END_ARRAY {
			// intensionally don't check for previous comma
			// if StrictComma {
			// 	return fmt.Errorf("%w at %d, unexpected comma", ErrUnexpectedSep, start-1)
			// }
			iter.head = skip(iter.data, iter.head+1)
			return nil
		}
		if !cb(idx) {
			iter.Error = ErrInterrupt
			return nil
		}
		if iter.Error != nil {
			return iter.Error
		}
		idx++
		iter.head = skip(iter.data, iter.head)
		if iter.head >= len(iter.data) || iter.data[iter.head] != ',' {
			break
		}
		iter.head = skip(iter.data, iter.head+1)
	}
	if iter.head >= len(iter.data) {
		iter.Error = fmt.Errorf("%w while reading array, expecting END_ARRAY", ErrEarlyEOF)
		return iter.Error
	}
	if iter.data[iter.head] != ']' {
		iter.Error = fmt.Errorf("%w at %d, expected END_ARRAY", ErrUnexpectedToken, iter.head)
		return iter.Error
	}
	iter.head++
	return nil
}
