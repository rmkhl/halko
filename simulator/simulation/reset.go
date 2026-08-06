// Package simulation owns the simulator's run lifecycle: returning every
// element and the physics state to their configured starting values once a
// program ends, so the next run starts from a clean slate.
package simulation

import (
	"sync"

	"github.com/rmkhl/halko/simulator/faults"
	"github.com/rmkhl/halko/simulator/physics"
	"github.com/rmkhl/halko/types/log"
)

// idleMessage is what the control unit shows once no program is running.
const idleMessage = "idle"

// Heater is the heated element: it holds a temperature and a power state.
type Heater interface {
	SetTemperature(float32)
	TurnOn(bool)
}

// TemperatureSetter is an element whose temperature the physics engine drives.
type TemperatureSetter interface {
	SetTemperature(float32)
}

// Switch is an element with only a power state.
type Switch interface {
	TurnOn(bool)
}

// Resetter returns the simulation to its initial state between programs.
type Resetter struct {
	Heater Heater
	Wood   TemperatureSetter
	Fan    Switch
	Steam  Switch
	// PhysicsState is written here under mu, but the tick loop in
	// simulator/main.go writes the same struct on every tick with no lock.
	// That race predates this type — the old HTTP display handler had it
	// identically — and is not fixed here. mu guards Resetter's own fields
	// (lastMessage); it does not make PhysicsState safe for concurrent use.
	PhysicsState        *physics.SimulationState
	Faults              *faults.Injector
	InitialKilnTemp     float32
	InitialMaterialTemp float32
	EnvironmentTemp     float32

	mu          sync.Mutex
	lastMessage string
}

// OnDisplayMessage resets the simulation when the display text becomes
// "idle" after having been something else, which is how the control unit
// announces that a program has stopped. Repeated idle messages are not
// transitions and do nothing.
func (r *Resetter) OnDisplayMessage(text string) {
	r.mu.Lock()
	transitioned := text == idleMessage && r.lastMessage != idleMessage
	r.lastMessage = text
	r.mu.Unlock()

	if transitioned {
		r.Reset()
	}
}

// Reset puts every element, the physics state and the failure schedule back
// to their starting values.
func (r *Resetter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Info("Resetting simulation to initial state (Kiln: %.1f°C, Material: %.1f°C, Environment: %.1f°C)",
		r.InitialKilnTemp, r.InitialMaterialTemp, r.EnvironmentTemp)

	r.Heater.SetTemperature(r.InitialKilnTemp)
	r.Wood.SetTemperature(r.InitialMaterialTemp)

	r.Heater.TurnOn(false)
	r.Fan.TurnOn(false)
	r.Steam.TurnOn(false)

	r.PhysicsState.KilnTemp = r.InitialKilnTemp
	r.PhysicsState.MaterialTemp = r.InitialMaterialTemp
	r.PhysicsState.EnvironmentTemp = r.EnvironmentTemp
	r.PhysicsState.HeaterIsOn = false
	r.PhysicsState.FanIsOn = false
	r.PhysicsState.SteamIsOn = false

	// Replay the failure schedule from the start on the next run.
	r.Faults.Reset()

	log.Info("Simulation reset complete")
}
