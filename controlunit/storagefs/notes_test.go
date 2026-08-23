package storagefs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rmkhl/halko/types"
)

func testStorage(t *testing.T) *ExecutorFileStorage {
	t.Helper()
	storage, err := NewExecutorFileStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage: %v", err)
	}
	return storage
}

func note(text string, at int64) types.RunNote {
	return types.RunNote{
		Time:         at,
		Step:         "Lammitys",
		Temperatures: types.TemperatureStatus{Material: 42.5, Kiln: 55.0},
		Text:         text,
	}
}

// The notes file is the durable copy: what Add writes must be readable back as
// the same notes, since the history view and the PDF read the file, never the
// writer.
func TestNotesWriterWritesWhatItWasGiven(t *testing.T) {
	storage := testStorage(t)
	writer := NewNotesWriter(storage, "run@2026-08-23T10:00:00+03:00")

	if err := writer.Add(note("smells of resin", 1000)); err != nil {
		t.Fatalf("Add: %v", err)
	}

	path := filepath.Join(storage.runningPath, "run@2026-08-23T10:00:00+03:00.notes")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading notes file: %v", err)
	}

	var stored []types.RunNote
	if err := json.Unmarshal(content, &stored); err != nil {
		t.Fatalf("notes file is not a JSON array: %v (%s)", err, content)
	}
	if len(stored) != 1 {
		t.Fatalf("stored %d notes, want 1", len(stored))
	}
	if stored[0].Text != "smells of resin" || stored[0].Step != "Lammitys" {
		t.Errorf("stored note = %+v", stored[0])
	}
	if stored[0].Temperatures.Kiln != 55.0 {
		t.Errorf("stored kiln temperature = %v, want 55", stored[0].Temperatures.Kiln)
	}
}

