package core

import (
	"fmt"
	"strconv"
	"text/template"
)

// mathFuncs — Math
//
// Domain: math
//
// Arithmetic on values that may arrive as integers, floats or numeric strings:
// add, subtract, multiply, divide, modulo, absolute value, minimum, maximum
// and clamping. Division and modulo by zero are errors.
func mathFuncs() template.FuncMap {
	return template.FuncMap{
		"add":   Add,
		"sub":   Sub,
		"mul":   Mul,
		"div":   Div,
		"mod":   Mod,
		"min":   Min,
		"max":   Max,
		"clamp": Clamp,
		"abs":   Abs,
	}
}

// Add adds two numbers. Returns `int64` when the result is whole, `float64`
// otherwise.
//
// Keywords: math, arithmetic, add, sum, number, plus
//
// Example:
//
//	{{ add 8080 1000 }}  → 9080
func Add(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("add: non-numeric arguments %v, %v", a, b)
	}
	return nativeNumber(af + bf), nil
}

// Sub subtracts the second number from the first.
//
// Keywords: math, arithmetic, subtract, difference, minus, number
//
// Example:
//
//	{{ sub .deployment.spec.replicas 1 }}  → 2
func Sub(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("sub: non-numeric arguments %v, %v", a, b)
	}
	return nativeNumber(af - bf), nil
}

// Mul multiplies two numbers.
//
// Keywords: math, arithmetic, multiply, product, times, number
//
// Example:
//
//	{{ mul .deployment.spec.replicas 2 }}  → 6
func Mul(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("mul: non-numeric arguments %v, %v", a, b)
	}
	return nativeNumber(af * bf), nil
}

// Div divides the first number by the second. Returns an error on division by
// zero.
//
// Keywords: math, arithmetic, divide, quotient, number, ratio
//
// Example:
//
//	{{ div 8 4 }}  → 2
func Div(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("div: non-numeric arguments %v, %v", a, b)
	}
	if bf == 0 {
		return nil, fmt.Errorf("div: division by zero")
	}
	return nativeNumber(af / bf), nil
}

// Mod returns the remainder of dividing the first number by the second, as
// an integer (`int64`). An error when the second is zero.
//
// Keywords: math, arithmetic, modulo, remainder, integer, modulus
//
// Example:
//
//	{{ mod 8080 100 }}  → 80
func Mod(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("mod: non-numeric arguments %v, %v", a, b)
	}
	if bf == 0 {
		return nil, fmt.Errorf("mod: division by zero")
	}
	return int64(af) % int64(bf), nil
}

// Min returns the smaller of two numbers. Use to cap a value at an upper
// bound.
//
// Keywords: math, minimum, clamp, cap, bound, number
//
// Example:
//
//	{{ min .deployment.spec.replicas 10 }}  → 3
func Min(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("min: non-numeric arguments %v, %v", a, b)
	}
	if af < bf {
		return nativeNumber(af), nil
	}
	return nativeNumber(bf), nil
}

// Max returns the larger of two numbers. Use to enforce a minimum floor.
//
// Keywords: math, maximum, clamp, floor, bound, number
//
// Example:
//
//	{{ max .deployment.spec.replicas 5 }}  → 5
func Max(a, b interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	bf, be := anyToFloat(b)
	if ae != nil || be != nil {
		return nil, fmt.Errorf("max: non-numeric arguments %v, %v", a, b)
	}
	if af > bf {
		return nativeNumber(af), nil
	}
	return nativeNumber(bf), nil
}

// Clamp constrains a value to the range `[lo, hi]`. Equivalent to `max(lo,
// min(hi, val))`.
//
// Keywords: math, clamp, range, bounds, constrain, number, limit
//
// Example:
//
//	{{ clamp 25 1 20 }}  → 20
//	{{ clamp 0 1 20 }}   → 1
func Clamp(val, lo, hi interface{}) (interface{}, error) {
	vf, ve := anyToFloat(val)
	lf, le := anyToFloat(lo)
	hf, he := anyToFloat(hi)
	if ve != nil || le != nil || he != nil {
		return nil, fmt.Errorf("clamp: non-numeric arguments")
	}
	if vf < lf {
		return nativeNumber(lf), nil
	}
	if vf > hf {
		return nativeNumber(hf), nil
	}
	return nativeNumber(vf), nil
}

// Abs returns the absolute value of a number.
//
// Keywords: math, absolute, positive, number
//
// Example:
//
//	{{ abs -300 }}  → 300
func Abs(a interface{}) (interface{}, error) {
	af, ae := anyToFloat(a)
	if ae != nil {
		return nil, fmt.Errorf("abs: non-numeric argument %v", a)
	}
	if af < 0 {
		af = -af
	}
	return nativeNumber(af), nil
}

// ── internal ──────────────────────────────────────────────────────────────────

// anyToFloat converts any numeric-ish value to float64.
func anyToFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	}
	return 0, fmt.Errorf("cannot convert %T to number", v)
}

// nativeNumber returns int64 when f is a whole number, float64 otherwise.
// This keeps integer arithmetic returning integers — important for fields
// like spec.replicas that the API server expects as integers.
func nativeNumber(f float64) interface{} {
	if f == float64(int64(f)) {
		return int64(f)
	}
	return f
}
