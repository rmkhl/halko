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
	p := Power{name: name, power: engine.NewPower(name)}
	return &p
}

func (h *Power) TurnOn(initialState bool) {
	h.power.Start(initialState)
}

func (h *Power) TurnOff() {
	h.power.Stop()
}

func (h *Power) SwitchTo(upcoming bool) {
	h.power.SwitchTo(upcoming)
}

func (h *Power) IsOn() bool {
	return h.power.IsRunning()
}

func (h *Power) Name() string {
	return h.name
}

func (h *Power) CurrentCycle() uint8 {
	running, turnedOn := h.power.Info()

	if running && turnedOn {
		return 1
	}

	return 0
}

func (h *Power) Info() (bool, bool) {
	return h.power.Info()
}

func (h *Power) Tick() {
	h.power.Tick()
}
