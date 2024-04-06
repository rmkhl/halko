package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type Phases struct{}

func (p *Phases) ByID(id string) (*domain.Phase, error) {
	phase, err := byID(phases, id, parsePhase)
	return phase, transformError(err)
}

func (p *Phases) All() ([]*domain.Phase, error) {
	phases, err := all(phases, parsePhase)
	return phases, transformError(err)
}

func parsePhase(data []byte) (*domain.Phase, error) {
	var phase domain.Phase

	if err := json.Unmarshal(data, &phase); err != nil {
		return nil, err
	}

	return &phase, nil
}