// Notes are a record of a run in order, so the file must read back oldest
// first no matter how many were taken.
func TestNotesWriterKeepsTheOrderTheyWereWritten(t *testing.T) {
	storage := testStorage(t)
	writer := NewNotesWriter(storage, "run")

	for i, text := range []string{"first", "second", "third"} {
		if err := writer.Add(note(text, int64(1000+i))); err != nil {
			t.Fatalf("Add(%s): %v", text, err)
		}
	}

	content, err := os.ReadFile(filepath.Join(storage.runningPath, "run.notes"))
	if err != nil {
		t.Fatalf("reading notes file: %v", err)
	}
	var stored []types.RunNote
	if err := json.Unmarshal(content, &stored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := []string{"first", "second", "third"}
	if len(stored) != len(want) {
		t.Fatalf("stored %d notes, want %d", len(stored), len(want))
	}
	for i, text := range want {
		if stored[i].Text != text {
			t.Errorf("note %d = %q, want %q", i, stored[i].Text, text)
		}
	}
}

// A run with no notes must leave no file behind, so MoveToHistory and
// DeleteExecutedProgram have nothing to move or delete for it.
func TestNotesWriterCreatesNoFileUntilAskedTo(t *testing.T) {
	storage := testStorage(t)
	NewNotesWriter(storage, "run")

	if _, err := os.Stat(filepath.Join(storage.runningPath, "run.notes")); !os.IsNotExist(err) {
		t.Errorf("notes file exists before any note was added (err = %v)", err)
	}
}

// The runner closes the writer immediately before moving the run's files to
// history. A note accepted after that would recreate the file in running/ and
// strand it there, so a closed writer must refuse.
func TestNotesWriterRefusesAfterClose(t *testing.T) {
	storage := testStorage(t)
	writer := NewNotesWriter(storage, "run")
	writer.Close()

	if err := writer.Add(note("too late", 2000)); !errors.Is(err, ErrNotesClosed) {
		t.Errorf("Add after Close = %v, want ErrNotesClosed", err)
	}
	if _, err := os.Stat(filepath.Join(storage.runningPath, "run.notes")); !os.IsNotExist(err) {
		t.Errorf("Add after Close created the notes file (err = %v)", err)
	}
}

// The writer serves the running view's listing, so it must report what it has
// without a file read.
func TestNotesWriterReportsItsNotes(t *testing.T) {
	storage := testStorage(t)
	writer := NewNotesWriter(storage, "run")
	_ = writer.Add(note("one", 1000))
	_ = writer.Add(note("two", 1001))

	notes := writer.Notes()

	if len(notes) != 2 || notes[0].Text != "one" || notes[1].Text != "two" {
		t.Errorf("Notes() = %+v", notes)
	}
}

// The notes file has to travel with the rest of the run's artifacts, or a run
// is annotated right up until it finishes and then reads back blank.
func TestMoveToHistoryTakesTheNotes(t *testing.T) {
	storage := testStorage(t)
	const run = "run@2026-08-23T10:00:00+03:00"
	writer := NewNotesWriter(storage, run)
	if err := writer.Add(note("kept", 1000)); err != nil {
		t.Fatalf("Add: %v", err)
	}
	writer.Close()

	if err := storage.MoveToHistory(run); err != nil {
		t.Fatalf("MoveToHistory: %v", err)
	}

	if _, err := os.Stat(filepath.Join(storage.runningPath, run+".notes")); !os.IsNotExist(err) {
		t.Errorf("notes file still in running/ (err = %v)", err)
	}
	notes, err := storage.LoadRunNotes(run)
	if err != nil {
		t.Fatalf("LoadRunNotes: %v", err)
	}
	if len(notes) != 1 || notes[0].Text != "kept" {
		t.Errorf("LoadRunNotes = %+v, want one note %q", notes, "kept")
	}
}

// Most runs have no notes, so the move must not start failing for them.
func TestMoveToHistorySucceedsWithNoNotes(t *testing.T) {
	storage := testStorage(t)
	const run = "quiet-run"

	if err := storage.MoveToHistory(run); err != nil {
		t.Fatalf("MoveToHistory with no notes: %v", err)
	}
}

// A run nobody annotated is not an error to read; the history view asks for
// every run's notes the same way.
func TestLoadRunNotesOfARunWithNoneIsEmpty(t *testing.T) {
	storage := testStorage(t)

	notes, err := storage.LoadRunNotes("never-annotated")

	if err != nil {
		t.Fatalf("LoadRunNotes: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("LoadRunNotes = %+v, want empty", notes)
	}
}

// The name reaches this from a URL path, so it gets the same treatment as
// every other name-keyed storage call.
func TestLoadRunNotesRejectsAnEscapingName(t *testing.T) {
	storage := testStorage(t)

	if _, err := storage.LoadRunNotes("../../etc/passwd"); !errors.Is(err, types.ErrInvalidStorageName) {
		t.Errorf("LoadRunNotes with an escaping name = %v, want ErrInvalidStorageName", err)
	}
}

// Deleting a run deletes everything about it.
func TestDeleteExecutedProgramTakesTheNotes(t *testing.T) {
	storage := testStorage(t)
	const run = "doomed"
	writer := NewNotesWriter(storage, run)
	_ = writer.Add(note("gone soon", 1000))
	writer.Close()
	if err := storage.MoveToHistory(run); err != nil {
		t.Fatalf("MoveToHistory: %v", err)
	}

	_ = storage.DeleteExecutedProgram(run)

	if _, err := os.Stat(filepath.Join(storage.notesPath, run+".json")); !os.IsNotExist(err) {
		t.Errorf("notes file survived deletion (err = %v)", err)
	}
}

// The reason the running file is not called <run>.notes.json: anything
// matching running/*.json is recovered as an abandoned run on the next start.
func TestNotesFileIsNotMistakenForARunningProgram(t *testing.T) {
	storage := testStorage(t)
	writer := NewNotesWriter(storage, "run@2026-08-23T10:00:00+03:00")
	_ = writer.Add(note("hello", 1000))

	running, err := storage.ListRunningPrograms()

	if err != nil {
		t.Fatalf("ListRunningPrograms: %v", err)
	}
	if len(running) != 0 {
		t.Errorf("ListRunningPrograms = %v, want none - the notes file was picked up as a run", running)
	}
}
