package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type (
	phase struct {
		*domain.Phase
	}

	phases struct{}
)

func (p *phases) ByName(name string) (*domain.Phase, error) {
	phase, err := byName(name, new(phase))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCast[domain.Phase](phase)
}

func (p *phases) All() ([]*domain.Phase, error) {
	phases, err := all(new(phase))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCastSlice[domain.Phase](phases)
}

func (p *phases) CreateOrUpdate(name string, pp *domain.Phase) (*domain.Phase, error) {
	ppp, err := save(name, &phase{pp})
	if err != nil {
		return nil, transformError(err)
	}
	cast, err := runtimeCast[phase](ppp)
	if err != nil {
		return nil, transformError(err)
	}
	return cast.Phase, nil
}

func (p *phase) name() string {
	return string(p.Name)
}

func (p *phase) setName(name string) {
	p.Name = domain.Name(name)
}

func (p *phase) unmarshalJSON(data []byte) (any, error) {
	var phase domain.Phase

	if err := json.Unmarshal(data, &phase); err != nil {
		return nil, err
	}

	return &phase, nil
}
