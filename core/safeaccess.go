package core

import "text/template"

// safeAccessFuncs — Safe access
//
// Domain: safeaccess
//
// Read a nested field and fall back to a default when it is absent or of the
// wrong type, typed as a string, an integer or a boolean.
func safeAccessFuncs() template.FuncMap {
	return template.FuncMap{
		"getOr":       GetOr,
		"getStringOr": GetStringOr,
		"getIntOr":    GetIntOr,
		"getBoolOr":   GetBoolOr,
	}
}

// GetOr returns `val` if non-empty, otherwise `def`. The generic version —
// same semantics as `default` but reads more naturally when chaining with
// other functions.
//
// Keywords: safe, default, fallback, access, nil, absent, get
//
// Example:
//
//	{{ getOr .spec.workers 4 }}  → 4
func GetOr(val interface{}, def interface{}) interface{} {
	if Empty(val) {
		return def
	}
	return val
}

// GetStringOr returns `val` as a string if it is a non-empty string, otherwise
// returns `def`.
//
// Keywords: safe, default, fallback, string, typed, absent, get
//
// Example:
//
//	{{ getStringOr .spec.sidecarImage "envoyproxy/envoy:v1.30" }}  → "envoyproxy/envoy:v1.30"
func GetStringOr(val interface{}, def string) string {
	if s, ok := val.(string); ok && s != "" {
		return s
	}
	return def
}

// GetIntOr returns `val` as `int` if it is a numeric type (`int`, `int64`,
// `float64`), otherwise returns `def`.
//
// Keywords: safe, default, fallback, integer, typed, absent, get, number
//
// Example:
//
//	{{ getIntOr .deployment.spec.replicas 1 }}  → 3
//	{{ getIntOr .spec.workers 1 }}             → 1
func GetIntOr(val interface{}, def int) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return def
}

// GetBoolOr returns `val` as `bool` if it is a `bool`, otherwise returns `def`.
//
// Keywords: safe, default, fallback, boolean, typed, absent, get, bool
//
// Example:
//
//	{{ getBoolOr .spec.debug false }}    → true
//	{{ getBoolOr .spec.suspend false }}  → false
func GetBoolOr(val interface{}, def bool) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return def
}
