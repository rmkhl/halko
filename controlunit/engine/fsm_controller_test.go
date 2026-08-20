package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rmkhl/halko/types"
)

// TestHeatUpModulatesDeltaSteam drives a heating step whose steam uses delta
// control against a real psuController, proving heat_up actually modulates
// steam off when the kiln runs ahead of the material and back on when it falls
// behind. Steam only behaves this way because enterState builds its controller
// from the step; hardwiring it to a constant would still pass every
// program-validation test while leaving steam on for the whole climb.
func TestHeatUpModulatesDeltaSteam(t *testing.T) {
	var mu sync.Mutex
	commanded := map[string]uint8{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		psuName := strings.TrimPrefix(r.URL.Path, "/")
		var cmd PowerCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Errorf("decoding power command for %q: %v", psuName, err)
		}
		mu.Lock()
		commanded[psuName] = cmd.Percent
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	program := &types.Program{
		ProgramName: "delta steam",
		ProgramSteps: []types.ProgramStep{{
			Name:              "heat",
			StepType:          types.StepTypeHeating,
			TargetTemperature: 100,
			Heater:            &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(0)},
			Fan:               &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(100)},
			Steam: &types.PowerPidSettings{
				Type:     types.PowerSettingTypeDelta,
				MinDelta: f32(5),
				MaxDelta: f32(10),
			},
		}},
	}

	fsm := &programFSMController{
		state:               fsmStateHeatUp,
		program:             program,
		psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
	}
	handler := &heatUpStateHandler{fsm: fsm}
	handler.enterState()

	// steamAt runs one tick with the given readings and reports what steam was
	// commanded to. Material stays well below the 100°C target throughout, so
	// the step never completes.
	steamAt := func(kiln, material float32) uint8 {
		fsm.temperatures.reading.Kiln = kiln
		fsm.temperatures.reading.Material = material
		fsm.temperatures.updated = 0
		fsm.currentTemperatures.updated = 1
		handler.executeState()
		mu.Lock()
		defer mu.Unlock()
		return commanded[psuSteam]
	}

	// Kiln 12°C ahead of the material, past the 10°C upper bound: steam cuts out.
	if got := steamAt(62, 50); got != 0 {
		t.Errorf("steam at delta 12 = %d%%, want 0%% (above max delta)", got)
	}
	// Back to 4°C, under the 5°C lower bound: steam comes back on.
	if got := steamAt(54, 50); got != 100 {
		t.Errorf("steam at delta 4 = %d%%, want 100%% (below min delta)", got)
	}
	// Inside the band the controller holds its previous state.
	if got := steamAt(57, 50); got != 100 {
		t.Errorf("steam at delta 7 = %d%%, want 100%% (hysteresis holds)", got)
	}
}

// stepDuration builds the Runtime a time-limited step carries.
func stepDuration(seconds int) *types.StepDuration {
	return &types.StepDuration{Duration: time.Duration(seconds) * time.Second}
}

