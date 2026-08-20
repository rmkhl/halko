package engine

import (
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

const (
	fsmStateStart           fsmState = "start"
	fsmStateNextProgramStep fsmState = "next_program_step"
	fsmStateIdle            fsmState = "idle"
	fsmStateWaiting         fsmState = "waiting"
	fsmStateEqualize        fsmState = "equalize"
	fsmStateSteamPrewarm    fsmState = "steam_prewarm"
	fsmStateHeatUp          fsmState = "heat_up"
	fsmStateAcclimate       fsmState = "acclimate"
	fsmStateCoolDown        fsmState = "cool_down"
	fsmStateFailed          fsmState = "failed"
)

type (
	fsmState string

	fsmStateHandler interface {
		executeState() fsmState
		enterState()
	}

	startStateHandler struct {
		fsm *programFSMController
	}

	nextProgramStepHandler struct {
		fsm *programFSMController
	}

	waitingStateHandler struct {
		fsm *programFSMController
	}

	equalizeStateHandler struct {
		fsm *programFSMController
	}

	steamPrewarmStateHandler struct {
		fsm *programFSMController
	}

	heatUpStateHandler struct {
		fsm         *programFSMController
		fanPower    PowerController
		heaterPower PowerController
		steamPower  PowerController
	}

	acclimateStateHandler struct {
		fsm         *programFSMController
		fanPower    PowerController
		heaterPower PowerController
		steamPower  PowerController
		// Resolved when the step is entered. elapsed is a second count, so
		// keeping the runtime in the same unit avoids converting the step's
		// duration on every tick.
		runtimeSeconds int64
	}

	coolDownStateHandler struct {
		fsm         *programFSMController
		fanPower    PowerController
		heaterPower PowerController
		steamPower  PowerController
		// Resolved when the step is entered. A cooling step's runtime is
		// optional, so hasRuntimeLimit says whether runtimeSeconds means
		// anything.
		runtimeSeconds  int64
		hasRuntimeLimit bool
	}

	failedStateHandler struct {
		fsm *programFSMController
	}

	idleStateHandler struct {
		fsm *programFSMController
	}

	fsmPSUStatus struct {
		updated int64
		reading psuReadings
	}

	programFSMController struct {
		state   fsmState
		started int64
		stopped int64

		program       *types.Program
		numberOfSteps int
		step          int
		stepStarted   int64

		psuController *psuController

		psuStatus    fsmPSUStatus
		temperatures fsmTemperatures

		// These are updated by the runner on regular bases.
		// No lock needed as the runner is the only one updating them
		// and will not update them while executeStep is running.
		currentPSUStatus    *fsmPSUStatus
		currentTemperatures *fsmTemperatures

		stateHandlers map[fsmState]fsmStateHandler
		stepToState   map[types.StepType]fsmState
		defaults      *types.Defaults
	}
)

func (h *startStateHandler) executeState() fsmState {
	log.Debug("FSM: start state - transitioning to waiting")
	return fsmStateWaiting
}

func (h *startStateHandler) enterState() {
	h.fsm.stopped = 0
	log.Info("FSM: Entered start state - program started at %s", time.Unix(h.fsm.started, 0).Format(time.RFC3339))
}

func (h *waitingStateHandler) executeState() fsmState {
	// Make sure we have received updated temperature and psu status
	sensorsValid := h.fsm.currentTemperatures.kilnValidAt >= h.fsm.started &&
		h.fsm.currentTemperatures.materialValidAt >= h.fsm.started
	if h.fsm.currentPSUStatus.updated >= h.fsm.started && sensorsValid {
		log.Debug("FSM: waiting state - sensors ready (PSU: %s, Temp: %s), transitioning to the first step",
			time.Unix(h.fsm.currentPSUStatus.updated, 0).Format(time.RFC3339),
			time.Unix(h.fsm.currentTemperatures.updated, 0).Format(time.RFC3339))
		return fsmStateNextProgramStep
	}
	log.Trace("FSM: waiting state - waiting for sensors (PSU: %s >= %s: %v, Temp: %s >= %s: %v)",
		time.Unix(h.fsm.currentPSUStatus.updated, 0).Format(time.RFC3339),
		time.Unix(h.fsm.started, 0).Format(time.RFC3339),
		h.fsm.currentPSUStatus.updated >= h.fsm.started,
		time.Unix(h.fsm.currentTemperatures.updated, 0).Format(time.RFC3339),
		time.Unix(h.fsm.started, 0).Format(time.RFC3339),
		h.fsm.currentTemperatures.updated >= h.fsm.started)
	return fsmStateWaiting
}

func (h *waitingStateHandler) enterState() {
	h.fsm.step = -1
	h.fsm.stepStarted = time.Now().Unix()
	log.Info("FSM: Entered waiting state - waiting for initial sensor data")
}

// The equalization step holds until the kiln and the material are within the
// program's band, so every program starts from a known state. It never adds
// heat: a kiln above the band is cooled by exhausting warm air with the fan,
// and a kiln below it is simply waited out - the wood is the larger thermal
// mass and warms the air up to meet it.
func (h *equalizeStateHandler) executeState() fsmState {
	delta := *h.fsm.program.Equalize.Delta
	kiln := h.fsm.currentTemperatures.reading.Kiln
	material := h.fsm.currentTemperatures.reading.Material
	gap := kiln - material

	fan := uint8(0)
	next := fsmStateNextProgramStep
	switch {
	case gap > delta:
		log.Trace("FSM: equalize - kiln (%.1f°C) above material (%.1f°C) by more than %.1f°C, exhausting",
			kiln, material, delta)
		fan = 100
		next = fsmStateEqualize
	case gap < -delta:
		log.Trace("FSM: equalize - kiln (%.1f°C) below material (%.1f°C) by more than %.1f°C, waiting for the wood to warm the air",
			kiln, material, delta)
		next = fsmStateEqualize
	default:
		log.Info("FSM: equalize - kiln (%.1f°C) and material (%.1f°C) within %.1f°C, program may start",
			kiln, material, delta)
	}

	// Every channel, every tick. The power unit's idle watchdog zeroes
	// anything it has not been told about within max_idle_time.
	h.fsm.psuController.setPower(psuFan, fan)
	h.fsm.psuController.setPower(psuOven, 0)
	h.fsm.psuController.setPower(psuSteam, 0)

	return next
}

func (h *equalizeStateHandler) enterState() {
	h.fsm.stepStarted = time.Now().Unix()
	log.Info("FSM: Entered equalize state - band %.1f°C", *h.fsm.program.Equalize.Delta)
}

// The steam generator is a water reservoir with a heating element: it produces
// nothing until it boils, several minutes in, and nothing measures it. This
// step infers the moment it starts producing from the kiln reading. It runs
// with the heater and the fan off, entered with the kiln and the material
// already inside the band, so the wood can push the air to its own temperature
// and no further - a kiln climbing past the top of the band is receiving heat
// from the steam.
//
// The step therefore hands over with the kiln a few degrees above the wood,
// outside the equalization band. That is deliberate: the program's first step
// is a heating step, which bands the kiln above the material anyway, so the
// hand-over lands inside the band that step wants.
func (h *steamPrewarmStateHandler) executeState() fsmState {
	delta := *h.fsm.program.Equalize.Delta
	kiln := h.fsm.currentTemperatures.reading.Kiln
	material := h.fsm.currentTemperatures.reading.Material

	// Every channel, every tick, so the power unit's idle watchdog cannot cut
	// the steam out from under the step it is waiting on.
	h.fsm.psuController.setPower(psuSteam, 100)
	h.fsm.psuController.setPower(psuOven, 0)
	h.fsm.psuController.setPower(psuFan, 0)

	if kiln > material+delta {
		log.Info("FSM: steam warm-up - kiln (%.1f°C) climbing past material (%.1f°C) + %.1f°C, the generator is producing",
			kiln, material, delta)
		return fsmStateNextProgramStep
	}

	elapsed := time.Now().Unix() - h.fsm.stepStarted
	if elapsed >= h.fsm.defaults.Equalize.SteamPrewarmTimeoutSeconds {
		log.Error("FSM: steam warm-up - no rise past %.1f°C above the material in %ds (limit %ds), the reservoir is empty or the element is dead - failing program",
			delta, elapsed, h.fsm.defaults.Equalize.SteamPrewarmTimeoutSeconds)
		return fsmStateFailed
	}

	log.Trace("FSM: steam warm-up - waiting for the generator (kiln %.1f°C, material %.1f°C, %ds / %ds)",
		kiln, material, elapsed, h.fsm.defaults.Equalize.SteamPrewarmTimeoutSeconds)
	return fsmStateSteamPrewarm
}

func (h *steamPrewarmStateHandler) enterState() {
	h.fsm.stepStarted = time.Now().Unix()
	log.Info("FSM: Entered steam warm-up state - waiting up to %ds for the generator to start producing",
		h.fsm.defaults.Equalize.SteamPrewarmTimeoutSeconds)
}

func (h *nextProgramStepHandler) executeState() fsmState {
	// Note this assumes that before first call fsm.steps is set to -1
	h.fsm.step++
	// End of the program reached
	if h.fsm.step >= h.fsm.numberOfSteps {
		log.Info("FSM: All steps completed (step %d >= %d), transitioning to idle",
			h.fsm.step, h.fsm.numberOfSteps)
		return fsmStateIdle
	}
	h.fsm.stepStarted = time.Now().Unix()
	nextState := h.fsm.stepToState[h.fsm.program.ProgramSteps[h.fsm.step].StepType]
	log.Info("FSM: Moving to step %d/%d: '%s' (type: %s, target: %d°C) - transitioning to %s",
		h.fsm.step+1, h.fsm.numberOfSteps,
		h.fsm.program.ProgramSteps[h.fsm.step].Name,
		h.fsm.program.ProgramSteps[h.fsm.step].StepType,
		h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature,
		nextState)
	return nextState
}

func (h *nextProgramStepHandler) enterState() {
	h.fsm.stepStarted = time.Now().Unix()
	log.Debug("FSM: Entered next_program_step state")
}

func (h *heatUpStateHandler) executeState() fsmState {
	// If the target temperature is reached, we can move to the next step
	if h.fsm.temperatures.reading.Material >= float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature) {
		log.Info("FSM: heat_up - target temperature reached (material: %.1f°C >= target: %d°C)",
			h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
		return fsmStateNextProgramStep
	}
	log.Trace("FSM: heat_up - heating (material: %.1f°C / target: %d°C)",
		h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		log.Debug("FSM: heat_up - updating power (kiln: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		heaterPower := h.heaterPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, heaterPower)
		log.Trace("FSM: heat_up - heater power: %d%%", heaterPower)

		fanResult := h.fanPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuFan, fanResult)
		log.Trace("FSM: heat_up - fan power: %d%%", fanResult)

		steamResult := h.steamPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuSteam, steamResult)
		log.Trace("FSM: heat_up - steam power: %d%%", steamResult)

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateHeatUp
}

func (h *heatUpStateHandler) enterState() {
	log.Info("FSM: Entered heat_up state - target: %d°C", h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	step := &h.fsm.program.ProgramSteps[h.fsm.step]
	h.fanPower = NewPowerController(step.StepType, 0, step.Fan)
	h.heaterPower = NewPowerController(step.StepType, float32(step.TargetTemperature), step.Heater)
	h.steamPower = NewPowerController(step.StepType, 0, step.Steam)
}

func (h *acclimateStateHandler) executeState() fsmState {
	// Once we have been acclimating long enough, we can move to the next step
	elapsed := time.Now().Unix() - h.fsm.stepStarted
	if elapsed >= h.runtimeSeconds {
		log.Info("FSM: acclimate - runtime complete (%ds / %ds)", elapsed, h.runtimeSeconds)
		return fsmStateNextProgramStep
	}
	log.Trace("FSM: acclimate - maintaining temperature (%ds / %ds, material: %.1f°C, target: %d°C)",
		elapsed, h.runtimeSeconds, h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		log.Debug("FSM: acclimate - updating power (kiln: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuSteam, h.steamPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateAcclimate
}

func (h *acclimateStateHandler) enterState() {
	step := &h.fsm.program.ProgramSteps[h.fsm.step]
	h.runtimeSeconds = int64(step.Runtime.Seconds())
	log.Info("FSM: Entered acclimate state - target: %d°C, duration: %ds",
		step.TargetTemperature, h.runtimeSeconds)
	h.fanPower = NewPowerController(step.StepType, 0, step.Fan)
	h.heaterPower = NewPowerController(step.StepType, float32(step.TargetTemperature), step.Heater)
	h.steamPower = NewPowerController(step.StepType, 0, step.Steam)
}

func (h *coolDownStateHandler) executeState() fsmState {
	// If we have been cooling down long enough, we can move to the next step
	elapsed := time.Now().Unix() - h.fsm.stepStarted
	if h.hasRuntimeLimit && elapsed >= h.runtimeSeconds {
		log.Info("FSM: cool_down - runtime limit reached (%ds / %ds)", elapsed, h.runtimeSeconds)
		return fsmStateNextProgramStep
	}
	// If the wood has cooled enough (or we have reached the time limit), we can move to the next step
	if h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature != 0 && h.fsm.temperatures.reading.Material <= float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature) {
		log.Info("FSM: cool_down - target temperature reached (material: %.1f°C <= target: %d°C)",
			h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
		return fsmStateNextProgramStep
	}
	log.Trace("FSM: cool_down - cooling (material: %.1f°C / target: %d°C, elapsed: %ds)",
		h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, elapsed)
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		log.Debug("FSM: cool_down - updating power (kiln: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuSteam, h.steamPower.Update(h.fsm.temperatures.reading.Kiln, h.fsm.temperatures.reading.Material))

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateCoolDown
}

func (h *coolDownStateHandler) enterState() {
	step := &h.fsm.program.ProgramSteps[h.fsm.step]
	h.hasRuntimeLimit = step.Runtime != nil
	if h.hasRuntimeLimit {
		h.runtimeSeconds = int64(step.Runtime.Seconds())
	}
	log.Info("FSM: Entered cool_down state - target: %d°C", step.TargetTemperature)
	h.fanPower = NewPowerController(step.StepType, 0, step.Fan)
	h.heaterPower = NewPowerController(step.StepType, float32(step.TargetTemperature), step.Heater)
	h.steamPower = NewPowerController(step.StepType, 0, step.Steam)
}

func (h *failedStateHandler) executeState() fsmState {
	// This is an end state, do not automatically transition from idle state
	return fsmStateFailed
}

func (h *failedStateHandler) enterState() {
	log.Error("FSM: Entered failed state - shutting down")
	h.fsm.shutdown()
}

func (h *idleStateHandler) executeState() fsmState {
	// This is an end state, do not automatically transition from idle state
	return fsmStateIdle
}

func (h *idleStateHandler) enterState() {
	log.Info("FSM: Entered idle state - program complete, shutting down")
	h.fsm.shutdown()
}

func newProgramFSMController(psuController *psuController, psuStatus *fsmPSUStatus, temperatures *fsmTemperatures, defaults *types.Defaults) *programFSMController {
	controller := &programFSMController{
		psuController:       psuController,
		currentPSUStatus:    psuStatus,
		currentTemperatures: temperatures,
		stepToState: map[types.StepType]fsmState{
			types.StepTypeHeating:      fsmStateHeatUp,
			types.StepTypeAcclimate:    fsmStateAcclimate,
			types.StepTypeCooling:      fsmStateCoolDown,
			types.StepTypeEqualize:     fsmStateEqualize,
			types.StepTypeSteamPrewarm: fsmStateSteamPrewarm,
		},
		defaults: defaults,
	}
	controller.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateStart:           &startStateHandler{fsm: controller},
		fsmStateNextProgramStep: &nextProgramStepHandler{fsm: controller},
		fsmStateIdle:            &idleStateHandler{fsm: controller},
		fsmStateWaiting:         &waitingStateHandler{fsm: controller},
		fsmStateEqualize:        &equalizeStateHandler{fsm: controller},
		fsmStateSteamPrewarm:    &steamPrewarmStateHandler{fsm: controller},
		fsmStateHeatUp:          &heatUpStateHandler{fsm: controller},
		fsmStateAcclimate:       &acclimateStateHandler{fsm: controller},
		fsmStateCoolDown:        &coolDownStateHandler{fsm: controller},
		fsmStateFailed:          &failedStateHandler{fsm: controller},
	}
	return controller
}

func (p *programFSMController) executeTick() {
	p.executeTickAt(time.Now().Unix())
}

func (p *programFSMController) executeTickAt(now int64) {
	// Reached the end of the program
	if p.Completed() {
		log.Trace("FSM: executeTick - program completed, no action")
		return
	}

	// A sensor that has stopped reporting valid readings leaves the
	// controllers working from a frozen value, so stop the program and
	// switch everything off rather than keep heating blind.
	if sensor, seconds := p.currentTemperatures.invalidFor(now, p.started); seconds > p.defaults.SensorTimeoutSeconds {
		log.Error("FSM: no valid %s temperature for %ds (limit %ds) - failing program",
			sensor, seconds, p.defaults.SensorTimeoutSeconds)
		p.state = fsmStateFailed
		p.stepStarted = now
		p.stateHandlers[p.state].enterState()
		return
	}

	previousState := p.state
	p.state = p.stateHandlers[p.state].executeState()
	if p.state != previousState {
		log.Info("FSM: State transition: %s -> %s", previousState, p.state)
		p.stepStarted = time.Now().Unix()
		p.stateHandlers[p.state].enterState()
	} else {
		log.Trace("FSM: executeTick - remaining in state %s", p.state)
	}

	// Last thing record the current psu and temperature status
	// for the next tick execution
	p.psuStatus = *p.currentPSUStatus
	p.temperatures = *p.currentTemperatures
}

// Shutdown the program. If the program has not completed normally we need to turn off all power.
func (p *programFSMController) shutdown() {
	if p.stopped == 0 {
		p.stopped = time.Now().Unix()
		if p.psuController == nil {
			log.Debug("FSM: Shutdown with no power controller attached")
			return
		}
		log.Info("FSM: Shutting down - turning off all power")
		p.psuController.setPower(psuOven, 0)
		p.psuController.setPower(psuFan, 0)
		p.psuController.setPower(psuSteam, 0)
		log.Debug("FSM: Shutdown complete at %d", p.stopped)
	}
}

func (p *programFSMController) Start(program *types.Program, startTime int64) {
	p.program = program
	p.state = fsmStateStart
	p.numberOfSteps = len(program.ProgramSteps)
	p.started = startTime
	log.Info("FSM: Starting program '%s' with %d steps at %s", program.ProgramName, p.numberOfSteps, time.Unix(startTime, 0).Format(time.RFC3339))
	p.stateHandlers[p.state].enterState()
}

func (p *programFSMController) Completed() bool {
	return p.state == fsmStateFailed || p.state == fsmStateIdle
}

func (p *programFSMController) Failed() bool {
	return p.state == fsmStateFailed
}

func (p *programFSMController) UpdateStatus(status *types.ExecutionStatus) {
	status.StartedAt = p.started
	status.CurrentStepStartedAt = p.stepStarted

	// Set current step name based on state
	switch {
	case p.step >= 0 && p.step < p.numberOfSteps:
		status.CurrentStep = p.program.ProgramSteps[p.step].Name
	case p.step < 0:
		status.CurrentStep = "Waiting"
	default:
		status.CurrentStep = "Completed"
	}

	status.Temperatures.Material = p.temperatures.reading.Material
	status.Temperatures.Kiln = p.temperatures.reading.Kiln
	status.Temperatures.MaterialDie = p.temperatures.reading.MaterialDie
	status.Temperatures.KilnPrimaryDie = p.temperatures.reading.KilnPrimaryDie
	status.Temperatures.KilnSecondaryDie = p.temperatures.reading.KilnSecondaryDie
	status.PowerStatus.Heater = int8(p.psuStatus.reading.Heater.Percent)
	status.PowerStatus.Fan = int8(p.psuStatus.reading.Fan.Percent)
	status.PowerStatus.Steam = int8(p.psuStatus.reading.Steam.Percent)
}
