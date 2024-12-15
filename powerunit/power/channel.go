package power

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rmkhl/halko/powerunit/shelly"
)

type (
	channel struct {
		m            sync.RWMutex
		s            ShellyInterface
		shellyID     shelly.ID
		cycleUpdated chan struct{}
		cancel       chan struct{}
		currentCycle float32
		nextCycle    float32
	}
)

const (
	floatingMinute  = float32(time.Minute)
	timeoutDuration = 5 * time.Minute
)

func newChannel(shelly ShellyInterface, shellyID shelly.ID) *channel {
	return &channel{sync.RWMutex{}, shelly, shellyID, make(chan struct{}), make(chan struct{}), 0, 0}
}

func (c *channel) Start() error {
	go c.timeoutHandler()

	for {
		c.m.RLock()
		onTime := time.Duration(c.currentCycle * floatingMinute)
		offTime := time.Duration((1 - c.currentCycle) * floatingMinute)
		c.m.RUnlock()

		if onTime > 0 {
			on := time.NewTimer(time.Duration(onTime))
			c.s.SetState(shelly.On, c.shellyID)
			err := c.handleTimeout(on)
			if err != nil {
				return err
			}
		}
		if offTime > 0 {
			off := time.NewTimer(offTime)
			c.s.SetState(shelly.Off, c.shellyID)
			err := c.handleTimeout(off)
			if err != nil {
				return err
			}
		}

		c.m.Lock()
		c.currentCycle = c.nextCycle
		c.nextCycle = 0
		c.m.Unlock()
	}
}

func (c *channel) Stop() {
	c.cancel <- struct{}{}
}

func (c *channel) handleTimeout(t *time.Timer) error {
	select {
	case <-t.C:
		return nil
	case <-c.cancel:
		c.shutDown()
		return errors.New("cancel signal received")
	}
}

func (c *channel) shutDown() {
	c.s.SetState(shelly.Off, c.shellyID)
	c.m.Lock()
	defer c.m.Unlock()
	c.currentCycle = 0
	c.nextCycle = 0
}

func (c *channel) UpdateCycle(cycle uint8) error {
	ratio := float32(cycle) / float32(100)
	if ratio > 1 || ratio < 0 {
		return fmt.Errorf("cycle percentage %v out of bounds", cycle)
	}
	c.m.Lock()
	defer c.m.Unlock()
	if c.currentCycle == 0 {
		c.currentCycle = ratio
	} else {
		c.nextCycle = ratio
	}
	return nil
}

func (c *channel) timeoutHandler() {
	timeout := time.NewTimer(timeoutDuration)

	for {
		select {
		case <-timeout.C:
			c.cancel <- struct{}{}
			return
		case <-c.cycleUpdated:
			timeout.Reset(timeoutDuration)
		}
	}
}
