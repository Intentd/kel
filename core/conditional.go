package core

import (
	"fmt"
	"text/template"
)

// conditionalFuncs — Conditional
//
// Domain: conditional
//
// Choose between values inside an expression: defaults for empty values, the
// first non-empty of several, and ternaries, so a template does not need an if
// block for a one-line choice.
func conditionalFuncs() template.FuncMap {
	return template.FuncMap{
		"ternary":     Ternary,
		"boolTernary": BoolTernary,
		"boolDefault": BoolDefault,
		"eqTernary":   EqTernary,
		"coalesce":    Coalesce,
		"default":     Default,
		"empty":       Empty,
		"notEmpty":    NotEmpty,
	}
}

// EqTernary returns `trueVal` when `val` equals `target` (string comparison),
// `falseVal` otherwise. Shorthand for the `boolTernary (eq val target) trueVal
// falseVal` pattern — useful when branching on a string field value such as a
// status string, a mode flag, or a cross-CRD `found` result.
//
// Keywords: conditional, branch, equality, string, ternary, compare, match
//
// Example:
//
//	{{ eqTernary .spec.mode "production" "strict" "permissive" }}  → "strict"
func EqTernary(val, target, trueVal, falseVal interface{}) interface{} {
	if fmt.Sprintf("%v", val) == fmt.Sprintf("%v", target) {
		return trueVal
	}
	return falseVal
}

// Ternary returns `trueVal` when `condition` is truthy, `falseVal` otherwise.
// Condition truthiness follows Go rules: non-empty string, non-zero number,
// non-empty slice/map, `true` bool.
//
// Keywords: conditional, branch, ternary, if-else, logic, truthy
//
// Example:
//
//	{{ ternary .spec.debug "debug" "info" }}  → "debug"
func Ternary(condition interface{}, trueVal, falseVal interface{}) interface{} {
	if isTruthy(condition) {
		return trueVal
	}
	return falseVal
}

// Coalesce returns the first non-empty value from a variadic list. Useful when
// a field can come from multiple sources with a final fallback.
//
// Keywords: conditional, coalesce, fallback, first, multiple, chain
//
// Example:
//
//	{{ coalesce .spec.sidecarImage .spec.defaultSidecarImage "envoyproxy/envoy:v1.30" }}  → "envoyproxy/envoy:v1.30"
func Coalesce(vals ...interface{}) interface{} {
	for _, v := range vals {
		if !Empty(v) {
			return v
		}
	}
	return nil
}

// func Default(dflt, val interface{}) interface{} {
// 	if Empty(val) {
// 		return dflt
// 	}
// 	return val
// }

// Default returns `val` if non-empty, otherwise returns `def`. "Empty" means
// nil, `""`, `0`, `false`, empty slice, or empty map.
//
// Keywords: conditional, default, fallback, absent, empty, nil
//
// Example:
//
//	{{ default "info" .spec.logLevel }}  → "info"
//	{{ .spec.logLevel | default "info" }}  → "info"
func Default(dflt, val interface{}) interface{} {
	if Empty(val) {
		return dflt
	}
	return val
}

// Empty returns `true` when the value is empty (nil, `""`, `0`, `false`, empty
// slice, empty map, or `"<no value>"`).
//
// Keywords: conditional, empty, check, boolean, nil, absent
//
// Example:
//
//	{{ empty .spec.logLevel }}  → true
func Empty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == "" || val == "<no value>"
	case bool:
		return !val
	case int:
		return val == 0
	case int32:
		return val == 0
	case int64:
		return val == 0
	case float32:
		return val == 0
	case float64:
		return val == 0
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	}
	return fmt.Sprintf("%v", v) == ""
}

// NotEmpty returns `true` when the value is non-empty. The inverse of `empty`.
//
// Keywords: conditional, empty, check, boolean, present, exists
//
// Example:
//
//	{{ notEmpty .spec.mode }}  → true
func NotEmpty(v interface{}) bool {
	return !Empty(v)
}

// func isTruthy(v interface{}) bool {
// 	return !Empty(v)
// }

func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != ""
	case int, int8, int16, int32, int64:
		return val != 0
	case uint, uint8, uint16, uint32, uint64:
		return val != 0
	case float32:
		return val != 0.0
	case float64:
		return val != 0.0
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	default:
		// For any other type (e.g. struct), treat as true
		return true
	}
}

// BoolTernary is like `ternary` but requires a strict `bool` argument. Use this
// when the field is declared as a boolean in the CRD schema — it avoids the
// truthiness ambiguity of `ternary` for values like `0`, `""`, or `[]`.
//
// Keywords: conditional, branch, boolean, ternary, if-else, strict
//
// Example:
//
//	{{ boolTernary .spec.debug "debug" "info" }}  → "debug"
func BoolTernary(cond bool, trueVal, falseVal interface{}) interface{} {
	if cond {
		return trueVal
	}
	return falseVal
}

// BoolDefault returns the field's boolean value if it is a `bool`, otherwise
// returns `def`. Useful when a boolean field may be absent (nil).
//
// Keywords: conditional, boolean, default, fallback, absent, nil
//
// Example:
//
//	{{ boolDefault .spec.suspend false }}  → false
func BoolDefault(v interface{}, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}
