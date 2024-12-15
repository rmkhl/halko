package power

import (
	"context"
	"fmt"

	"github.com/rmkhl/halko/powerunit/shelly"
	"golang.org/x/sync/errgroup"
)

type (
	ShellyInterface interface {
		GetState(id shelly.ID) (shelly.PowerState, error)
		SetState(state shelly.PowerState, id shelly.ID) (shelly.PowerState, error)
	}

	Controller struct {
		s         ShellyInterface
		fan       *channel
		heater    *channel
		humidifer *channel
	}

	States map[shelly.ID]shelly.PowerState
)

func New(s ShellyInterface) *Controller {
	return &Controller{s, newChannel(s, shelly.Fan), newChannel(s, shelly.Heater), newChannel(s, shelly.Humidifier)}
}

func (c *Controller) Start() error {
	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			c.fan.Stop()
			return ctx.Err()
		default:
			return c.fan.Start()
		}
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			c.heater.Stop()
			return ctx.Err()
		default:
			return c.heater.Start()
		}
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			c.humidifer.Stop()
			return ctx.Err()
		default:
			return c.humidifer.Start()
		}
	})

	return eg.Wait()
}

func (c *Controller) GetState(id shelly.ID) (shelly.PowerState, error) {
	return c.s.GetState(id)
}

func (c *Controller) GetAllStates() (States, error) {
	states := States{}
	for _, id := range []shelly.ID{shelly.Fan, shelly.Heater, shelly.Humidifier} {
		state, err := c.s.GetState(id)
		if err != nil {
			return nil, err
		}
		states[id] = state
	}
	return states, nil
}

func (c *Controller) SetCycle(percentage uint8, id shelly.ID) error {
	var ch *channel
	switch id {
	case shelly.Fan:
		ch = c.fan
	case shelly.Heater:
		ch = c.heater
	case shelly.Humidifier:
		ch = c.humidifer
	default:
		return fmt.Errorf("unknown channel id %d", id)
	}
	return ch.UpdateCycle(percentage)
}
