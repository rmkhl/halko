package storagefs

import (
	"encoding/csv"
	"os"
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

// readLog returns the CSV rows written so far, header included.
func readLog(t *testing.T, path string) [][]string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open log: %v", err)
	}
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse log: %v", err)
	}
	return rows
}

func statusAt(step string, kiln, material float32) *types.ExecutionStatus {
	return &types.ExecutionStatus{
		CurrentStep:          step,
		CurrentStepStartedAt: time.Now().Unix(),
		Temperatures:         types.TemperatureStatus{Kiln: kiln, Material: material},
		PowerStatus:          types.PSUStatus{Heater: 80, Fan: 50, Humidifier: 0},
	}
}

func TestExecutionLogWritesItsHeaderOnCreation(t *testing.T) {
	storage := newTestStorage(t)

	writer := NewExecutionLogWriter(storage, runName, 60, time.Now().Unix())
	if writer == nil {
		t.Fatal("expected a writer")
	}
	defer writer.Close()

	rows := readLog(t, storage.GetRunningLogPath(runName))
	if len(rows) != 1 {
		t.Fatalf("expected only a header row, got %d rows", len(rows))
	}

	want := []string{"time", "step", "steptime", "material", "kiln", "heater", "fan", "humidifier"}
	if len(rows[0]) != len(want) {
		t.Fatalf("expected %d columns, got %v", len(want), rows[0])
	}
	for i, column := range want {
		if rows[0][i] != column {
			t.Fatalf("column %d: expected %q, got %q", i, column, rows[0][i])
		}
	}
}

func TestExecutionLogSkipsNonRunningSteps(t *testing.T) {
	storage := newTestStorage(t)

	writer := NewExecutionLogWriter(storage, runName, 60, time.Now().Unix())
	defer writer.Close()

	for _, step := range []string{"", "Waiting", "Completed"} {
		writer.AddLine(statusAt(step, 50, 40))
	}

	rows := readLog(t, storage.GetRunningLogPath(runName))
	if len(rows) != 1 {
		t.Fatalf("expected nothing but the header, got %v", rows)
	}
}

// The resolution throttles steady state, but a step change must always be
// recorded or the log would not show when the run moved on.
func TestExecutionLogWritesOnEveryStepChange(t *testing.T) {
	storage := newTestStorage(t)

	// A resolution far longer than the test runs: only step changes get through.
	writer := NewExecutionLogWriter(storage, runName, 3600, time.Now().Unix())
	defer writer.Close()

	writer.AddLine(statusAt(stepHeating, 50, 40))
	writer.AddLine(statusAt(stepHeating, 51, 41))
	writer.AddLine(statusAt("Acclimate", 52, 42))
	writer.AddLine(statusAt("Acclimate", 53, 43))
	writer.AddLine(statusAt("Cooling", 54, 44))

	rows := readLog(t, storage.GetRunningLogPath(runName))

	// Header plus one row per step change.
	if len(rows) != 4 {
		t.Fatalf("expected a header and 3 step rows, got %d rows: %v", len(rows), rows)
	}
	for i, step := range []string{stepHeating, "Acclimate", "Cooling"} {
		if rows[i+1][1] != step {
			t.Fatalf("row %d: expected step %q, got %q", i+1, step, rows[i+1][1])
		}
	}
}

func TestExecutionLogWritesWithinTheSameStepOnceTheResolutionElapses(t *testing.T) {
	storage := newTestStorage(t)

	// A resolution of 0 means every call is due.
	writer := NewExecutionLogWriter(storage, runName, 0, time.Now().Unix())
	defer writer.Close()

	writer.AddLine(statusAt(stepHeating, 50, 40))
	writer.AddLine(statusAt(stepHeating, 51, 41))
	writer.AddLine(statusAt(stepHeating, 52, 42))

	rows := readLog(t, storage.GetRunningLogPath(runName))
	if len(rows) != 4 {
		t.Fatalf("expected a header and 3 rows, got %d rows: %v", len(rows), rows)
	}
}

func TestExecutionLogRecordsTemperaturesAndPower(t *testing.T) {
	storage := newTestStorage(t)

	startedAt := time.Now().Unix()
	writer := NewExecutionLogWriter(storage, runName, 0, startedAt)
	defer writer.Close()

	writer.AddLine(statusAt(stepHeating, 55.57, 44.42))

	rows := readLog(t, storage.GetRunningLogPath(runName))
	if len(rows) != 2 {
		t.Fatalf("expected a header and one row, got %v", rows)
	}

	row := rows[1]
	for i, want := range map[int]string{
		1: "Heating",
		3: "44.4", // material, one decimal
		4: "55.6", // kiln, one decimal
		5: "80",   // heater
		6: "50",   // fan
		7: "0",    // humidifier
	} {
		if row[i] != want {
			t.Fatalf("column %d: expected %q, got %q (row %v)", i, want, row[i], row)
		}
	}
}

func TestExecutionLogGetStartTime(t *testing.T) {
	storage := newTestStorage(t)

	startedAt := time.Now().Unix()
	writer := NewExecutionLogWriter(storage, runName, 60, startedAt)
	defer writer.Close()

	if got := writer.GetStartTime(); got != startedAt {
		t.Fatalf("expected start time %d, got %d", startedAt, got)
	}
}

// A run that ends while a status update is in flight must not take the service
// down with it.
func TestExecutionLogIsSafeAfterClose(t *testing.T) {
	storage := newTestStorage(t)

	writer := NewExecutionLogWriter(storage, runName, 0, time.Now().Unix())
	writer.AddLine(statusAt(stepHeating, 50, 40))
	writer.Close()

	writer.AddLine(statusAt(stepHeating, 51, 41))
	writer.Close()

	rows := readLog(t, storage.GetRunningLogPath(runName))
	if len(rows) != 2 {
		t.Fatalf("expected the log to stop at 2 rows, got %v", rows)
	}
}

// NewExecutionLogWriter returns nil when it cannot create the file, and every
// method has to tolerate that.
func TestNilExecutionLogWriterIsSafe(t *testing.T) {
	var writer *ExecutionLogWriter

	writer.AddLine(statusAt(stepHeating, 50, 40))
	writer.Close()

	if got := writer.GetStartTime(); got != 0 {
		t.Fatalf("expected 0 from a nil writer, got %d", got)
	}
}
