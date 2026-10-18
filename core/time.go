package core

import (
	"fmt"
	"math"
	"text/template"
	"time"

	"github.com/robfig/cron/v3"
)

// timeFuncs — Time
//
// Domain: time
//
// Reason about timestamps (RFC3339) and durations (Go units plus d, w, mo and
// y): elapsed and remaining time, formatting, time windows and weekdays,
// duration arithmetic, and the next run of a cron schedule.
func timeFuncs() template.FuncMap {
	return template.FuncMap{
		"timeAgo":         TimeAgo,
		"timeSince":       TimeSince,
		"timeUntil":       TimeUntil,
		"isExpired":       IsExpired,
		"timeFormat":      TimeFormat,
		"durationSeconds": DurationSeconds,
		"durationAdd":     DurationAdd,
		"durationValid":   DurationValid,
		"weekday":         Weekday,
		"weekend":         Weekend,
		"timeInWindow":    TimeInWindow,
		"timeNotInWindow": TimeNotInWindow,
		"nextCron":        NextCron,
	}
}

// parseTime attempts to parse a timestamp string in RFC3339 or common formats.
func parseTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// TimeAgo returns a human-readable elapsed-time string from a timestamp.
// Outputs seconds, minutes, hours, or days depending on magnitude.
//
// Keywords: time, ago, elapsed, human, readable, display, status, string
//
// Example:
//
//	{{ timeAgo .deployment.metadata.creationTimestamp }}  → "3d ago"
func TimeAgo(s string) string {
	t, ok := parseTime(fmt.Sprint(s))
	if !ok {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// TimeSince returns the number of seconds elapsed since a timestamp as an
// integer, for comparing against a threshold.
//
// Keywords: time, since, seconds, elapsed, duration, int, compare
//
// Example:
//
//	{{ timeSince .deployment.metadata.creationTimestamp }}  → 259200
func TimeSince(s string) int64 {
	t, ok := parseTime(fmt.Sprint(s))
	if !ok {
		return 0
	}
	return int64(math.Round(time.Since(t).Seconds()))
}

// TimeUntil returns a Go duration string representing the time remaining until
// a timestamp. Returns `"0s"` when the timestamp is in the past or
// unparseable. Useful as a requeue delay, so a controller wakes up exactly
// when the timestamp passes.
//
// Keywords: time, until, remaining, duration, expiry, cert, requeue, schedule
//
// Example:
//
//	{{ timeUntil .certificate.status.notAfter }}  → "719h43m12s"
func TimeUntil(s string) string {
	t, ok := parseTime(fmt.Sprint(s))
	if !ok {
		return "0s"
	}
	d := time.Until(t)
	if d <= 0 {
		return "0s"
	}
	return d.Truncate(time.Second).String()
}

// IsExpired returns `true` when a timestamp plus a duration is in the past.
// Use it to decide when something is due for rotation or cleanup. Duration is
// an extended duration string — Go's units (`"30m"`, `"24h"`) plus
// `d`/`w`/`mo`/`y` (`"30d"`, `"1y"`).
//
// Keywords: time, expired, expiry, rotation, ttl, boolean, when, condition
//
// Example:
//
//	{{ isExpired "2026-01-15T10:30:00Z" "30d" }}  → true
func IsExpired(timestamp interface{}, duration string) bool {
	t, ok := parseTime(fmt.Sprint(timestamp))
	if !ok {
		return false
	}
	d, err := ParseTimeDuration(duration)
	if err != nil {
		return false
	}
	return time.Now().After(t.Add(d))
}

// TimeFormat reformats a timestamp string using Go's time format layout.
// Returns `""` for unparseable input.
//
// Keywords: time, format, layout, display, readable, string, date
//
// Example:
//
//	{{ timeFormat "2026-04-13T09:30:00Z" "Jan 2, 2006" }}  → "Apr 13, 2026"
func TimeFormat(timestamp interface{}, layout string) string {
	t, ok := parseTime(fmt.Sprint(timestamp))
	if !ok {
		return ""
	}
	return t.UTC().Format(layout)
}

// DurationSeconds parses an extended duration string (Go units plus d/w/mo/y)
// and returns the total seconds. 0 for an empty string; an error for one that
// does not parse.
//
// Keywords: time, duration, seconds, parse, int, convert
//
// Example:
//
//	{{ durationSeconds "5m" }}    → 300
//	{{ durationSeconds "1h30m" }} → 5400
//	{{ durationSeconds "7d" }}    → 604800
func DurationSeconds(d string) (int64, error) {
	if d == "" {
		return 0, nil
	}
	dur, err := ParseTimeDuration(d)
	if err != nil {
		return 0, err
	}
	return int64(dur.Seconds()), nil
}

// DurationAdd adds two extended duration strings (Go units plus d/w/mo/y) and
// returns the result as a canonical Go duration string. An error when either
// does not parse.
//
// Keywords: time, duration, add, combine, string
//
// Example:
//
//	{{ durationAdd "5m" "30s" }}  → "5m30s"
//	{{ durationAdd "1h" "90m" }}  → "2h30m0s"
//	{{ durationAdd "1d" "12h" }}  → "36h0m0s"
func DurationAdd(a, b string) (string, error) {
	da, err := ParseTimeDuration(a)
	if err != nil {
		return "", err
	}
	db, err := ParseTimeDuration(b)
	if err != nil {
		return "", err
	}
	return (da + db).String(), nil
}

// DurationValid returns `true` when the string is a valid extended duration
// (Go units plus `d`/`w`/`mo`/`y`). Use in validation rules to reject
// malformed duration fields.
//
// Keywords: time, duration, valid, validate, boolean, check
//
// Example:
//
//	{{ durationValid "30d" }}    → true
//	{{ durationValid "thirty" }}  → false
func DurationValid(s string) bool {
	if s == "" {
		return false
	}
	_, err := ParseTimeDuration(s)
	return err == nil
}

// Weekday returns `true` when the current UTC day is Monday through Friday.
// Use it to act only on business days.
//
// Keywords: time, weekday, businessday, monday, friday, boolean, window
//
// Example:
//
//	{{ if weekday }}business day{{ end }}
func Weekday() bool {
	wd := time.Now().UTC().Weekday()
	return wd >= time.Monday && wd <= time.Friday
}

// Weekend returns `true` when the current UTC day is Saturday or Sunday. The
// exact complement of `weekday`.
//
// Keywords: time, weekend, saturday, sunday, boolean, offpeak
//
// Example:
//
//	{{ if weekend }}scaled down for the weekend{{ end }}
func Weekend() bool {
	wd := time.Now().UTC().Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

// TimeInWindow returns `true` when the current UTC time falls within the
// window `[after, before)`. Both arguments must be `"HH:MM"` strings. Returns
// `false` for malformed input.
//
// Keywords: time, window, businesshours, after, before, boolean, schedule
//
// Example:
//
//	{{ timeInWindow "09:00" "18:00" }}
//	{{ and (timeInWindow "09:00" "18:00") weekday }}
func TimeInWindow(after, before string) bool {
	now := time.Now().UTC()
	a, err := parseHHMMNote(after, now)
	if err != nil {
		return false
	}
	b, err := parseHHMMNote(before, now)
	if err != nil {
		return false
	}
	return !now.Before(a) && now.Before(b)
}

// TimeNotInWindow returns `true` when the current UTC time is outside the
// window `[after, before)`. The exact complement of `timeInWindow`. Useful for
// maintenance-window logic where the resource should exist only when the
// window is closed.
//
// Keywords: time, window, maintenance, outside, boolean, complement
//
// Example:
//
//	{{ timeNotInWindow "02:00" "04:00" }}
func TimeNotInWindow(after, before string) bool {
	return !TimeInWindow(after, before)
}

// NextCron returns the next scheduled fire time for a standard 5-field cron
// expression as an RFC3339 string. An error for an expression that does not
// parse.
//
// Keywords: time, cron, next, schedule, rfc3339, string, transition
//
// Example:
//
//	{{ nextCron "0 9 * * 1" }}  → "2026-07-14T09:00:00Z"  (next Monday 09:00 UTC)
//	{{ nextCron "0 2 * * 0" }}  → next Sunday 02:00 UTC
func NextCron(expr string) (string, error) {
	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		return "", fmt.Errorf("cron expression %q: %w", expr, err)
	}
	return schedule.Next(time.Now().UTC()).UTC().Format(time.RFC3339), nil
}

// parseHHMMNote parses a "HH:MM" string anchored to the date of now (UTC).
func parseHHMMNote(s string, now time.Time) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time %q: expected HH:MM", s)
	}
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC), nil
}
