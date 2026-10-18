// pkg/note/listmap_test.go
package core

import (
	"reflect"
	"sort"
	"testing"
)

func TestFunctionListHas(t *testing.T) {
	tests := []struct {
		name string
		list interface{}
		val  interface{}
		want bool
	}{
		{"present", []interface{}{"a", "b", "c"}, "b", true},
		{"absent", []interface{}{1, 2, 3}, 4, false},
		{"not a slice", "string", "a", false},
		{"nil", nil, "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListHas(tt.list, tt.val); got != tt.want {
				t.Errorf("listHas(%v, %v) = %v, want %v", tt.list, tt.val, got, tt.want)
			}
		})
	}
}

func TestFunctionListGet(t *testing.T) {
	tests := []struct {
		name  string
		list  interface{}
		index int
		want  interface{}
	}{
		{"valid", []interface{}{"a", "b", "c"}, 1, "b"},
		{"out of range", []interface{}{"a"}, 2, nil},
		{"negative index", []interface{}{"a"}, -1, nil},
		{"not a slice", "string", 0, nil},
		{"nil", nil, 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListGet(tt.list, tt.index); got != tt.want {
				t.Errorf("listGet(%v, %d) = %v, want %v", tt.list, tt.index, got, tt.want)
			}
		})
	}
}

func TestFunctionListLen(t *testing.T) {
	tests := []struct {
		name string
		list interface{}
		want int
	}{
		{"slice length 3", []interface{}{1, 2, 3}, 3},
		{"empty slice", []interface{}{}, 0},
		{"not a slice", "string", 0},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListLen(tt.list); got != tt.want {
				t.Errorf("listLen(%v) = %d, want %d", tt.list, got, tt.want)
			}
		})
	}
}

func TestFunctionMapGet(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		key  string
		want interface{}
	}{
		{"valid", map[string]interface{}{"a": 1, "b": 2}, "b", 2},
		{"missing key", map[string]interface{}{"a": 1}, "b", nil},
		{"not a map", "string", "a", nil},
		{"nil", nil, "a", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapGet(tt.m, tt.key); got != tt.want {
				t.Errorf("mapGet(%v, %q) = %v, want %v", tt.m, tt.key, got, tt.want)
			}
		})
	}
}

func TestFunctionMapKeys(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		want []string
	}{
		{"valid", map[string]interface{}{"a": 1, "b": 2}, []string{"a", "b"}},
		{"empty map", map[string]interface{}{}, []string{}},
		{"not a map", "string", []string{}},
		{"nil", nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeys(tt.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapKeys(%v) = %v, want %v", tt.m, got, tt.want)
			}
		})
	}
}

func TestFunctionMapValues(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		want []interface{}
	}{
		{"valid", map[string]interface{}{"a": 1, "b": 2}, []interface{}{1, 2}},
		{"empty map", map[string]interface{}{}, []interface{}{}},
		{"not a map", "string", []interface{}{}},
		{"nil", nil, []interface{}{}},
	}
	toInts := func(vs []interface{}) []int {
		out := make([]int, len(vs))
		for i, v := range vs {
			out[i] = v.(int)
		}
		return out
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapValues(tt.m)
			if len(got) != len(tt.want) {
				t.Errorf("mapValues(%v) len = %d, want %d", tt.m, len(got), len(tt.want))
				return
			}
			if len(got) == 0 {
				return
			}
			// Map iteration order is non-deterministic; sort numerically before comparing.
			gotInts, wantInts := toInts(got), toInts(tt.want)
			sort.Ints(gotInts)
			sort.Ints(wantInts)
			if !reflect.DeepEqual(gotInts, wantInts) {
				t.Errorf("mapValues(%v) = %v, want %v", tt.m, gotInts, wantInts)
			}
		})
	}
}

func TestFunctionMapPick(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		keys []string
		want map[string]interface{}
	}{
		{"pick existing", map[string]interface{}{"a": 1, "b": 2, "c": 3}, []string{"a", "c"}, map[string]interface{}{"a": 1, "c": 3}},
		{"pick missing", map[string]interface{}{"a": 1}, []string{"b"}, map[string]interface{}{}},
		{"not a map", "string", []string{"a"}, map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapPick(tt.m, tt.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapPick(%v, %v) = %v, want %v", tt.m, tt.keys, got, tt.want)
			}
		})
	}
}

func TestFunctionMapOmit(t *testing.T) {
	tests := []struct {
		name string
		m    interface{}
		keys []string
		want map[string]interface{}
	}{
		{"omit existing", map[string]interface{}{"a": 1, "b": 2, "c": 3}, []string{"b"}, map[string]interface{}{"a": 1, "c": 3}},
		{"omit missing", map[string]interface{}{"a": 1}, []string{"b"}, map[string]interface{}{"a": 1}},
		{"not a map", "string", []string{"a"}, map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapOmit(tt.m, tt.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapOmit(%v, %v) = %v, want %v", tt.m, tt.keys, got, tt.want)
			}
		})
	}
}
