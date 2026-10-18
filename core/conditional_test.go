package core

import (
	"reflect"
	"testing"
)

func TestFunctionTernary(t *testing.T) {
	tests := []struct {
		name     string
		cond     interface{}
		trueVal  interface{}
		falseVal interface{}
		want     interface{}
	}{
		{"truthy string", "hello", "yes", "no", "yes"},
		{"empty string", "", "yes", "no", "no"},
		{"truthy int", 1, "yes", "no", "yes"},
		{"zero int", 0, "yes", "no", "no"},
		{"truthy bool true", true, "yes", "no", "yes"},
		{"truthy bool false", false, "yes", "no", "no"},
		{"non-empty slice", []interface{}{1}, "yes", "no", "yes"},
		{"empty slice", []interface{}{}, "yes", "no", "no"},
		{"nil", nil, "yes", "no", "no"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ternary(tt.cond, tt.trueVal, tt.falseVal)
			if got != tt.want {
				t.Errorf("tenary(%v, %v, %v) = %v, want %v", tt.cond, tt.trueVal, tt.falseVal, got, tt.want)
			}
		})
	}
}

func TestFunctionCoalesce(t *testing.T) {
	tests := []struct {
		name string
		vals []interface{}
		want interface{}
	}{
		{"first non-empty", []interface{}{"", nil, "value", "other"}, "value"},
		{"all empty", []interface{}{"", nil, 0, false}, nil},
		{"single value", []interface{}{"hello"}, "hello"},
		{"empty list", []interface{}{}, nil},
		{"numbers", []interface{}{0, 0, 5, 10}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Coalesce(tt.vals...)
			if got != tt.want {
				t.Errorf("coalesce(%v) = %v, want %v", tt.vals, got, tt.want)
			}
		})
	}
}

func TestFunctionDefault(t *testing.T) {
	// noteDefault(dflt, val): returns val if non-empty, else dflt.
	// In pipeline: {{ val | default dflt }} → noteDefault(dflt, val)
	tests := []struct {
		name string
		dflt interface{}
		val  interface{}
		want interface{}
	}{
		{"non-empty string", "default", "foo", "foo"},
		{"empty string", "default", "", "default"},
		{"nil", "default", nil, "default"},
		{"zero int", 100, 0, 100},
		{"non-zero int", 100, 42, 42},
		{"false bool", true, false, true},
		{"true bool", false, true, true},
		{"empty slice", "default", []interface{}{}, "default"},
		{"non-empty slice", "default", []interface{}{1}, []interface{}{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Default(tt.dflt, tt.val)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultFunc(%v, %v) = %v, want %v", tt.dflt, tt.val, got, tt.want)
			}
		})
	}
}

func TestFunctionEmpty(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"<no value>", "<no value>", true},
		{"non-empty string", "hello", false},
		{"zero int", 0, true},
		{"non-zero int", 5, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"empty slice", []interface{}{}, true},
		{"non-empty slice", []interface{}{1}, false},
		{"empty map", map[string]interface{}{}, true},
		{"non-empty map", map[string]interface{}{"a": 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Empty(tt.v); got != tt.want {
				t.Errorf("empty(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestFunctionNotEmpty(t *testing.T) {
	// Just inverse of empty
	if NotEmpty("") != false {
		t.Error("notEmpty('') should be false")
	}
	if NotEmpty("hello") != true {
		t.Error("notEmpty('hello') should be true")
	}
}

func TestFunctionIsTruthy(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"nil", nil, false},
		{"empty string", "", false},
		{"non-empty string", "test", true},
		{"zero int", 0, false},
		{"non-zero int", 1, true},
		{"zero float", 0.0, false},
		{"non-zero float", 1.5, true},
		{"false bool", false, false},
		{"true bool", true, true},
		{"empty slice", []interface{}{}, false},
		{"non-empty slice", []interface{}{1}, true},
		{"empty map", map[string]interface{}{}, false},
		{"non-empty map", map[string]interface{}{"a": 1}, true},
		{"other type", struct{}{}, true}, // non-nil struct -> true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTruthy(tt.v); got != tt.want {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestFunctionBoolTernary(t *testing.T) {
	if got := BoolTernary(true, "yes", "no"); got != "yes" {
		t.Errorf("boolTernary(true) = %v, want yes", got)
	}
	if got := BoolTernary(false, "yes", "no"); got != "no" {
		t.Errorf("boolTernary(false) = %v, want no", got)
	}
}

func TestFunctionBoolDefault(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		def  bool
		want bool
	}{
		{"bool true", true, false, true},
		{"bool false", false, true, false},
		{"non-bool string", "true", false, false},
		{"non-bool int", 1, true, true},
		{"nil", nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolDefault(tt.v, tt.def); got != tt.want {
				t.Errorf("boolDefault(%v, %v) = %v, want %v", tt.v, tt.def, got, tt.want)
			}
		})
	}
}
