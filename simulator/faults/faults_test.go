package faults

import (
	"math/rand/v2"
	"testing"
	"time"
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
