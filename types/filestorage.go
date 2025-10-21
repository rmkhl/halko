package types

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrProgramExists       = errors.New("program exists")
	ErrProgramDoesNotExist = errors.New("program does not exist")
)

type FileStorage struct {
	BasePath string
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	return &FileStorage{BasePath: basePath}, nil
}

func (storage *FileStorage) ListPrograms(searchPath string) ([]string, error) {
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

func (storage *FileStorage) LoadProgram(filePath string) (*Program, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	content, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var program Program
	err = json.Unmarshal(content, &program)
	if err != nil {
		return nil, err
	}

	return &program, nil
}

func (storage *FileStorage) SaveProgram(filePath string, program *Program) error {
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

func (storage *FileStorage) DeleteProgram(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return ErrProgramDoesNotExist
	}

	return os.Remove(filePath)
}
