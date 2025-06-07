package power

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rmkhl/halko/powerunit/shelly"
)

type (
	ShellyInterface interface {
		GetState(id shelly.ID) (shelly.PowerState, error)
		SetState(state shelly.PowerState, id shelly.ID) (shelly.PowerState, error)
	}

	Controller struct {
		s          ShellyInterface
		fan        *channel
		heater     *channel
		humidifier *channel
		errChan    chan error
	}

	States map[shelly.ID]shelly.PowerState
)

func New(s ShellyInterface) *Controller {
	errChan := make(chan error, 3)
	return &Controller{s, newChannel(s, shelly.Fan, errChan), newChannel(s, shelly.Heater, errChan), newChannel(s, shelly.Humidifier, errChan), errChan}
}

func worker(ctx context.Context, channel *channel, wg *sync.WaitGroup) {
	defer wg.Done()
	go channel.Start()
	<-ctx.Done()
	channel.Stop()
}

func (c *Controller) Stop() {
	c.errChan <- fmt.Errorf("controller stopped")
}

func (c *Controller) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

	go worker(ctx, c.fan, &wg)
	go worker(ctx, c.heater, &wg)
	go worker(ctx, c.humidifier, &wg)

	go func() {
		err := <-c.errChan
		log.Printf("%s", err)
		cancel()
	}()

	wg.Wait()
	log.Printf("channels stopped")
	return nil
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
		ch = c.humidifier
	default:
		return fmt.Errorf("unknown channel id %d", id)
	}
	return ch.UpdateCycle(percentage)
}
