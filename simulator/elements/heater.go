package elements

import (
	"sync"
)

type (
	Heater struct {
		*Power
		mutex       sync.RWMutex
		temperature float32
		minTemp     float32
		wood        *Wood
	}
)

func NewHeater(name string, minTemp float32, material *Wood) *Heater {
	h := Heater{Power: NewPower(name), temperature: minTemp, minTemp: minTemp, wood: material}
	return &h
}

func (h *Heater) Tick() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.wood.AmbientTemperature(h.temperature)
	h.power.Tick()

	_, isOn := h.power.Info()
	if isOn {
		h.temperature += 0.1
		return
	}
	h.temperature = max(h.minTemp, h.temperature-0.01)
}

// Implement TemperatureSensor interface.
func (h *Heater) Temperature() float32 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.temperature
}
