package gendoc

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	fns, domains, err := Parse("testdata/good")
	if err != nil {
		t.Fatal(err)
	}
	if len(domains) != 1 || domains[0] != (Domain{Name: "greet", Title: "Greetings", Intro: "Say hello, with a hyphen-ated word."}) {
		t.Errorf("domains = %+v", domains)
	}
	if len(fns) != 1 {
		t.Fatalf("fns = %+v", fns)
	}
	f := fns[0]
	if f.Name != "hello" || f.GoName != "Hello" || f.Domain != "greet" {
		t.Errorf("names = %q %q %q", f.Name, f.GoName, f.Domain)
	}
	if want := "hello returns a greeting for name.\n\n- A list item stays on its own line."; f.Description != want {
		t.Errorf("Description = %q, want %q", f.Description, want)
	}
	if f.Example != `{{ hello "kel" }}  → hello kel` {
		t.Errorf("Example = %q", f.Example)
	}
	if strings.Join(f.Keywords, ",") != "greet,hello" || strings.Join(f.SeeAlso, ",") != "hello" {
		t.Errorf("Keywords = %v, SeeAlso = %v", f.Keywords, f.SeeAlso)
	}
}

func TestParse_ReportsEveryProblem(t *testing.T) {
	_, _, err := Parse("testdata/bad")
	if err == nil {
		t.Fatal("want an error")
	}
	for _, want := range []string{
		`builder badFuncs: doc comment must start with "badFuncs — <Title>"`,
		`"upper" is registered as ToUpper, which is not declared in this package`,
		`NoDoc ("nodoc"): no doc comment`,
		`NoKeywords ("noKw"): no Keywords line`,
		`See also "missing" is not a registered function`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q:\n%v", want, err)
		}
	}
}
