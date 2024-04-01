package filesystem

import "github.com/rmkhl/halko/configurator/domain"

type Programs struct{}

func (db Programs) Current() (*domain.Program, error) {
	return nil, nil
}
