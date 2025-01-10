package jsontk

import (
	"fmt"
	"strings"
	"testing"
)

func TestIterator(t *testing.T) {
	iter := Iterator{data: []byte(`{"test":1,"xx":true,
	"vv"
	: false, "ww": {"a": 1}, "zz": [1,2,3]}`)}
	if err := iter.NextObject(func(key *Token) bool {
		fmt.Println(key.String())
		typ, _, _ := iter.Next()
		switch typ {
		case BEGIN_ARRAY:
			iter.NextArray(func(idx int) bool {
				fmt.Println("\t", idx)
				iter.Next()
				return true
			})
		case BEGIN_OBJECT:
			iter.NextObject(func(key *Token) bool {
				fmt.Println("\t", key.String())
				iter.Next()
				return true
			})
		}
		return true
	}); err != nil {
		t.Error(err)
	}
}

func TestIteratorNextObject(t *testing.T) {
	sb := strings.Builder{}
	sb.WriteRune('{')
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf(`"key_%d": "value_%d"`, i, i))
		sb.WriteRune(',')
	}
	sb.WriteRune('}')
	s := sb.String()
	bs := []byte(s)
	key := fmt.Sprintf("key_%d", 50)
	expectedValue := fmt.Sprintf("value_%d", 50)
	iter := Iterator{}

	iter.Reset(bs)
	iter.NextObject(func(k *Token) bool {
		if k.EqualString(key) {
			iter.NextToken(k)
			if k.Type != STRING || !k.EqualString(expectedValue) {
				t.Error(fmt.Errorf("unexpected value; got %q; want %q", string(k.Value), expectedValue))
			}
			return false
		}
		iter.Skip()
		return true
	})
}
