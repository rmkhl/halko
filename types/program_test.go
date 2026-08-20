package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func f32(v float32) *float32 { return &v }
func b(v bool) *bool         { return &v }
func u8(v uint8) *uint8      { return &v }

// validProgram builds the smallest program that passes validation: a heating
// step, an acclimate holding the same target, then a cooling step.
func validProgram() Program {
	return Program{ProgramSteps: []ProgramStep{
		steamHeatingStep(150, steamDelta()),
		steamAcclimateStep(150, &PowerPidSettings{Power: u8(0)}),
		steamCoolingStep(30, &PowerPidSettings{Power: u8(0)}),
	}}
}

// templateDefaults loads the shipped defaults, the same source the other
// defaults-driven tests in this file use.
func templateDefaults(t *testing.T) *Defaults {
	t.Helper()
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}
	return config.ControlUnitConfig.Defaults
}

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
				Heater:            &PowerPidSettings{MinDelta: f32(-1), MaxDelta: f32(3)},
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
			err := tt.step.Validate(100)
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
	if err := program.ProgramSteps[0].Validate(100); err != nil {
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

// Only delta control bounds the kiln/material delta during heating; simple runs
// the heater open-loop and lets the kiln run arbitrarily far ahead.
func TestHeatingStepRequiresDeltaControlledHeater(t *testing.T) {
	tests := []struct {
		name    string
		heater  *PowerPidSettings
		wantErr bool
	}{
		{"delta accepted", &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(10)}, false},
		{"simple rejected", &PowerPidSettings{Power: u8(100)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := heatingStep(&PowerPidSettings{Power: u8(0)})
			step.Heater = tt.heater

			err := step.Validate(100)
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
// target+heater.min_delta (the floor of its delta band), and an acclimate holds
// the kiln in a band around its own target, so it hands over at the floor of
// that band.
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
			// An acclimate holds the kiln down to target+min_delta, so a target
			// sitting exactly on the ceiling still hands over below it.
			name: "acclimate at the ceiling hands over below it",
			steps: []ProgramStep{
				steamHeatingStep(100, steamDelta()),
				steamAcclimateStep(100, steamOff()),
				steamHeatingStep(160, steamConstant()),
				steamCoolingStep(30, steamOff()),
			},
			wantErr: true,
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

// acclimateStep builds a complete acclimate step whose heater the caller picks.
func acclimateStep(heater *PowerPidSettings) ProgramStep {
	step := steamAcclimateStep(150, &PowerPidSettings{Power: u8(0)})
	step.Heater = heater
	step.Fan = &PowerPidSettings{Power: u8(100)}
	return step
}

// An acclimate holds the kiln in a band around the step target, so its deltas
// are offsets from the target rather than from the material. min_delta is the
// floor the air may sag to before the heater fires and must be negative; a zero
// one puts the floor on the stop point and chatters. max_delta caps how far
// above target the air may be driven and must be positive; a zero one leaves
// the heater unable to push the air past target to recover the wood.
func TestAcclimateStepRequiresTargetReferencedDeltaBand(t *testing.T) {
	tests := []struct {
		name    string
		heater  *PowerPidSettings
		wantErr bool
	}{
		{"band straddling target accepted", &PowerPidSettings{MinDelta: f32(-1), MaxDelta: f32(3)}, false},
		{"zero min_delta rejected", &PowerPidSettings{MinDelta: f32(0), MaxDelta: f32(3)}, true},
		{"positive min_delta rejected", &PowerPidSettings{MinDelta: f32(1), MaxDelta: f32(3)}, true},
		{"zero max_delta rejected", &PowerPidSettings{MinDelta: f32(-1), MaxDelta: f32(0)}, true},
		{"negative max_delta rejected", &PowerPidSettings{MinDelta: f32(-3), MaxDelta: f32(-1)}, true},
		{"simple rejected", &PowerPidSettings{Power: u8(100)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := acclimateStep(tt.heater)
			err := step.Validate(100)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation to fail, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected validation to pass, got %v", err)
			}
		})
	}
}

// A reversed or collapsed band leaves the controller with a bound it can never
// cross, silently disabling it, so it is rejected wherever deltas are used.
func TestDeltaBandRequiresMinBelowMax(t *testing.T) {
	tests := []struct {
		name    string
		heater  *PowerPidSettings
		wantErr bool
	}{
		{"ordered band accepted", &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(10)}, false},
		{"reversed band rejected", &PowerPidSettings{MinDelta: f32(10), MaxDelta: f32(5)}, true},
		{"collapsed band rejected", &PowerPidSettings{MinDelta: f32(5), MaxDelta: f32(5)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := heatingStep(&PowerPidSettings{Power: u8(0)})
			step.Heater = tt.heater
			err := step.Validate(100)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation to fail, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected validation to pass, got %v", err)
			}
		})
	}
}

// An acclimate with no heater block must come out of ApplyDefaults with a
// target-referenced delta band.
func TestApplyDefaultsGivesAcclimateADeltaBand(t *testing.T) {
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	// No heater and no steam: ApplyDefaults must supply both.
	program := Program{ProgramSteps: []ProgramStep{steamAcclimateStep(150, nil)}}
	program.ApplyDefaults(config.ControlUnitConfig.Defaults)

	heater := program.ProgramSteps[0].Heater
	if heater.MinDelta == nil || heater.MaxDelta == nil {
		t.Fatal("acclimate heater did not default to a delta band")
	}
	if *heater.MinDelta >= 0 {
		t.Fatalf("default min delta = %v, want negative", *heater.MinDelta)
	}
	if *heater.MaxDelta <= 0 {
		t.Fatalf("default max delta = %v, want positive", *heater.MaxDelta)
	}
	if err := program.ProgramSteps[0].Validate(100); err != nil {
		t.Fatalf("defaulted acclimate step failed validation: %v", err)
	}
}

// The ceiling on a step's target is a property of the kiln, so it comes from
// the configured defaults rather than a constant in the validator.
func TestTargetTemperatureCeilingComesFromDefaults(t *testing.T) {
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}
	defaults := config.ControlUnitConfig.Defaults
	lowered := uint8(120)
	defaults.MaxTargetTemperature = &lowered

	program := Program{ProgramSteps: []ProgramStep{
		steamHeatingStep(150, steamDelta()),
		steamAcclimateStep(150, &PowerPidSettings{Power: u8(0)}),
		steamCoolingStep(30, &PowerPidSettings{Power: u8(0)}),
	}}
	program.ApplyDefaults(defaults)

	if err := program.Validate(); err == nil {
		t.Fatal("expected a 150C step to be rejected under a 120C ceiling, got nil")
	}
}

