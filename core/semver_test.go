package core

import (
	"bytes"
	"testing"
	"text/template"
)

// The semver template functions are the Go helpers registered directly: a
// template gets the value, or an error for a version that does not parse.
func TestSemverFuncsInTemplates(t *testing.T) {
	render := func(expr string) (string, error) {
		tmpl, err := template.New("t").Funcs(semverFuncs()).Parse(expr)
		if err != nil {
			return "", err
		}
		var b bytes.Buffer
		err = tmpl.Execute(&b, nil)
		return b.String(), err
	}
	cases := map[string]string{
		`{{ semverMajor "1.2.3" }}`:                                            "1",
		`{{ semverMinor "v1.2.3" }}`:                                           "2",
		`{{ semverPatch "1.2.3-rc.1" }}`:                                       "3",
		`{{ semverValid "1.2.3" }}`:                                            "true",
		`{{ semverValid "banana" }}`:                                           "false",
		`{{ semverCompare "1.2.3" "1.10.0" }}`:                                 "-1",
		`{{ semverBump "1.2.3" "minor" }}`:                                     "1.3.0",
		`{{ semverConstraint "1.2.3" ">= 1.0.0" }}`:                            "true",
		`{{ if semverConstraint "0.9.0" ">= 1.0.0" }}yes{{ else }}no{{ end }}`: "no",
	}
	for expr, want := range cases {
		got, err := render(expr)
		if err != nil || got != want {
			t.Errorf("%s = %q, %v; want %q", expr, got, err, want)
		}
	}
	for _, expr := range []string{`{{ semverMajor "banana" }}`, `{{ semverBump "1.2.3" "sideways" }}`, `{{ semverConstraint "1.2.3" "not a constraint" }}`} {
		if _, err := render(expr); err == nil {
			t.Errorf("%s: want an error", expr)
		}
	}
}
