package kelcel

import (
	"fmt"
	"sort"
	"text/template"

	"cel.dev/cel-go/cel"

	"github.com/intentd/kel"
	"github.com/intentd/kel/ecosystem"
	"github.com/intentd/kel/kube"
)

// excluded names the kube and ecosystem functions not offered, and why.
var excluded = map[string]string{
	"podRestartedWithin":      "reads the clock: a CEL expression must give the same answer every time",
	"cronJobNextScheduleTime": "reads the clock: a CEL expression must give the same answer every time",
}

// renamed gives the CEL name of a KEL function whose own name is a reserved
// word in CEL. Every other function keeps its KEL name.
var renamed = map[string]string{
	"namespace": "namespaceOf",
}

// celName is the name a KEL function has in CEL.
func celName(kelName string) string {
	if n, ok := renamed[kelName]; ok {
		return n
	}
	return kelName
}

// coreDomains are KEL's general domains offered too: CEL has nothing for
// cron expressions or semantic versions, and these are transforms people
// would otherwise write themselves. CEL's own strings, lists, maps, math and
// conversions cover the rest of core.
var coreDomains = map[string]bool{"cron": true, "semver": true}

// offered is every kube and ecosystem function, and those of coreDomains, by
// KEL name, except the excluded ones.
func offered() template.FuncMap {
	out := template.FuncMap{}
	add := func(name string, fn any) {
		if _, skip := excluded[name]; !skip {
			out[name] = fn
		}
	}
	for _, funcs := range []template.FuncMap{kube.Functions(), ecosystem.Functions()} {
		for name, fn := range funcs {
			add(name, fn)
		}
	}
	all := kel.FuncMap()
	for _, f := range kel.Builtin().Functions {
		if coreDomains[f.Domain] {
			add(f.Name, all[f.Name])
		}
	}
	return out
}

// Library returns KEL's functions as a CEL library.
func Library() cel.EnvOption {
	return cel.Lib(library{})
}

// Functions returns the functions Library offers: each name in CEL, mapped to
// the KEL function it calls. They differ only where KEL's name is reserved in
// CEL (namespace is namespaceOf).
func Functions() map[string]string {
	out := map[string]string{}
	for name := range offered() {
		out[celName(name)] = name
	}
	return out
}

// Excluded returns the kube and ecosystem functions Library does not offer,
// each with the reason.
func Excluded() map[string]string {
	out := make(map[string]string, len(excluded))
	for name, why := range excluded {
		out[name] = why
	}
	return out
}

type library struct{}

func (library) LibraryName() string {
	return "github.com/intentd/kel"
}

func (library) CompileOptions() []cel.EnvOption {
	funcs := offered()
	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}
	sort.Strings(names)

	opts := make([]cel.EnvOption, 0, len(names))
	for _, name := range names {
		opt, err := function(celName(name), funcs[name])
		if err != nil {
			panic(fmt.Sprintf("kelcel: %s: %v", name, err))
		}
		opts = append(opts, opt)
	}
	return opts
}

func (library) ProgramOptions() []cel.ProgramOption {
	return nil
}
