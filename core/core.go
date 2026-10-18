package core

import (
	"text/template"

	"github.com/intentd/kel/internal/funcmap"
)

// Function builds
func Functions() template.FuncMap {
	m := template.FuncMap{}
	register(m, asFuncs())
	register(m, cronFuncs())
	register(m, domainFuncs())
	register(m, timeFuncs())
	register(m, dataFuncs())
	register(m, stringFuncs())
	register(m, mathFuncs())
	register(m, typeFuncs())
	register(m, conditionalFuncs())
	register(m, randomFuncs())
	register(m, listMapFuncs())
	register(m, safeAccessFuncs())
	register(m, semverFuncs())
	register(m, netFuncs())
	register(m, validationFuncs())
	return m
}

// register merges src into dst.
func register(dst, src template.FuncMap) {
	funcmap.Register(dst, src)
}
