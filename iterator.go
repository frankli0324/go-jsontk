package jsontk

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
	*iter = Iterator{data: data}
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
	typ, length, err := next(iter.data, iter.head)
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
	switch typ := iter.Peek(); typ {
	case BEGIN_ARRAY, BEGIN_OBJECT:
		iter.skipContainer()
		if iter.Error != nil {
			return INVALID, 0, 0
		}
		return typ, iter.head, 0
	case INVALID, END_ARRAY, END_OBJECT:
		return INVALID, iter.head, 0
	default:
		typ, length, err := next(iter.data, iter.head)
		iter.Error = err
		if err == nil {
			iter.head += length
		}
		return typ, iter.head, length
	}
}

var act = [256]int8{'"': 2, '{': 1, '[': 1, '}': -1, ']': -1}

func (iter *Iterator) skipContainer() {
	data := iter.data
	i := iter.head + 1
	for depth := 1; depth > 0; {
		var a int8
		for i < len(data) {
			if a = act[data[i]]; a != 0 {
				break
			}
			i++
			for i+4 <= len(data) && act[data[i]]|act[data[i+1]]|act[data[i+2]]|act[data[i+3]] == 0 {
				i += 4
			}
		}
		if i >= len(data) {
			iter.Error = ErrEarlyEOF.at(i, "unexpected EOF while skipping value")
			break
		}
		if a != 2 {
			depth += int(a)
			i++
			continue
		}
		var length int
		if _, length, iter.Error = next(data, i); iter.Error != nil {
			break
		}
		i += length
	}
	iter.head = i
}

// NextObject iterates over the next value as an object, assuming that it is one.
// One MUST be aware that the "key" callback parameter is only valid before next call to ANY method on [Iterator],
// even within the callback body
func (iter *Iterator) NextObject(cb func(key *Token) bool) error {
	if iter.Error != nil {
		return iter.Error
	}
	iter.head = skipTo(iter.data, '{', iter.head)
	if iter.head >= len(iter.data) {
		iter.Error = ErrEarlyEOF.at(iter.head, "while reading object")
		return iter.Error
	}
	if iter.data[iter.head] != '{' {
		iter.Error = ErrUnexpectedToken.at(iter.head, "expected BEGIN_OBJECT")
		return iter.Error
	}
	iter.head++
	for {
		iter.head = skip(iter.data, iter.head)
		if iter.head >= len(iter.data) {
			iter.Error = ErrEarlyEOF.at(iter.head, "while reading object, expecting object key or END_OBJECT")
			return iter.Error
		}
		currentType, length, errOnce := next(iter.data, iter.head)
		iter.Error = errOnce
		if currentType != STRING {
			if currentType == END_OBJECT {
				iter.head++
				return nil
			}
			if iter.Error == nil {
				iter.Error = ErrUnexpectedToken.at(iter.head, "expected string key")
			}
			return iter.Error
		}
		iter.key = Token{Type: KEY, Value: iter.data[iter.head : iter.head+length]}
		iter.head = skipTo(iter.data, ':', iter.head+length)
		if iter.head >= len(iter.data) || iter.data[iter.head] != ':' {
			iter.Error = ErrUnexpectedToken.at(iter.head, "expected colon")
			return iter.Error
		}
		iter.head++
		var interrupted bool
		if cb == nil {
			iter.Skip()
		} else {
			interrupted = !cb(&iter.key)
		}
		if iter.Error != nil {
			return iter.Error
		}
		if interrupted {
			iter.Error = ErrInterrupt
			return nil
		}

		iter.head = skipTo(iter.data, ',', iter.head)
		if iter.head >= len(iter.data) {
			iter.Error = ErrEarlyEOF.at(iter.head, "while reading object, expecting comma or END_OBJECT")
			return iter.Error
		}
		if iter.data[iter.head] != ',' {
			if iter.data[iter.head] != '}' {
				iter.Error = ErrUnexpectedToken.at(iter.head, "expected comma or END_OBJECT")
				return iter.Error
			}
			iter.head++
			return nil
		}
		iter.head++
	}
}

func (iter *Iterator) NextArray(cb func(idx int) bool) error {
	if iter.Error != nil {
		return iter.Error
	}
	iter.head = skipTo(iter.data, '[', iter.head)
	if iter.head >= len(iter.data) {
		iter.Error = ErrEarlyEOF.at(iter.head, "while reading array")
		return iter.Error
	}
	if iter.data[iter.head] != '[' {
		iter.Error = ErrUnexpectedToken.at(iter.head, "expected BEGIN_ARRAY")
		return iter.Error
	}
	iter.head++

	for idx := 0; ; idx++ {
		iter.head = skip(iter.data, iter.head)
		if iter.head >= len(iter.data) {
			iter.Error = ErrEarlyEOF.at(iter.head, "while reading array, expecting element or END_ARRAY")
			return iter.Error
		}
		if iter.data[iter.head] == ']' { // [] | [1,]
			iter.head = skip(iter.data, iter.head+1)
			return nil
		}
		var interrupted bool
		if cb == nil {
			iter.Skip()
		} else {
			interrupted = !cb(idx)
		}
		if iter.Error != nil {
			return iter.Error
		}
		if interrupted {
			iter.Error = ErrInterrupt
			return nil
		}
		iter.head = skipTo(iter.data, ',', iter.head)
		if iter.head >= len(iter.data) {
			iter.Error = ErrEarlyEOF.at(iter.head, "while reading array, expecting comma or END_ARRAY")
			return iter.Error
		}
		if iter.data[iter.head] != ',' {
			if iter.data[iter.head] != ']' {
				iter.Error = ErrUnexpectedToken.at(iter.head, "expected comma or END_ARRAY")
				return iter.Error
			}
			iter.head++
			return nil
		}
		iter.head++
	}
}
