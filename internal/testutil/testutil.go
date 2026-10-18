// Package testutil holds what KEL's tests share: reading fixtures as an API
// client would, and turning them into the typed client-go objects Go
// programs hold. Only tests import it.
package testutil

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	utiljson "k8s.io/apimachinery/pkg/util/json"
)

// Testdata returns the path of a file under the repository's testdata
// directory, wherever the calling test runs from.
func Testdata(elem ...string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(append([]string{filepath.Dir(file), "..", "..", "testdata"}, elem...)...)
}

// ReadJSON decodes a JSON file under testdata into v the way a dynamic client
// does: whole numbers as int64, others as float64.
func ReadJSON(t testing.TB, v interface{}, elem ...string) {
	t.Helper()
	raw, err := os.ReadFile(Testdata(elem...))
	if err != nil {
		t.Fatal(err)
	}
	if err := utiljson.Unmarshal(raw, v); err != nil {
		t.Fatal(err)
	}
}

// Typed converts obj, a map in the shape the API returns, to a pointer to a
// new value of proto's type (such as corev1.Pod{}), with TypeMeta empty as on
// objects from a typed list call. A field the type doesn't have fails the
// test: the fixture isn't a valid object of that type.
func Typed(t testing.TB, obj interface{}, proto interface{}) interface{} {
	t.Helper()
	return typed(t, obj, reflect.TypeOf(proto)).Interface()
}

// TypedList converts a list of maps to a slice of proto's type ([]corev1.Pod),
// as the Items of a typed list call.
func TypedList(t testing.TB, items []interface{}, proto interface{}) interface{} {
	t.Helper()
	typ := reflect.TypeOf(proto)
	s := reflect.MakeSlice(reflect.SliceOf(typ), 0, len(items))
	for _, it := range items {
		s = reflect.Append(s, typed(t, it, typ).Elem())
	}
	return s.Interface()
}

func typed(t testing.TB, obj interface{}, typ reflect.Type) reflect.Value {
	t.Helper()
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	ptr := reflect.New(typ)
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(ptr.Interface()); err != nil {
		t.Fatalf("not a valid %s: %v", typ, err)
	}
	if tm := ptr.Elem().FieldByName("TypeMeta"); tm.IsValid() {
		tm.Set(reflect.Zero(tm.Type()))
	}
	return ptr
}
