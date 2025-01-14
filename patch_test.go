package jsontk

import (
	"testing"
)

func TestPatch(t *testing.T) {
	t.Run("ReplaceSimpleStringValue", func(t *testing.T) {
		got, count := Patch([]byte(`{"name": "old"}`), "$.name", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"name": "new"}`, 1)
	})

	t.Run("ReplaceNumberValue", func(t *testing.T) {
		got, count := Patch([]byte(`{"age": 20}`), "$.age", func([]byte) []byte {
			return []byte("30")
		})
		assertPatch(t, got, count, `{"age": 30}`, 1)
	})

	t.Run("ReplaceNestedObjectValue", func(t *testing.T) {
		got, count := Patch([]byte(`{"user": {"name": "old", "age": 20}}`), "$.user.name", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"user": {"name": "new", "age": 20}}`, 1)
	})

	t.Run("ReplaceWithLongerValue", func(t *testing.T) {
		got, count := Patch([]byte(`{"name": "old"}`), "$.name", func([]byte) []byte {
			return []byte(`"very long new value"`)
		})
		assertPatch(t, got, count, `{"name": "very long new value"}`, 1)
	})

	t.Run("ReplaceWithShorterValue", func(t *testing.T) {
		got, count := Patch([]byte(`{"name": "very long old value"}`), "$.name", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"name": "new"}`, 1)
	})

	t.Run("DeepNestedPath", func(t *testing.T) {
		got, count := Patch([]byte(`{"a": {"b": {"c": {"d": "old"}}}}`), "$.a.b.c.d", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"a": {"b": {"c": {"d": "new"}}}}`, 1)
	})

	t.Run("InvalidPath", func(t *testing.T) {
		got, count := Patch([]byte(`{"name": "old"}`), "$.nonexistent", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"name": "old"}`, 0)
	})

	t.Run("InvalidJsonPathFormat", func(t *testing.T) {
		got, count := Patch([]byte(`{"name": "old"}`), "invalid.path", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `{"name": "old"}`, 0)
	})

	t.Run("ReplaceMultipleValuesWithWildcard", func(t *testing.T) {
		got, count := Patch([]byte(`[{"name": "old"}, {"name": "old"}]`), "$.*.name", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `[{"name": "new"}, {"name": "new"}]`, 2)
	})

	t.Run("ReplaceSpecificArrayElement", func(t *testing.T) {
		got, count := Patch([]byte(`[{"name": "old"}, {"name": "old"}]`), "$[0].name", func([]byte) []byte {
			return []byte(`"new"`)
		})
		assertPatch(t, got, count, `[{"name": "new"}, {"name": "old"}]`, 1)
	})

	t.Run("MultipleSequentialReplacements", func(t *testing.T) {
		json := `{"a": "old", "b": "old", "c": "old"}`
		result := []byte(json)
		totalCount := 0

		for _, path := range []string{"$.a", "$.b", "$.c"} {
			var count int
			result, count = Patch(result, path, func([]byte) []byte {
				return []byte(`"new"`)
			})
			totalCount += count
		}

		assertPatch(t, result, totalCount, `{"a": "new", "b": "new", "c": "new"}`, 3)
	})
}

func assertPatch(t *testing.T, got []byte, gotCount int, want string, wantCount int) {
	t.Helper()

	if gotCount != wantCount {
		t.Errorf("count = %v, want %v", gotCount, wantCount)
		return
	}

	if wantCount > 0 && string(got) != want {
		t.Errorf("got %v, want %v", string(got), want)
	}
}
