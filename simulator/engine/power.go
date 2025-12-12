package engine

import (
	"sync"

	"github.com/rmkhl/halko/types/log"
)

type (
	Power struct {
		mutex    sync.RWMutex
		tick     int
		running  bool
		current  bool
		upcoming bool
	}
)

func NewPower() *Power {
	p := Power{tick: 0, current: false, upcoming: false, running: false}
	return &p
}

func (p *Power) Start(initialState bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Debug("Power: Starting with initial state: %v", initialState)
	p.running = true
	p.current = initialState
	p.upcoming = false
	p.tick = 0
}

func (p *Power) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Debug("Power: Stopping (was running: %v, current: %v)", p.running, p.current)
	p.running = false
	p.current = false
	p.upcoming = false
	p.tick = 0
}

func (p *Power) SwitchTo(upcoming bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.upcoming != upcoming {
		log.Debug("Power: Switching upcoming state from %v to %v (current: %v)", p.upcoming, upcoming, p.current)
	}
	p.upcoming = upcoming
}

func (p *Power) Tick() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Not currently running, nothing to advance
	if !p.running {
		p.current = false
		p.tick = 0
		return
	}

	if p.tick < 9 {
		p.tick++
		log.Trace("Power: Tick %d/10, current: %v, upcoming: %v", p.tick, p.current, p.upcoming)
	} else {
		if p.current != p.upcoming {
			log.Info("Power: State transition at tick 10 - %v -> %v", p.current, p.upcoming)
		}
		p.current = p.upcoming
		p.tick = 0
	}
}

func (p *Power) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

func (p *Power) Info() (running bool, tick bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.running, p.current
}
