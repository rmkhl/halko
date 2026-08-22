package main

import (
	"strings"
	"testing"

	"github.com/rmkhl/halko/types"
)

func describedProgram() *types.Program {
	return &types.Program{
		ProgramName: "birch",
		ProgramSteps: []types.ProgramStep{
			{Name: "warm up", StepType: types.StepTypeHeating, TargetTemperature: 100},
		},
	}
}

// The description is the only part of a program that says why it looks the way
// it does, so `validate --verbose` has to show it alongside the structure.
func TestDescribeProgramShowsTheDescription(t *testing.T) {
	program := describedProgram()
	program.Description = "Birch 50mm\nGentle schedule, run it after the pine"

	out := describeProgram(program)

	for _, want := range []string{"birch", "Birch 50mm", "Gentle schedule", "warm up", "heating", "100"} {
		if !strings.Contains(out, want) {
			t.Errorf("description does not mention %q:\n%s", want, out)
		}
	}
}

func TestDescribeProgramOmitsAnAbsentDescription(t *testing.T) {
	out := describeProgram(describedProgram())

	if strings.Contains(out, "Description") {
		t.Errorf("description of a program without one still has a Description line:\n%s", out)
	}
}
