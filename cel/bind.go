package kelcel

import (
	"fmt"
	"reflect"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// form is one way to call a KEL function from CEL: the Go types of the
// arguments CEL passes, and whether the last one is the variadic slice.
type form struct {
	suffix string
	params []reflect.Type
	slice  bool
}

// function declares a KEL function in CEL from its Go signature. Each form is
// a global overload and, when its first argument is an object, a method too.
// A trailing variadic parameter gives three forms: without it, with one value,
// and with a list.
func function(name string, fn interface{}) (cel.EnvOption, error) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	if ft.NumOut() == 0 || ft.NumOut() > 2 || (ft.NumOut() == 2 && ft.Out(1) != errorType) {
		return nil, fmt.Errorf("results %s: want a value and at most an error", ft)
	}
	result, err := celType(ft.Out(0))
	if err != nil {
		return nil, fmt.Errorf("result: %w", err)
	}

	params := make([]reflect.Type, ft.NumIn())
	for i := range params {
		params[i] = ft.In(i)
	}
	forms := []form{{params: params}}
	if ft.IsVariadic() {
		base, last := params[:len(params)-1], params[len(params)-1]
		forms = []form{
			{suffix: "_list", params: params, slice: true},
			{suffix: "_one", params: append(append([]reflect.Type(nil), base...), last.Elem())},
		}
		if len(base) > 0 {
			forms = append(forms, form{suffix: "_none", params: base})
		}
	}

	var overloads []cel.FunctionOpt
	for _, f := range forms {
		args := make([]*cel.Type, len(f.params))
		for i, p := range f.params {
			if args[i], err = celType(p); err != nil {
				return nil, fmt.Errorf("argument %d: %w", i+1, err)
			}
		}
		f := f
		binding := cel.FunctionBinding(func(vals ...ref.Val) ref.Val { return call(fv, f, vals) })
		id := "kel_" + name + f.suffix
		overloads = append(overloads, cel.Overload(id, args, result, binding))
		if len(f.params) > 0 && f.params[0].Kind() == reflect.Interface {
			overloads = append(overloads, cel.MemberOverload(id+"_member", args, result, binding))
		}
	}
	return cel.Function(name, overloads...), nil
}

// celType is the CEL type for a Go parameter or result.
func celType(t reflect.Type) (*cel.Type, error) {
	switch t.Kind() {
	case reflect.Bool:
		return cel.BoolType, nil
	case reflect.String:
		return cel.StringType, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cel.IntType, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cel.UintType, nil
	case reflect.Float32, reflect.Float64:
		return cel.DoubleType, nil
	case reflect.Interface:
		return cel.DynType, nil
	case reflect.Map:
		return cel.MapType(cel.StringType, cel.DynType), nil
	case reflect.Slice, reflect.Array:
		elem, err := celType(t.Elem())
		if err != nil {
			return nil, err
		}
		return cel.ListType(elem), nil
	}
	return nil, fmt.Errorf("no CEL type for %s", t)
}

// call runs a KEL function with CEL values converted to the form's Go
// parameters, and converts its result back.
func call(fv reflect.Value, f form, vals []ref.Val) ref.Val {
	in := make([]reflect.Value, len(vals))
	for i, v := range vals {
		native, err := toGo(v, f.params[i])
		if err != nil {
			return types.NewErr("%v", err)
		}
		in[i] = native
	}
	var out []reflect.Value
	if f.slice {
		out = fv.CallSlice(in)
	} else {
		out = fv.Call(in)
	}
	if len(out) == 2 && !out[1].IsNil() {
		return types.NewErr("%v", out[1].Interface())
	}
	return types.DefaultTypeAdapter.NativeToValue(toPlain(out[0]))
}

// toGo converts a CEL value to a Go parameter. KEL reads objects as the API
// returns them, plain maps and lists, so a parameter that takes any value gets
// those, all the way down.
func toGo(v ref.Val, t reflect.Type) (reflect.Value, error) {
	if t.Kind() == reflect.Interface {
		native := Native(v)
		if native == nil {
			return reflect.Zero(t), nil
		}
		return reflect.ValueOf(&native).Elem(), nil
	}
	native, err := v.ConvertToNative(t)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(native), nil
}

// Native returns a CEL value as plain Go values, all the way down: bool,
// int64, uint64, float64, string, nil, []interface{} and
// map[string]interface{}, the shapes Kubernetes objects are made of. An object
// handed to CEL from Go is still its Go map underneath and comes back as it
// is; maps and lists built in the expression are converted element by
// element. Use it to read an expression's result, or to pass CEL values to Go
// functions that read objects.
func Native(v ref.Val) interface{} {
	switch native := v.Value().(type) {
	case map[string]interface{}, []interface{}:
		return native
	}
	switch c := v.(type) {
	case traits.Mapper:
		out := map[string]interface{}{}
		for it := c.Iterator(); it.HasNext() == types.True; {
			k := it.Next()
			out[fmt.Sprint(k.Value())] = Native(c.Get(k))
		}
		return out
	case traits.Lister:
		n, _ := c.Size().Value().(int64)
		out := make([]interface{}, 0, n)
		for i := int64(0); i < n; i++ {
			out = append(out, Native(c.Get(types.Int(i))))
		}
		return out
	}
	if v == types.NullValue {
		return nil
	}
	return v.Value()
}

var plainMapType = reflect.TypeOf(map[string]interface{}{})

// toPlain turns named types (KEL's State and Map) into the plain ones CEL's
// type adapter understands.
func toPlain(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return v.Bool()
	case reflect.Map:
		if v.Type() != plainMapType && v.Type().ConvertibleTo(plainMapType) {
			return v.Convert(plainMapType).Interface()
		}
	}
	return v.Interface()
}
