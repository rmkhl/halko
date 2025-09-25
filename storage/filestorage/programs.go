package filestorage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/types"
)

var (
	ErrProgramExists       = errors.New("program exists")
	ErrProgramDoesNotExist = errors.New("program does not exist")
)

type FileStorage struct {
	basePath             string
	executedProgramsPath string
	statusPath           string
	logPath              string
	programPath          string
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	storage := FileStorage{basePath: basePath}

	storage.executedProgramsPath = filepath.Join(storage.basePath, "history")
	err := os.MkdirAll(storage.executedProgramsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	storage.programPath = filepath.Join(storage.basePath, "programs")
	err = os.MkdirAll(storage.programPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	storage.statusPath = filepath.Join(storage.executedProgramsPath, "status")
	err = os.MkdirAll(storage.statusPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	storage.logPath = filepath.Join(storage.executedProgramsPath, "logs")
	err = os.MkdirAll(storage.logPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func (storage *FileStorage) ListExecutedPrograms() ([]string, error) {
	searchPath := filepath.Join(storage.executedProgramsPath, "*.json")
	return storage.listPrograms(searchPath)
}

func (storage *FileStorage) listPrograms(searchPath string) ([]string, error) {
	programs := []string{}

	files, err := filepath.Glob(searchPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		programs = append(programs, fileName[:len(fileName)-5])
	}

	return programs, nil
}

func (storage *FileStorage) LoadExecutedProgram(programName string) (*types.Program, error) {
	filePath := filepath.Join(storage.executedProgramsPath, programName+".json")
	return storage.loadProgram(filePath)
}

func (storage *FileStorage) loadProgram(filePath string) (*types.Program, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	content, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var program types.Program
	err = json.Unmarshal(content, &program)
	if err != nil {
		return nil, err
	}

	return &program, nil
}

func (storage *FileStorage) saveProgram(filePath string, program *types.Program) error {
	jsonFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	content, err := json.Marshal(program)
	if err != nil {
		return err
	}
	_, err = jsonFile.Write(content)
	return err
}

func (storage *FileStorage) CreateExecutedProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.executedProgramsPath, programName+".json")

	_, err := os.Stat(filePath)
	if err == nil {
		return ErrProgramExists
	}
	if !errors.Is(err, os.ErrNotExist) {
		return ErrProgramExists
	}

	return storage.saveProgram(filePath, program)
}

func (storage *FileStorage) DeleteExecutedProgram(programName string) error {
	filePath := filepath.Join(storage.executedProgramsPath, programName+".json")

	_, err := os.Stat(filePath)
	if err != nil {
		return ErrProgramDoesNotExist
	}

	return os.Remove(filePath)
}

func (storage *FileStorage) ListStoredPrograms() ([]string, error) {
	searchPath := filepath.Join(storage.programPath, "*.json")
	return storage.listPrograms(searchPath)
}

func (storage *FileStorage) LoadStoredProgram(programName string) (*types.Program, error) {
	filePath := filepath.Join(storage.programPath, programName+".json")
	return storage.loadProgram(filePath)
}

func (storage *FileStorage) SaveStoredProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.programPath, programName+".json")
	return storage.saveProgram(filePath, program)
}

func (storage *FileStorage) CreateStoredProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.programPath, programName+".json")

	_, err := os.Stat(filePath)
	if err == nil {
		return ErrProgramExists
	}
	if !errors.Is(err, os.ErrNotExist) {
		return ErrProgramExists
	}

	return storage.saveProgram(filePath, program)
}

func (storage *FileStorage) UpdateStoredProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.programPath, programName+".json")

	_, err := os.Stat(filePath)
	if err != nil {
		return ErrProgramDoesNotExist
	}

	return storage.saveProgram(filePath, program)
}

func (storage *FileStorage) DeleteStoredProgram(programName string) error {
	filePath := filepath.Join(storage.programPath, programName+".json")

	_, err := os.Stat(filePath)
	if err != nil {
		return ErrProgramDoesNotExist
	}

	return os.Remove(filePath)
}
