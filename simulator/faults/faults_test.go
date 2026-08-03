package faults

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

// testRNG returns a deterministic source so a seeded run always produces the
// same sensor ordering and the same pattern of dropouts.
func testRNG() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}

func TestObserveDoesNotArmWhileKilnIsNotHotter(t *testing.T) {
	i := newWithRNG(true, testRNG())
	now := time.Now()

	i.Observe(20.0, 20.0, now) // equal: not yet heating
	if i.armed {
		t.Fatalf("expected injector to stay disarmed on equal temperatures")
	}

	i.Observe(19.0, 20.0, now) // kiln colder than material
	if i.armed {
		t.Fatalf("expected injector to stay disarmed while kiln is colder")
	}
}

func TestObserveArmsOnceKilnExceedsMaterial(t *testing.T) {
	i := newWithRNG(true, testRNG())
	start := time.Now()

	i.Observe(20.1, 20.0, start)
	if !i.armed {
		t.Fatalf("expected injector to arm when kiln exceeds material")
	}
	if !i.armedAt.Equal(start) {
		t.Fatalf("expected arm time %v, got %v", start, i.armedAt)
	}

	// Cooling drops the kiln back below the material; the clock must not
	// restart, so a later arming attempt keeps the original timestamp.
	i.Observe(30.0, 25.0, start.Add(time.Minute))
	if !i.armedAt.Equal(start) {
		t.Fatalf("expected arm time to stay %v, got %v", start, i.armedAt)
	}
}

func TestDisabledInjectorNeverArms(t *testing.T) {
	i := newWithRNG(false, testRNG())

	i.Observe(90.0, 20.0, time.Now())
	if i.armed {
		t.Fatalf("expected a disabled injector to stay disarmed")
	}
}

func TestResetDisarms(t *testing.T) {
	i := newWithRNG(true, testRNG())
	start := time.Now()

	i.Observe(30.0, 20.0, start)
	i.Reset()
	if i.armed {
		t.Fatalf("expected Reset to disarm the injector")
	}

	// A later run re-arms and restarts the clock from the new instant.
	restart := start.Add(10 * time.Minute)
	i.Observe(30.0, 20.0, restart)
	if !i.armed || !i.armedAt.Equal(restart) {
		t.Fatalf("expected re-arm at %v, got armed=%v at %v", restart, i.armed, i.armedAt)
	}
}

func TestNilInjectorIsSafe(_ *testing.T) {
	var i *Injector

	i.Observe(90.0, 20.0, time.Now())
	i.Reset()
}

// armedInjector returns an injector armed at the returned instant, with a
// deterministic source.
func armedInjector(t *testing.T) (*Injector, time.Time) {
	t.Helper()

	i := newWithRNG(true, testRNG())
	start := time.Now()
	i.Observe(30.0, 20.0, start)
	return i, start
}

// readings is a fresh response map with both sensors reporting plausible
// values, matching what the simulator's temperature handler builds.
func readings() types.TemperatureResponse {
	return types.TemperatureResponse{"kiln": 50.0, "material": 30.0}
}

func invalid(r types.TemperatureResponse, sensor string) bool {
	return r[sensor] == types.InvalidTemperatureReading
}

func TestApplyLeavesReadingsAloneBeforeArming(t *testing.T) {
	i := newWithRNG(true, testRNG())

	r := readings()
	i.Apply(r, time.Now().Add(time.Hour))
	if invalid(r, "kiln") || invalid(r, "material") {
		t.Fatalf("expected untouched readings before arming, got %v", r)
	}
}

func TestApplyLeavesReadingsAloneInTheFirstMinute(t *testing.T) {
	i, start := armedInjector(t)

	for _, elapsed := range []time.Duration{0, 30 * time.Second, 59 * time.Second} {
		r := readings()
		i.Apply(r, start.Add(elapsed))
		if invalid(r, "kiln") || invalid(r, "material") {
			t.Fatalf("expected untouched readings at %v, got %v", elapsed, r)
		}
	}
}

