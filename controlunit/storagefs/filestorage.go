package storagefs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type ExecutorFileStorage struct {
	*types.FileStorage
	executedProgramsPath string
	statusPath           string
	logPath              string
	runningPath          string
}

func NewExecutorFileStorage(basePath string) (*ExecutorFileStorage, error) {
	log.Info("Creating ExecutorFileStorage with basePath: %s", basePath)
	baseStorage, err := types.NewFileStorage(basePath)
	if err != nil {
		log.Error("Failed to create base FileStorage: %v", err)
		return nil, err
	}

	executorStorage := &ExecutorFileStorage{
		FileStorage: baseStorage,
	}

	// Initialize executor-specific paths
	executorStorage.executedProgramsPath = filepath.Join(baseStorage.BasePath, "history")
	log.Debug("Creating executed programs directory: %s", executorStorage.executedProgramsPath)
	err = os.MkdirAll(executorStorage.executedProgramsPath, os.ModePerm)
	if err != nil {
		log.Error("Failed to create executed programs directory: %v", err)
		return nil, err
	}

	executorStorage.statusPath = filepath.Join(executorStorage.executedProgramsPath, "status")
	log.Debug("Creating status directory: %s", executorStorage.statusPath)
	err = os.MkdirAll(executorStorage.statusPath, os.ModePerm)
	if err != nil {
		log.Error("Failed to create status directory: %v", err)
		return nil, err
	}

	executorStorage.logPath = filepath.Join(executorStorage.executedProgramsPath, "logs")
	log.Debug("Creating logs directory: %s", executorStorage.logPath)
	err = os.MkdirAll(executorStorage.logPath, os.ModePerm)
	if err != nil {
		log.Error("Failed to create logs directory: %v", err)
		return nil, err
	}

	executorStorage.runningPath = filepath.Join(baseStorage.BasePath, "running")
	log.Debug("Creating running directory: %s", executorStorage.runningPath)
	err = os.MkdirAll(executorStorage.runningPath, os.ModePerm)
	if err != nil {
		log.Error("Failed to create running directory: %v", err)
		return nil, err
	}

	log.Info("Successfully created ExecutorFileStorage with paths - history: %s, status: %s, logs: %s, running: %s",
		executorStorage.executedProgramsPath, executorStorage.statusPath, executorStorage.logPath, executorStorage.runningPath)
	return executorStorage, nil
}

func (storage *ExecutorFileStorage) UpdateState(name string, status types.ProgramState) error {
	log.Debug("Updating state for program '%s' to '%s'", name, status)
	filePath := filepath.Join(storage.statusPath, name+".txt")

	statusFile, err := os.Create(filePath)
	if err != nil {
		log.Error("Failed to create status file for program '%s': %v", name, err)
		return err
	}
	defer statusFile.Close()

	_, err = statusFile.WriteString(string(status))
	if err != nil {
		log.Error("Failed to write status to file for program '%s': %v", name, err)
		return err
	}
	log.Info("Successfully updated state for program '%s' to '%s'", name, status)
	return nil
}

func (storage *ExecutorFileStorage) LoadState(name string) (types.ProgramState, int64, error) {
	log.Debug("Loading state for program '%s'", name)
	statusFilePath := filepath.Join(storage.statusPath, name+".txt")

	fileStatus, err := os.Stat(statusFilePath)
	if err != nil {
		log.Debug("Status file not found for program '%s': %v", name, err)
		return types.ProgramStateUnknown, 0, err
	}

	status, err := os.ReadFile(statusFilePath)
	if err != nil {
		log.Error("Failed to read status file for program '%s': %v", name, err)
		return types.ProgramStateUnknown, 0, err
	}
	log.Debug("Successfully loaded state for program '%s': %s", name, status)
	return types.ProgramState(status), fileStatus.ModTime().Unix(), nil
}

func (storage *ExecutorFileStorage) MaybeDeleteState(name string) {
	statusFilePath := filepath.Join(storage.statusPath, name+".txt")
	os.Remove(statusFilePath)
}

func (storage *ExecutorFileStorage) MaybeDeleteExecutionLog(name string) {
	filePath := filepath.Join(storage.logPath, name+".csv")
	os.Remove(filePath)
}

func (storage *ExecutorFileStorage) GetLogPath(name string) string {
	return filepath.Join(storage.logPath, name+".csv")
}

func (storage *ExecutorFileStorage) GetRunningLogPath(name string) string {
	return filepath.Join(storage.runningPath, name+".csv")
}

func (storage *ExecutorFileStorage) ListExecutedPrograms() ([]string, error) {
	searchPath := filepath.Join(storage.executedProgramsPath, "*.json")
	return storage.ListPrograms(searchPath)
}

func (storage *ExecutorFileStorage) LoadExecutedProgram(programName string) (*types.Program, error) {
	filePath := filepath.Join(storage.executedProgramsPath, programName+".json")
	return storage.LoadProgram(filePath)
}

