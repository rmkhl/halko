package database

import "github.com/rmkhl/halko/configurator/domain"

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
