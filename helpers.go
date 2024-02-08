package jsontk

import (
	"errors"
	"strconv"
)

func (j JSON) Index(i int) JSON {
	if len(j) == 0 || j[0].Type != BEGIN_ARRAY {
		return nil
	}
	ctr, idx := -1, -2
	for iv, v := range j {
		if v.Type == BEGIN_ARRAY || v.Type == BEGIN_OBJECT {
			ctr++
		}
		if v.Type == END_ARRAY || v.Type == END_OBJECT {
			ctr--
		}
		if ctr == 0 {
			idx++
		} else if ctr > 0 {
			continue
		} else if ctr < 0 {
			break
		}
		if idx == i {
			return j[iv:]
		}
	}
	return nil
}

func (j JSON) Get(key string) JSON {
	if len(j) == 0 || j[0].Type != BEGIN_OBJECT {
		return nil
	}
	ctr := -1
	for i, v := range j {
		if v.Type == BEGIN_ARRAY || v.Type == BEGIN_OBJECT {
			ctr++
		}
		if v.Type == END_ARRAY || v.Type == END_OBJECT {
			ctr--
		}
		if ctr > 0 || v.Type != KEY {
			continue
		}
		if ctr < 0 {
			break
		}
		rk, err := unquoteBytes(v.Value)
		if err != nil {
			continue
		}
		if key == rk {
			return j[i+1:]
		}
	}
	return nil
}

func (j JSON) Int64() (int64, error) {
	if len(j) == 0 || j[0].Type != NUMBER {
		return 0, errors.New("invalid")
	}
	return strconv.ParseInt(string(j[0].Value), 10, 64)
}

func (j JSON) String() (string, error) {
	if len(j) == 0 || j[0].Type != STRING {
		return "", errors.New("invalid")
	}
	return unquoteBytes(j[0].Value)
}
