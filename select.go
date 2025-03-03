package jsontk

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

var minInt = -1 << (32<<(^uint(0)>>63) - 1)

func (iter *Iterator) Select(path string, cb func(iter *Iterator)) error {
	selectors, err := parseJSONPath(path)
	if err != nil {
		return err
	}
	traverse(iter, selectors, cb)
	return iter.Error
}

func traverse(iter *Iterator, sel []selector, f func(iter *Iterator)) {
	if len(sel) == 0 {
		f(iter)
		return
	}
	switch iter.Peek() {
	case BEGIN_OBJECT:
		iter.NextObject(func(key *Token) bool {
			if sel[0] == recursive {
				save := iter.head
				traverse(iter, sel, f)
				iter.head = save
			}
			if sel[0].SelectObj(key, iter) {
				traverse(iter, sel[1:], f)
			} else {
				iter.Skip()
			}
			return true
		})
	case BEGIN_ARRAY:
		if traverseInversedArr(iter, sel[0], f) {
			return
		}
		iter.NextArray(func(idx int) bool {
			if sel[0] == recursive {
				save := iter.head
				traverse(iter, sel, f)
				iter.head = save
			}
			if sel[0].SelectArr(idx, iter) {
				traverse(iter, sel[1:], f)
			} else {
				iter.Skip()
			}
			return true
		})
	default:
		iter.Skip()
	}
}

func traverseInversedArr(iter *Iterator, sel selector, f func(iter *Iterator)) bool {
	switch sel := sel.(type) {
	case indexSelector:
		if sel >= 0 {
			return false
		}
	case *arrSliceSelector:
		if sel.start >= 0 && (sel.end == -1 || sel.end >= 0) && sel.step >= 0 {
			return false
		}
	default:
		return false
	}
	indexes := make([]int, 0, 10)
	if err := iter.NextArray(func(idx int) bool {
		_, i, _ := iter.Skip()
		indexes = append(indexes, i)
		return true
	}); err != nil {
		return false
	}
	after := iter.head
	switch sel := sel.(type) {
	case indexSelector:
		iter.head = indexes[len(indexes)+int(sel)]
		f(iter)
	case *arrSliceSelector:
		s, e := sel.start, sel.end
		if s < 0 {
			s += len(indexes)
		}
		if e == minInt {
			e = -1
		} else if e < 0 {
			e += len(indexes)
		}
		for ; (e-s)*sel.step > 0; s += sel.step {
			if s >= 0 && s < len(indexes) {
				iter.head = indexes[s]
				f(iter)
			}
		}
	}
	if iter.Error == nil {
		iter.head = after
	}
	return true
}

type as [8]uint32

func (as as) c(c byte) bool {
	return as[c/32]&(1<<(c%32)) != 0
}

var emptyChar = func() (as as) {
	for _, c := range " \t\r\n" {
		as[c/32] |= 1 << (c % 32)
	}
	return
}()

var alphaDigitSlash = func() (as as) {
	for _, c := range "0123456789_" {
		as[c/32] |= 1 << (c % 32)
	}
	for c := 'A'; c < 'Z'; c++ {
		as[c/32] |= 1 << (c % 32)
	}
	for c := 'a'; c < 'z'; c++ {
		as[c/32] |= 1 << (c % 32)
	}
	return
}()

func parseJSONPathBracket(b string) (end int, ret selector, err error) {
	var segs = []selector{nil}[:0]
	for end < len(b) {
		for end < len(b) && emptyChar.c(b[end]) {
			end++
		}
		if end == len(b) {
			return end, nil, ErrInvalidJsonpath
		}
		switch b[end] {
		case '*':
			segs = append(segs, wildcardSelector{})
			end++
		case '\'', '"':
			endStr, strEsc := end+1, false
			for endStr < len(b) && (b[endStr] != b[end] || strEsc) {
				if b[endStr] == '\\' {
					strEsc = !strEsc
				} else {
					strEsc = false
				}
				endStr++
			}
			if b[endStr] != b[end] {
				return end, nil, fmt.Errorf("%w: invalid name selector", ErrInvalidJsonpath)
			}
			endStr++
			str := []byte(b[end:endStr])[:endStr-end]
			if str[0] == '\'' { // make sure that unquote will not complain about unescaped quotes
				str[0], str[len(str)-1] = '"', '"'
			}
			unquoted, ok := unquote(str)
			if !ok {
				return end, nil, fmt.Errorf("%w: invalid name selector, unquote error", ErrInvalidJsonpath)
			}
			segs = append(segs, nameSelector(unquoted))
			end = endStr
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', ':':
			cnt := 0
			var nums = [3]int{0, -1, 1}
			var isdefault [3]bool
			for {
				numStop, num := end, 0
				for numStop < len(b) && b[numStop] == '-' || b[numStop] >= '0' && b[numStop] <= '9' {
					numStop++
				}
				if numStop == end {
					isdefault[cnt] = true
				} else if num, err = strconv.Atoi(b[end:numStop]); err != nil {
					return
				} else {
					nums[cnt] = num
					if cnt == 2 && num < 0 {
						if isdefault[0] {
							nums[0] = -1
						}
						if isdefault[1] {
							nums[1] = minInt
						}
					}
					end = numStop
				}
				for end < len(b) && emptyChar.c(b[end]) {
					end++
				}
				if end == len(b) || b[end] != ':' || cnt == 2 {
					break
				}
				cnt, end = cnt+1, end+1
				for end < len(b) && emptyChar.c(b[end]) {
					end++
				}
			}
			switch cnt {
			case 0:
				segs = append(segs, indexSelector(nums[0]))
			case 1, 2:
				segs = append(segs, &arrSliceSelector{nums[0], nums[1], nums[2]})
			}
		case '?':
			err = fmt.Errorf("%w: filter selector is currently unsupported", ErrInvalidJsonpath)
			return
		default:
			err = ErrInvalidJsonpath
			return
		}
		for end < len(b) && emptyChar.c(b[end]) {
			end++
		}
		if end == len(b) {
			err = fmt.Errorf("%w: expecting ] but early EOF", ErrInvalidJsonpath)
			return
		}
		if b[end] == ']' {
			if len(segs) > 1 {
				ret = combineSelector(segs)
			} else if len(segs) == 1 {
				ret = segs[0]
			}
			return end + 1, ret, err
		}
		if b[end] != ',' {
			c, idx := b[end], end
			err = fmt.Errorf("%w: expecting ] or , but got %c at %d", ErrInvalidJsonpath, c, idx)
			return
		}
		end++
	}
	err = fmt.Errorf("%w: early EOF", ErrInvalidJsonpath)
	return
}