// The temperature steam stops being able to heat the kiln past is a property of
// the installation, so it comes from the configured defaults rather than a
// constant in the validator.
func TestSteamCeilingComesFromDefaults(t *testing.T) {
	config, err := LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}
	defaults := config.ControlUnitConfig.Defaults

	// An acclimate at 90C with steam held on: rejected under a 100C ceiling,
	// accepted once the ceiling is configured below it.
	program := func() Program {
		return Program{ProgramSteps: []ProgramStep{
			steamHeatingStep(90, steamDelta()),
			steamAcclimateStep(90, &PowerPidSettings{Power: u8(50)}),
			steamCoolingStep(30, &PowerPidSettings{Power: u8(0)}),
		}}
	}

	high := uint8(100)
	defaults.SteamCeiling = &high
	p := program()
	p.ApplyDefaults(defaults)
	if err := p.Validate(); err == nil {
		t.Fatal("expected steam-on acclimate below the ceiling to be rejected")
	}

	low := uint8(80)
	defaults.SteamCeiling = &low
	p = program()
	p.ApplyDefaults(defaults)
	if err := p.Validate(); err != nil {
		t.Fatalf("expected acclimate above a lowered ceiling to pass, got %v", err)
	}
}

// A program that says nothing about equalization still gets a band, so the
// customer's existing programs keep working and pick up the installation's
// configured default.
func TestEqualizeDeltaFallsBackToDefaults(t *testing.T) {
	defaults := templateDefaults(t)
	program := validProgram()
	program.ApplyDefaults(defaults)

	if program.Equalize == nil {
		t.Fatal("ApplyDefaults left Equalize nil")
	}
	if program.Equalize.Delta == nil || *program.Equalize.Delta != *defaults.Equalize.Delta {
		t.Errorf("delta = %v, want the configured default %v", program.Equalize.Delta, *defaults.Equalize.Delta)
	}
	if program.Equalize.SteamPrewarm == nil || *program.Equalize.SteamPrewarm != *defaults.Equalize.SteamPrewarm {
		t.Errorf("steam_prewarm = %v, want the configured default %v",
			program.Equalize.SteamPrewarm, *defaults.Equalize.SteamPrewarm)
	}
}

