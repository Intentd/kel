package ecosystem

import (
	"text/template"

	"github.com/intentd/kel/internal/funcmap"
)

func Functions() template.FuncMap {
	m := template.FuncMap{}
	register(m, validationFuncs())
	register(m, prometheusFuncs())
	return m
}

// register merges src into dst.
func register(dst, src template.FuncMap) {
	funcmap.Register(dst, src)
}
