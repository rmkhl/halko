package storagefs

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type (
	ExecutionLogWriter struct {
		storage    *ExecutorFileStorage
		name       string
		file       *os.File
		csvWriter  *csv.Writer
		resolution int64
		startedAt  int64
		lastUpdate int64
		lastStep   string
	}
)

func NewExecutionLogWriter(fileStorage *ExecutorFileStorage, name string, resolution int64) *ExecutionLogWriter {
	log.Info("Creating execution log writer for program '%s'", name)
	filePath := filepath.Join(fileStorage.runningPath, name+".csv")
	logFile, err := os.Create(filePath)
	if err != nil {
		log.Error("Failed to create execution log file for program '%s': %v", name, err)
		return nil
	}
	writer := ExecutionLogWriter{
		storage:    fileStorage,
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
	log.Debug("Successfully created execution log writer for program '%s'", name)
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

	// Log if: step changed OR resolution time has elapsed
	stepChanged := status.CurrentStep != writer.lastStep
	timeElapsed := now-writer.lastUpdate >= writer.resolution

	if !stepChanged && !timeElapsed {
		log.Trace("Execution log: Skipping line - step unchanged and resolution not met (last update %ds ago)", now-writer.lastUpdate)
		return
	}

	log.Trace("Execution log: Adding line for step '%s' (changed=%v, elapsed=%v)", status.CurrentStep, stepChanged, timeElapsed)
	_ = writer.csvWriter.Write([]string{
		strconv.FormatInt(now-writer.startedAt, 10),
		status.CurrentStep,
		strconv.FormatInt(now-status.CurrentStepStartedAt, 10),
		fmt.Sprintf("%.1f", status.Temperatures.Material),
		fmt.Sprintf("%.1f", status.Temperatures.Oven),
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
	log.Debug("Closing execution log writer for program '%s'", writer.name)
	_ = writer.file.Close()
	writer.csvWriter = nil
	writer.file = nil
	log.Debug("Successfully closed execution log writer for program '%s'", writer.name)
}
