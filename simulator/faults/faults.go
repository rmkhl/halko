// Package faults injects temperature sensor failures into the simulator's
// readings, so the control unit's invalid-reading handling and its failsafe
// can be exercised without a broken probe on real hardware.
package faults

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/rmkhl/halko/types/log"
)

// Stage boundaries, measured from the instant the injector arms.
const (
	// dropoutsBegin is when sensors start failing intermittently.
	dropoutsBegin = 60 * time.Second
	// firstSensorLostAt is when one randomly chosen sensor starts failing
	// on every read.
	firstSensorLostAt = 120 * time.Second
	// secondSensorLostAt is when the remaining sensor joins it, which
	// should trip the control unit's failsafe two minutes later.
	secondSensorLostAt = 180 * time.Second

	// dropoutProbability is the chance of any one sensor reporting an
	// invalid reading on a given read during the intermittent stage.
	dropoutProbability = 0.2
)

// Injector rewrites temperature readings to types.InvalidTemperatureReading on
// an escalating schedule. It stays inert until Observe sees the kiln rise
// above the material, which is the simulator's proxy for a program heating.
//
// The zero value is not usable; construct one with New. All methods are safe
// to call on a nil or disabled Injector, in which case they do nothing.
type Injector struct {
	mu      sync.Mutex
	enabled bool
	rng     *rand.Rand
	armed   bool
	armedAt time.Time
}

// New returns an Injector seeded from the clock. When enabled is false the
// returned Injector never alters a reading.
func New(enabled bool) *Injector {
	seed := uint64(time.Now().UnixNano())
	return newWithRNG(enabled, rand.New(rand.NewPCG(seed, seed>>32)))
}

// newWithRNG builds an Injector over a caller-supplied source, so tests can
// drive a deterministic sequence.
func newWithRNG(enabled bool, rng *rand.Rand) *Injector {
	return &Injector{enabled: enabled, rng: rng}
}

// Observe starts the failure clock the first time the kiln reads hotter than
// the material. Once armed the injector stays armed, so the cooling phase —
// where the kiln falls back below the material — does not restart the
// schedule.
func (i *Injector) Observe(kiln, material float32, now time.Time) {
	if i == nil || !i.enabled {
		return
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if i.armed || kiln <= material {
		return
	}

	i.armed = true
	i.armedAt = now
	log.Info("Sensor failure injection armed (kiln %.1f°C above material %.1f°C): dropouts in %v, one sensor lost at %v, both at %v",
		kiln, material, dropoutsBegin, firstSensorLostAt, secondSensorLostAt)
}

// Reset disarms the injector and clears every failure, so the next run replays
// the sequence from the start.
func (i *Injector) Reset() {
	if i == nil || !i.enabled {
		return
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.armed = false
	i.armedAt = time.Time{}
	log.Info("Sensor failure injection reset")
}
