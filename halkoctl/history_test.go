package main

import (
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

func TestFormatDurationLong(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0s"},
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes and seconds", 125 * time.Second, "2m 5s"},
		{"hours, minutes and seconds", 7385 * time.Second, "2h 3m 5s"},
		{"sub-second rounds down", 900 * time.Millisecond, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDurationLong(tt.d); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFormatPowerControl(t *testing.T) {
	power := uint8(60)
	minDelta, maxDelta := float32(2.5), float32(8.0)

	tests := []struct {
		name     string
		settings *types.PowerPidSettings
		want     string
	}{
		{"simple", &types.PowerPidSettings{Power: &power}, "Simple (60%)"},
		{
			"delta",
			&types.PowerPidSettings{MinDelta: &minDelta, MaxDelta: &maxDelta},
			"Delta (min: 2.5°C, max: 8.0°C)",
		},
		{"nothing set", &types.PowerPidSettings{}, "Not specified"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPowerControl(tt.settings); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
