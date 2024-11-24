package elements

import (
	"github.com/rmkhl/halko/simulator/engine"
)

type (
	Power struct {
		name  string
		power *engine.Power
	}
)

func NewPower(name string) *Power {
	p := Power{name: name, power: new(engine.Power)}
	return &p
}

// Implement PowerManager interface.
func (h *Power) TurnOn(cycle *engine.Cycle) {
	h.power.Start(cycle)
}

func (h *Power) TurnOff() {
	h.power.Stop()
}

func (h *Power) SwitchTo(cycle *engine.Cycle) {
	h.power.UpdateCycle(cycle)
}

// Implement PowerSensor interface.
func (h *Power) IsOn() bool {
	return h.power.IsRunning()
}

func (h *Power) Name() string {
	return h.name
}

func (h *Power) CurrentCycle() uint8 {
	percentage, _ := h.power.CycleInfo()

	return percentage
}

func (h *Power) Tick() {
	h.power.Tick()
}