func TestApplyDropsOutAboutOneReadInFive(t *testing.T) {
	i, start := armedInjector(t)
	at := start.Add(90 * time.Second)

	const samples = 10000
	failures := map[string]int{}
	for n := 0; n < samples; n++ {
		r := readings()
		i.Apply(r, at)
		for _, sensor := range sensorNames {
			if invalid(r, sensor) {
				failures[sensor]++
			}
		}
	}

	// Both sensors roll independently, so each lands near 20% of samples.
	for _, sensor := range sensorNames {
		if failures[sensor] < 1700 || failures[sensor] > 2300 {
			t.Fatalf("expected ~%d dropouts for %s, got %d",
				int(samples*dropoutProbability), sensor, failures[sensor])
		}
	}
}

func TestApplyLosesOneSensorAtTwoMinutes(t *testing.T) {
	i, start := armedInjector(t)
	at := start.Add(150 * time.Second)

	// The first read past the boundary picks the lost sensor; from then on
	// it is invalid on every read.
	r := readings()
	i.Apply(r, at)
	if len(i.lost) != 1 {
		t.Fatalf("expected exactly one lost sensor, got %v", i.lost)
	}
	lost := i.lost[0]

	survivorValid := 0
	for n := 0; n < 100; n++ {
		r := readings()
		i.Apply(r, at)
		if !invalid(r, lost) {
			t.Fatalf("expected %s to be invalid on every read, got %v", lost, r)
		}
		for _, sensor := range sensorNames {
			if sensor != lost && !invalid(r, sensor) {
				survivorValid++
			}
		}
	}

	// The other sensor is still only dropping out intermittently, so across
	// 100 reads it is valid most of the time.
	if survivorValid < 50 {
		t.Fatalf("expected the surviving sensor to stay mostly valid, got %d/100 valid", survivorValid)
	}
}

func TestApplyLosesBothSensorsAtThreeMinutes(t *testing.T) {
	i, start := armedInjector(t)
	at := start.Add(181 * time.Second)

	for n := 0; n < 100; n++ {
		r := readings()
		i.Apply(r, at)
		if !invalid(r, "kiln") || !invalid(r, "material") {
			t.Fatalf("expected both sensors invalid, got %v", r)
		}
	}
}

func TestResetClearsLostSensors(t *testing.T) {
	i, start := armedInjector(t)

	r := readings()
	i.Apply(r, start.Add(200*time.Second))
	if len(i.lost) != 2 {
		t.Fatalf("expected both sensors lost, got %v", i.lost)
	}

	i.Reset()
	if len(i.lost) != 0 {
		t.Fatalf("expected Reset to clear lost sensors, got %v", i.lost)
	}

	// A second run starts clean: freshly armed, nothing fails yet.
	restart := start.Add(time.Hour)
	i.Observe(30.0, 20.0, restart)
	r = readings()
	i.Apply(r, restart.Add(10*time.Second))
	if invalid(r, "kiln") || invalid(r, "material") {
		t.Fatalf("expected untouched readings after reset, got %v", r)
	}
}

func TestDisabledInjectorNeverAltersReadings(t *testing.T) {
	i := newWithRNG(false, testRNG())
	start := time.Now()
	i.Observe(30.0, 20.0, start)

	r := readings()
	i.Apply(r, start.Add(time.Hour))
	if invalid(r, "kiln") || invalid(r, "material") {
		t.Fatalf("expected a disabled injector to leave readings alone, got %v", r)
	}
}

func TestNilInjectorApplyIsSafe(t *testing.T) {
	var i *Injector

	r := readings()
	i.Apply(r, time.Now())
	if invalid(r, "kiln") || invalid(r, "material") {
		t.Fatalf("expected a nil injector to leave readings alone, got %v", r)
	}
}