// A time-limited step ends when its runtime has elapsed and not before. The
// handlers resolve that runtime when the step is entered rather than on every
// tick, so entering the state is what has to make the value available.
func TestTimedStepsEndOnTheirRuntime(t *testing.T) {
	tests := []struct {
		name    string
		state   fsmState
		step    types.ProgramStep
		handler func(*programFSMController) fsmStateHandler
	}{
		{
			name:  "acclimate",
			state: fsmStateAcclimate,
			step: types.ProgramStep{
				Name: "hold", StepType: types.StepTypeAcclimate, TargetTemperature: 100,
				Runtime: stepDuration(600),
				Heater:  &types.PowerPidSettings{Type: types.PowerSettingTypeDelta, MinDelta: f32(-1), MaxDelta: f32(3)},
				Fan:     &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(100)},
				Steam:   &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(0)},
			},
			handler: func(f *programFSMController) fsmStateHandler { return &acclimateStateHandler{fsm: f} },
		},
		{
			name:  "cooling",
			state: fsmStateCoolDown,
			step: types.ProgramStep{
				Name: "cool", StepType: types.StepTypeCooling, TargetTemperature: 30,
				Runtime: stepDuration(600),
				Heater:  &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(0)},
				Fan:     &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(100)},
				Steam:   &types.PowerPidSettings{Type: types.PowerSettingTypeSimple, Power: u8(0)},
			},
			handler: func(f *programFSMController) fsmStateHandler { return &coolDownStateHandler{fsm: f} },
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := &programFSMController{
				state:               tt.state,
				program:             &types.Program{ProgramSteps: []types.ProgramStep{tt.step}},
				psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
				currentPSUStatus:    &fsmPSUStatus{},
				currentTemperatures: &fsmTemperatures{},
			}
			// Material stays above a cooling target and below an acclimate one,
			// so only the runtime can end either step.
			fsm.temperatures.reading.Material = 50
			handler := tt.handler(fsm)

			fsm.stepStarted = time.Now().Unix()
			handler.enterState()
			if got := handler.executeState(); got != tt.state {
				t.Fatalf("step ended immediately: %v", got)
			}

			fsm.stepStarted = time.Now().Unix() - 599
			if got := handler.executeState(); got != tt.state {
				t.Fatalf("step ended one second early: %v", got)
			}

			fsm.stepStarted = time.Now().Unix() - 600
			if got := handler.executeState(); got != fsmStateNextProgramStep {
				t.Fatalf("step did not end on its runtime: %v", got)
			}
		})
	}
}

// equalizeProgram builds a program carrying the startup steps the engine
// synthesizes, with the given band.
func equalizeProgram(delta float32, prewarm bool) *types.Program {
	steps := []types.ProgramStep{{Name: "Equalize", StepType: types.StepTypeEqualize}}
	if prewarm {
		steps = append(steps, types.ProgramStep{Name: "Steam warm-up", StepType: types.StepTypeSteamPrewarm})
	}
	return &types.Program{
		ProgramName:  "equalize",
		Equalize:     &types.EqualizeSettings{Delta: &delta, SteamPrewarm: &prewarm},
		ProgramSteps: steps,
	}
}

// The equalization step never adds heat. A kiln above the band is cooled with
// the fan; a kiln below it is left alone, because the wood warms the air on its
// own. Both cases command every channel on every tick - the power unit's idle
// watchdog zeroes anything it has not heard about within max_idle_time, which
// is how the old preheat state silently lost its fan partway through.
func TestEqualizeDrivesTheRightChannels(t *testing.T) {
	var mu sync.Mutex
	commanded := map[string][]uint8{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		psuName := strings.TrimPrefix(r.URL.Path, "/")
		var cmd PowerCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			t.Errorf("decoding power command for %q: %v", psuName, err)
		}
		mu.Lock()
		commanded[psuName] = append(commanded[psuName], cmd.Percent)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	tests := []struct {
		name                string
		kiln, material      float32
		wantState           fsmState
		wantFan, wantHeater uint8
		wantSteam           uint8
	}{
		{"kiln above the band", 30, 20, fsmStateEqualize, 100, 0, 0},
		{"kiln below the band", 20, 30, fsmStateEqualize, 0, 0, 0},
		{"inside the band", 21, 20, fsmStateNextProgramStep, 0, 0, 0},
		{"inside the band, kiln low", 19, 20, fsmStateNextProgramStep, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			commanded = map[string][]uint8{}
			mu.Unlock()

			fsm := &programFSMController{
				state:               fsmStateEqualize,
				program:             equalizeProgram(2, false),
				psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
				currentPSUStatus:    &fsmPSUStatus{},
				currentTemperatures: &fsmTemperatures{},
			}
			fsm.currentTemperatures.reading.Kiln = tt.kiln
			fsm.currentTemperatures.reading.Material = tt.material

			handler := &equalizeStateHandler{fsm: fsm}
			handler.enterState()
			if got := handler.executeState(); got != tt.wantState {
				t.Fatalf("state = %v, want %v", got, tt.wantState)
			}

			mu.Lock()
			defer mu.Unlock()
			for psu, want := range map[string]uint8{psuFan: tt.wantFan, psuOven: tt.wantHeater, psuSteam: tt.wantSteam} {
				got := commanded[psu]
				if len(got) == 0 {
					t.Errorf("%s was never commanded; the watchdog will zero it", psu)
					continue
				}
				if got[len(got)-1] != want {
					t.Errorf("%s = %d%%, want %d%%", psu, got[len(got)-1], want)
				}
			}
		})
	}
}

