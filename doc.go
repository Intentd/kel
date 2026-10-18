// Package kel is KEL, the Kubernetes Expression Library: functions that know
// Kubernetes, for any Go template or expression engine. It is a library, not
// a language: no parser, no grammar, no runtime of its own.
//
// FuncMap returns every function as a template.FuncMap, and Builtin describes
// them: domain, description, example and keywords, from their doc comments.
// Register adds your own functions beside KEL's, and Override replaces one
// on purpose.
//
// The functions live in three packages, by what they read:
//
//   - core: plain values (strings, numbers, time, cron, collections, …)
//   - kube: Kubernetes objects, in the shape the API returns them
//   - ecosystem: what the tools around Kubernetes produce, such as a
//     Prometheus query result or a container image reference
//
// Every function takes its inputs as arguments and does no I/O: a function
// about pods is given the pods.
package kel
