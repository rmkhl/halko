package database

import (
	"errors"

	"github.com/rmkhl/halko/configurator/domain"
)

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrNotFound              = errors.New("not found")
	ErrUnexpectedReturnValue = errors.New("unexpected return value")
)

type Interface struct {
	Phases   Phases
	Programs Programs
}

type Entity[T any] interface {
	All() ([]T, error)
	ByID(id string) (T, error)
	CreateOrUpdate(T) (T, error)
}

type Programs interface {
	Entity[*domain.Program]
}

type Phases interface {
	Entity[*domain.Phase]
}
