package core

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

// cronFuncs — Cron
//
// Domain: cron
//
// Work with five-field cron expressions (minute hour day-of-month month day-
// of-week): read single fields, build and normalise expressions, convert
// between the string form and a map with one key per field, describe a
// schedule in words, and check that an expression is valid.
func cronFuncs() template.FuncMap {
	return template.FuncMap{
		"cronMinute":    CronMinute,
		"cronHour":      CronHour,
		"cronDom":       CronDom,
		"cronMonth":     CronMonth,
		"cronDow":       CronDow,
		"cronField":     CronField,
		"cronExpr":      CronExpr,
		"cronValid":     CronValid,
		"cronFromMap":   CronFromMapStrict,
		"cronToMap":     CronToMap,
		"cronFromAny":   CronFromMapAny,
		"cronNormalize": CronNormalize,
		"cronDescribe":  CronDescribe,
	}
}

// CronMinute returns the minute field of a cron expression; "*" for an
// empty expression, and an error when it does not have five fields.
//
// Keywords: cron, minute
//
// Example:
//
//	{{ cronMinute "*/5 2 * * 1" }}  → "*/5"
func CronMinute(expr string) (string, error) { return CronField(expr, 0) }

// CronHour returns the hour field of a cron expression; "*" for an
// empty expression, and an error when it does not have five fields.
//
// Keywords: cron, hour
//
// Example:
//
//	{{ cronHour "*/5 2 * * 1" }}  → "2"
func CronHour(expr string) (string, error) { return CronField(expr, 1) }

// CronDom returns the day-of-month field of a cron expression; "*" for an
// empty expression, and an error when it does not have five fields.
//
// Keywords: cron, dom
//
// Example:
//
//	{{ cronDom "0 0 15 * *" }}  → "15"
func CronDom(expr string) (string, error) { return CronField(expr, 2) }

// CronMonth returns the month field of a cron expression; "*" for an
// empty expression, and an error when it does not have five fields.
//
// Keywords: cron, month
//
// Example:
//
//	{{ cronMonth "0 0 1 6 *" }}  → "6"
func CronMonth(expr string) (string, error) { return CronField(expr, 3) }

// CronDow returns the day-of-week field of a cron expression; "*" for an
// empty expression, and an error when it does not have five fields.
//
// Keywords: cron, dow
//
// Example:
//
//	{{ cronDow "0 0 * * 1" }}  → "1"
func CronDow(expr string) (string, error) { return CronField(expr, 4) }

// CronField extracts a field by index (0–4). The general form of the five
// named extractors above.
//
// Keywords: cron, field
//
// Example:
//
//	{{ cronField "*/5 2 * * 1" 1 }}  → "2"
func CronField(expr string, pos int) (string, error) {
	if strings.TrimSpace(expr) == "" {
		return "*", nil
	}
	if pos < 0 || pos > 4 {
		return "*", fmt.Errorf("cronField: position %d out of range [0-4]", pos)
	}
	expanded := ExpandCronMacro(expr)
	parts := strings.Fields(expanded)
	if len(parts) != 5 {
		return "*", fmt.Errorf(
			"cronField: expected 5 fields in %q, got %d",
			expr, len(parts),
		)
	}
	return parts[pos], nil
}

// CronExpr builds a five-field cron expression from five explicit string
// parts. Empty parts default to `*`.
//
// Keywords: cron, schedule, build, expression, fields, construct
//
// Example:
//
//	{{ cronExpr "*/15" "" "" "" "1-5" }}  → "*/15 * * * 1-5"
func CronExpr(minute, hour, dom, month, dow string) string {
	return fmt.Sprintf("%s %s %s %s %s",
		starIfEmpty(minute), starIfEmpty(hour),
		starIfEmpty(dom), starIfEmpty(month), starIfEmpty(dow),
	)
}

func starIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "*"
	}
	return s
}

// CronFromMapStrict converts a schedule **map** to a five-field cron string.
// Reads keys `minute`, `hour`, `dayOfMonth`, `month`, `dayOfWeek`; absent keys
// default to `*`. Errors if the input is not a map — use `cronFromAny` when
// the input may be a string.
//
// Keywords: cron, schedule, map, convert, fields, expression
//
// Example:
//
//	{{ cronFromMap .spec.schedule }}  → "*/5 0 * * 1"
func CronFromMapStrict(v interface{}) (string, error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cronFromMap: expected map, got %T — use cronFromAny if the input may be a cron string", v)
	}
	return cronFromMap(m), nil
}

// CronFromMapAny converts a schedule value to a five-field cron string.
// Accepts either shape:
//
// Keywords: cron, schedule, convert, string, normalize, any
//
// Example:
//
//	{{ cronFromAny .spec.schedule }}  → "*/5 0 * * 1"
//	{{ cronFromAny "0 2 * * *" }}    → "0 2 * * *"
func CronFromMapAny(v interface{}) string {
	switch m := v.(type) {
	case map[string]interface{}:
		return cronFromMap(m)
	case string:
		return CronNormalize(m)
	default:
		return "* * * * *"
	}
}

// CronToMap converts a cron string to the structured map shape. For use in
// **conversion path specs only** — the result is a nested map, not a plain
// string.
//
// Keywords: cron, schedule, parse, fields, map, convert, structured
//
// Example:
//
//	{{ (cronToMap "*/5 0 * * 1").minute }}  → "*/5"
func CronToMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	expr := fmt.Sprintf("%v", v)
	out := map[string]interface{}{
		"minute": "*", "hour": "*", "dayOfMonth": "*", "month": "*", "dayOfWeek": "*",
	}
	if strings.TrimSpace(expr) == "" {
		return out
	}
	expanded := ExpandCronMacro(expr)
	parts := strings.Fields(expanded)
	if len(parts) != 5 {
		return out
	}
	out["minute"] = parts[0]
	out["hour"] = parts[1]
	out["dayOfMonth"] = parts[2]
	out["month"] = parts[3]
	out["dayOfWeek"] = parts[4]
	return out
}

