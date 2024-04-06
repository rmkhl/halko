package database

import (
	"errors"

	"github.com/rmkhl/halko/configurator/domain"
)

var (
	ErrNotFound = errors.New("not found")
)

type Interface struct {
	Cycles   Cycles
	Phases   Phases
	Programs Programs
}

type Fetchable[T any] interface {
	ByID(id string) (T, error)
	All() ([]T, error)
}

type Programs interface {
	Fetchable[*domain.Program]
	Current() (*domain.Program, error)
}

type Phases interface {
	Fetchable[*domain.Phase]
}

type Cycles interface {
	Fetchable[*domain.Cycle]
}
