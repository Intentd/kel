package kel_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"

	"github.com/intentd/kel"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		name string
		expr string
		err  string // a substring of the error; "" means valid
	}{
		{"one argument", `{{ conditionsState .obj }}`, ""},
		{"too many", `{{ conditionsState .obj "Ready" }}`, "conditionsState takes 1 argument, got 2"},
		{"too few", `{{ hasCondition .obj }}`, "hasCondition takes 2 arguments, got 1"},
		{"piped value counts", `{{ .obj | conditionsState }}`, ""},
		{"piped value too many", `{{ .obj | hasCondition "Ready" "x" }}`, "hasCondition takes 2 arguments, got 3"},
		{"parenthesised", `{{ firstState (conditionsState .obj) (expectState .v true) }}`, "expectState takes 3 arguments, got 2"},
		{"variadic, none", `{{ firstState }}`, ""},
		{"variadic, several", `{{ list 1 2 3 }}`, ""},
		{"inside if and range", `{{ if hasCondition .obj "Ready" }}{{ range .items }}{{ name . "x" }}{{ end }}{{ end }}`, "name takes 1 argument, got 2"},
		{"inside with, else", `{{ with .obj }}ok{{ else }}{{ phase }}{{ end }}`, "phase takes 1 argument, got 0"},
		{"template built-ins are left to Go", `{{ printf "%d" 1 2 }} {{ eq 1 1 }}`, ""},
		{"unknown function", `{{ notAFunction .obj }}`, `function "notAFunction" not defined`},
		{"syntax", `{{ conditionsState .obj `, "unclosed action"},
		{"no template", `just text`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := kel.Check(c.expr, nil)
			if c.err == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), c.err)
			}
		})
	}
}

func TestCheckWithOwnFunctions(t *testing.T) {
	funcs := kel.FuncMap()
	kel.Register(funcs, template.FuncMap{"greet": func(name string) string { return "hi " + name }})

	assert.NoError(t, kel.Check(`{{ greet "you" }} {{ conditionsState .obj }}`, funcs))
	err := kel.Check(`{{ greet }}`, funcs)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "greet takes 1 argument, got 0")
	}
	// The error says where: the template's line and column.
	assert.True(t, strings.Contains(err.Error(), ":1:"), err.Error())
}

func TestCheckLeavesKELUnchanged(t *testing.T) {
	before := len(kel.FuncMap())
	_ = kel.Check(`{{ conditionsState .obj }}`, nil)
	assert.Equal(t, before, len(kel.FuncMap()))
}
