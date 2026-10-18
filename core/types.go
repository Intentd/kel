package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// typeFuncs — Types
//
// Domain: types
//
// Inspect and convert types: what type a value has (map, slice, string,
// number, bool, null), whether it is empty, and conversion to strings,
// numbers, booleans, lists and JSON.
func typeFuncs() template.FuncMap {
	return template.FuncMap{
		"toInt":    ToInt,
		"toFloat":  ToFloat,
		"toBool":   ToBool,
		"toString": ToString,
		"toJSON":   ToJSON,
		"toList":   ToList,

		// typeOf v — returns the type name as a string
		// "string", "number", "bool", "map", "slice", "null", "unknown"
		"typeOf": TypeOf,

		// len v — returns element count or string length
		// Overrides the built-in template len with one that handles maps
		"len": KelLen,

		// Shorthand type checks
		"typeMap":     TypeMap,
		"typeList":    TypeList,
		"typeString":  TypeString,
		"typeNumber":  TypeNumber,
		"typeBool":    TypeBool,
		"typeNull":    TypeNull,
		"isEmpty":     IsEmpty,
		"isScalar":    IsScalar,
		"isComposite": IsComposite,
	}
}

// ToInt converts any value to `int64`. Truncates floats.
//
// Keywords: type, convert, integer, coerce, number, cast
//
// Example:
//
//	{{ toInt "3" }}  → 3
//	{{ toInt 3.7 }}  → 3
func ToInt(v interface{}) (int64, error) {
	if b, ok := v.(bool); ok {
		if b {
			return 1, nil
		}
		return 0, nil
	}
	f, err := anyToFloat(v)
	if err != nil {
		return 0, fmt.Errorf("toInt: %w", err)
	}
	return int64(f), nil
}

// ToFloat converts any value to `float64`.
//
// Keywords: type, convert, float, coerce, number, cast, decimal
//
// Example:
//
//	{{ toFloat "0.75" }}  → 0.75
func ToFloat(v interface{}) (float64, error) {
	return anyToFloat(v)
}

// ToBool converts a value to `bool`. Returns an error for unrecognized
// strings.
//
// Keywords: type, convert, boolean, coerce, parse, cast, bool
//
// Example:
//
//	{{ toBool "yes" }}  → true
func ToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case int, int32, int64, float32, float64:
		f, _ := anyToFloat(v)
		return f != 0, nil
	case string:
		switch val {
		case "true", "True", "TRUE", "1", "yes", "YES", "Yes", "on", "ON":
			return true, nil
		case "false", "False", "FALSE", "0", "no", "NO", "No", "off", "OFF", "":
			return false, nil
		}
		return false, fmt.Errorf("toBool: unrecognised value %q", val)
	}
	return false, fmt.Errorf("toBool: cannot convert %T", v)
}

// ToString converts any value to its string representation. Uses
// `fmt.Sprintf("%v", v)`.
//
// Keywords: type, convert, string, coerce, format, cast, stringify
//
// Example:
//
//	{{ toString .deployment.spec.replicas }}  → "3"
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// ToJSON marshals any value to a JSON string. Useful for embedding structured
// data in annotations, ConfigMap values, or status fields.
//
// Keywords: data, json, marshal, encode, serialize, annotation, configmap, string
//
// Example:
//
//	{{ toJSON .spec.selector }}  → {"app":"web","tier":"frontend"}
func ToJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("toJson: %w", err)
	}
	return string(b), nil
}

// ToList converts a comma-separated string to a list of trimmed strings.
// Returns an empty list for empty strings. Use it where a comma-separated
// value, such as an annotation, needs to be treated as a list.
//
// Keywords: type, convert, list, slice, split, parse, csv
//
// Example:
//
//	{{ toList "a, b, c" }}  → [a b c]
func ToList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// TypeOf returns the runtime type name of any value. The canonical way to
// branch on field type.
//
// Keywords: type, inspect, runtime, branch, kind, detect
//
// Example:
//
//	{{ typeOf .spec.schedule }}  → "map"
//	{{ typeOf .spec.regions }}   → "slice"
func TypeOf(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case string:
		return "string"
	case float64, float32:
		return "number"
	case int, int32, int64, uint, uint32, uint64:
		return "number"
	case bool:
		return "bool"
	case map[string]interface{}:
		return "map"
	case []interface{}:
		return "slice"
	default:
		return "unknown"
	}
}