func TestProgramDeltaWinsOverDefaults(t *testing.T) {
	program := validProgram()
	program.Equalize = &EqualizeSettings{Delta: f32(0.5), SteamPrewarm: b(true)}
	program.ApplyDefaults(templateDefaults(t))

	if *program.Equalize.Delta != 0.5 {
		t.Errorf("delta = %v, want the program's own 0.5", *program.Equalize.Delta)
	}
	if program.Equalize.SteamPrewarm == nil || !*program.Equalize.SteamPrewarm {
		t.Error("the program's own steam_prewarm was dropped")
	}
}

// A band of zero or less can never be satisfied, so it is a program that would
// wait forever rather than one that starts promptly.
func TestEqualizeDeltaMustBePositive(t *testing.T) {
	for _, delta := range []float32{0, -1} {
		program := validProgram()
		program.Equalize = &EqualizeSettings{Delta: f32(delta)}
		program.ApplyDefaults(templateDefaults(t))

		if err := program.Validate(); err == nil {
			t.Errorf("delta %v was accepted, want an error", delta)
		}
	}
}

// The startup steps are the control unit's to create. A program naming them
// would bypass the rules the authored steps are held to.
func TestStartupStepTypesCannotBeAuthored(t *testing.T) {
	for _, stepType := range []StepType{StepTypeEqualize, StepTypeSteamPrewarm} {
		program := validProgram()
		program.ProgramSteps[0].StepType = stepType
		program.ApplyDefaults(templateDefaults(t))

		if err := program.Validate(); err == nil {
			t.Errorf("step type %q was accepted in an authored program", stepType)
		}
	}
}

func TestPrependStartupSteps(t *testing.T) {
	tests := []struct {
		name      string
		prewarm   bool
		wantTypes []StepType
	}{
		{
			name:      "without steam prewarm",
			prewarm:   false,
			wantTypes: []StepType{StepTypeEqualize, StepTypeHeating},
		},
		{
			name:      "with steam prewarm",
			prewarm:   true,
			wantTypes: []StepType{StepTypeEqualize, StepTypeSteamPrewarm, StepTypeHeating},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := validProgram()
			program.Equalize = &EqualizeSettings{SteamPrewarm: b(tt.prewarm)}
			program.ApplyDefaults(templateDefaults(t))
			if err := program.Validate(); err != nil {
				t.Fatalf("the authored program did not validate: %v", err)
			}
			authored := len(program.ProgramSteps)

			program.PrependStartupSteps()

			added := len(tt.wantTypes) - 1
			if len(program.ProgramSteps) != authored+added {
				t.Fatalf("got %d steps, want %d", len(program.ProgramSteps), authored+added)
			}
			for i, want := range tt.wantTypes {
				if program.ProgramSteps[i].StepType != want {
					t.Errorf("step %d is %q, want %q", i, program.ProgramSteps[i].StepType, want)
				}
			}
			if program.ProgramSteps[0].Name == "" {
				t.Error("the synthesized equalize step has no name; status and the execution log key on it")
			}
		})
	}
}

// Expansion must not be able to rescue a program whose own first step is not a
// heating step: the rules apply to what the operator authored.
func TestValidationSeesTheAuthoredProgram(t *testing.T) {
	program := validProgram()
	program.ProgramSteps[0] = steamCoolingStep(30, &PowerPidSettings{Power: u8(0)})
	program.ApplyDefaults(templateDefaults(t))

	if err := program.Validate(); err == nil {
		t.Fatal("a program starting with a cooling step was accepted")
	}
}
