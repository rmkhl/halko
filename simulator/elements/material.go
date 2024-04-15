package elements

import (
	"sync"
)

type (
	Wood struct {
		mutex       sync.Mutex
		temperature float32
		min_temp    float32
		max_temp    float32
	}
)

func NewWood(min_temp float32, max_temp float32) *Wood {
	w := Wood{min_temp: min_temp, max_temp: max_temp, temperature: min_temp}
	return &w
}

func (w *Wood) TargetReached() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.temperature >= w.max_temp
}

// Implement the HeatedMaterial interface
func (w *Wood) AmbientTemperature(temperature float32) {
	var effective_delta float32 = 0.0

	w.mutex.Lock()
	defer w.mutex.Unlock()

	delta := temperature - w.temperature
	switch {
	case delta > 0.0:
		effective_delta = 0.01
	case delta < 0.0:
		effective_delta = -0.01
	}
	w.temperature = max(w.min_temp, w.temperature+effective_delta)
}

// Implement the TemperatureSensor interface
func (w *Wood) Temperature() float32 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.temperature
}