// KelLen returns the length of a string, slice, or map. Overrides Go's
// built-in `len` with one that handles all three types uniformly.
//
// Keywords: type, length, count, string, slice, map, size
//
// Example:
//
//	{{ len .spec.regions }}            → 2
//	{{ len .deployment.metadata.name }}  → 3
func KelLen(v interface{}) int {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case string:
		return len(t)
	case []interface{}:
		return len(t)
	case map[string]interface{}:
		return len(t)
	default:
		return 0
	}
}

// IsEmpty returns `true` when the value is nil, an empty string, an empty
// slice, or an empty map.
//
// Keywords: type, empty, nil, check, boolean, absent
//
// Example:
//
//	{{ isEmpty .spec.logLevel }}  → true
//	{{ isEmpty .spec.regions }}   → false
func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return t == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	default:
		return false
	}
}

// IsScalar reports whether v is a single value: a string, number or bool.
//
// Keywords: type, scalar, predicate, string, number, bool
//
// Example:
//
//	{{ isScalar .deployment.spec.replicas }}  → true
//	{{ isScalar .spec.selector }}  → false
func IsScalar(v interface{}) bool {
	switch TypeOf(v) {
	case "string", "number", "bool":
		return true
	default:
		return false
	}
}

// IsComposite reports whether v holds other values: a map or a slice.
//
// Keywords: type, composite, predicate, map, list, slice
//
// Example:
//
//	{{ isComposite .spec.selector }}  → true
//	{{ isComposite .spec.replicas }}  → false
func IsComposite(v interface{}) bool {
	switch TypeOf(v) {
	case "map", "slice":
		return true
	default:
		return false
	}
}

// TypeMap reports whether v is a map (a YAML object); the same as eq (typeOf v) "map".
//
// Keywords: type, predicate, map, check
//
// Example:
//
//	{{ if typeMap .spec.schedule }}…{{ end }}
func TypeMap(v interface{}) bool {
	return TypeOf(v) == "map"
}

// TypeList reports whether v is a list (a YAML array); the same as eq (typeOf v) "slice".
//
// Keywords: type, predicate, list, check
//
// Example:
//
//	{{ if typeList .spec.schedule }}…{{ end }}
func TypeList(v interface{}) bool {
	return TypeOf(v) == "slice"
}

// TypeString reports whether v is a string; the same as eq (typeOf v) "string".
//
// Keywords: type, predicate, string, check
//
// Example:
//
//	{{ if typeString .spec.schedule }}…{{ end }}
func TypeString(v interface{}) bool {
	return TypeOf(v) == "string"
}

// TypeNumber reports whether v is a number; the same as eq (typeOf v) "number".
//
// Keywords: type, predicate, number, check
//
// Example:
//
//	{{ if typeNumber .spec.schedule }}…{{ end }}
func TypeNumber(v interface{}) bool {
	return TypeOf(v) == "number"
}

// TypeBool reports whether v is a boolean; the same as eq (typeOf v) "bool".
//
// Keywords: type, predicate, bool, check
//
// Example:
//
//	{{ if typeBool .spec.schedule }}…{{ end }}
func TypeBool(v interface{}) bool {
	return TypeOf(v) == "bool"
}

// TypeNull reports whether v is null or absent; the same as eq (typeOf v) "null".
//
// Keywords: type, predicate, null, check
//
// Example:
//
//	{{ if typeNull .spec.schedule }}…{{ end }}
func TypeNull(v interface{}) bool {
	return TypeOf(v) == "null"
}
