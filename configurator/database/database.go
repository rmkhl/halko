package database

import "github.com/rmkhl/halko/configurator/domain"

type Interface struct {
	Programs Programs
}

type Programs interface {
	Current() (*domain.Program, error)
}
