package kel

import (
	"slices"
	"strings"
)

//go:generate go run ./hack/generate

// FunctionInfo describes one function, from its doc comment.
type FunctionInfo struct {
	Name        string   // template name, e.g. "cronValid"
	Domain      string   // e.g. "cron"
	Description string   // what it does
	Example     string   // a template using it
	SeeAlso     []string // related functions
	Keywords    []string // search terms
}

// DomainInfo describes one group of functions, from its builder's doc comment.
type DomainInfo struct {
	Name  string // e.g. "cron"
	Title string // e.g. "Cron"
	Intro string // what the functions are for
}

// Registry is a set of documented functions and their domains. Builtin
// returns KEL's; a program that adds functions can merge its own with With.
type Registry struct {
	Functions []FunctionInfo // sorted by domain, then name
	Domains   []DomainInfo   // sorted by name
}

// Builtin returns the registry of KEL's functions.
func Builtin() Registry {
	return Registry{Functions: BuiltinFunctions, Domains: BuiltinDomains}
}

// With returns r merged with other. A function in other replaces the one in
// r with the same name; a domain in both keeps r's title and both intros.
func (r Registry) With(other Registry) Registry {
	out := Registry{}
	names := map[string]bool{}
	for _, f := range other.Functions {
		names[f.Name] = true
	}
	for _, f := range r.Functions {
		if !names[f.Name] {
			out.Functions = append(out.Functions, f)
		}
	}
	out.Functions = append(out.Functions, other.Functions...)
	slices.SortFunc(out.Functions, func(a, b FunctionInfo) int {
		if c := strings.Compare(a.Domain, b.Domain); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})

	out.Domains = slices.Clone(r.Domains)
	for _, d := range other.Domains {
		i := slices.IndexFunc(out.Domains, func(x DomainInfo) bool { return x.Name == d.Name })
		switch {
		case i < 0:
			out.Domains = append(out.Domains, d)
		case d.Intro != "" && !strings.Contains(out.Domains[i].Intro, d.Intro):
			out.Domains[i].Intro = strings.TrimSpace(out.Domains[i].Intro + "\n\n" + d.Intro)
		}
	}
	slices.SortFunc(out.Domains, func(a, b DomainInfo) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// ByDomain returns the functions in domain (case-insensitive).
func (r Registry) ByDomain(domain string) []FunctionInfo {
	var out []FunctionInfo
	for _, f := range r.Functions {
		if strings.EqualFold(f.Domain, domain) {
			out = append(out, f)
		}
	}
	return out
}

// Get returns the function called name (case-insensitive).
func (r Registry) Get(name string) (FunctionInfo, bool) {
	for _, f := range r.Functions {
		if strings.EqualFold(f.Name, name) {
			return f, true
		}
	}
	return FunctionInfo{}, false
}

// Domain returns the domain called name (case-insensitive).
func (r Registry) Domain(name string) (DomainInfo, bool) {
	for _, d := range r.Domains {
		if strings.EqualFold(d.Name, name) {
			return d, true
		}
	}
	return DomainInfo{}, false
}

// Search returns the functions whose name, description, example or a
// keyword contains term (case-insensitive).
func (r Registry) Search(term string) []FunctionInfo {
	term = strings.ToLower(term)
	var out []FunctionInfo
	for _, f := range r.Functions {
		if strings.Contains(strings.ToLower(f.Name), term) ||
			strings.Contains(strings.ToLower(f.Description), term) ||
			strings.Contains(strings.ToLower(f.Example), term) ||
			slices.ContainsFunc(f.Keywords, func(k string) bool { return strings.Contains(strings.ToLower(k), term) }) {
			out = append(out, f)
		}
	}
	return out
}
