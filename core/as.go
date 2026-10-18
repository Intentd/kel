package core

import (
	"encoding/json"
	"text/template"

	"gopkg.in/yaml.v3"
)

// asFuncs — Collections
//
// Domain: collections
//
// Convert a value that may arrive in several shapes (a native list or map, or
// the same as a JSON or YAML string) into a list, a map or a string, so the
// functions after it see one shape.
func asFuncs() template.FuncMap {
	return template.FuncMap{
		"asList":   AsList,
		"asMap":    AsMap,
		"asString": AsString,
	}
}

// AsList converts input to `[]interface{}`. Accepts native slice, YAML list
// string, or JSON array string.
//
// Keywords: list, convert, parse, json, yaml, slice, coerce
//
// Example:
//
//	{{ asList "[\"us-east-1\",\"eu-west-1\"]" }}  → [us-east-1 eu-west-1]
//	{{ len (asList .spec.regions) }}  → 2
func AsList(input any) []any {
	switch v := input.(type) {

	case []any:
		return v

	case string:
		// Try YAML
		var y any
		if err := yaml.Unmarshal([]byte(v), &y); err == nil {
			if list, ok := y.([]any); ok {
				return list
			}
		}

		// Try JSON
		var j []any
		if err := json.Unmarshal([]byte(v), &j); err == nil {
			return j
		}

		return []any{}

	default:
		return []any{}
	}
}

// AsMap converts input to `map[string]interface{}`. Accepts native map, YAML
// map string, or JSON object string.
//
// Keywords: map, convert, parse, json, yaml, coerce, object
//
// Example:
//
//	{{ (asMap "{\"tier\":\"web\"}").tier }}  → "web"
func AsMap(input any) map[string]any {
	switch v := input.(type) {

	case map[string]any:
		return v

	case string:
		// Try YAML
		var y any
		if err := yaml.Unmarshal([]byte(v), &y); err == nil {
			if m, ok := y.(map[string]any); ok {
				return m
			}
		}

		// Try JSON
		var j map[string]any
		if err := json.Unmarshal([]byte(v), &j); err == nil {
			return j
		}

		return map[string]any{}

	default:
		return map[string]any{}
	}
}

// AsString converts any value to a string. For maps and slices this produces a
// JSON encoding.
//
// Keywords: string, convert, serialize, json, coerce, stringify
//
// Example:
//
//	{{ asString 3 }}                   → "3"
//	{{ asString .deployment.spec.replicas }}  → "3"
func AsString(input any) string {
	switch v := input.(type) {

	case string:
		return v

	case []byte:
		return string(v)

	default:
		// Try JSON encode → string
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}
		return ""
	}
}
