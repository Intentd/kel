// Package anymap reads the map[string]interface{} values KEL passes around,
// including named types built on it, such as kube.Map.
package anymap

import "reflect"

var mapType = reflect.TypeOf(map[string]interface{}(nil))

// Of returns v as a map[string]interface{}: v itself, or v converted when its
// type is a named type with that underlying type. false for anything else.
func Of(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || !rv.Type().ConvertibleTo(mapType) || rv.Kind() != reflect.Map {
		return nil, false
	}
	return rv.Convert(mapType).Interface().(map[string]interface{}), true
}

// Strings returns the string values of m by key, as labels and annotations
// are. Values that are not strings are left out. Never nil.
func Strings(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
