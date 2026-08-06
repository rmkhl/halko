package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}

func TestIsJSONFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"object", `{"name":"test"}`, true},
		{"array", `[1,2,3]`, true},
		{"leading whitespace", "\n  {\"name\":\"test\"}", true},
		{"plain text", "hello", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTemp(t, "candidate.txt", tt.content)
			if got := isJSONFile(path); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestIsJSONFileOnAMissingFile(t *testing.T) {
	if isJSONFile(filepath.Join(t.TempDir(), "nope.json")) {
		t.Fatal("expected false for a file that does not exist")
	}
}

func TestLoadProgramFromFile(t *testing.T) {
	path := writeTemp(t, "program.json",
		`{"name":"Oak drying","steps":[{"name":"Heating","type":"heating","temperature_target":60}]}`)

	program, err := loadProgramFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.ProgramName != "Oak drying" {
		t.Fatalf("expected program name %q, got %q", "Oak drying", program.ProgramName)
	}
	if len(program.ProgramSteps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(program.ProgramSteps))
	}
}

// A program with no name of its own is named after the file.
func TestLoadProgramFromFileFallsBackToTheFilename(t *testing.T) {
	path := writeTemp(t, "oak-run.json", `{"steps":[]}`)

	program, err := loadProgramFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.ProgramName != "oak-run" {
		t.Fatalf("expected program name %q, got %q", "oak-run", program.ProgramName)
	}
}

func TestLoadProgramFromFileRejectsNonJSON(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"not json at all", "notes.txt", "just some notes"},
		{"malformed json", "program.json", `{"name":`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTemp(t, tt.filename, tt.content)
			if _, err := loadProgramFromFile(path); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestLoadProgramFromFileReportsAMissingFile(t *testing.T) {
	if _, err := loadProgramFromFile(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected an error for a file that does not exist")
	}
}
