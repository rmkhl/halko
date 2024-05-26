package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type (
	program struct {
		*domain.Program
	}

	programs struct{}
)

func (p *programs) ByName(name string) (*domain.Program, error) {
	prog, err := byName(name, new(program))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCast[domain.Program](prog)
}

func (p *programs) All() ([]*domain.Program, error) {
	progs, err := all(new(program))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCastSlice[domain.Program](progs)
}

func (p *programs) CreateOrUpdate(pp *domain.Program) (*domain.Program, error) {
	ppp, err := save(&program{pp})
	if err != nil {
		return nil, transformError(err)
	}
	cast, err := runtimeCast[program](ppp)
	if err != nil {
		return nil, transformError(err)
	}
	return cast.Program, nil
}

func (p *program) name() string {
	return string(p.Name)
}

func (p *program) setName(name string) {
	p.Name = domain.Name(name)
}

func (p *program) unmarshalJSON(data []byte) (any, error) {
	var prog domain.Program

	if err := json.Unmarshal(data, &prog); err != nil {
		return nil, err
	}

	return &prog, nil
}
