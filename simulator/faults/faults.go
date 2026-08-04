// Package faults injects probe failures into the simulator's ESP32
// emulation, so the sensor unit's degraded-probe handling and the control
// unit's failsafe can be exercised without a broken probe on real hardware.
package faults

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// Stage boundaries, measured from the instant the injector arms.
const (
	// dropoutsBegin is when probes start failing intermittently.
	dropoutsBegin = 60 * time.Second
	// firstSensorLostAt is when one randomly chosen probe starts failing on
	// every read.
	firstSensorLostAt = 120 * time.Second
	// secondSensorLostAt is when a second probe joins it.
	secondSensorLostAt = 180 * time.Second
	// thirdSensorLostAt is when the last probe goes too.
	thirdSensorLostAt = 240 * time.Second

	// dropoutProbability is the chance of any one probe reporting an
	// invalid reading on a given read during the intermittent stage.
	dropoutProbability = 0.2
)

// sensorNames are the three probes the ESP32 firmware reports, in the order
// it prints them. Must match sensorName[3] in
// sensorunit/esp32/sensorunit/sensorunit.ino:69.
var sensorNames = []string{"KilnPrimary", "KilnSecondary", "Wood"}

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
	// lost holds the sensors that fail on every read, in the order they
	// were lost.
	lost []string
	// dropoutsLogged keeps the intermittent-stage announcement to one line
	// rather than one per read.
	dropoutsLogged bool
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
	log.Info("Sensor failure injection armed (kiln %.1f°C above material %.1f°C): dropouts in %v, one probe lost at %v, a second at %v, the last at %v",
		kiln, material, dropoutsBegin, firstSensorLostAt, secondSensorLostAt, thirdSensorLostAt)
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
	i.lost = nil
	i.dropoutsLogged = false
	log.Info("Sensor failure injection reset")
}

// Apply rewrites readings that the current stage of the schedule says have
// failed, replacing them with types.InvalidTemperatureReading. It mutates the
// map in place and does nothing until the injector is armed.
func (i *Injector) Apply(readings types.TemperatureResponse, now time.Time) {
	if i == nil || !i.enabled {
		return
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if !i.armed {
		return
	}

	elapsed := now.Sub(i.armedAt)
	if elapsed >= dropoutsBegin && !i.dropoutsLogged {
		i.dropoutsLogged = true
		log.Info("Sensor failure injection: intermittent dropouts begin (%.0f%% chance per sensor per read)",
			dropoutProbability*100)
	}
	i.loseSensors(lostByElapsed(elapsed))

	for _, sensor := range sensorNames {
		if _, ok := readings[sensor]; !ok {
			continue
		}
		if i.isLost(sensor) || (elapsed >= dropoutsBegin && i.rng.Float64() < dropoutProbability) {
			readings[sensor] = types.InvalidTemperatureReading
			log.Debug("Sensor failure injection: reporting %s as invalid", sensor)
		}
	}
}

// lostByElapsed returns how many probes should be failing permanently by the
// given point in the schedule.
func lostByElapsed(elapsed time.Duration) int {
	switch {
	case elapsed >= thirdSensorLostAt:
		return 3
	case elapsed >= secondSensorLostAt:
		return 2
	case elapsed >= firstSensorLostAt:
		return 1
	default:
		return 0
	}
}

// loseSensors picks sensors at random until the wanted number are failing
// permanently. Callers hold i.mu.
func (i *Injector) loseSensors(wanted int) {
	for len(i.lost) < wanted {
		remaining := make([]string, 0, len(sensorNames))
		for _, sensor := range sensorNames {
			if !i.isLost(sensor) {
				remaining = append(remaining, sensor)
			}
		}
		if len(remaining) == 0 {
			return
		}

		sensor := remaining[i.rng.IntN(len(remaining))]
		i.lost = append(i.lost, sensor)
		log.Info("Sensor failure injection: %s sensor lost, now failing on every read (%d of %d lost)",
			sensor, len(i.lost), len(sensorNames))
	}
}

// isLost reports whether a sensor is failing permanently. Callers hold i.mu.
func (i *Injector) isLost(sensor string) bool {
	for _, name := range i.lost {
		if name == sensor {
			return true
		}
	}
	return false
}
