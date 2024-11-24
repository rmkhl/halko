package engine

import (
	"sync"
)

type (
	Power struct {
		mutex    sync.RWMutex
		tick     int
		current  *Cycle
		upcoming *Cycle
	}
)

func NewPower() *Power {
	p := Power{tick: 0, current: nil, upcoming: nil}
	return &p
}

func (p *Power) Start(cycle *Cycle) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.current = cycle
	p.upcoming = nil
	p.tick = 0
}

func (p *Power) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.current = nil
	p.upcoming = nil
	p.tick = 0
}

func (p *Power) UpdateCycle(cycle *Cycle) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.current == nil {
		p.current = cycle
		return
	}

	if p.upcoming == nil {
		p.upcoming = cycle
	}
}

func (p *Power) Tick() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Not currently running, nothing to advance
	if p.current == nil {
		p.tick = 0
		return
	}

	if p.tick < 9 {
		p.tick++
	} else {
		if p.upcoming != nil {
			p.current = p.upcoming
			p.upcoming = nil
		}
		p.tick = 0
	}
}

func (p *Power) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.current != nil
}

func (p *Power) CycleInfo() (percentage uint8, tick bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.current == nil {
		return 0, false
	}

	return p.current.percentage, p.current.ticks[p.tick]
}
