package database

import (
	"errors"

	"github.com/rmkhl/halko/types"
)

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrNotFound              = errors.New("not found")
	ErrUnexpectedReturnValue = errors.New("unexpected return value")
)

type Interface struct {
	Programs Programs
}

type Entity[T any] interface {
	All() ([]T, error)
	ByName(name string) (T, error)
	CreateOrUpdate(string, T) (T, error)
}

type Programs interface {
	Entity[*types.Program]
}
