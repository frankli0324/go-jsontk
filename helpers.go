package jsontk

import (
	"encoding/json"
	"errors"
	"strconv"
)

func (j *JSON) Type() TokenType {
	if j == nil || len(j.store) == 0 {
		return INVALID
	}
	return j.store[0].Type
}

func (j *JSON) Index(i int) *JSON {
	if j.Type() != BEGIN_ARRAY {
		return nil
	}
	ctr, idx := -1, -2
	for iv, v := range j.store {
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
			return &JSON{j.store[iv:]}
		}
	}
	return nil
}

func (j *JSON) Get(key string) *JSON {
	if j.Type() != BEGIN_OBJECT {
		return nil
	}
	ctr, lastKey := -1, -1
	for i, v := range j.store {
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
		if rk, ok := unquote(v.Value); ok && key == rk {
			lastKey = i + 1
		}
	}
	if lastKey == -1 {
		return nil
	}
	return &JSON{j.store[lastKey:]}
}

func (j *JSON) BatchGet(into map[string]*JSON) int {
	if j.Type() != BEGIN_OBJECT {
		return 0
	}
	ctr, cnt := -1, 0
	for i, v := range j.store {
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
		if rk, ok := unquote(v.Value); ok {
			if v, ok := into[rk]; ok {
				into[rk] = &JSON{j.store[i+1:]}
				if v == nil {
					cnt++
				}
			}
		}
	}
	return cnt
}

func (j *JSON) Number() (json.Number, error) {
	if j.Type() != NUMBER {
		return "", errors.New("invalid")
	}
	return json.Number(j.store[0].Value), nil
}

func (j *JSON) Int64() (int64, error) {
	if j.Type() != NUMBER {
		return 0, errors.New("invalid")
	}
	return strconv.ParseInt(string(j.store[0].Value), 10, 64)
}

func (j *JSON) Float64() (float64, error) {
	if j.Type() != NUMBER {
		return 0, errors.New("invalid")
	}
	return strconv.ParseFloat(string(j.store[0].Value), 64)
}

func (j *JSON) String() (string, error) {
	if j.Type() != STRING {
		return "", errors.New("invalid")
	}
	s, ok := unquote(j.store[0].Value)
	if !ok {
		return "", errors.New("invalid")
	}
	return s, nil
}

func (j *JSON) Bool() (bool, error) {
	if j.Type() != BOOLEAN {
		return false, errors.New("invalid")
	}
	// since it's successfully tokenized, values should be always certain
	return j.store[0].Value[0] == 't', nil
}

func (j *JSON) IsNull() bool {
	return j.Type() == NULL
}
