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

func b(s string) []byte {
	return []byte(s)
}

func TestSelect(t *testing.T) {
	t.Run("SelectPositiveSlice", func(t *testing.T) {
		expt := Expectation([][]byte{b("{\"test\"\n\t\t: 1}"), b("{}"), b("4")})
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
		expt := Expectation([][]byte{b("3")})
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
		expt := Expectation([][]byte{b("1"), b("2"), b("[3]"), b("4")})
		var iter Iterator
		iter.Reset([]byte(`{
	"test1": 1, "test2": 2, "test3\"": [3], "test4'": 4
}`))
		iter.Select(`$['test1',"test2", 'test3"', "test4'"]`, func(iter *Iterator) {
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
		expt := Expectation([][]byte{b("1"), b("3")})
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
}
