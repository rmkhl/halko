package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func f32(v float32) *float32 { return &v }
func u8(v uint8) *uint8      { return &v }

// steamDelta builds a steam setting that asks for delta control.
func steamDelta() *PowerPidSettings {
	return &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(10)}
}

// heatingStep builds a valid heating step whose steam setting the caller picks.
func heatingStep(steam *PowerPidSettings) ProgramStep {
	return ProgramStep{
		Name:              "heat",
		StepType:          StepTypeHeating,
		TargetTemperature: 100,
		Heater:            &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(10)},
		Fan:               &PowerPidSettings{Power: u8(100)},
		Steam:             steam,
	}
}

func TestProgramValidation(t *testing.T) {
	// Load defaults from the template config
	// LoadConfig already validates the configuration structure
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "Delta Program",
			filename: "example-program-delta.json",
		},
		{
			name:     "PID Program",
			filename: "example-program-pid.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Updated path to go up one level to reach the example directory
			examplePath := filepath.Join("..", "example", tt.filename)
			data, err := os.ReadFile(examplePath)
			if err != nil {
				t.Fatalf("Failed to read example file %s: %v", tt.filename, err)
			}

			var program Program
			err = json.Unmarshal(data, &program)
			if err != nil {
				t.Fatalf("Failed to unmarshal program from %s: %v", tt.filename, err)
			}

			// Apply defaults before validation
			program.ApplyDefaults(config.ControlUnitConfig.Defaults)

			err = program.Validate()
			if err != nil {
				t.Errorf("Program validation failed for %s: %v", tt.filename, err)
			}
		})
	}
}

// Steam is a self-limiting heat source: it cannot drive the kiln past ~100°C.
// Below that it is the dominant contributor to kiln temperature, so a heating
// step may modulate it against the kiln/material delta. Acclimate and cooling
// keep steam on constant power, where it serves its moisture role.
func TestSteamDeltaControlAcceptedOnlyInHeatingSteps(t *testing.T) {
	oneHour := StepDuration{time.Hour}

	tests := []struct {
		name    string
		step    ProgramStep
		wantErr bool
	}{
		{
			name:    "heating accepts delta steam",
			step:    heatingStep(steamDelta()),
			wantErr: false,
		},
		{
			name: "acclimate rejects delta steam",
			step: ProgramStep{
				Name:              "hold",
				StepType:          StepTypeAcclimate,
				TargetTemperature: 100,
				Runtime:           &oneHour,
				Heater:            &PowerPidSettings{Pid: &PidSettings{Kp: 1}},
				Fan:               &PowerPidSettings{Power: u8(100)},
				Steam:             steamDelta(),
			},
			wantErr: true,
		},
		{
			name: "cooling rejects delta steam",
			step: ProgramStep{
				Name:              "cool",
				StepType:          StepTypeCooling,
				TargetTemperature: 40,
				Heater:            &PowerPidSettings{Power: u8(0)},
				Fan:               &PowerPidSettings{Power: u8(100)},
				Steam:             steamDelta(),
			},
			wantErr: true,
		},
		{
			name:    "heating still accepts simple steam",
			step:    heatingStep(&PowerPidSettings{Power: u8(100)}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.step.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation to fail, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected validation to pass, got %v", err)
			}
		})
	}
}

// ApplyDefaults fills in a zero power for steam when a step says nothing about
// it. It must not do that when the step asked for delta control, or the step
// ends up with two control methods and fails validation.
func TestApplyDefaultsLeavesSteamDeltaControlIntact(t *testing.T) {
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	program := Program{
		ProgramName:  "steam delta",
		ProgramSteps: []ProgramStep{heatingStep(steamDelta())},
	}

	program.ApplyDefaults(config.ControlUnitConfig.Defaults)

	steam := program.ProgramSteps[0].Steam
	if steam.Power != nil {
		t.Errorf("ApplyDefaults set steam power to %d, clobbering delta control", *steam.Power)
	}
	if steam.MinDelta == nil || steam.MaxDelta == nil {
		t.Fatal("ApplyDefaults dropped the steam deltas")
	}

	// Validate the step rather than the program: whole-program validation also
	// enforces step count and ordering, which this test says nothing about.
	if err := program.ProgramSteps[0].Validate(); err != nil {
		t.Errorf("step validation failed after applying defaults: %v", err)
	}
}

func steamHeatingStep(target uint8, steam *PowerPidSettings) ProgramStep {
	return ProgramStep{
		Name:              "heat",
		StepType:          StepTypeHeating,
		TargetTemperature: target,
		Steam:             steam,
	}
}

func steamCoolingStep(target uint8, steam *PowerPidSettings) ProgramStep {
	return ProgramStep{
		Name:              "cool",
		StepType:          StepTypeCooling,
		TargetTemperature: target,
		Heater:            &PowerPidSettings{Power: u8(0)},
		Steam:             steam,
	}
}

func steamAcclimateStep(target uint8, steam *PowerPidSettings) ProgramStep {
	return ProgramStep{
		Name:              "hold",
		StepType:          StepTypeAcclimate,
		TargetTemperature: target,
		Runtime:           &StepDuration{time.Hour},
		Steam:             steam,
	}
}

