package kelcel

import (
	"fmt"

	"cel.dev/cel-go/cel"
)

// Env returns a CEL environment with KEL's library, each of vars declared as
// a variable of any type (objects, lists, values), and opts added: other
// libraries, extensions, further variables.
//
//	env, err := kelcel.Env([]string{"object"})
//	ast, iss := env.Compile(`conditionsState(object) == "healthy"`)
func Env(vars []string, opts ...cel.EnvOption) (*cel.Env, error) {
	all := make([]cel.EnvOption, 0, len(vars)+len(opts)+1)
	all = append(all, Library())
	for _, v := range vars {
		all = append(all, cel.Variable(v, cel.DynType))
	}
	all = append(all, opts...)
	env, err := cel.NewEnv(all...)
	if err != nil {
		return nil, fmt.Errorf("CEL environment: %w", err)
	}
	return env, nil
}

// Check reports whether expr compiles in Env(vars, opts...), without running
// it: a syntax error, an unknown function or variable, or a call with the
// wrong number or type of arguments fails here. It is the CEL counterpart of
// kel.Check for templates.
//
//	err := kelcel.Check(`conditionsState(object, "Ready")`, []string{"object"})
//	// found no matching overload for 'conditionsState'
func Check(expr string, vars []string, opts ...cel.EnvOption) error {
	env, err := Env(vars, opts...)
	if err != nil {
		return err
	}
	if _, iss := env.Compile(expr); iss.Err() != nil {
		return fmt.Errorf("compiling %q: %w", expr, iss.Err())
	}
	return nil
}
