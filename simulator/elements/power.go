package elements

import (
	"github.com/rmkhl/halko/simulator/types"
)

type (
	Power struct {
		name  string
		power *types.Power
	}
)

func NewPower(name string) *Power {
	p := Power{name: name, power: new(types.Power)}
	return &p
}

// Implement PowerManager interface
func (h *Power) TurnOn(cycle *types.Cycle) {
	h.power.Start(cycle)
}

func (h *Power) TurnOff() {
	h.power.Stop()
}

func (h *Power) SwitchTo(cycle *types.Cycle) {
	h.power.UpdateCycle(cycle)
}

// Implement PowerSensor interface
func (h *Power) IsOn() bool {
	return h.power.IsRunning()
}

func (h *Power) Name() string {
	return h.name
}

func (h *Power) CurrentCycle() string {
	var name, _, _ = h.power.CycleInfo()

	return name
}

func (h *Power) Tick() {
	h.power.Tick()
}
