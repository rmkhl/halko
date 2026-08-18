package engine

import (
	"fmt"
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

func f32(v float32) *float32 { return &v }
func u8(v uint8) *uint8      { return &v }

func TestNewPowerControllerSelection(t *testing.T) {
	deltaSettings := &types.PowerPidSettings{
		Type:     types.PowerSettingTypeDelta,
		MinDelta: f32(5),
		MaxDelta: f32(20),
	}

	tests := []struct {
		name     string
		stepType types.StepType
		target   float32
		settings *types.PowerPidSettings
		wantType any
	}{
		{"simple", types.StepTypeCooling, 0,
			&types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(40)},
			&simplePowerController{}},
		{"heating delta", types.StepTypeHeating, 80, deltaSettings, &heatingDeltaController{}},
		{"acclimate delta", types.StepTypeAcclimate, 80, deltaSettings, &acclimateDeltaController{}},
		{"pid", types.StepTypeAcclimate, 80,
			&types.PowerPidSettings{Type: types.PowerSettingTypePid, Pid: &types.PidSettings{Kp: 1}},
			&pidPowerController{}},
		// fail-safes: anything unresolvable heats at 0%
		{"delta on cooling step", types.StepTypeCooling, 0, deltaSettings, &simplePowerController{}},
		{"nil settings", types.StepTypeHeating, 80, nil, &simplePowerController{}},
		{"no method resolved", types.StepTypeHeating, 80, &types.PowerPidSettings{}, &simplePowerController{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPowerController(tt.stepType, tt.target, tt.settings)
			if fmt.Sprintf("%T", got) != fmt.Sprintf("%T", tt.wantType) {
				t.Fatalf("NewPowerController() = %T, want %T", got, tt.wantType)
			}
		})
	}
}

func TestNewPowerControllerFailSafeReturnsZero(t *testing.T) {
	c := NewPowerController(types.StepTypeHeating, 80, nil)
	runSequence(t, c, []reading{{kiln: 0, material: 0, want: 0}})
}

func TestNewPowerControllerSimpleUsesConfiguredPower(t *testing.T) {
	c := NewPowerController(types.StepTypeCooling, 0,
		&types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(35)})
	runSequence(t, c, []reading{{kiln: 10, material: 10, want: 35}})
}

// TestAcclimateHoldsKilnBandAroundTarget covers the target-referenced control
// law: the kiln is held in [target+minDelta, target+maxDelta] and the material
// gates only when the heater is allowed to stop.
// The acclimate controller does two different jobs, each with its own
// reference and its own half of the delta band. Heating the kiln closes a
// distance to the target, so it bands the kiln in [target+minDelta, target].
// Heating the material drives the gradient that pushes heat into the wood, so
// it bands the kiln in [material, material+maxDelta]. Which job is in effect is
// decided by whether the wood has reached target.
func TestAcclimateHeatsTheKilnTowardsTarget(t *testing.T) {
	tests := []struct {
		name string
		seq  []reading
	}{
		{
			name: "bands the kiln between target+minDelta and target",
			seq: []reading{
				{kiln: 149.0, material: 151.0, want: 100}, // sagged to target-1
				{kiln: 149.5, material: 151.0, want: 100}, // inside the band, keep heating
				{kiln: 150.0, material: 151.0, want: 0},   // reached target
				{kiln: 149.5, material: 151.0, want: 0},   // inside the band, stay off
			},
		},
		{
			name: "entering hot from a heating step settles onto target",
			seq: []reading{
				{kiln: 155.0, material: 150.0, want: 0}, // handover at target+min_delta_heating
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &acclimateDeltaController{target: 150, minDelta: -1, maxDelta: 3}
			runSequence(t, c, tt.seq)
		})
	}
}

func TestAcclimateHeatsTheMaterialWithinTheGradientLimit(t *testing.T) {
	tests := []struct {
		name string
		seq  []reading
	}{
		{
			name: "bands the kiln between the material and material+maxDelta",
			seq: []reading{
				{kiln: 148.0, material: 149.0, want: 100}, // kiln at or below the wood
				{kiln: 150.0, material: 149.0, want: 100}, // climbing inside the band
				{kiln: 152.0, material: 149.0, want: 0},   // reached the gradient limit
				{kiln: 150.0, material: 149.0, want: 0},   // falling through the band, stay off
				{kiln: 149.0, material: 149.0, want: 100}, // back down to the wood, heat again
			},
		},
		{
			name: "never adds heat to a kiln already past the gradient limit",
			seq: []reading{
				{kiln: 153.0, material: 149.0, want: 0}, // gap 4 exceeds maxDelta 3
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &acclimateDeltaController{target: 150, minDelta: -1, maxDelta: 3}
			runSequence(t, c, tt.seq)
		})
	}
}

// The two bands meet at the target: below it the kiln may sit up to maxDelta
// above the wood, at it the kiln belongs in [target+minDelta, target]. So the
// moment the wood arrives, a kiln still up in the heating band is above its new
// one and the heater stops.
func TestAcclimateStopsWhenTheMaterialReachesTarget(t *testing.T) {
	c := &acclimateDeltaController{target: 150, minDelta: -1, maxDelta: 3}
	runSequence(t, c, []reading{
		{kiln: 149.0, material: 149.0, want: 100}, // heating the wood
		{kiln: 151.0, material: 149.5, want: 100}, // still below target, still climbing
		{kiln: 151.0, material: 150.0, want: 0},   // wood arrived, kiln above its new band
	})
}

// Replays the decay half of a real 150C hold recorded on the Pi. The kiln sags
// well below target long before the wood does, so heating the kiln is the job
// that fires - the previous material-referenced controller waited for the wood
// to drop, and with a negative min_delta its envelope could never re-arm.
func TestAcclimateFiresOnKilnSagNotMaterialSag(t *testing.T) {
	c := &acclimateDeltaController{target: 150, minDelta: -1, maxDelta: 3}
	runSequence(t, c, []reading{
		{kiln: 152.8, material: 150.2, want: 0},   // pulse just ended
		{kiln: 153.2, material: 150.8, want: 0},   // coasting on residual element heat
		{kiln: 151.5, material: 152.0, want: 0},   // kiln now below the wood, both above target
		{kiln: 150.2, material: 151.8, want: 0},   // still inside the kiln band
		{kiln: 149.5, material: 151.0, want: 0},   // still inside the kiln band
		{kiln: 148.5, material: 150.8, want: 100}, // sagged past target-1, heat
	})
}

func TestPidControllerAdvancesItsSampleClock(t *testing.T) {
	c := NewPidController(&types.PidSettings{Kp: 1, Ki: 1})
	start := time.Now().Unix() - 60
	c.State.PreviousUpdate = start

	c.Update(100, 90)

	if c.State.PreviousUpdate == start {
		t.Fatal("PreviousUpdate did not advance: sampleInterval grows without bound")
	}
}

// Two updates inside the same second leave no elapsed time to differentiate
// over. Dividing by it yields Inf or NaN, and converting that to int is
// undefined in Go, so the controller must decline to act instead.
func TestPidControllerIgnoresSubSecondUpdates(t *testing.T) {
	c := NewPidController(&types.PidSettings{Kp: 1, Ki: 1, Kd: 1})
	c.State.PreviousUpdate = time.Now().Unix()

	if got := c.Update(100, 90); got != 0 {
		t.Fatalf("Update within the same second = %v, want 0", got)
	}
	if d := c.State.CurrentErrorDerivative; d != d || d > 1e30 || d < -1e30 {
		t.Fatalf("derivative poisoned: %v", d)
	}
}
