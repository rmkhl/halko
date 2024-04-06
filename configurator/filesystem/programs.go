package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type Programs struct{}

func (p *Programs) Current() (*domain.Program, error) {
	prog, err := byID(programs, "current", parseProgram)
	return prog, transformError(err)
}

func (p *Programs) ByID(id string) (*domain.Program, error) {
	prog, err := byID(programs, id, parseProgram)
	return prog, transformError(err)
}

func (p *Programs) All() ([]*domain.Program, error) {
	progs, err := all(programs, parseProgram)
	return progs, transformError(err)
}

func parseProgram(data []byte) (*domain.Program, error) {
	var prog domain.Program

	if err := json.Unmarshal(data, &prog); err != nil {
		return nil, err
	}

	return &prog, nil
}
