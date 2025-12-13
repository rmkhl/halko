package engine

import (
	"sync"

	"github.com/rmkhl/halko/types/log"
)

type (
	Power struct {
		name     string
		mutex    sync.RWMutex
		tick     int
		running  bool
		current  bool
		upcoming bool
	}
)

func NewPower(name string) *Power {
	p := Power{name: name, tick: 0, current: false, upcoming: false, running: false}
	return &p
}

func (p *Power) Start(initialState bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Debug("Power[%s]: Starting with initial state: %v", p.name, initialState)
	p.running = true
	p.current = initialState
	p.upcoming = false
	p.tick = 0
}

func (p *Power) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Debug("Power[%s]: Stopping (was running: %v, current: %v)", p.name, p.running, p.current)
	p.running = false
	p.current = false
	p.upcoming = false
	p.tick = 0
}

func (p *Power) SwitchTo(upcoming bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.upcoming != upcoming {
		log.Debug("Power[%s]: Switching upcoming state from %v to %v (current: %v)", p.name, p.upcoming, upcoming, p.current)
	}
	p.upcoming = upcoming
}

func (p *Power) Tick() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Not currently running, nothing to advance
	if !p.running {
		log.Trace("Power[%s]: Not running, skipping tick", p.name)
		p.current = false
		p.tick = 0
		return
	}

	if p.tick < 9 {
		p.tick++
		log.Trace("Power[%s]: Tick %d/10, current: %v, upcoming: %v", p.name, p.tick, p.current, p.upcoming)
	} else {
		if p.current != p.upcoming {
			log.Info("Power[%s]: State transition at tick 10 - %v -> %v", p.name, p.current, p.upcoming)
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
