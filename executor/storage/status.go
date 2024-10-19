package storage

import (
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types"
)

type (
	StateWriter struct {
		storage *ProgramStorage
		name    string
	}
)

func NewStateWriter(storage *ProgramStorage, name string) *StateWriter {
	return &StateWriter{storage: storage, name: name}
}

func (statusWriter *StateWriter) UpdateState(status types.ProgramState) error {
	filePath := filepath.Join(statusWriter.storage.statusPath, statusWriter.name+".txt")

	statusFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer statusFile.Close()

	_, err = statusFile.WriteString(string(status))
	if err != nil {
		return err
	}
	return nil
}

func (storage *ProgramStorage) MaybeDeleteState(name string) {
	statusFilePath := filepath.Join(storage.statusPath, name+".txt")
	os.Remove(statusFilePath)
}

// Retuns saved state and time it was saved.
func (storage *ProgramStorage) LoadState(name string) (types.ProgramState, int64, error) {
	statusFilePath := filepath.Join(storage.statusPath, name+".txt")

	fileStatus, err := os.Stat(statusFilePath)
	if err != nil {
		return types.ProgramStateUnknown, 0, err
	}

	status, err := os.ReadFile(statusFilePath)
	if err != nil {
		return types.ProgramStateUnknown, 0, err
	}
	return types.ProgramState(status), fileStatus.ModTime().Unix(), nil
}
