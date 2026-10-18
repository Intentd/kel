package kube

import (
	"encoding/json"
	"reflect"

	utiljson "k8s.io/apimachinery/pkg/util/json"
)

// Helpers for reading Kubernetes objects in their native shape: the
// map[string]interface{} an unstructured or dynamic client returns, or JSON
// decoded into a map.
//
// Every function reads its input through object or objectList, so the other
// shapes Go callers have work too: *unstructured.Unstructured and its lists,
// and typed client-go objects such as *corev1.Pod, []corev1.Pod or
// *corev1.PodList. They are converted to the map shape first.

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
// (with items) as the API returns for a list call, in any shape object
// accepts. Anything else is empty.
func objectList(v interface{}) []map[string]interface{} {
	var items []interface{}
	switch t := v.(type) {
	case nil:
		return []map[string]interface{}{}
	case []interface{}:
		items = t
	case []map[string]interface{}:
		return t
	case map[string]interface{}:
		items, _ = t["items"].([]interface{})
	default:
		items = goItems(reflect.ValueOf(v))
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		if m := object(it); m != nil {
			out = append(out, m)
		}
	}
	return out
}

// goItems returns the elements of a Go slice, or of the Items field of a Go
// list type (*corev1.PodList, *unstructured.UnstructuredList); nil otherwise.
func goItems(rv reflect.Value) []interface{} {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		rv = rv.FieldByName("Items")
	}
	if rv.Kind() != reflect.Slice {
		return nil
	}
	out := make([]interface{}, rv.Len())
	for i := range out {
		el := rv.Index(i)
		if el.Kind() == reflect.Struct && el.CanAddr() {
			el = el.Addr() // convert in place; no copy of each object
		}
		out[i] = el.Interface()
	}
	return out
}

// object returns v as an object, nil when it is not one. A map is returned
// as is. An unstructured object gives its content. A typed object is
// converted to a map; when its TypeMeta is empty, as it is on objects from a
// typed list call, kind is set from the Go type's name (Pod, Deployment).
func object(v interface{}) map[string]interface{} {
	switch t := v.(type) {
	case nil:
		return nil
	case map[string]interface{}:
		return t
	case interface{ UnstructuredContent() map[string]interface{} }:
		if rv := reflect.ValueOf(t); rv.Kind() == reflect.Ptr && rv.IsNil() {
			return nil
		}
		return t.UnstructuredContent()
	}
	return typedObject(reflect.ValueOf(v))
}

// typedObject converts a Go struct, or a pointer to one, to its map shape.
func typedObject(rv reflect.Value) (m map[string]interface{}) {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	ptr := rv
	if rv.CanAddr() {
		ptr = rv.Addr()
	} else {
		ptr = reflect.New(rv.Type())
		ptr.Elem().Set(rv)
	}
	// unstructured.Unstructured passed by value: use its content.
	if u, ok := ptr.Interface().(interface{ UnstructuredContent() map[string]interface{} }); ok {
		return u.UnstructuredContent()
	}
	// A typed object's JSON is exactly what the API returns for it. The
	// apimachinery decoder gives whole numbers as int64, as a dynamic client does.
	b, err := json.Marshal(ptr.Interface())
	if err != nil || utiljson.Unmarshal(b, &m) != nil {
		return nil
	}
	if _, isAPIType := rv.Type().FieldByName("TypeMeta"); isAPIType {
		if k, _ := m["kind"].(string); k == "" {
			m["kind"] = rv.Type().Name()
		}
	}
	return m
}

// asObject is object with the comma-ok form of a type assertion.
func asObject(v interface{}) (map[string]interface{}, bool) {
	m := object(v)
	return m, m != nil
}
