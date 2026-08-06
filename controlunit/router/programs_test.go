package router

import (
	"testing"
	"time"
)

// Run names carry their start time as "<program name>@<RFC3339>", which is what
// orders the history list.
func TestStartTimeFromName(t *testing.T) {
	started := time.Date(2026, 8, 5, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		run  string
		want int64
	}{
		{"name and timestamp", "Oak drying@" + started.Format(time.RFC3339), started.Unix()},
		{"no timestamp", "Oak drying", 0},
		{"unparsable timestamp", "Oak drying@yesterday", 0},
		{"empty timestamp", "Oak drying@", 0},
		{"more than one separator", "Oak@drying@" + started.Format(time.RFC3339), 0},
		{"empty name", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := startTimeFromName(tt.run); got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestStartTimeFromNameKeepsTheZone(t *testing.T) {
	// A timestamp with an offset resolves to the same instant as its UTC form.
	withOffset := "Oak drying@2026-08-05T17:30:00+03:00"
	utc := "Oak drying@2026-08-05T14:30:00Z"

	if got, want := startTimeFromName(withOffset), startTimeFromName(utc); got != want {
		t.Fatalf("expected both forms to give the same instant, got %d and %d", got, want)
	}
}
