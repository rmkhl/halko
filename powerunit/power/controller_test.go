package power

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rmkhl/halko/powerunit/shelly"
)

// relayRecorder is a real HTTP server speaking enough of the Shelly RPC API for
// the controller to drive it, while remembering what each relay was told to do.
type relayRecorder struct {
	mu     sync.Mutex
	states [shelly.NumberOfDevices]bool
	writes int
}

func (r *relayRecorder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.Atoi(req.URL.Query().Get("id"))
		if err != nil || id < 0 || id >= shelly.NumberOfDevices {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":-103,"message":"bad id"}`))
			return
		}

		r.mu.Lock()
		if on := req.URL.Query().Get("on"); on != "" {
			r.states[id] = on == "true"
			r.writes++
		}
		output := r.states[id]
		r.mu.Unlock()

		_, _ = w.Write([]byte(`{"output":` + strconv.FormatBool(output) + `}`))
	}
}

func (r *relayRecorder) isOn(id int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.states[id]
}

// newTestController builds a controller wired to a real Shelly client and
// server, already in the state Start() leaves behind (every relay off, every
// percentage zero) so tests can drive processTick directly instead of waiting
// on the real ticker.
func newTestController(t *testing.T, maxIdleTime time.Duration) (*Controller, *relayRecorder) {
	t.Helper()

	recorder := &relayRecorder{}
	server := httptest.NewServer(recorder.handler())
	t.Cleanup(server.Close)

	c := New(maxIdleTime, 100*time.Second, shelly.New(server.URL))
	t.Cleanup(c.cancel)

	for i := range shelly.NumberOfDevices {
		c.powerStates[i].percentage = 0
		c.powerStates[i].currentState = shelly.Off
	}
	c.lastCommand = time.Now()

	return c, recorder
}

// runCycle advances the controller through a full 100-tick cycle, calling back
// after each tick with the tick number that was just processed.
func runCycle(t *testing.T, c *Controller, after func(tick int)) {
	t.Helper()

	for tick := range 100 {
		if err := c.processTick(); err != nil {
			t.Fatalf("tick %d: unexpected error: %v", tick, err)
		}
		if after != nil {
			after(tick)
		}
	}
}

func TestZeroPercentNeverTurnsOn(t *testing.T) {
	c, relays := newTestController(t, time.Hour)

	runCycle(t, c, func(tick int) {
		if relays.isOn(0) {
			t.Fatalf("device at 0%% was on at tick %d", tick)
		}
	})
}

func TestFullPercentStaysOnForTheWholeCycle(t *testing.T) {
	c, relays := newTestController(t, time.Hour)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{100, 0, 0})

	runCycle(t, c, func(tick int) {
		if !relays.isOn(0) {
			t.Fatalf("device at 100%% was off at tick %d", tick)
		}
	})
}

func TestPartialPercentIsOnForThatShareOfTheCycle(t *testing.T) {
	c, relays := newTestController(t, time.Hour)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{30, 0, 0})

	onTicks := 0
	runCycle(t, c, func(tick int) {
		on := relays.isOn(0)
		if on {
			onTicks++
		}

		// On for ticks 0-29, off from tick 30 to the end of the cycle.
		if want := tick < 30; on != want {
			t.Fatalf("tick %d: expected on=%v, got on=%v", tick, want, on)
		}
	})

	if onTicks != 30 {
		t.Fatalf("expected 30 on-ticks in a cycle at 30%%, got %d", onTicks)
	}
}

func TestEachDeviceRunsItsOwnDutyCycle(t *testing.T) {
	c, relays := newTestController(t, time.Hour)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 50, 0})

	runCycle(t, c, func(tick int) {
		for id, want := range map[int]bool{0: tick < 10, 1: tick < 50, 2: false} {
			if got := relays.isOn(id); got != want {
				t.Fatalf("tick %d, device %d: expected on=%v, got on=%v", tick, id, want, got)
			}
		}
	})
}

