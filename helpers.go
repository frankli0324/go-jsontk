package jsontk

import (
	"errors"
	"strconv"
)

func (j JSON) Index(i int) JSON {
	// TODO
	return nil
}

func (j JSON) Get(key string) JSON {
	if len(j) == 0 || j[0].Type != BEGIN_OBJECT {
		return nil
	}
	ctr := 0
	for i, v := range j {
		if v.Type == BEGIN_ARRAY || v.Type == BEGIN_OBJECT {
			ctr++
		}
		if v.Type == END_ARRAY || v.Type == END_OBJECT {
			ctr--
		}
		if ctr != 0 {
			continue
		}
		rk, err := unquoteBytes(v.Value)
		if err != nil {
			continue
		}
		if v.Type == KEY && key == rk {
			return j[i:]
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
