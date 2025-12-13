package elements

import (
	"sync"

	"github.com/rmkhl/halko/types/log"
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

func NewHeater(name string, initialTemp float32, envTemp float32, material *Wood) *Heater {
	h := Heater{Power: NewPower(name), temperature: initialTemp, minTemp: envTemp, wood: material}
	return &h
}

func (h *Heater) Tick() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.wood.AmbientTemperature(h.temperature)
	h.power.Tick()

	oldTemp := h.temperature
	_, isOn := h.power.Info()
	if isOn {
		h.temperature += 0.1
		log.Trace("Heater '%s': Heating - %.1f°C -> %.1f°C", h.Name(), oldTemp, h.temperature)
		return
	}
	h.temperature = max(h.minTemp, h.temperature-0.01)
	if h.temperature != oldTemp {
		log.Trace("Heater '%s': Cooling - %.1f°C -> %.1f°C (min: %.1f°C)", h.Name(), oldTemp, h.temperature, h.minTemp)
	}
}

func (h *Heater) Temperature() float32 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.temperature
}
