package elements

import (
	"sync"
)

type (
	Wood struct {
		mutex       sync.RWMutex
		temperature float32
		minTemp     float32
	}
)

func NewWood(minTemp float32) *Wood {
	w := Wood{minTemp: minTemp, temperature: minTemp}
	return &w
}

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

func (w *Wood) Temperature() float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.temperature
}