// Only delta control bounds the kiln/material delta during heating: simple runs
// the heater open-loop and PID regulates the kiln to target while ignoring the
// material entirely, so either lets the kiln run arbitrarily far ahead.
func TestHeatingStepRequiresDeltaControlledHeater(t *testing.T) {
	tests := []struct {
		name    string
		heater  *PowerPidSettings
		wantErr bool
	}{
		{"delta accepted", &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(10)}, false},
		{"simple rejected", &PowerPidSettings{Power: u8(100)}, true},
		{"pid rejected", &PowerPidSettings{Pid: &PidSettings{Kp: 1}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := heatingStep(&PowerPidSettings{Power: u8(0)})
			step.Heater = tt.heater

			err := step.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation to fail, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected validation to pass, got %v", err)
			}
		})
	}
}

// Steam cannot heat the kiln past its ceiling, but below it steam outruns the
// heater and opens a kiln/material delta the heater cannot close by backing
// off. Constant steam is therefore allowed only once the kiln is at or above
// the ceiling for the whole step. A heating step hands over at
// target+heater.min_delta (the floor of its delta band); an acclimate settles
// the kiln onto the material, so it hands over at target.
func TestSteamMayOnlyRunOpenLoopAboveTheKilnCeiling(t *testing.T) {
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	steamOff := func() *PowerPidSettings { return &PowerPidSettings{Power: u8(0)} }
	steamConstant := func() *PowerPidSettings { return &PowerPidSettings{Power: u8(50)} }

	tests := []struct {
		name    string
		steps   []ProgramStep
		wantErr bool
	}{
		{
			name: "first heating step cannot run steam open-loop",
			steps: []ProgramStep{
				steamHeatingStep(100, steamConstant()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: true,
		},
		{
			name: "first heating step may delta-control steam",
			steps: []ProgramStep{
				steamHeatingStep(100, steamDelta()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: false,
		},
		{
			name: "first heating step may switch steam off",
			steps: []ProgramStep{
				steamHeatingStep(100, steamOff()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: false,
		},
		{
			name: "constant steam rejected when the kiln enters below the ceiling",
			steps: []ProgramStep{
				// hands over at 60 + 5 = 65, well under the ceiling
				steamHeatingStep(60, steamOff()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: true,
		},
		{
			name: "handover uses the floor of the delta band, not its top",
			steps: []ProgramStep{
				// 92 + 5 = 97 is below the ceiling even though 92 + 10 is not
				steamHeatingStep(92, steamOff()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: true,
		},
		{
			name: "constant steam accepted once the kiln clears the ceiling",
			steps: []ProgramStep{
				// 95 + 5 = 100 reaches the ceiling exactly
				steamHeatingStep(95, steamOff()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: false,
		},
		{
			name: "acclimate below the ceiling must switch steam off",
			steps: []ProgramStep{
				steamHeatingStep(90, steamDelta()),
				steamAcclimateStep(90, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: true,
		},
		{
			name: "acclimate above the ceiling may hold steam constant",
			steps: []ProgramStep{
				steamHeatingStep(160, steamDelta()),
				steamAcclimateStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: false,
		},
		{
			name: "cooling must switch steam off",
			steps: []ProgramStep{
				steamHeatingStep(160, steamDelta()),
				steamAcclimateStep(160, steamConstant()),
				steamCoolingStep(30, steamConstant()),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := Program{ProgramName: "steam rules", ProgramSteps: tt.steps}
			program.ApplyDefaults(config.ControlUnitConfig.Defaults)

			err := program.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation to fail, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected validation to pass, got %v", err)
			}
		})
	}
}

func TestProgramValidationWithCopy(t *testing.T) {
	// Load defaults from the template config
	// LoadConfig already validates the configuration structure
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// Test the validation pattern used in startNewProgram, which validates a
	// copy so the stored program is left as the author wrote it. Storing a
	// program does not validate at all - only starting one does.
	examplePath := filepath.Join("..", "example", "example-program-delta.json")
	data, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example file: %v", err)
	}

	var originalProgram Program
	err = json.Unmarshal(data, &originalProgram)
	if err != nil {
		t.Fatalf("Failed to unmarshal program: %v", err)
	}

	// Store the original state to verify it's not modified
	originalJSON, err := json.Marshal(originalProgram)
	if err != nil {
		t.Fatalf("Failed to marshal original program: %v", err)
	}

	// Create a deep copy for validation (simulating the router behavior)
	programCopy, err := originalProgram.Duplicate()
	if err != nil {
		t.Fatalf("Failed to duplicate program: %v", err)
	}

	// Apply defaults to the copy and validate
	programCopy.ApplyDefaults(config.ControlUnitConfig.Defaults)
	err = programCopy.Validate()
	if err != nil {
		t.Errorf("Program validation failed on copy: %v", err)
	}

	// Verify the original program was not modified
	currentJSON, err := json.Marshal(originalProgram)
	if err != nil {
		t.Fatalf("Failed to marshal current program: %v", err)
	}
	if string(originalJSON) != string(currentJSON) {
		t.Error("Original program was unexpectedly modified during validation")
	}

	// Verify the copy has defaults applied
	if !programCopy.DefaultsApplied {
		t.Error("Copy should have defaults applied")
	}

	// Verify the original does not have defaults applied
	if originalProgram.DefaultsApplied {
		t.Error("Original program should not have defaults applied")
	}
}
