package storagefs

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type ProgramStorage struct {
	*types.FileStorage
	programPath string
}

func NewProgramStorage(basePath string) (*ProgramStorage, error) {
	log.Info("Creating ProgramStorage with basePath: %s", basePath)
	baseStorage, err := types.NewFileStorage(basePath)
	if err != nil {
		log.Error("Failed to create base FileStorage: %v", err)
		return nil, err
	}

	programStorage := &ProgramStorage{
		FileStorage: baseStorage,
	}

	programStorage.programPath = filepath.Join(baseStorage.BasePath, "programs")
	log.Debug("Creating programs directory: %s", programStorage.programPath)
	err = os.MkdirAll(programStorage.programPath, os.ModePerm)
	if err != nil {
		log.Error("Failed to create programs directory: %v", err)
		return nil, err
	}

	log.Info("Successfully created ProgramStorage with programs path: %s", programStorage.programPath)
	return programStorage, nil
}

func (storage *ProgramStorage) ListStoredPrograms() ([]string, error) {
	searchPath := filepath.Join(storage.programPath, "*.json")
	return storage.ListPrograms(searchPath)
}

func (storage *ProgramStorage) LoadStoredProgram(programName string) (*types.Program, error) {
	log.Debug("Loading stored program: %s", programName)
	filePath := filepath.Join(storage.programPath, programName+".json")
	program, err := storage.LoadProgram(filePath)
	if err != nil {
		log.Error("Failed to load stored program '%s': %v", programName, err)
		return nil, err
	}
	log.Debug("Successfully loaded stored program: %s", programName)
	return program, nil
}

func (storage *ProgramStorage) SaveStoredProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.programPath, programName+".json")
	return storage.SaveProgram(filePath, program)
}

func (storage *ProgramStorage) CreateStoredProgram(programName string, program *types.Program) error {
	log.Info("Creating stored program: %s", programName)
	filePath := filepath.Join(storage.programPath, programName+".json")

	_, err := os.Stat(filePath)
	if err == nil {
		log.Warning("Program '%s' already exists", programName)
		return types.ErrProgramExists
	}
	if !errors.Is(err, os.ErrNotExist) {
		log.Error("Error checking if program '%s' exists: %v", programName, err)
		return types.ErrProgramExists
	}

	err = storage.SaveProgram(filePath, program)
	if err != nil {
		log.Error("Failed to create stored program '%s': %v", programName, err)
		return err
	}
	log.Info("Successfully created stored program: %s", programName)
	return nil
}

func (storage *ProgramStorage) UpdateStoredProgram(programName string, program *types.Program) error {
	log.Info("Updating stored program: %s", programName)
	filePath := filepath.Join(storage.programPath, programName+".json")

	_, err := os.Stat(filePath)
	if err != nil {
		log.Warning("Program '%s' does not exist for update", programName)
		return types.ErrProgramDoesNotExist
	}

	err = storage.SaveProgram(filePath, program)
	if err != nil {
		log.Error("Failed to update stored program '%s': %v", programName, err)
		return err
	}
	log.Info("Successfully updated stored program: %s", programName)
	return nil
}

func (storage *ProgramStorage) DeleteStoredProgram(programName string) error {
	log.Info("Deleting stored program: %s", programName)
	filePath := filepath.Join(storage.programPath, programName+".json")
	err := storage.DeleteProgram(filePath)
	if err != nil {
		log.Error("Failed to delete stored program '%s': %v", programName, err)
		return err
	}
	log.Info("Successfully deleted stored program: %s", programName)
	return nil
}
