package elements

import (
	"sync"
)

type (
	Wood struct {
		mutex       sync.RWMutex
		temperature float32
		minTemp     float32
		maxTemp     float32
	}
)

func NewWood(minTemp, maxTemp float32) *Wood {
	w := Wood{minTemp: minTemp, maxTemp: maxTemp, temperature: minTemp}
	return &w
}

func (w *Wood) TargetReached() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.temperature >= w.maxTemp
}

// Implement the HeatedMaterial interface.
func (w *Wood) AmbientTemperature(temperature float32) {
	var effectiveDelta float32

	w.mutex.Lock()
	defer w.mutex.Unlock()

	delta := temperature - w.temperature
	switch {
	case delta > 0.0:
		effectiveDelta = 0.01
	case delta < 0.0:
		effectiveDelta = -0.01
	}
	w.temperature = max(w.minTemp, w.temperature+effectiveDelta)
}

// Implement the TemperatureSensor interface.
func (w *Wood) Temperature() float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.temperature
}
