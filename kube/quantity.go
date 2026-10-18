package kube

import (
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/api/resource"
)

// quantityFuncs — Quantity
//
// Domain: quantity
//
// Work with Kubernetes resource quantities ("500m", "2Gi"): parse, compare,
// add and subtract them, keeping the canonical form.
func quantityFuncs() template.FuncMap {
	return template.FuncMap{
		"parseQuantity":  ParseQuantity,
		"formatQuantity": FormatQuantity,
		"sumQuantity":    SumQuantity,
		"minusQuantity":  SubtractQuantity,
	}
}

// ParseQuantity converts a Kubernetes quantity string to `float64`. Useful as
// input to math functions for arithmetic.
//
// Keywords: quantity, resource, cpu, memory, parse, float, kubernetes, budget
//
// Example:
//
//	{{ parseQuantity "500m" }}  → 0.5
//	{{ parseQuantity "1Gi" }}   → 1.073741824e+09
func ParseQuantity(q string) (float64, error) {
	parsed, err := resource.ParseQuantity(q)
	if err != nil {
		return 0, fmt.Errorf("parseQuantity: %q is not a valid Kubernetes quantity: %w", q, err)
	}
	return parsed.AsApproximateFloat64(), nil
}

// FormatQuantity converts a `float64` back to a canonical Kubernetes quantity
// string. Sub-unit CPU fractions are expressed in milli-cores.
//
// Keywords: quantity, resource, cpu, memory, format, string, kubernetes, canonical
//
// Example:
//
//	{{ formatQuantity 0.5 }}         → "500m"
//	{{ formatQuantity 1073741824 }}  → "1Gi"
func FormatQuantity(f float64) (string, error) {
	// Check if the value is an integer and >= 1 — likely a memory quantity
	if f == float64(int64(f)) && f >= 1 {
		q := resource.NewQuantity(int64(f), resource.BinarySI)
		return q.String(), nil
	}
	// Otherwise treat as CPU (milli‑units)
	millis := int64(f * 1000)
	q := resource.NewMilliQuantity(millis, resource.DecimalSI)
	return q.String(), nil
}

// SumQuantity adds two Kubernetes quantity strings and returns the canonical
// string sum. Both operands must be the same dimension (CPU or memory).
//
// Keywords: quantity, resource, cpu, memory, add, arithmetic, sum, budget
//
// Example:
//
//	{{ sumQuantity "100m" "200m" }}  → "300m"
//	{{ sumQuantity "1Gi" "512Mi" }}  → "1536Mi"
func SumQuantity(a, b string) (string, error) {
	qa, err := resource.ParseQuantity(a)
	if err != nil {
		return "", fmt.Errorf("sumQuantity: %q: %w", a, err)
	}
	qb, err := resource.ParseQuantity(b)
	if err != nil {
		return "", fmt.Errorf("sumQuantity: %q: %w", b, err)
	}
	qa.Add(qb)
	return qa.String(), nil
}

// SubtractQuantity subtracts the second Kubernetes quantity from the first and
// returns the canonical string representation of the difference. Both operands
// must be the same dimension (CPU or memory).
//
// Keywords: quantity, resource, cpu, memory, subtract, arithmetic, difference, budget
//
// Example:
//
//	{{ minusQuantity "1Gi" "512Mi" }}  → "512Mi"
func SubtractQuantity(a, b string) (string, error) {
	qa, err := resource.ParseQuantity(a)
	if err != nil {
		return "", fmt.Errorf("subtractQuantity: %q: %w", a, err)
	}
	qb, err := resource.ParseQuantity(b)
	if err != nil {
		return "", fmt.Errorf("subtractQuantity: %q: %w", b, err)
	}
	qa.Sub(qb)
	return qa.String(), nil
}
