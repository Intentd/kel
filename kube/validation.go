package kube

import (
	"text/template"

	"k8s.io/apimachinery/pkg/api/validate/content"
)

// kubeValidationFuncs — Kubernetes validation
//
// Domain: kubevalidation
//
// Check values against the rules the API server applies to names and metadata:
// label keys and values, annotation keys, and DNS-1123 subdomains, so a bad
// value is caught with a clear message before the API server rejects it.
func kubeValidationFuncs() template.FuncMap {
	return template.FuncMap{
		"isValidLabelValue":    IsValidLabelValue,
		"isValidLabelKey":      IsValidLabelKey,
		"isValidAnnotationKey": IsValidAnnotationKey,
		"isDNS1123Subdomain":   IsDNS1123Subdomain,
	}
}

// IsValidLabelValue reports whether a string is a syntactically valid
// Kubernetes label value: max 63 characters, alphanumeric, may contain `-`,
// `_`, `.`, and must start and end with an alphanumeric character. An empty
// string is valid — Kubernetes treats an absent label value as permitted.
//
// Keywords: kubernetes, label, value, validate, valid, boolean, format
//
// Example:
//
//	{{ isValidLabelValue "team-payments" }}  → true
//	{{ isValidLabelValue "team payments" }}  → false
func IsValidLabelValue(value string) bool {
	return len(content.IsLabelValue(value)) == 0
}

// IsValidLabelKey reports whether a string is a syntactically valid Kubernetes
// label key: `[prefix/]name` — `name` is alphanumeric (max 63 chars, may
// contain `-`, `_`, `.`), and `prefix`, if present, must be a valid DNS
// subdomain (max 253 chars).
//
// Keywords: kubernetes, label, key, validate, valid, boolean, format, prefix
//
// Example:
//
//	{{ isValidLabelKey "app.kubernetes.io/name" }}  → true
func IsValidLabelKey(key string) bool {
	return len(content.IsLabelKey(key)) == 0
}

// IsValidAnnotationKey reports whether a string is a syntactically valid
// Kubernetes annotation key — the same `[prefix/]name` format as a label key.
// Annotation values have no Kubernetes format restriction of their own (that's
// what distinguishes them from labels), so there is no
// `isValidAnnotationValue` function.
//
// Keywords: kubernetes, annotation, key, validate, valid, boolean, format, prefix
//
// Example:
//
//	{{ isValidAnnotationKey "example.com/owner" }}  → true
func IsValidAnnotationKey(key string) bool {
	return len(content.IsLabelKey(key)) == 0
}

// IsDNS1123Subdomain reports whether a string is a syntactically valid
// DNS-1123 subdomain: lowercase alphanumeric segments (may contain `-`)
// separated by `.`, max 253 characters. This is the format Kubernetes requires
// for object names in most resource types, and for the prefix portion of a
// label/annotation key.
//
// Keywords: kubernetes, dns, subdomain, name, validate, valid, boolean, format
//
// Example:
//
//	{{ isDNS1123Subdomain "api.example.com" }}  → true
//	{{ isDNS1123Subdomain "API_Example" }}      → false
func IsDNS1123Subdomain(value string) bool {
	return len(content.IsDNS1123Subdomain(value)) == 0
}
