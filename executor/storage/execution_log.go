package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rmkhl/halko/types"
)

type (
	ExecutionLogWriter struct {
		storage     *ProgramStorage
		name        string
		file        *os.File
		csvWriter   *csv.Writer
		resolution  int64
		started_at  int64
		last_update int64
		last_step   string
	}
)

func NewExecutionLogWriter(storage *ProgramStorage, name string, resolution int64) *ExecutionLogWriter {
	filePath := filepath.Join(storage.logPath, name+".csv")
	logFile, err := os.Create(filePath)
	if err != nil {
		return nil
	}
	writer := ExecutionLogWriter{
		storage:     storage,
		name:        name,
		file:        logFile,
		csvWriter:   csv.NewWriter(logFile),
		resolution:  resolution,
		last_update: 0,
		last_step:   "",
		started_at:  time.Now().Unix(),
	}
	writer.csvWriter.Write([]string{
		"time",
		"step",
		"steptime",
		"material",
		"oven",
		"heater",
		"fan",
		"humidifier",
	})
	writer.csvWriter.Flush()
	return &writer
}

func (writer *ExecutionLogWriter) AddLine(status *types.ExecutionStatus) {
	if writer == nil {
		return
	}
	if writer.csvWriter == nil {
		return
	}
	now := time.Now().Unix()
	if now-writer.last_update > writer.resolution && status.CurrentStep == writer.last_step {
		return
	}
	writer.csvWriter.Write([]string{
		fmt.Sprintf("%d", now-writer.started_at),
		status.CurrentStep,
		fmt.Sprintf("%d", now-status.CurrentStepStartedAt),
		fmt.Sprintf("%g", status.Temperatures.Material),
		fmt.Sprintf("%g", status.Temperatures.Oven),
		fmt.Sprintf("%d", status.PowerStatus.Heater),
		fmt.Sprintf("%d", status.PowerStatus.Fan),
		fmt.Sprintf("%d", status.PowerStatus.Humidifier),
	})
	writer.csvWriter.Flush()
	writer.last_update = now
	writer.last_step = status.CurrentStep
}

func (writer *ExecutionLogWriter) Close() {
	if writer == nil {
		return
	}
	_ = writer.file.Close()
	writer.csvWriter = nil
	writer.file = nil
}

func (storage *ProgramStorage) MaybeDeleteExecutionLog(name string) {
	filePath := filepath.Join(storage.logPath, name+".csv")
	os.Remove(filePath)
}
