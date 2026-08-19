package types

import "testing"

// Names become filenames under base_path. Anything that is not a single plain
// segment can escape it once joined, so the rule is a whitelist: one segment,
// nothing else.
func TestValidateStorageNameAcceptsPlainNames(t *testing.T) {
	names := []string{
		"Oak drying",
		"Oak drying@2026-08-19T14:30:00Z",
		"program_1",
		"a.b",
	}

	for _, name := range names {
		if err := ValidateStorageName(name); err != nil {
			t.Errorf("expected %q to be accepted, got %v", name, err)
		}
	}
}

func TestValidateStorageNameRejectsEscapes(t *testing.T) {
	names := map[string]string{
		"parent":         "..",
		"current":        ".",
		"relative path":  "../../etc/passwd",
		"leading slash":  "/etc/passwd",
		"embedded slash": "runs/../../etc/passwd",
		"trailing slash": "oak/",
		"backslash":      `..\..\windows`,
		"empty":          "",
		"nul byte":       "oak\x00.json",
	}

	for label, name := range names {
		if err := ValidateStorageName(name); err == nil {
			t.Errorf("%s: expected %q to be rejected, got nil", label, name)
		}
	}
}
