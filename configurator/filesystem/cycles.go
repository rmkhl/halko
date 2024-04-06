package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type Cycles struct{}

func (c *Cycles) ByID(id string) (*domain.Cycle, error) {
	cycle, err := byID(cycles, id, parseCycle)
	return cycle, transformError(err)
}

func (c *Cycles) All() ([]*domain.Cycle, error) {
	cycles, err := all(cycles, parseCycle)
	return cycles, transformError(err)
}

func parseCycle(data []byte) (*domain.Cycle, error) {
	var cycle domain.Cycle

	if err := json.Unmarshal(data, &cycle); err != nil {
		return nil, err
	}

	return &cycle, nil
}
