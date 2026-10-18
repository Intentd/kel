// pkg/note/time_test.go
package core

import (
	"testing"
	"time"
)

func TestTimeUntil(t *testing.T) {
	future := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	past := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	if got := TimeUntil(future); got == "0s" {
		t.Errorf("timeUntil(future) = %q, want non-zero duration", got)
	}
	if got := TimeUntil(past); got != "0s" {
		t.Errorf("timeUntil(past) = %q, want \"0s\"", got)
	}
	if got := TimeUntil(""); got != "0s" {
		t.Errorf("timeUntil(\"\") = %q, want \"0s\"", got)
	}
	if got := TimeUntil("not-a-time"); got != "0s" {
		t.Errorf("timeUntil(invalid) = %q, want \"0s\"", got)
	}
}

func TestWeekdayWeekend(t *testing.T) {
	// These are time-dependent; just assert they are mutually exclusive.
	w := Weekday()
	we := Weekend()
	if w == we {
		t.Errorf("weekday=%v and weekend=%v must be mutually exclusive", w, we)
	}
}

func TestTimeInWindow(t *testing.T) {
	tests := []struct {
		after, before string
		wantErr       bool // malformed → false, no panic
	}{
		{"09:00", "18:00", false},
		{"00:00", "23:59", false},
		{"bad", "18:00", false},
		{"09:00", "bad", false},
		{"", "", false},
	}
	for _, tc := range tests {
		got := TimeInWindow(tc.after, tc.before)
		inv := TimeNotInWindow(tc.after, tc.before)
		if !tc.wantErr && got == inv && tc.after != "" && tc.before != "" {
			// For valid inputs the two must be complements — but only when
			// we have non-empty inputs (empty inputs both return false).
			t.Errorf("timeInWindow(%q,%q)=%v and timeNotInWindow=%v must differ",
				tc.after, tc.before, got, inv)
		}
	}
}

func TestTimeInWindowComplement(t *testing.T) {
	// A full-day window must always be true; its complement always false.
	if !TimeInWindow("00:00", "23:59") {
		t.Error("timeInWindow(00:00, 23:59) should always be true")
	}
	if TimeNotInWindow("00:00", "23:59") {
		t.Error("timeNotInWindow(00:00, 23:59) should always be false")
	}
}

func TestNextCron(t *testing.T) {
	got, err := NextCron("0 9 * * 1") // every Monday 09:00
	if err != nil {
		t.Fatalf("nextCron: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, got); err != nil {
		t.Errorf("nextCron result %q is not RFC3339: %v", got, err)
	}
}

func TestNextCronInvalid(t *testing.T) {
	if _, err := NextCron("not a cron"); err == nil {
		t.Error("expected an error for an invalid cron expression")
	}
	if _, err := NextCron(""); err == nil {
		t.Error("expected an error for an empty cron expression")
	}
}

func TestNextCronIsFuture(t *testing.T) {
	got, err := NextCron("* * * * *") // every minute
	if err != nil {
		t.Fatalf("nextCron: %v", err)
	}
	ts, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !ts.After(time.Now().UTC()) {
		t.Errorf("nextCron result %s should be in the future", got)
	}
}
