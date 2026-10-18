// Package kube holds KEL's functions on Kubernetes objects, in the shape the
// API returns them: the map an unstructured or dynamic client gives, or JSON
// decoded into a map. Functions about related objects (a Deployment's pods,
// an object's events) take those objects as an argument; where they come
// from is the caller's business.
package kube
