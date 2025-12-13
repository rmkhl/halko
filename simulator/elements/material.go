package elements

import (
	"sync"

	"github.com/rmkhl/halko/types/log"
)

type (
	Wood struct {
		mutex       sync.RWMutex
		temperature float32
		minTemp     float32
	}
)

func NewWood(initialTemp float32, envTemp float32) *Wood {
	w := Wood{minTemp: envTemp, temperature: initialTemp}
	return &w
}

func (w *Wood) AmbientTemperature(temperature float32) {
	var effectiveDelta float32

	w.mutex.Lock()
	defer w.mutex.Unlock()

	oldTemp := w.temperature
	delta := temperature - w.temperature
	switch {
	case delta > 0.0:
		effectiveDelta = 0.01
	case delta < 0.0:
		effectiveDelta = -0.01
	}
	w.temperature = max(w.minTemp, w.temperature+effectiveDelta)
	if w.temperature != oldTemp {
		log.Trace("Wood: Temperature change %.1f°C -> %.1f°C (ambient: %.1f°C, delta: %.2f°C)",
			oldTemp, w.temperature, temperature, delta)
	}
}

func (w *Wood) Temperature() float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.temperature
}
