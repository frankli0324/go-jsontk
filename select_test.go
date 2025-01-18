package jsontk

import (
	"bytes"
	"testing"
)

type Expectation [][]byte

func (e *Expectation) Next(data []byte) ([]byte, bool) {
	if len(*e) == 0 {
		return nil, data == nil
	}
	if !bytes.Equal(data, (*e)[0]) {
		return (*e)[0], false
	}
	*e = (*e)[1:]
	return nil, true
}

func b(s ...string) [][]byte {
	ret := make([][]byte, len(s))
	for i, s := range s {
		if s == "" {
			ret[i] = nil
		} else {
			ret[i] = []byte(s)
		}
	}
	return ret
}

func TestSelect(t *testing.T) {
	t.Run("SelectPositiveSlice", func(t *testing.T) {
		expt := Expectation(b("{\"test\"\n\t\t: 1}", "{}", "4"))
		var iter Iterator
		iter.Reset([]byte(`[{"test"
		: 1},1,{},3,4,5]`))
		iter.Select("$[:10:2]", func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SelectMultipleNestedKeys", func(t *testing.T) {
		expt := Expectation(b("3"))
		var iter Iterator
		iter.Reset([]byte(`{
	"test1": 1, "test2": 2, "test3\"": [3], "test4'": 4
}`))
		iter.Select(`$['test1',"test2", 'test3"', "test4'"][0]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SelectMultipleKeys", func(t *testing.T) {
		expt := Expectation(b(`"1"`, "2", "[3]", "4", "5", "6"))
		var iter Iterator
		iter.Reset([]byte(`{
	"test1": "1", "test2": 2, "test3\"": [3], "test4'": 4,  "tes\'t5": 5, "tes\"t6": 6
}`))
		iter.Select(`$['test1',"test2", 'test3"', "test4'", 'tes\'t5', "tes\"t6"]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("RecursiveSelection", func(t *testing.T) {
		expt := Expectation(b("1", "3"))
		var iter Iterator
		iter.Reset([]byte(`{"test1": 1, "test2": {"vvv": [1,2,3]}, "test3\"": [3], "test4'": 4}`))
		iter.Select(`$..[0]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("IndexBackward", func(t *testing.T) {
		expt := Expectation(b("2"))
		var iter Iterator
		iter.Reset([]byte(`[1  ,2,  3]`))
		iter.Select(`$[-2]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SliceForward", func(t *testing.T) {
		expt := Expectation(b("2", "3"))
		var iter Iterator
		iter.Reset([]byte(`[1  ,2,  3]`))
		iter.Select(`$[1:]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SliceForwardStep", func(t *testing.T) {
		expt := Expectation(b("2", "4", "6"))
		var iter Iterator
		iter.Reset([]byte(`[1  ,2,  3, 4,5 , 6]`))
		iter.Select(`$[1::2]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SliceBackward", func(t *testing.T) {
		expt := Expectation(b("6", "5", "4", "3", "2", "1"))
		var iter Iterator
		iter.Reset([]byte(`[1  ,2,  3, 4,5 , 6]`))
		iter.Select(`$[::-1]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
	t.Run("SliceBackwardSkipStep", func(t *testing.T) {
		expt := Expectation(b("6", "4"))
		var iter Iterator
		iter.Reset([]byte(`[1  ,2,  3, 4,5 , 6]`))
		iter.Select(`$[5:1:-2]`, func(iter *Iterator) {
			_, i, l := iter.Skip()
			if v, ok := expt.Next(iter.data[i : i+l]); !ok {
				t.Errorf("result mismatch, expected %s, got %s", string(v), string(iter.data[i:i+l]))
			}
		})
		if _, ok := expt.Next(nil); !ok {
			t.Errorf("didn't match all expectations, %d remaining", len(expt))
		}
	})
}
