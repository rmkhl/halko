package storagefs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// ErrNotesClosed is returned once the run has ended and its files have been
// handed to MoveToHistory. Writing after that would recreate the notes file
// under running/ with nothing left to move it.
var ErrNotesClosed = errors.New("run has ended")

// NotesWriter owns a run's notes file, in the same way StateWriter owns its
// status file and ExecutionLogWriter its CSV. It holds the notes in memory and
// rewrites the file whole on every addition: a run collects a handful of notes
// and every reader wants all of them, so appending in place would buy nothing
// and cost the ability to write atomically.
type NotesWriter struct {
	mu     sync.Mutex
	path   string
	name   string
	notes  []types.RunNote
	closed bool
}

// NewNotesWriter prepares to write, but creates no file. A run nobody
// annotated must leave nothing behind, so that MoveToHistory and
// DeleteExecutedProgram have nothing to find for it.
func NewNotesWriter(fileStorage *ExecutorFileStorage, name string) *NotesWriter {
	return &NotesWriter{
		path: filepath.Join(fileStorage.runningPath, name+".notes"),
		name: name,
	}
}

func (writer *NotesWriter) Add(note types.RunNote) error {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.closed {
		return ErrNotesClosed
	}

	writer.notes = append(writer.notes, note)
	if err := writeNotesFile(writer.path, writer.notes); err != nil {
		// Keep the file and the in-memory list in step: a note that did not
		// reach the disk must not be reported back as recorded.
		writer.notes = writer.notes[:len(writer.notes)-1]
		log.Error("Failed to write notes for run '%s': %v", writer.name, err)
		return err
	}
	log.Info("Recorded note for run '%s'", writer.name)
	return nil
}

// Notes returns a copy, so a caller ranging over it cannot be tripped by a
// concurrent Add reallocating the slice.
func (writer *NotesWriter) Notes() []types.RunNote {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	notes := make([]types.RunNote, len(writer.notes))
	copy(notes, writer.notes)
	return notes
}

func (writer *NotesWriter) Close() {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	writer.closed = true
}

// writeNotesFile replaces the file in one step. A crash mid-write leaves the
// previous complete file rather than a truncated one. The temp file
// deliberately keeps the .notes prefix so it cannot match the running/*.json
// glob either.
func writeNotesFile(path string, notes []types.RunNote) error {
	data, err := json.Marshal(notes)
	if err != nil {
		return err
	}

	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0644); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
