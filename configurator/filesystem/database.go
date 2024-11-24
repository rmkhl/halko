package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/utils"
	"github.com/rmkhl/halko/types"
)

const (
	basePath = "/app/fsdb"
)

type entity string

const (
	ePrograms entity = "programs"
	eError    entity = "error"
)

func New() *database.Interface {
	return &database.Interface{
		Programs: &programs{},
	}
}

func byName(name string, o Object) (any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}
	return readFromFile(o, filenameByName(e, name))
}

func all(o Object) ([]any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}
	filenames, err := filepathsByEntity(e)
	if err != nil {
		return nil, err
	}
	data := make([]any, 0, len(filenames))
	for _, fn := range filenames {
		item, err := readFromFile(o, fn)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, nil
}

func readFromFile(o Object, filepath string) (any, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return o.unmarshalJSON(b)
}

func save(name string, o Object) (any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}

	if name != "" {
		if err := deleteFile(name, e); err != nil {
			return nil, err
		}
	}

	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	fn := filenameByName(e, o.name())
	if err = os.WriteFile(fn, data, 0664); err != nil {
		return nil, err
	}

	return o, nil
}

func deleteFile(name string, e entity) error {
	filename := filenameByName(e, name)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func filenameByName(entity entity, name string) string {
	return fmt.Sprintf("%s/%s.json", pathByEntity(entity), name)
}

func pathByEntity(entity entity) string {
	return fmt.Sprintf("%s/%s", basePath, entity)
}

func filepathsByEntity(entity entity) ([]string, error) {
	dirName := pathByEntity(entity)
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
			return !item.IsDir() && strings.HasSuffix(item.Name(), ".json")
		},
	)
	return utils.Map(files, func(f fs.DirEntry) string {
		return fmt.Sprintf("%s/%s", dirName, f.Name())
	}), nil
}

func transformError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return database.ErrNotFound
	}
	return err
}

func entityForType(t any) (entity, bool) {
	switch t.(type) {
	case *types.Program, *program:
		return ePrograms, true
	default:
		return eError, false
	}
}

type Object interface {
	name() string
	setName(name string)
	unmarshalJSON(data []byte) (any, error)
}

func runtimeCast[T any](data any) (*T, error) {
	cast, ok := data.(*T)
	if !ok {
		return nil, database.ErrUnexpectedReturnValue
	}
	return cast, nil
}

func runtimeCastSlice[T any](data []any) ([]*T, error) {
	cast := make([]*T, 0, len(data))
	for _, d := range data {
		v, ok := d.(*T)
		if !ok {
			return nil, database.ErrUnexpectedReturnValue
		}
		cast = append(cast, v)
	}
	return cast, nil
}
