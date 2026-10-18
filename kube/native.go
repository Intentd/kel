package kube

// Helpers for reading Kubernetes objects in their native shape: the
// map[string]interface{} an unstructured or dynamic client returns, or
// JSON decoded into a map.

// nested returns the value at path in obj, nil when any step is missing or
// not a map.
func nested(obj map[string]interface{}, path ...string) interface{} {
	var cur interface{} = obj
	for _, p := range path {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = m[p]
	}
	return cur
}

// nestedString returns the string at path in obj, "" when absent.
func nestedString(obj map[string]interface{}, path ...string) string {
	s, _ := nested(obj, path...).(string)
	return s
}

// objectList returns the objects in v: a list of objects, or a List object
// (with items) as the API returns for a list call. Anything else is empty.
func objectList(v interface{}) []map[string]interface{} {
	var items []interface{}
	switch t := v.(type) {
	case []interface{}:
		items = t
	case []map[string]interface{}:
		return t
	case map[string]interface{}:
		items, _ = t["items"].([]interface{})
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

// object returns v as an object, nil when it is not one.
func object(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}
