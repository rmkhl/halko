package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/rmkhl/halko/executor/types"
)

var (
	ErrProgramExists       = errors.New("program exists")
	ErrProgramDoesNotExist = errors.New("program does not exist")
)

type ProgramStorage struct {
	basePath     string
	programsPath string
}

func NewProgramStorage(basePath string) (*ProgramStorage, error) {
	storage := ProgramStorage{basePath: basePath}

	storage.programsPath = filepath.Join(storage.basePath, "programs")
	err := os.MkdirAll(storage.programsPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func (storage *ProgramStorage) ListPrograms() ([]string, error) {
	programs := []string{}

	files, err := filepath.Glob(filepath.Join(storage.programsPath, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		programs = append(programs, fileName[:len(fileName)-5])
	}

	return programs, nil
}

func (storage *ProgramStorage) LoadProgram(programName string) (*types.Program, error) {
	filePath := filepath.Join(storage.programsPath, programName+".json")

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

func (storage *ProgramStorage) saveProgram(filePath string, program *types.Program) error {
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

func (storage *ProgramStorage) CreateProgram(programName string, program *types.Program) error {
	filePath := filepath.Join(storage.programsPath, programName+".json")

	_, err := os.Stat(filePath)
	if err == nil {
		return ErrProgramExists
	}
	if !errors.Is(err, os.ErrNotExist) {
		return ErrProgramExists
	}

	return storage.saveProgram(filePath, program)
}

func (storage *ProgramStorage) DeleteProgram(programName string) error {
	filePath := filepath.Join(storage.programsPath, programName+".json")

	_, err := os.Stat(filePath)
	if err != nil {
		return ErrProgramDoesNotExist
	}

	return os.Remove(filePath)
}
