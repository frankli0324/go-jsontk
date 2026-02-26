package json

import (
	"encoding/json"
	"reflect"
	"testing"
)

func assert(t *testing.T, b bool) {
	t.Helper()
	if !b {
		t.FailNow()
	}
}

type Group struct {
	Name   string           `json:"name"`
	Tags   []string         `json:"tags"`
	Scores map[string]int   `json:"scores"`
	People []map[string]any `json:"people"`
}

func TestJSONUnmarshal_MapAndSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Group
		wantErr bool
	}{
		{
			name: "valid with map and slice",
			input: `{
                "name": "Team Alpha",
                "tags": ["go", "json", "test"],
                "scores": {"alice": 10, "bob": 8},
                "people": [
                    {"name": "Alice", "age": 25},
                    {"name": "Bob", "age": 30}
                ]
            }`,
			want: Group{
				Name:   "Team Alpha",
				Tags:   []string{"go", "json", "test"},
				Scores: map[string]int{"alice": 10, "bob": 8},
				People: []map[string]any{
					{"name": "Alice", "age": float64(25)},
					{"name": "Bob", "age": float64(30)},
				},
			},
		},
		{
			name: "empty arrays and maps",
			input: `{
                "name": "Empty Team",
                "tags": [],
                "scores": {},
                "people": []
            }`,
			want: Group{
				Name:   "Empty Team",
				Tags:   []string{},
				Scores: map[string]int{},
				People: []map[string]any{},
			},
		},
		{
			name:  "missing optional fields",
			input: `{"name": "Partial Team"}`,
			want: Group{
				Name: "Partial Team",
			},
		},
		{
			name: "type mismatch in slice element",
			input: `{
                "name": "Broken Team",
                "tags": [1, 2, 3]
            }`,
			wantErr: true,
		},
		{
			name: "type mismatch in map value",
			input: `{
                "name": "Team Beta",
                "scores": {"alice": "ten"}
            }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Group
			err := Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: got err=%v, wantErr=%v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshal mismatch:\n got  %+v\n want %+v", got, tt.want)
			}
		})
	}
}

type Matrix struct {
	Fixed [3]int    `json:"fixed"`
	Words [2]string `json:"words"`
}

func TestJSONUnmarshal_ArrayTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Matrix
		wantErr bool
	}{
		{
			name:  "valid array with exact length",
			input: `{"fixed": [1, 2, 3], "words": ["a", "b"]}`,
			want:  Matrix{Fixed: [3]int{1, 2, 3}, Words: [2]string{"a", "b"}},
		},
		{
			name:  "too few elements (fills remaining with zero)",
			input: `{"fixed": [1, 2], "words": ["x", "y"]}`,
			want:  Matrix{Fixed: [3]int{1, 2, 0}, Words: [2]string{"x", "y"}},
		},
		{
			name:  "too many elements (extra ignored)",
			input: `{"fixed": [1, 2, 3, 4], "words": ["ok", "no", "extra"]}`,
			want:  Matrix{Fixed: [3]int{1, 2, 3}, Words: [2]string{"ok", "no"}},
		},
		{
			name:    "type mismatch inside array",
			input:   `{"fixed": [1, "oops", 3], "words": ["x", "y"]}`,
			wantErr: true,
		},
		{
			name:  "missing field (array left zeroed)",
			input: `{"words": ["hello", "world"]}`,
			want:  Matrix{Words: [2]string{"hello", "world"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Matrix
			err := Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error status: got err=%v, wantErr=%v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshal mismatch:\n got  %+v\n want %+v", got, tt.want)
			}
		})
	}
}

// Example struct for realistic test data
type User struct {
	ID      int                `json:"id"`
	Name    string             `json:"name"`
	Tags    []string           `json:"tags"`
	Scores  map[string]float64 `json:"scores"`
	Address struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"address"`
}

var (
	testJSON = []byte(`{
        "id": 42,
        "name": "Alice",
        "tags": ["go", "json", "bench"],
        "scores": {"math": 98.5, "english": 88.2},
        "address": {"city": "Seattle", "state": "WA"}
    }`)
)

func Benchmark(b *testing.B) {
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var u User
			if err := json.Unmarshal(testJSON, &u); err != nil {
				b.Fatalf("stdjson.Unmarshal failed: %v", err)
			}
		}
	})
	b.Run("jsontk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var u User
			if err := Unmarshal(testJSON, &u); err != nil {
				b.Fatalf("custom Unmarshal failed: %v", err)
			}
		}
	})
}

func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestStrangeBehavior(t *testing.T) {
	t.Run("input non-empty slice", func(t *testing.T) {
		v := []int{1, 2, 3, 4}
		assert(t, json.Unmarshal([]byte("[5,6]"), &v) == nil)
		assert(t, sliceEqual(v[:4], []int{5, 6, 3, 4}))
		v = []int{1, 2, 3, 4}
		assert(t, Unmarshal([]byte("[5,6]"), &v) == nil)
		assert(t, sliceEqual(v[:4], []int{5, 6, 3, 4}))
	})
	t.Run("input non-empty mismatch slice", func(t *testing.T) {
		v := []int{1, 2, 3, 4}
		assert(t, json.Unmarshal([]byte("[5,\"6\"]"), &v) != nil)
		assert(t, sliceEqual(v[:4], []int{5, 2, 3, 4}))
		v = []int{1, 2, 3, 4}
		assert(t, Unmarshal([]byte("[5,\"6\"]"), &v) != nil)
		assert(t, sliceEqual(v[:4], []int{5, 2, 3, 4}))
	})
	t.Run("input non-empty map", func(t *testing.T) {
		v := map[string]int{"test": 1}
		assert(t, json.Unmarshal([]byte(`{"test2":2}`), &v) == nil)
		assert(t, v["test"] == 1)
	})
	t.Run("assign null to int", func(t *testing.T) {
		v := 1
		assert(t, json.Unmarshal([]byte(`null`), &v) == nil)
		assert(t, v == 1)
	})
	t.Run("assign null to map inside struct", func(t *testing.T) {
		m := map[string]int{"test": 1}
		v := struct {
			M map[string]int
		}{M: m}
		assert(t, json.Unmarshal([]byte(`{"M":null}`), &v) == nil)
		assert(t, v.M == nil)
		assert(t, m["test"] == 1)
	})
}
