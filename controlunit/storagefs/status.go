package storagefs

import (
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type (
	StateWriter struct {
		storage   *ExecutorFileStorage
		name      string
		isRunning bool
	}
)

func NewStateWriter(fileStorage *ExecutorFileStorage, name string) *StateWriter {
	return &StateWriter{storage: fileStorage, name: name, isRunning: true}
}

func (statusWriter *StateWriter) UpdateState(status types.ProgramState) error {
	// Determine file path based on whether program is still running
	var filePath string
	if statusWriter.isRunning {
		filePath = filepath.Join(statusWriter.storage.runningPath, statusWriter.name+".txt")
	} else {
		filePath = filepath.Join(statusWriter.storage.statusPath, statusWriter.name+".txt")
	}

	log.Debug("Updating state for program '%s' to '%s' in %s", statusWriter.name, status, filePath)
	statusFile, err := os.Create(filePath)
	if err != nil {
		log.Error("Failed to create status file for program '%s': %v", statusWriter.name, err)
		return err
	}
	defer statusFile.Close()

	_, err = statusFile.WriteString(string(status))
	if err != nil {
		log.Error("Failed to write status to file for program '%s': %v", statusWriter.name, err)
		return err
	}
	log.Info("Successfully updated state for program '%s' to '%s'", statusWriter.name, status)
	return nil
}

func (statusWriter *StateWriter) MarkCompleted() {
	statusWriter.isRunning = false
}
