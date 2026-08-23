package engine

import (
	"errors"
	"testing"
)

// Both note calls are reachable from HTTP at any time, including when no run
// has ever been started. They must answer, not panic on a nil runner.
func TestAddRunNoteWithNothingRunning(t *testing.T) {
	engine := &ControlEngine{}

	_, err := engine.AddRunNote("anybody there")

	if !errors.Is(err, ErrNoProgramRunning) {
		t.Errorf("AddRunNote with no run = %v, want ErrNoProgramRunning", err)
	}
}

func TestRunNotesWithNothingRunning(t *testing.T) {
	engine := &ControlEngine{}

	_, err := engine.RunNotes()

	if !errors.Is(err, ErrNoProgramRunning) {
		t.Errorf("RunNotes with no run = %v, want ErrNoProgramRunning", err)
	}
}
