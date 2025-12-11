package storagefs

import (
	"github.com/rmkhl/halko/types"
)

type (
	StateWriter struct {
		storage *ExecutorFileStorage
		name    string
	}
)

func NewStateWriter(fileStorage *ExecutorFileStorage, name string) *StateWriter {
	return &StateWriter{storage: fileStorage, name: name}
}

func (statusWriter *StateWriter) UpdateState(status types.ProgramState) error {
	return statusWriter.storage.UpdateState(statusWriter.name, status)
}
