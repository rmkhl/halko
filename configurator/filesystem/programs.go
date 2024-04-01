package filesystem

import "github.com/rmkhl/halko/configurator/domain"

type Programs struct{}

func (p *Programs) Current() (*domain.Program, error) {
	return nil, nil
}

func (p *Programs) ByID(id string) (*domain.Program, error) {
	return byID(programs, id, parseProgram)
}

func (p *Programs) All() ([]*domain.Program, error) {
	return all(programs, parseProgram)
}

func parseProgram(data []byte) (*domain.Program, error) {
	return nil, nil
}
