package good

import "text/template"

// greetFuncs — Greetings
//
// Domain: greet
//
// Say hello, with a hyphen-
// ated word.
func greetFuncs() template.FuncMap {
	return template.FuncMap{"hello": Hello}
}

// Hello returns a greeting for name.
//
// - A list item stays on its own line.
//
// Keywords: greet, hello
//
// Example:
//
//	{{ hello "kel" }}  → hello kel
//
// See also: hello
func Hello(name string) string { return "hello " + name }
