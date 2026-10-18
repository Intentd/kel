package kel

import (
	"maps"
	"text/template"

	"github.com/intentd/kel/core"
	"github.com/intentd/kel/ecosystem"
	"github.com/intentd/kel/internal/funcmap"
	"github.com/intentd/kel/kube"
)

// funcs is the package-level function map.
// Built once at init time and never handed out, so callers cannot change it.
var funcs = buildFunctions()

// FuncMap returns KEL's functions as a template.FuncMap: a copy, so a caller can
// add its own (with Register) without changing KEL's.
//
//	tmpl, err := template.New("f").
//	    Option("missingkey=zero").
//	    Funcs(kel.FuncMap()).
//	    Parse(value)
func FuncMap() template.FuncMap {
	return maps.Clone(funcs)
}

// buildFunctions constructs the complete function map from all domains.
// Called once at package init.
func buildFunctions() template.FuncMap {
	m := template.FuncMap{}
	Register(m, kube.Functions())
	Register(m, core.Functions())
	Register(m, ecosystem.Functions())
	return m
}

// Register adds src's functions to dst; a name already in dst panics, since
// two functions would answer to one name. Use it to add your own functions
// beside KEL's.
func Register(dst, src template.FuncMap) { funcmap.Register(dst, src) }

// Override replaces functions already in dst with src's, on purpose; a name
// not in dst panics.
func Override(dst, src template.FuncMap) { funcmap.Override(dst, src) }
