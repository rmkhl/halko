package engine

import (
	"sync"
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

	p.running = true
	p.current = initialState
	p.upcoming = false
	p.tick = 0
}

func (p *Power) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.running = false
	p.current = false
	p.upcoming = false
	p.tick = 0
}

func (p *Power) SwitchTo(upcoming bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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
	} else {
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
