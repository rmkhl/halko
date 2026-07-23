package router

import (
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestKilnSelectorSticksWithinHysteresis(t *testing.T) {
	var s kilnSelector

	// First poll: primary is higher, so primary is selected.
	if got := s.Select(20.0, 19.8); got != 20.0 {
		t.Fatalf("expected primary value 20.0, got %v", got)
	}

	// Secondary edges higher, but within the hysteresis margin: the
	// selection must not flip, so the primary value is still reported.
	if got := s.Select(20.0, 20.3); got != 20.0 {
		t.Fatalf("expected to stick with primary value 20.0, got %v", got)
	}
}

func TestKilnSelectorSwitchesBeyondHysteresis(t *testing.T) {
	var s kilnSelector

	s.Select(20.0, 19.8) // selects primary

	// Secondary exceeds primary by more than the margin: switch.
	if got := s.Select(20.0, 21.0); got != 21.0 {
		t.Fatalf("expected switch to secondary value 21.0, got %v", got)
	}

	// Selection is sticky in the other direction too: primary slightly
	// higher again, but within the margin, so secondary is still reported.
	if got := s.Select(20.4, 20.2); got != 20.2 {
		t.Fatalf("expected to stick with secondary value 20.2, got %v", got)
	}
}

func TestKilnSelectorSingleValidSensor(t *testing.T) {
	var s kilnSelector

	if got := s.Select(types.InvalidTemperatureReading, 19.5); got != 19.5 {
		t.Fatalf("expected secondary value 19.5, got %v", got)
	}

	// Primary comes back but within margin of the now-selected secondary:
	// stay on secondary.
	if got := s.Select(19.6, 19.5); got != 19.5 {
		t.Fatalf("expected to stick with secondary value 19.5, got %v", got)
	}

	if got := s.Select(19.5, types.InvalidTemperatureReading); got != 19.5 {
		t.Fatalf("expected primary value 19.5, got %v", got)
	}
}

func TestKilnSelectorBothInvalid(t *testing.T) {
	var s kilnSelector

	if got := s.Select(types.InvalidTemperatureReading, types.InvalidTemperatureReading); got != types.InvalidTemperatureReading {
		t.Fatalf("expected invalid reading sentinel, got %v", got)
	}
}