func parseJSONPath(path string) ([]selector, error) {
	if !strings.HasPrefix(path, "$") {
		return nil, ErrInvalidJsonpath
	}
	var selectors []selector
	path = path[1:]
	for len(path) > 1 {
		switch path[0] {
		case '.':
			switch path[1] {
			case '.':
				selectors = append(selectors, recursive)
				if len(path) > 3 && path[2] == '[' {
					path = path[2:]
				} else {
					path = path[1:]
				}
			case '*':
				selectors = append(selectors, wildcardSelector{})
				path = path[2:]
			default:
				idx := strings.IndexFunc(path[1:], func(r rune) bool {
					if r < utf8.RuneSelf && alphaDigitSlash.c(byte(r)) {
						return false
					}
					if r >= utf8.RuneSelf && r != utf8.RuneError {
						return false
					}
					return true
				})
				if idx == -1 {
					idx = len(path) - 1
				}
				selectors = append(selectors, nameSelector(path[1:idx+1]))
				path = path[idx+1:]
			}
		case '[':
			end, ret, err := parseJSONPathBracket(path[1:])
			if err != nil {
				return selectors, err
			}
			selectors = append(selectors, ret)
			path = path[end+1:]
		default:
			return selectors, ErrInvalidJsonpath
		}
	}
	if len(path) != 0 {
		return selectors, ErrInvalidJsonpath
	}
	return selectors, nil
}

type selector interface {
	SelectArr(idx int, iter *Iterator) bool
	SelectObj(key *Token, iter *Iterator) bool
}

type combineSelector []selector

func (n combineSelector) SelectArr(idx int, iter *Iterator) bool {
	for _, n := range n {
		if n.SelectArr(idx, iter) {
			return true
		}
	}
	return false
}
func (n combineSelector) SelectObj(key *Token, iter *Iterator) bool {
	for _, n := range n {
		if n.SelectObj(key, iter) {
			return true
		}
	}
	return false
}

type nameSelector string

func (n nameSelector) SelectArr(idx int, iter *Iterator) bool {
	return false
}
func (n nameSelector) SelectObj(key *Token, iter *Iterator) bool {
	return key.EqualString(string(n))
}

type indexSelector int

func (i indexSelector) SelectArr(idx int, iter *Iterator) bool {
	return idx == int(i)
}
func (i indexSelector) SelectObj(key *Token, iter *Iterator) bool {
	return false
}

// arrSliceSelector represents a selector for array slices
type arrSliceSelector struct{ start, end, step int }

func (s *arrSliceSelector) SelectArr(idx int, iter *Iterator) bool {
	if s.step < 0 { // currently only positive step is supported
		return false
	}
	if idx < s.start || s.end != -1 && idx >= s.end {
		return false
	}
	return (idx-s.start)%s.step == 0
}
func (i *arrSliceSelector) SelectObj(key *Token, iter *Iterator) bool {
	return false
}

type wildcardSelector struct{}

func (wildcardSelector) SelectArr(idx int, iter *Iterator) bool {
	return true
}
func (wildcardSelector) SelectObj(key *Token, iter *Iterator) bool {
	return true
}

type recursiveSelector struct{}

var recursive recursiveSelector

func (recursiveSelector) SelectArr(idx int, iter *Iterator) bool {
	return true
}
func (recursiveSelector) SelectObj(key *Token, iter *Iterator) bool {
	return true
}
