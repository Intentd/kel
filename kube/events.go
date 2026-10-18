package kube

import "text/template"

// eventFuncs — Warnings
//
// Domain: warnings
//
// Answer questions about Kubernetes Events, as the API returns them (core/v1
// or events.k8s.io/v1), passed as a list or an EventList. Only Warning events
// count, so a caller can pass all of an object's events.
func eventFuncs() template.FuncMap {
	return template.FuncMap{
		"hasWarnings":     HasWarnings,
		"warningCount":    WarningCount,
		"firstWarning":    FirstWarning,
		"warningMessages": WarningMessages,
		"warningReasons":  WarningReasons,
	}
}

// HasWarnings reports whether any of the events is a Warning.
//
// Keywords: events, warnings, kubernetes, boolean, alert, health
//
// Example:
//
//	{{ hasWarnings .events }}
func HasWarnings(events interface{}) bool {
	return len(warnings(events)) > 0
}

// WarningCount returns the number of Warning events.
//
// Keywords: events, warnings, count, int, total
//
// Example:
//
//	{{ warningCount .events }}  → 2
func WarningCount(events interface{}) int {
	return len(warnings(events))
}

// FirstWarning returns the message of the first Warning event; "" when there
// is none.
//
// Keywords: events, warnings, message, string, first, alert
//
// Example:
//
//	{{ firstWarning .events }}  → "Failed to pull image registry.example.com/web:2.0: not found"
func FirstWarning(events interface{}) string {
	ws := warnings(events)
	if len(ws) == 0 {
		return ""
	}
	return eventMessage(ws[0])
}

// WarningMessages returns the message of each Warning event, in order.
//
// Keywords: events, warnings, messages, list, string, all
//
// Example:
//
//	{{ warningMessages .events }}
//	→ [Back-off restarting failed container Liveness probe failed]
func WarningMessages(events interface{}) []string {
	var out []string
	for _, e := range warnings(events) {
		if m := eventMessage(e); m != "" {
			out = append(out, m)
		}
	}
	return out
}

// WarningReasons returns the reason of each Warning event, in order.
//
// Keywords: events, warnings, reasons, list, string, deduplicated
//
// Example:
//
//	{{ warningReasons .events }}  → [Failed BackOff]
func WarningReasons(events interface{}) []string {
	var out []string
	for _, e := range warnings(events) {
		if r := nestedString(e, "reason"); r != "" {
			out = append(out, r)
		}
	}
	return out
}

// warnings returns the Warning events in v.
func warnings(v interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, e := range objectList(v) {
		if nestedString(e, "type") == "Warning" {
			out = append(out, e)
		}
	}
	return out
}

// eventMessage reads the message: "message" in core/v1, "note" in
// events.k8s.io/v1.
func eventMessage(e map[string]interface{}) string {
	if m := nestedString(e, "message"); m != "" {
		return m
	}
	return nestedString(e, "note")
}
