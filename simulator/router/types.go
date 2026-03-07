package router

import (
	"sync"

	"github.com/rmkhl/halko/simulator/physics"
)

// SimulationResetter holds references for resetting the simulation to initial state
type SimulationResetter struct {
	Heater interface {
		SetTemperature(float32)
		TurnOn(bool)
	}
	Wood                interface{ SetTemperature(float32) }
	Fan                 interface{ TurnOn(bool) }
	Humidifier          interface{ TurnOn(bool) }
	PhysicsState        *physics.SimulationState
	InitialOvenTemp     float32
	InitialMaterialTemp float32
	EnvironmentTemp     float32
	Mutex               sync.Mutex
	LastMessage         string // Track last display message to avoid redundant resets
}

// Router holds methods for handling HTTP requests
type Router struct {
	Resetter *SimulationResetter
}
