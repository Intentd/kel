package kel_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/intentd/kel"
)

var (
	predicate = regexp.MustCompile(`^(is|has|all|any)[A-Z]|Detected$`)
	listName  = regexp.MustCompile(`(Names|IPs|Phases|Nodes|Reasons|Messages|Hosts|Taints|Modes|Endpoints|Types|Rules|Values|Ports|ReplicaSets)$`)
	countName = regexp.MustCompile(`(Count|Len)$`)
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

// Every KEL function has a Go-shaped signature, so the same function serves
// templates and Go callers: booleans, numbers and lists are their own types,
// never strings, and a function returns at most a value and an error.
func TestSignatures(t *testing.T) {
	for name, fn := range kel.FuncMap() {
		ft := reflect.TypeOf(fn)
		if ft.Kind() != reflect.Func {
			t.Errorf("%s: not a function", name)
			continue
		}
		switch ft.NumOut() {
		case 1:
		case 2:
			if ft.Out(1) != errorType {
				t.Errorf("%s: second result is %s, want error", name, ft.Out(1))
			}
		default:
			t.Errorf("%s: %d results, want a value and at most an error", name, ft.NumOut())
			continue
		}
		out := ft.Out(0)
		switch {
		case predicate.MatchString(name):
			if out.Kind() != reflect.Bool {
				t.Errorf("%s: a predicate returns %s, want bool", name, out)
			}
		case listName.MatchString(name) && out.Kind() != reflect.Slice:
			t.Errorf("%s: returns %s, want a list", name, out)
		case countName.MatchString(name) && !isNumber(out):
			t.Errorf("%s: a count returns %s, want a number", name, out)
		}
	}
}

func isNumber(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Uint64, reflect.Float64:
		return true
	}
	return false
}
