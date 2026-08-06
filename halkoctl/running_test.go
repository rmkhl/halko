package main

import "testing"

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 45, "45s"},
		{"exactly a minute", 60, "1m 0s"},
		{"minutes and seconds", 125, "2m 5s"},
		{"exactly an hour", 3600, "1h 0m 0s"},
		{"hours, minutes and seconds", 7385, "2h 3m 5s"},
		{"a long run", 86400, "24h 0m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.seconds); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
