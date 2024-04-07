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

func (p *phases) ByID(id string) (*domain.Phase, error) {
	phase, err := byID(id, new(phase))
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

func (p *phases) CreateOrUpdate(pp *domain.Phase) (*domain.Phase, error) {
	return nil, nil
}

func (p *phase) id() string {
	return string(p.ID)
}

func (p *phase) setID(id string) {
	p.ID = domain.ID(id)
}

func (p *phase) unmarshalJSON(data []byte) (any, error) {
	var phase domain.Phase

	if err := json.Unmarshal(data, &phase); err != nil {
		return nil, err
	}

	return &phase, nil
}
