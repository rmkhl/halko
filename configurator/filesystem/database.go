package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/domain"
	"github.com/rmkhl/halko/configurator/utils"
)

const (
	basePath = "/fsdb"
)

type entity string

const (
	cycles   entity = "cycles"
	phases   entity = "phases"
	programs entity = "programs"
)

func New() *database.Interface {
	return &database.Interface{
		Cycles:   &Cycles{},
		Phases:   &Phases{},
		Programs: &Programs{},
	}
}

func byID[T *domain.Program | *domain.Cycle | *domain.Phase](entity entity, id string, parseFn func([]byte) (T, error)) (T, error) {
	b, err := os.ReadFile(filenameByID(entity, id))
	if err != nil {
		return nil, err
	}
	return parseFn(b)
}

func all[T *domain.Program | *domain.Cycle | *domain.Phase](entity entity, parseFn func([]byte) (T, error)) ([]T, error) {
	filenames, err := filenamesByEntity(entity)
	if err != nil {
		return nil, err
	}
	data := make([]T, 0, len(filenames))
	for _, fn := range filenames {
		item, err := byID(entity, fn, parseFn)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, nil
}

func filenameByID(entity entity, id string) string {
	return fmt.Sprintf("%s/%s/%s", basePath, entity, id)
}

func filenamesByEntity(entity entity) ([]string, error) {
	dirName := fmt.Sprintf("%s/%s", basePath, entity)
	dir, err := os.Open(dirName)
	if err != nil {
		return nil, err
	}
	entries, err := dir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	files := utils.Filter(
		entries,
		func(item os.DirEntry) bool {
			return !item.IsDir()
		},
	)
	return utils.Map(files, func(f fs.DirEntry) string {
		return f.Name()
	}), nil
}

func transformError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return database.ErrNotFound

	}
	return err
}
