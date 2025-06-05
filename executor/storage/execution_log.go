package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rmkhl/halko/types"
)

type (
	ExecutionLogWriter struct {
		storage    *ProgramStorage
		name       string
		file       *os.File
		csvWriter  *csv.Writer
		resolution int64
		startedAt  int64
		lastUpdate int64
		lastStep   string
	}
)

func NewExecutionLogWriter(storage *ProgramStorage, name string, resolution int64) *ExecutionLogWriter {
	filePath := filepath.Join(storage.logPath, name+".csv")
	logFile, err := os.Create(filePath)
	if err != nil {
		return nil
	}
	writer := ExecutionLogWriter{
		storage:    storage,
		name:       name,
		file:       logFile,
		csvWriter:  csv.NewWriter(logFile),
		resolution: resolution,
		lastUpdate: 0,
		lastStep:   "",
		startedAt:  time.Now().Unix(),
	}
	_ = writer.csvWriter.Write([]string{
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
	if now-writer.lastUpdate > writer.resolution && status.CurrentStep == writer.lastStep {
		return
	}
	_ = writer.csvWriter.Write([]string{
		strconv.FormatInt(now-writer.startedAt, 10),
		status.CurrentStep,
		strconv.FormatInt(now-status.CurrentStepStartedAt, 10),
		fmt.Sprintf("%f", status.Temperatures.Material),
		fmt.Sprintf("%f", status.Temperatures.Oven),
		strconv.Itoa(int(status.PowerStatus.Heater)),
		strconv.Itoa(int(status.PowerStatus.Fan)),
		strconv.Itoa(int(status.PowerStatus.Humidifier)),
	})
	writer.csvWriter.Flush()
	writer.lastUpdate = now
	writer.lastStep = status.CurrentStep
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
