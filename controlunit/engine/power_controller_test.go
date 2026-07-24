package engine

import (
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

type reading struct {
	kiln     float32
	material float32
	want     uint8
}

// updater is satisfied by every power controller implementation.
type updater interface {
	Update(kilnTemperature, materialTemperature float32) uint8
}

func runSequence(t *testing.T, c updater, seq []reading) {
	t.Helper()
	for i, r := range seq {
		if got := c.Update(r.kiln, r.material); got != r.want {
			t.Fatalf("step %d: Update(kiln=%v, material=%v) = %d, want %d", i, r.kiln, r.material, got, r.want)
		}
	}
}

func TestHeatingDeltaHysteresis(t *testing.T) {
	tests := []struct {
		name string
		seq  []reading
	}{
		{
			name: "starts heating and cuts off at upper bound",
			seq: []reading{
				{kiln: 20, material: 20, want: 100}, // below band, heat
				{kiln: 39, material: 20, want: 100}, // inside band while rising, keep heating
				{kiln: 40, material: 20, want: 0},   // reached material+maxDelta, off
			},
		},
		{
			name: "stays off inside band until lower bound reached",
			seq: []reading{
				{kiln: 41, material: 20, want: 0},   // above band, off
				{kiln: 30, material: 20, want: 0},   // falling through band, stay off (hysteresis)
				{kiln: 26, material: 20, want: 0},   // still above material+minDelta, stay off
				{kiln: 25, material: 20, want: 100}, // reached material+minDelta, back on
				{kiln: 30, material: 20, want: 100}, // rising through band, keep heating
			},
		},
		{
			name: "band follows the material temperature",
			seq: []reading{
				{kiln: 50, material: 40, want: 100}, // inside band [45,60], initial state on
				{kiln: 60, material: 40, want: 0},   // reached 40+20, off
				{kiln: 60, material: 55, want: 100}, // material caught up: 55+5=60, on again
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &heatingDeltaController{minDelta: 5, maxDelta: 20, heaterOn: true}
			runSequence(t, c, tt.seq)
		})
	}
}

func TestAcclimateDeltaHoldsTarget(t *testing.T) {
	tests := []struct {
		name string
		seq  []reading
	}{
		{
			name: "never heats when material at or above target",
			seq: []reading{
				{kiln: 30, material: 60, want: 0}, // at target, kiln cold: no demand
				{kiln: 30, material: 61, want: 0}, // above target: never heat
			},
		},
		{
			name: "heats below target and stops the moment target is reached",
			seq: []reading{
				{kiln: 62, material: 59, want: 100}, // demand, kiln inside envelope
				{kiln: 62, material: 60, want: 0},   // material reached target, off immediately
			},
		},
		{
			name: "envelope breach cuts power mid-demand and re-arms at lower bound",
			seq: []reading{
				{kiln: 70, material: 55, want: 100}, // demand, inside envelope (55+20=75)
				{kiln: 75, material: 55, want: 0},   // breach at 55+20: off despite demand
				{kiln: 65, material: 55, want: 0},   // falling through band: still disarmed
				{kiln: 60, material: 55, want: 100}, // reached 55+5: re-armed, demand present
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &acclimateDeltaController{target: 60, minDelta: 5, maxDelta: 20, envelopeOK: true}
			runSequence(t, c, tt.seq)
		})
	}
}

func TestSimplePowerControllerReturnsFixedPower(t *testing.T) {
	c := &simplePowerController{power: 40}
	runSequence(t, c, []reading{
		{kiln: 10, material: 10, want: 40},
		{kiln: 90, material: 20, want: 40},
	})
}

func TestPidPowerControllerTracksItsOwnOutput(t *testing.T) {
	c := &pidPowerController{
		pid:    NewPidController(&types.PidSettings{Kp: 10}),
		target: 100,
	}
	// Backdate the PID state so Update sees a 1-second sample interval
	// (the first PidController.Update call otherwise just initializes state).
	c.pid.State.PreviousUpdate = time.Now().Unix() - 1

	// error = 100-90 = 10, Kp*10 = +100 applied to lastPower 0, clamped to 100
	if got := c.Update(90, 0); got != 100 {
		t.Fatalf("first Update = %d, want 100", got)
	}

	c.pid.State.PreviousUpdate = time.Now().Unix() - 1
	// error = 100-200 = -100, Kp*-100 = -1000 applied to lastPower 100, clamped to 0.
	// This fails if the implementation uses anything but its own last output as base.
	if got := c.Update(200, 0); got != 0 {
		t.Fatalf("second Update = %d, want 0", got)
	}
}