// A new command must not disturb the cycle in progress; it takes hold when the
// next cycle begins, so a relay never gets re-energised mid-cycle.
func TestNewPercentageTakesEffectAtTheNextCycle(t *testing.T) {
	c, relays := newTestController(t, time.Hour)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{20, 0, 0})

	// Advance past the point where 20% switches off.
	for range 25 {
		if err := c.processTick(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if relays.isOn(0) {
		t.Fatal("expected the device to be off at tick 25 of a 20% cycle")
	}

	// Raising it to 60% mid-cycle must not switch the relay back on now.
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{60, 0, 0})
	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if relays.isOn(0) {
		t.Fatal("percentage change switched the relay on mid-cycle")
	}

	// Run out the cycle; the new percentage applies from tick 0.
	for range 74 {
		if err := c.processTick(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if c.tickCount != 0 {
		t.Fatalf("expected to be back at tick 0, got %d", c.tickCount)
	}

	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !relays.isOn(0) {
		t.Fatal("expected the device on at the start of the cycle after raising it to 60%")
	}
}

// The watchdog is what cuts power if the control unit dies mid-run.
func TestIdleTimeoutZeroesPercentagesAndCutsPower(t *testing.T) {
	const maxIdle = 50 * time.Millisecond

	c, relays := newTestController(t, maxIdle)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{80, 80, 80})

	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !relays.isOn(0) {
		t.Fatal("expected the device on at the start of the cycle")
	}

	time.Sleep(2 * maxIdle)

	// The tick that notices the idle period zeroes the percentages...
	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.GetAllPercentages(); got != [shelly.NumberOfDevices]uint8{0, 0, 0} {
		t.Fatalf("expected all percentages zeroed after the idle timeout, got %v", got)
	}

	// ...and the following one switches the relays off.
	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for id := range shelly.NumberOfDevices {
		if relays.isOn(id) {
			t.Fatalf("device %d still on after the idle timeout", id)
		}
	}

	if !c.IsIdle() {
		t.Fatal("expected the controller to report itself idle")
	}
}

// Regression: status polling used to refresh lastCommand, so a webapp sitting
// on the running-program page held the watchdog off indefinitely.
func TestStatusPollingDoesNotHoldOffTheIdleTimeout(t *testing.T) {
	const maxIdle = 50 * time.Millisecond

	c, _ := newTestController(t, maxIdle)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{80, 0, 0})

	// Poll the status the way the webapp does while the idle period elapses.
	deadline := time.Now().Add(2 * maxIdle)
	for time.Now().Before(deadline) {
		c.GetAllPercentages()
		time.Sleep(5 * time.Millisecond)
	}

	if err := c.processTick(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := c.GetAllPercentages(); got != [shelly.NumberOfDevices]uint8{0, 0, 0} {
		t.Fatalf("status polling kept the watchdog from firing: percentages are %v", got)
	}
}

// Regression: switching a relay as part of the duty cycle used to refresh
// lastCommand. Because a running cycle actuates twice per cycle, the gap never
// grew past max_idle_time and the watchdog could not fire at any percentage
// between 1 and 99 — exactly the range a real run spends its time in.
func TestRunningDutyCycleDoesNotFeedTheWatchdog(t *testing.T) {
	// As in halko.cfg, the idle timeout is longer than one full cycle (70s vs
	// 60s there): ~120 ticks of 1ms here against a 100-tick cycle. That is what
	// lets twice-per-cycle actuation starve the watchdog indefinitely.
	const maxIdle = 120 * time.Millisecond

	c, _ := newTestController(t, maxIdle)
	c.SetAllPercentages([shelly.NumberOfDevices]uint8{50, 0, 0})

	// Two full cycles: relays actuate at ticks 0, 50, 100 and 150, never more
	// than half a cycle apart, while no command arrives at all.
	for range 200 {
		if err := c.processTick(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		time.Sleep(time.Millisecond)
	}

	if got := c.GetAllPercentages(); got != [shelly.NumberOfDevices]uint8{0, 0, 0} {
		t.Fatalf("duty-cycle switching kept the watchdog from firing: percentages are %v", got)
	}
}

func TestSetAllPercentagesClearsIdle(t *testing.T) {
	c, _ := newTestController(t, time.Hour)

	if !c.IsIdle() {
		t.Fatal("expected a fresh controller to report itself idle")
	}

	c.SetAllPercentages([shelly.NumberOfDevices]uint8{10, 0, 0})
	if c.IsIdle() {
		t.Fatal("expected a command to clear the idle flag")
	}
}

// Guards the locking around the shared state; meaningful under -race.
func TestConcurrentStatusAndCommands(t *testing.T) {
	c, _ := newTestController(t, time.Hour)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for range 100 {
				c.GetAllPercentages()
				c.IsIdle()
			}
		}()
		go func() {
			defer wg.Done()
			for i := range 100 {
				c.SetAllPercentages([shelly.NumberOfDevices]uint8{uint8(i % 101), 0, 0})
			}
		}()
	}
	wg.Wait()
}
