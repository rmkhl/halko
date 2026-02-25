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

func NewWood(initialTemp float32, envTemp float32) *Wood {
	w := Wood{minTemp: envTemp, temperature: initialTemp}
	return &w
}

func (w *Wood) Temperature() float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.temperature
}

func (w *Wood) SetTemperature(temp float32) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.temperature = temp
}

func (w *Wood) GetMinTemp() float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.minTemp
}
