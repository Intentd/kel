// Command generate writes kel_generated.go, KEL's registry of functions and
// domains, from the doc comments in core, kube and ecosystem. Run it from the
// KEL root:
//
//	go run ./hack/generate
package main

import (
	"fmt"
	"os"

	"github.com/intentd/kel/gendoc"
)

// Registry is the file this command writes.
var Registry = gendoc.Registry{
	Package:   "kel",
	Generator: "hack/generate",
	Functions: "BuiltinFunctions",
	Domains:   "BuiltinDomains",
}

func main() {
	fns, domains, err := gendoc.Parse("core", "kube", "ecosystem")
	if err != nil {
		fail("doc comments are incomplete:\n" + err.Error())
	}
	src, err := gendoc.Render(Registry, fns, domains)
	if err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile("kel_generated.go", src, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("wrote %d functions in %d domains\n", len(fns), len(domains))
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
