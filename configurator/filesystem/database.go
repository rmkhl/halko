package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rmkhl/halko/configurator/database"
	"github.com/rmkhl/halko/configurator/domain"
	"github.com/rmkhl/halko/configurator/utils"
)

const (
	basePath = "/fsdb"
)

type entity string

const (
	eCycles   entity = "cycles"
	ePhases   entity = "phases"
	ePrograms entity = "programs"
	eError    entity = "error"
)

func New() *database.Interface {
	return &database.Interface{
		Cycles:   &cycles{},
		Phases:   &phases{},
		Programs: &programs{},
	}
}

func byID(id string, o Object) (any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}
	return readFromFile(o, filenameByID(e, id))
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

func save(o Object) (any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}

	id := resolveAndUpdateID(o)

	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(filenameByID(e, id), data, 0664); err != nil {
		return nil, err
	}

	return o, nil
}

func resolveAndUpdateID(o Object) string {
	id := o.id()
	if !domain.ID(id).IsValid() {
		id = strconv.FormatInt(time.Now().Unix(), 10)
		o.setID(id)
	}
	return id
}

func filenameByID(entity entity, id string) string {
	return fmt.Sprintf("%s/%s.json", pathByEntity(entity), id)
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
	case *domain.Cycle, *cycle:
		return eCycles, true
	case *domain.Phase, *phase:
		return ePhases, true
	case *domain.Program, *program:
		return ePrograms, true
	default:
		return eError, false
	}
}

type Object interface {
	id() string
	setID(id string)
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
