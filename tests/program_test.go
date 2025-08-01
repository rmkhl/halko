package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestProgramValidation(t *testing.T) {
	// Load defaults from the template config
	config, err := types.ReadHalkoConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	if config.ExecutorConfig == nil || config.ExecutorConfig.Defaults == nil {
		t.Fatal("Config or defaults not found")
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

			var program types.Program
			err = json.Unmarshal(data, &program)
			if err != nil {
				t.Fatalf("Failed to unmarshal program from %s: %v", tt.filename, err)
			}

			// Apply defaults before validation
			program.ApplyDefaults(config.ExecutorConfig.Defaults)

			err = program.Validate()
			if err != nil {
				t.Errorf("Program validation failed for %s: %v", tt.filename, err)
			}
		})
	}
}

func TestProgramValidationWithCopy(t *testing.T) {
	// Load defaults from the template config
	config, err := types.ReadHalkoConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("Failed to read template config: %v", err)
	}

	if config.ExecutorConfig == nil || config.ExecutorConfig.Defaults == nil {
		t.Fatal("Config or defaults not found")
	}

	// Test the new validation pattern used in createProgram/updateProgram
	examplePath := filepath.Join("..", "example", "example-program-delta.json")
	data, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example file: %v", err)
	}

	var originalProgram types.Program
	err = json.Unmarshal(data, &originalProgram)
	if err != nil {
		t.Fatalf("Failed to unmarshal program: %v", err)
	}

	// Store the original state to verify it's not modified
	originalJSON, _ := json.Marshal(originalProgram)

	// Create a deep copy for validation (simulating the router behavior)
	programCopy, err := originalProgram.Duplicate()
	if err != nil {
		t.Fatalf("Failed to duplicate program: %v", err)
	}

	// Apply defaults to the copy and validate
	programCopy.ApplyDefaults(config.ExecutorConfig.Defaults)
	err = programCopy.Validate()
	if err != nil {
		t.Errorf("Program validation failed on copy: %v", err)
	}

	// Verify the original program was not modified
	currentJSON, _ := json.Marshal(originalProgram)
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
