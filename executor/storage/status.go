package storage

import (
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/executor/types"
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

func (storage *ProgramStorage) LoadState(name string) (types.ProgramState, error) {
	statusFilePath := filepath.Join(storage.statusPath, name+".txt")

	statusFile, err := os.Open(statusFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return types.ProgramStateUnknown, nil
		}
		return types.ProgramStateUnknown, err
	}
	defer statusFile.Close()

	status := make([]byte, 10)
	_, err = statusFile.Read(status)
	if err != nil {
		return types.ProgramStateUnknown, err
	}
	return types.ProgramState(status), nil
}
