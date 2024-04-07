package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
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
	b, err := os.ReadFile(filenameByID(e, id))
	if err != nil {
		return nil, err
	}
	return o.unmarshalJSON(b)
}

func all(o Object) ([]any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}
	filenames, err := filenamesByEntity(e)
	if err != nil {
		return nil, err
	}
	data := make([]any, 0, len(filenames))
	for _, fn := range filenames {
		item, err := byID(fn, o)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	return data, nil
}

func save(o Object) (any, error) {
	e, ok := entityForType(o)
	if !ok {
		return nil, database.ErrInvalidInput
	}

	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	id := resolveAndUpdateID(o)
	if err = os.WriteFile(filenameByID(e, id), data, 0664); err != nil {
		return nil, err
	}

	return o, nil
}

func resolveAndUpdateID(o Object) string {
	id := o.id()
	if !domain.ID(o.id()).IsValid() {
		id = strconv.FormatInt(time.Now().Unix(), 10)
		o.setID(id)
	}
	return id
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
