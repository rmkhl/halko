package elements

import (
	"sync"
)

type (
	Heater struct {
		*Power
		mutex       sync.Mutex
		temperature float32
		min_temp    float32
		max_temp    float32
		wood        *Wood
	}
)

func NewHeater(name string, min_temp float32, max_temp float32, material *Wood) *Heater {
	h := Heater{Power: NewPower(name), temperature: min_temp, min_temp: min_temp, max_temp: max_temp, wood: material}
	return &h
}

func (h *Heater) Tick() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.wood.AmbientTemperature(h.temperature)
	h.power.Tick()

	_, _, is_on := h.power.CycleInfo()
	if is_on {
		h.temperature = min(h.max_temp, h.temperature+0.1)
		return
	}
	h.temperature = max(h.min_temp, h.temperature-0.01)
}

// Implement TemperatureSensor interface
func (h *Heater) Temperature() float32 {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.temperature
}
