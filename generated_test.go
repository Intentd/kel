package kel_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/intentd/kel/gendoc"
)

// TestGeneratedIsCurrent fails when a doc comment changed without
// regenerating: run go generate.
func TestGeneratedIsCurrent(t *testing.T) {
	fns, domains, err := gendoc.Parse("core", "kube", "ecosystem")
	if err != nil {
		t.Fatalf("doc comments are incomplete:\n%v", err)
	}
	want, err := gendoc.Render(gendoc.Registry{
		Package: "kel", Generator: "hack/generate",
		Functions: "BuiltinFunctions", Domains: "BuiltinDomains",
	}, fns, domains)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("kel_generated.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("kel_generated.go is out of date: run go generate")
	}
}