func (storage *ExecutorFileStorage) CreateExecutedProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.runningPath, programName+".json")

	_, err := os.Stat(filePath)
	if err == nil {
		return types.ErrProgramExists
	}
	if !os.IsNotExist(err) {
		return types.ErrProgramExists
	}

	return storage.SaveProgram(filePath, program)
}

func (storage *ExecutorFileStorage) DeleteExecutedProgram(programName string) error {
	log.Info("Deleting executed program '%s' and all associated files", programName)
	var errors []string

	// Delete the program file
	filePath := filepath.Join(storage.executedProgramsPath, programName+".json")
	if err := storage.DeleteProgram(filePath); err != nil {
		log.Error("Failed to delete program file for '%s': %v", programName, err)
		errors = append(errors, "failed to delete program file: "+err.Error())
	} else {
		log.Debug("Successfully deleted program file for '%s'", programName)
	}

	// Delete the execution log
	logFilePath := filepath.Join(storage.logPath, programName+".csv")
	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to delete execution log for '%s': %v", programName, err)
		errors = append(errors, "failed to delete execution log: "+err.Error())
	} else {
		log.Debug("Successfully deleted execution log for '%s'", programName)
	}

	// Delete the state file
	statusFilePath := filepath.Join(storage.statusPath, programName+".txt")
	if err := os.Remove(statusFilePath); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to delete state file for '%s': %v", programName, err)
		errors = append(errors, "failed to delete state file: "+err.Error())
	} else {
		log.Debug("Successfully deleted state file for '%s'", programName)
	}

	// If there were any errors, combine them into a single error
	if len(errors) > 0 {
		log.Warning("Some deletions failed for program '%s': %s", programName, strings.Join(errors, "; "))
		return fmt.Errorf("deletion errors: %s", strings.Join(errors, "; "))
	}

	log.Info("Successfully deleted all files for executed program '%s'", programName)
	return nil
}

func (storage *ExecutorFileStorage) MoveToHistory(programName string) error {
	log.Info("Moving program '%s' from running to history", programName)
	var errors []string

	// Move program JSON file
	runningProgram := filepath.Join(storage.runningPath, programName+".json")
	historyProgram := filepath.Join(storage.executedProgramsPath, programName+".json")
	if err := os.Rename(runningProgram, historyProgram); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to move program file for '%s': %v", programName, err)
		errors = append(errors, "failed to move program file: "+err.Error())
	} else if err == nil {
		log.Debug("Moved program file for '%s' to history", programName)
	}

	// Move status file
	runningStatus := filepath.Join(storage.runningPath, programName+".txt")
	historyStatus := filepath.Join(storage.statusPath, programName+".txt")
	if err := os.Rename(runningStatus, historyStatus); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to move status file for '%s': %v", programName, err)
		errors = append(errors, "failed to move status file: "+err.Error())
	} else if err == nil {
		log.Debug("Moved status file for '%s' to history", programName)
	}

	// Move execution log
	runningLog := filepath.Join(storage.runningPath, programName+".csv")
	historyLog := filepath.Join(storage.logPath, programName+".csv")
	if err := os.Rename(runningLog, historyLog); err != nil && !os.IsNotExist(err) {
		log.Error("Failed to move execution log for '%s': %v", programName, err)
		errors = append(errors, "failed to move execution log: "+err.Error())
	} else if err == nil {
		log.Debug("Moved execution log for '%s' to history", programName)
	}

	if len(errors) > 0 {
		log.Warning("Some file moves failed for program '%s': %s", programName, strings.Join(errors, "; "))
		return fmt.Errorf("move errors: %s", strings.Join(errors, "; "))
	}

	log.Info("Successfully moved all files for program '%s' to history", programName)
	return nil
}

func (storage *ExecutorFileStorage) ListRunningPrograms() ([]string, error) {
	searchPath := filepath.Join(storage.runningPath, "*.json")
	return storage.ListPrograms(searchPath)
}

func (storage *ExecutorFileStorage) CleanupOrphanedRunning() error {
	log.Info("Checking for orphaned running programs")
	runningPrograms, err := storage.ListRunningPrograms()
	if err != nil {
		log.Error("Failed to list running programs: %v", err)
		return err
	}

	if len(runningPrograms) == 0 {
		log.Debug("No orphaned running programs found")
		return nil
	}

	log.Warning("Found %d orphaned running program(s), moving to history with 'canceled' status", len(runningPrograms))
	for _, programName := range runningPrograms {
		// Update status to canceled before moving
		statusPath := filepath.Join(storage.runningPath, programName+".txt")
		if err := os.WriteFile(statusPath, []byte(types.ProgramStateCanceled), 0644); err != nil {
			log.Warning("Failed to update status for orphaned program '%s': %v", programName, err)
		}

		if err := storage.MoveToHistory(programName); err != nil {
			log.Error("Failed to move orphaned program '%s' to history: %v", programName, err)
		}
	}

	return nil
}
