package storagefs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

// escapingName is what a caller sends to get out of the storage directory.
// filepath.Join cleans it into a real path above the base rather than
// rejecting it, so the storage layer has to refuse it itself.
const escapingName = "../../escaped"

func TestCreateStoredProgramRefusesToEscapeBasePath(t *testing.T) {
	base := t.TempDir()
	storage, err := NewProgramStorage(filepath.Join(base, "fsdb"))
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}

	err = storage.CreateStoredProgram(escapingName, testProgram(escapingName))

	if !errors.Is(err, types.ErrInvalidStorageName) {
		t.Fatalf("expected ErrInvalidStorageName, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(base, "escaped.json")); statErr == nil {
		t.Errorf("a file was written outside the base path")
	}
}

func TestExecutedProgramOperationsRefuseEscapingNames(t *testing.T) {
	storage := newTestStorage(t)

	if _, err := storage.LoadExecutedProgram(escapingName); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("LoadExecutedProgram: expected ErrInvalidStorageName, got %v", err)
	}
	if err := storage.DeleteExecutedProgram(escapingName); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("DeleteExecutedProgram: expected ErrInvalidStorageName, got %v", err)
	}
	if _, _, err := storage.LoadState(escapingName); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("LoadState: expected ErrInvalidStorageName, got %v", err)
	}
}

// The log path builders feed os.ReadFile directly in the router, so an
// escaping name there reads an arbitrary file rather than writing one.
func TestLogPathsRefuseEscapingNames(t *testing.T) {
	storage := newTestStorage(t)

	if _, err := storage.GetLogPath(escapingName); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("GetLogPath: expected ErrInvalidStorageName, got %v", err)
	}
	if _, err := storage.GetRunningLogPath(escapingName); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("GetRunningLogPath: expected ErrInvalidStorageName, got %v", err)
	}
}

// logPathOf and runningLogPathOf keep the existing tests readable now that the
// path builders validate their input and can fail.
func logPathOf(t *testing.T, storage *ExecutorFileStorage, name string) string {
	t.Helper()
	path, err := storage.GetLogPath(name)
	if err != nil {
		t.Fatalf("GetLogPath(%q): %v", name, err)
	}
	return path
}

func runningLogPathOf(t *testing.T, storage *ExecutorFileStorage, name string) string {
	t.Helper()
	path, err := storage.GetRunningLogPath(name)
	if err != nil {
		t.Fatalf("GetRunningLogPath(%q): %v", name, err)
	}
	return path
}
