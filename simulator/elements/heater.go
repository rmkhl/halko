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

func NewHeater(name string, initialTemp float32, envTemp float32, material *Wood) *Heater {
	h := Heater{Power: NewPower(name), temperature: initialTemp, minTemp: envTemp, wood: material}
	return &h
}

func (h *Heater) Tick() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Only advance power state machine - temperature handled by physics engine
	h.power.Tick()
}

func (h *Heater) Temperature() float32 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.temperature
}

func (h *Heater) SetTemperature(temp float32) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.temperature = temp
}

func (h *Heater) GetMinTemp() float32 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.minTemp
}