// Nothing measures the steam reservoir, so the step infers it from the kiln.
// With the heater and the fan off, the wood can push the air up to its own
// temperature and no further - so a kiln climbing past the top of the band has
// to be receiving heat from the only other thing switched on, the steam.
func TestSteamPrewarmProvesTheGeneratorIsProducing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	tests := []struct {
		name           string
		kiln, material float32
		want           fsmState
	}{
		{"still inside the band", 21, 20, fsmStateSteamPrewarm},
		{"exactly at the top of the band", 22, 20, fsmStateSteamPrewarm},
		{"climbing past the band", 22.5, 20, fsmStateNextProgramStep},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsm := &programFSMController{
				state:               fsmStateSteamPrewarm,
				program:             equalizeProgram(2, true),
				psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
				currentPSUStatus:    &fsmPSUStatus{},
				currentTemperatures: &fsmTemperatures{},
				defaults:            &types.Defaults{Equalize: &types.EqualizeDefaults{SteamPrewarmTimeoutSeconds: 1200}},
			}
			fsm.currentTemperatures.reading.Kiln = tt.kiln
			fsm.currentTemperatures.reading.Material = tt.material

			handler := &steamPrewarmStateHandler{fsm: fsm}
			fsm.stepStarted = time.Now().Unix()
			handler.enterState()

			if got := handler.executeState(); got != tt.want {
				t.Errorf("state = %v, want %v", got, tt.want)
			}
		})
	}
}

// An empty reservoir or a dead element produces no rise at all. The step fails
// the run rather than waiting forever with the element energised - and it does
// so before the program reaches a step that depends on steam.
func TestSteamPrewarmFailsOnItsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	fsm := &programFSMController{
		state:               fsmStateSteamPrewarm,
		program:             equalizeProgram(2, true),
		psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
		defaults:            &types.Defaults{Equalize: &types.EqualizeDefaults{SteamPrewarmTimeoutSeconds: 1200}},
	}
	fsm.currentTemperatures.reading.Kiln = 20
	fsm.currentTemperatures.reading.Material = 20

	handler := &steamPrewarmStateHandler{fsm: fsm}
	fsm.stepStarted = time.Now().Unix()
	handler.enterState()

	if got := handler.executeState(); got != fsmStateSteamPrewarm {
		t.Fatalf("state = %v, want the step to still be waiting", got)
	}

	fsm.stepStarted = time.Now().Unix() - 1199
	if got := handler.executeState(); got != fsmStateSteamPrewarm {
		t.Fatalf("state = %v, want the step to still be waiting one second before the timeout", got)
	}

	fsm.stepStarted = time.Now().Unix() - 1200
	if got := handler.executeState(); got != fsmStateFailed {
		t.Fatalf("state = %v, want the run to fail on the timeout", got)
	}
}

// The equalization step has no timeout on purpose: it never adds heat, so an
// unbounded wait is safe and the operator can cancel the run.
func TestEqualizeHasNoTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	fsm := &programFSMController{
		state:               fsmStateEqualize,
		program:             equalizeProgram(2, false),
		psuController:       &psuController{client: server.Client(), powerControlURL: server.URL},
		currentPSUStatus:    &fsmPSUStatus{},
		currentTemperatures: &fsmTemperatures{},
		defaults:            &types.Defaults{Equalize: &types.EqualizeDefaults{SteamPrewarmTimeoutSeconds: 1200}},
	}
	fsm.currentTemperatures.reading.Kiln = 50
	fsm.currentTemperatures.reading.Material = 20

	handler := &equalizeStateHandler{fsm: fsm}
	handler.enterState()
	fsm.stepStarted = time.Now().Unix() - 100000

	if got := handler.executeState(); got != fsmStateEqualize {
		t.Errorf("state = %v, want the step to keep waiting regardless of elapsed time", got)
	}
}
