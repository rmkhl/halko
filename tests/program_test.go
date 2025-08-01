package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

func TestProgramValidation(t *testing.T) {
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

			err = program.Validate()
			if err != nil {
				t.Errorf("Program validation failed for %s: %v", tt.filename, err)
			}
		})
	}
}
