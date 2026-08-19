package storagefs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

const (
	stepHeating = "Heating"
	runName     = "run-1"
)

func newTestStorage(t *testing.T) *ExecutorFileStorage {
	t.Helper()

	storage, err := NewExecutorFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	return storage
}

func testProgram(name string) *types.Program {
	return &types.Program{
		ProgramName: name,
		ProgramSteps: []types.ProgramStep{
			{Name: stepHeating, StepType: types.StepTypeHeating, TargetTemperature: 60},
		},
	}
}

// startRun puts a program into the running directory the way a real run does.
func startRun(t *testing.T, storage *ExecutorFileStorage, name string) {
	t.Helper()

	if err := storage.CreateExecutedProgram(name, testProgram(name)); err != nil {
		t.Fatalf("failed to create running program: %v", err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be gone, stat returned %v", path, err)
	}
}

func mustContain(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(content) != want {
		t.Fatalf("expected %s to contain %q, got %q", path, want, string(content))
	}
}

func TestNewExecutorFileStorageCreatesItsDirectories(t *testing.T) {
	base := t.TempDir()

	storage, err := NewExecutorFileStorage(base)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	for _, dir := range []string{
		storage.executedProgramsPath,
		storage.statusPath,
		storage.logPath,
		storage.runningPath,
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestCreateExecutedProgramRefusesToOverwrite(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	err := storage.CreateExecutedProgram(runName, testProgram(runName))
	if !errors.Is(err, types.ErrProgramExists) {
		t.Fatalf("expected ErrProgramExists, got %v", err)
	}
}

func TestListRunningProgramsSeesAStartedRun(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	running, err := storage.ListRunningPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(running) != 1 || running[0] != runName {
		t.Fatalf("expected [run-1], got %v", running)
	}

	// It is not history until it has been moved there.
	executed, err := storage.ListExecutedPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 0 {
		t.Fatalf("expected no executed programs yet, got %v", executed)
	}
}

func TestMoveToHistoryMovesEveryFile(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	// The files a real run leaves in running/: status and the execution log.
	writer := NewStateWriter(storage, runName)
	if err := writer.UpdateState(types.ProgramStateRunning); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}
	if err := os.WriteFile(runningLogPathOf(t, storage, runName), []byte("time,step\n"), 0o644); err != nil {
		t.Fatalf("failed to write log: %v", err)
	}

	if err := storage.MoveToHistory(runName); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	executed, err := storage.ListExecutedPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 1 || executed[0] != runName {
		t.Fatalf("expected [run-1] in history, got %v", executed)
	}

	mustContain(t, logPathOf(t, storage, runName), "time,step\n")

	running, err := storage.ListRunningPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(running) != 0 {
		t.Fatalf("expected running to be empty, got %v", running)
	}
	mustNotExist(t, runningLogPathOf(t, storage, runName))
}

// A run that failed before writing a log still has to be filed away.
func TestMoveToHistoryToleratesMissingFiles(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	if err := storage.MoveToHistory(runName); err != nil {
		t.Fatalf("expected a missing log and status to be tolerated, got %v", err)
	}

	executed, err := storage.ListExecutedPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 1 {
		t.Fatalf("expected the program in history, got %v", executed)
	}
}

// The order runner.go uses: the final state is written while the run still
// counts as running, MarkCompleted flips the target, and MoveToHistory then
// moves the running status file over. The final state has to survive that.
func TestFinalStateSurvivesTheMoveToHistory(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	writer := NewStateWriter(storage, runName)
	if err := writer.UpdateState(types.ProgramStateRunning); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}
	if err := writer.UpdateState(types.ProgramStateCompleted); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	writer.MarkCompleted()
	if err := storage.MoveToHistory(runName); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, _, err := storage.LoadState(runName)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if state != types.ProgramStateCompleted {
		t.Fatalf("expected state %q, got %q", types.ProgramStateCompleted, state)
	}
}

func TestLoadStateReportsUnknownForAnAbsentProgram(t *testing.T) {
	storage := newTestStorage(t)

	state, updatedAt, err := storage.LoadState("nope")
	if err == nil {
		t.Fatal("expected an error for a program with no state file")
	}
	if state != types.ProgramStateUnknown {
		t.Fatalf("expected %q, got %q", types.ProgramStateUnknown, state)
	}
	if updatedAt != 0 {
		t.Fatalf("expected no timestamp, got %d", updatedAt)
	}
}

func TestUpdateStateAndLoadStateRoundTrip(t *testing.T) {
	storage := newTestStorage(t)

	if err := storage.UpdateState(runName, types.ProgramStateFailed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, updatedAt, err := storage.LoadState(runName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != types.ProgramStateFailed {
		t.Fatalf("expected %q, got %q", types.ProgramStateFailed, state)
	}
	if updatedAt == 0 {
		t.Fatal("expected a modification timestamp")
	}
}

func TestDeleteExecutedProgramRemovesEveryFile(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	writer := NewStateWriter(storage, runName)
	if err := writer.UpdateState(types.ProgramStateCompleted); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}
	if err := os.WriteFile(runningLogPathOf(t, storage, runName), []byte("time,step\n"), 0o644); err != nil {
		t.Fatalf("failed to write log: %v", err)
	}
	if err := storage.MoveToHistory(runName); err != nil {
		t.Fatalf("failed to move to history: %v", err)
	}

	if err := storage.DeleteExecutedProgram(runName); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustNotExist(t, filepath.Join(storage.executedProgramsPath, runName+".json"))
	mustNotExist(t, logPathOf(t, storage, runName))
	mustNotExist(t, filepath.Join(storage.statusPath, runName+".txt"))
}

func TestDeleteExecutedProgramReportsAMissingProgram(t *testing.T) {
	storage := newTestStorage(t)

	// The log and status files are absent too, but only the missing program
	// itself is an error.
	if err := storage.DeleteExecutedProgram("nope"); err == nil {
		t.Fatal("expected an error for a program that was never there")
	}
}

// Crash recovery: anything still sitting in running/ at startup belongs to a
// run that never finished, and has to be filed as canceled.
func TestCleanupOrphanedRunningFilesTheRunAsCanceled(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)

	writer := NewStateWriter(storage, runName)
	if err := writer.UpdateState(types.ProgramStateRunning); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	if err := storage.CleanupOrphanedRunning(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	running, err := storage.ListRunningPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(running) != 0 {
		t.Fatalf("expected running to be empty, got %v", running)
	}

	executed, err := storage.ListExecutedPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 1 || executed[0] != runName {
		t.Fatalf("expected [run-1] in history, got %v", executed)
	}

	state, _, err := storage.LoadState(runName)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if state != types.ProgramStateCanceled {
		t.Fatalf("expected %q, got %q", types.ProgramStateCanceled, state)
	}
}

func TestCleanupOrphanedRunningIsANoOpOnACleanStart(t *testing.T) {
	storage := newTestStorage(t)

	if err := storage.CleanupOrphanedRunning(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	executed, err := storage.ListExecutedPrograms()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executed) != 0 {
		t.Fatalf("expected history to stay empty, got %v", executed)
	}
}

func TestLoadExecutedProgramReturnsWhatWasSaved(t *testing.T) {
	storage := newTestStorage(t)
	startRun(t, storage, runName)
	if err := storage.MoveToHistory(runName); err != nil {
		t.Fatalf("failed to move to history: %v", err)
	}

	program, err := storage.LoadExecutedProgram(runName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program.ProgramName != runName {
		t.Fatalf("expected program name run-1, got %q", program.ProgramName)
	}
	if len(program.ProgramSteps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(program.ProgramSteps))
	}
}
