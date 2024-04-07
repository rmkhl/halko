package filesystem

import (
	"encoding/json"

	"github.com/rmkhl/halko/configurator/domain"
)

type (
	cycle struct {
		*domain.Cycle
	}

	cycles struct{}
)

func (c *cycles) ByID(id string) (*domain.Cycle, error) {
	cycle, err := byID(id, new(cycle))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCast[domain.Cycle](cycle)
}

func (c *cycles) All() ([]*domain.Cycle, error) {
	cycles, err := all(new(cycle))
	if err != nil {
		return nil, transformError(err)
	}
	return runtimeCastSlice[domain.Cycle](cycles)
}

func (c *cycles) CreateOrUpdate(_ *domain.Cycle) (*domain.Cycle, error) {
	return nil, nil
}

func (c *cycle) id() string {
	return string(c.ID)
}

func (c *cycle) setID(id string) {
	c.ID = domain.ID(id)
}

func (c *cycle) unmarshalJSON(data []byte) (any, error) {
	var cycle domain.Cycle

	if err := json.Unmarshal(data, &cycle); err != nil {
		return nil, err
	}

	return &cycle, nil
}