// cronFromMap is kept as the unexported Go helper used internally.
func cronFromMap(m map[string]interface{}) string {
	get := func(key string) string {
		if v, ok := m[key]; ok {
			s := fmt.Sprintf("%v", v)
			if strings.TrimSpace(s) != "" {
				return s
			}
		}
		return "*"
	}
	return CronExpr(
		get("minute"),
		get("hour"),
		get("dayOfMonth"),
		get("month"),
		get("dayOfWeek"),
	)
}

// CronNormalize normalizes a cron string: expand `@`-macros, trim whitespace,
// ensure exactly five fields. Returns `"* * * * *"` for empty or malformed
// input.
//
// Keywords: cron, schedule, normalize, macro, expand, sanitize
//
// Example:
//
//	{{ cronNormalize "@daily" }}  → "0 0 * * *"
func CronNormalize(expr string) string {
	if strings.TrimSpace(expr) == "" {
		return "* * * * *"
	}

	expanded := ExpandCronMacro(expr)
	parts := strings.Fields(expanded)

	// If invalid, normalize to all "*"
	if len(parts) != 5 {
		return "* * * * *"
	}

	// Trim each field
	for i := range parts {
		if strings.TrimSpace(parts[i]) == "" {
			parts[i] = "*"
		}
	}

	return strings.Join(parts, " ")
}

// CronDescribe describes a cron expression in words, for display: "Every 5
// minutes", "At 02:00 on Monday through Friday". Expressions it cannot put
// into words come back as "Cron: <expression>"; "Invalid cron expression"
// when expr is not a five-field expression or a macro.
//
// Keywords: cron, schedule, describe, human-readable, display, status
//
// Example:
//
//	{{ cronDescribe "0 2 * * 1-5" }}  → "At 02:00 on Monday through Friday"
//	{{ cronDescribe "*/15 * * * *" }}  → "Every 15 minutes"
//
// See also: cronValid, nextCron
func CronDescribe(expr string) string {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "@") && len(strings.Fields(trimmed)) != 5) {
		return "Invalid cron expression"
	}
	parts := strings.Fields(CronNormalize(trimmed))
	if len(parts) != 5 {
		return "Invalid cron expression"
	}
	min, hr, dom, mon, dow := parts[0], parts[1], parts[2], parts[3], parts[4]
	fallback := "Cron: " + strings.Join(parts, " ")
	anyDay := dom == "*" && mon == "*" && dow == "*"

	switch {
	case anyDay && min == "*" && hr == "*":
		return "Every minute"
	case anyDay && hr == "*" && strings.HasPrefix(min, "*/"):
		return "Every " + strings.TrimPrefix(min, "*/") + " minutes"
	case anyDay && hr == "*" && isCronNumber(min):
		if min == "0" {
			return "Every hour"
		}
		return "At minute " + min + " of every hour"
	case anyDay && strings.HasPrefix(hr, "*/") && isCronNumber(min):
		every := "Every " + strings.TrimPrefix(hr, "*/") + " hours"
		if min == "0" {
			return every
		}
		return "At minute " + min + ", " + strings.ToLower(every[:1]) + every[1:]
	}

	if !isCronNumber(min) || !isCronNumber(hr) {
		return fallback
	}
	at := fmt.Sprintf("At %02s:%02s", hr, min)

	var when string
	switch {
	case dom == "*" && dow == "*":
		when = "every day"
	case dom == "*":
		days, ok := describeCronList(dow, weekdayName)
		if !ok {
			return fallback
		}
		when = "on " + days
	case dow == "*":
		if !isCronNumber(dom) {
			return fallback
		}
		when = "on day " + dom
		if mon == "*" {
			when += " of every month"
		}
	default:
		return fallback
	}
	if mon != "*" {
		months, ok := describeCronList(mon, monthName)
		if !ok {
			return fallback
		}
		when += " in " + months
	}
	return at + " " + when
}

func isCronNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

var (
	weekdays = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	months   = []string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
)

// weekdayName names a day-of-week field value: 0-7 (0 and 7 are Sunday) or
// MON..SUN.
func weekdayName(v string) (string, bool) {
	if n, err := strconv.Atoi(v); err == nil {
		if n >= 0 && n < len(weekdays) {
			return weekdays[n], true
		}
		return "", false
	}
	for _, d := range weekdays {
		if strings.EqualFold(d[:3], v) {
			return d, true
		}
	}
	return "", false
}

// monthName names a month field value: 1-12 or JAN..DEC.
func monthName(v string) (string, bool) {
	if n, err := strconv.Atoi(v); err == nil {
		if n >= 1 && n < len(months) {
			return months[n], true
		}
		return "", false
	}
	for _, m := range months[1:] {
		if strings.EqualFold(m[:3], v) {
			return m, true
		}
	}
	return "", false
}

// describeCronList names a list of values and ranges ("1-5", "1,3,5"):
// "Monday through Friday", "Monday, Wednesday and Friday".
func describeCronList(field string, name func(string) (string, bool)) (string, bool) {
	var items []string
	for _, item := range strings.Split(field, ",") {
		lo, hi, isRange := strings.Cut(item, "-")
		a, ok := name(lo)
		if !ok {
			return "", false
		}
		if !isRange {
			items = append(items, a)
			continue
		}
		b, ok := name(hi)
		if !ok {
			return "", false
		}
		items = append(items, a+" through "+b)
	}
	if len(items) == 1 {
		return items[0], true
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1], true
}
