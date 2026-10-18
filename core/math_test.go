// pkg/note/math_test.go
package core

import (
	"testing"
)

func TestFunctionAdd(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int + int", 1, 2, int64(3), false},
		{"int64 + int64", int64(5), int64(7), int64(12), false},
		{"float + float", 1.5, 2.5, int64(4), false},
		{"int + float", 3, 2.5, 5.5, false},
		{"string ints", "10", "20", int64(30), false},
		{"string floats", "1.2", "3.4", 4.6, false},
		{"non-numeric a", "abc", 5, nil, true},
		{"non-numeric b", 5, "def", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Add(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionSub(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int - int", 10, 3, int64(7), false},
		{"float - float", 5.5, 2.2, 3.3, false},
		{"neg result", 2, 5, int64(-3), false},
		{"string", "100", "30", int64(70), false},
		{"invalid a", "x", 10, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sub(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("sub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionMul(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int * int", 4, 5, int64(20), false},
		{"float * float", 2.5, 4.0, int64(10), false},
		{"mixed", 3, 1.5, 4.5, false},
		{"string", "2", "3", int64(6), false},
		{"invalid", "a", 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mul(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("mul() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("mul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionDiv(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int division", 10, 3, 3.3333333333333335, false}, // floats
		{"exact integer", 9, 3, int64(3), false},
		{"float", 7.5, 2.5, int64(3), false},
		{"division by zero", 1, 0, nil, true},
		{"string", "10", "2", int64(5), false},
		{"invalid", 5, "x", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Div(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("div() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("div() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionMod(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int mod", 10, 3, int64(1), false},
		{"negative", -10, 3, int64(-1), false},
		{"division by zero", 5, 0, nil, true},
		{"float converted", 7.5, 2.5, int64(1), false}, // floats truncated
		{"string", "15", "4", int64(3), false},
		{"invalid", "x", 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mod(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("mod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("mod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionMin(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int smaller first", 5, 10, int64(5), false},
		{"int smaller second", 20, 15, int64(15), false},
		{"float smaller first", 3.5, 4.2, 3.5, false},
		{"int vs float", 4, 3.9, 3.9, false},
		{"equal", 7, 7, int64(7), false},
		{"string", "100", "50", int64(50), false},
		{"invalid", "a", 10, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Min(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("min() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionMax(t *testing.T) {
	tests := []struct {
		name    string
		a, b    interface{}
		want    interface{}
		wantErr bool
	}{
		{"int larger first", 15, 10, int64(15), false},
		{"float larger second", 3.2, 4.5, 4.5, false},
		{"equal", 7, 7, int64(7), false},
		{"string", "200", "300", int64(300), false},
		{"invalid", "x", 5, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Max(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("max() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionClamp(t *testing.T) {
	tests := []struct {
		name        string
		val, lo, hi interface{}
		want        interface{}
		wantErr     bool
	}{
		{"within", 5, 1, 10, int64(5), false},
		{"below", 0, 1, 10, int64(1), false},
		{"above", 20, 1, 10, int64(10), false},
		{"float", 5.5, 1.0, 10.0, 5.5, false},
		{"string", "8", "1", "10", int64(8), false},
		{"invalid val", "x", 1, 10, nil, true},
		{"invalid lo", 5, "x", 10, nil, true},
		{"invalid hi", 5, 1, "y", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Clamp(tt.val, tt.lo, tt.hi)
			if (err != nil) != tt.wantErr {
				t.Errorf("clamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFunctionAbs(t *testing.T) {
	tests := []struct {
		name    string
		a       interface{}
		want    interface{}
		wantErr bool
	}{
		{"positive", 5, int64(5), false},
		{"negative", -5, int64(5), false},
		{"float positive", 3.2, 3.2, false},
		{"float negative", -4.8, 4.8, false},
		{"zero", 0, int64(0), false},
		{"string", "-123", int64(123), false},
		{"invalid", "abc", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Abs(tt.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("abs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("abs() = %v, want %v", got, tt.want)
			}
		})
	}
}
