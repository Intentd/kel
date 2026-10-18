package kube

import (
	"text/template"

	"github.com/intentd/kel/internal/funcmap"
)

func Functions() template.FuncMap {
	m := template.FuncMap{}
	register(m, kubernetesFuncs())
	register(m, kubeValidationFuncs())
	register(m, quantityFuncs())
	register(m, containerFuncs())
	register(m, serviceFuncs())
	register(m, replicaFuncs())
	register(m, podFuncs())
	register(m, jobFuncs())
	register(m, eventFuncs())
	register(m, pvcFuncs())
	register(m, ingressFuncs())
	register(m, hpaFuncs())
	register(m, nodeFuncs())
	register(m, replicaSetFuncs())
	register(m, statefulSetFuncs())
	register(m, fieldFuncs())
	return m
}

// register merges src into dst.
func register(dst, src template.FuncMap) {
	funcmap.Register(dst, src)
}
