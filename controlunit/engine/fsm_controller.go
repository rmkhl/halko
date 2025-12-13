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
	fsmStatePreHeat         fsmState = "preheat"
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

	preHeatStateHandler struct {
		fsm *programFSMController
	}

	heatUpStateHandler struct {
		fsm             *programFSMController
		fanPower        *PowerController
		heaterPower     *PowerController
		humidifierPower *PowerController
	}

	acclimateStateHandler struct {
		fsm             *programFSMController
		fanPower        *PowerController
		heaterPower     *PowerController
		humidifierPower *PowerController
	}

	coolDownStateHandler struct {
		fsm             *programFSMController
		fanPower        *PowerController
		heaterPower     *PowerController
		humidifierPower *PowerController
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

	fsmTemperatures struct {
		updated int64
		reading temperatureReadings
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
	h.fsm.started = time.Now().Unix()
	h.fsm.stopped = 0
	log.Info("FSM: Entered start state - program started at %s", time.Unix(h.fsm.started, 0).Format(time.RFC3339))
}

func (h *waitingStateHandler) executeState() fsmState {
	// Make sure we have received updated temperature and psu status
	if h.fsm.currentPSUStatus.updated >= h.fsm.started && h.fsm.currentTemperatures.updated >= h.fsm.started {
		log.Debug("FSM: waiting state - sensors ready (PSU: %s, Temp: %s), transitioning to preheat",
			time.Unix(h.fsm.currentPSUStatus.updated, 0).Format(time.RFC3339),
			time.Unix(h.fsm.currentTemperatures.updated, 0).Format(time.RFC3339))
		return fsmStatePreHeat
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
	log.Info("FSM: Entered waiting state - waiting for initial sensor data")
}

func (h *preHeatStateHandler) executeState() fsmState {
	// If the material is warmer than the oven, we start heating the oven immediately
	if h.fsm.currentTemperatures.reading.Oven < h.fsm.currentTemperatures.reading.Material {
		log.Trace("FSM: preheat state - oven (%.1f°C) < material (%.1f°C), heating required",
			h.fsm.currentTemperatures.reading.Oven, h.fsm.currentTemperatures.reading.Material)
		if h.fsm.psuStatus.reading.Heater.Percent == 0 {
			log.Debug("FSM: preheat state - setting heater to 100%%")
			h.fsm.psuController.setPower(psuOven, 100)
		}
		return fsmStatePreHeat
	}
	// Once the oven is at the same or higher temperature than the material
	// we can start the program
	log.Debug("FSM: preheat state - oven (%.1f°C) >= material (%.1f°C), ready to start program",
		h.fsm.currentTemperatures.reading.Oven, h.fsm.currentTemperatures.reading.Material)
	return fsmStateNextProgramStep
}

func (h *preHeatStateHandler) enterState() {
	// For preheat we turn on the fan
	log.Info("FSM: Entered preheat state - setting fan to 50%%")
	h.fsm.psuController.setPower(psuFan, 50)
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
		log.Debug("FSM: heat_up - updating power (oven: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		heaterPower := h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, heaterPower)
		log.Trace("FSM: heat_up - heater power: %d%%", heaterPower)

		fanPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Fan != nil && h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power != nil {
			fanPower = *h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power
		}
		fanResult := h.fanPower.Update(fanPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuFan, fanResult)
		log.Trace("FSM: heat_up - fan power: %d%%", fanResult)

		humidifierPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Humidifier != nil && h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power != nil {
			humidifierPower = *h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power
		}
		humidifierResult := h.humidifierPower.Update(humidifierPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuHumidifier, humidifierResult)
		log.Trace("FSM: heat_up - humidifier power: %d%%", humidifierResult)

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateHeatUp
}

func (h *heatUpStateHandler) enterState() {
	log.Info("FSM: Entered heat_up state - target: %d°C", h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	h.fanPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Fan)
	h.heaterPower = NewPowerController(float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), h.fsm.program.ProgramSteps[h.fsm.step].Heater)
	h.humidifierPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Humidifier)
}

func (h *acclimateStateHandler) executeState() fsmState {
	// Once we have been acclimating long enough, we can move to the next step
	elapsed := time.Now().Unix() - h.fsm.stepStarted
	required := int64(h.fsm.program.ProgramSteps[h.fsm.step].Runtime.Seconds())
	if elapsed >= required {
		log.Info("FSM: acclimate - runtime complete (%ds / %ds)", elapsed, required)
		return fsmStateNextProgramStep
	}
	log.Trace("FSM: acclimate - maintaining temperature (%ds / %ds, material: %.1f°C, target: %d°C)",
		elapsed, required, h.fsm.temperatures.reading.Material, h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		log.Debug("FSM: acclimate - updating power (oven: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		fanPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Fan != nil && h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power != nil {
			fanPower = *h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power
		}
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(fanPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		humidifierPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Humidifier != nil && h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power != nil {
			humidifierPower = *h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power
		}
		h.fsm.psuController.setPower(psuHumidifier, h.humidifierPower.Update(humidifierPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateAcclimate
}

func (h *acclimateStateHandler) enterState() {
	log.Info("FSM: Entered acclimate state - target: %d°C, duration: %.0fs",
		h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature,
		h.fsm.program.ProgramSteps[h.fsm.step].Runtime.Seconds())
	h.fanPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Fan)
	h.heaterPower = NewPowerController(float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), h.fsm.program.ProgramSteps[h.fsm.step].Heater)
	h.humidifierPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Humidifier)
}

func (h *coolDownStateHandler) executeState() fsmState {
	// If we have been cooling down long enough, we can move to the next step
	elapsed := time.Now().Unix() - h.fsm.stepStarted
	if h.fsm.program.ProgramSteps[h.fsm.step].Runtime != nil {
		required := int64(h.fsm.program.ProgramSteps[h.fsm.step].Runtime.Seconds())
		if elapsed >= required {
			log.Info("FSM: cool_down - runtime limit reached (%ds / %ds)", elapsed, required)
			return fsmStateNextProgramStep
		}
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
		log.Debug("FSM: cool_down - updating power (oven: %.1f°C, material: %.1f°C)",
			h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material)
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		fanPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Fan != nil && h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power != nil {
			fanPower = *h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power
		}
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(fanPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		humidifierPower := uint8(0)
		if h.fsm.program.ProgramSteps[h.fsm.step].Humidifier != nil && h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power != nil {
			humidifierPower = *h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power
		}
		h.fsm.psuController.setPower(psuHumidifier, h.humidifierPower.Update(humidifierPower, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))

		// Mark these temperature readings as processed
		h.fsm.temperatures.updated = h.fsm.currentTemperatures.updated
	}
	return fsmStateCoolDown
}

func (h *coolDownStateHandler) enterState() {
	log.Info("FSM: Entered cool_down state - target: %d°C", h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature)
	h.fanPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Fan)
	h.heaterPower = NewPowerController(float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), h.fsm.program.ProgramSteps[h.fsm.step].Heater)
	h.humidifierPower = NewPowerController(0, h.fsm.program.ProgramSteps[h.fsm.step].Humidifier)
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
			types.StepTypeHeating:   fsmStateHeatUp,
			types.StepTypeAcclimate: fsmStateAcclimate,
			types.StepTypeCooling:   fsmStateCoolDown,
		},
		defaults: defaults,
	}
	controller.stateHandlers = map[fsmState]fsmStateHandler{
		fsmStateStart:           &startStateHandler{fsm: controller},
		fsmStateNextProgramStep: &nextProgramStepHandler{fsm: controller},
		fsmStateIdle:            &idleStateHandler{fsm: controller},
		fsmStateWaiting:         &waitingStateHandler{fsm: controller},
		fsmStatePreHeat:         &preHeatStateHandler{fsm: controller},
		fsmStateHeatUp:          &heatUpStateHandler{fsm: controller},
		fsmStateAcclimate:       &acclimateStateHandler{fsm: controller},
		fsmStateCoolDown:        &coolDownStateHandler{fsm: controller},
		fsmStateFailed:          &failedStateHandler{fsm: controller},
	}
	return controller
}

func (p *programFSMController) executeTick() {
	// Reached the end of the program
	if p.Completed() {
		log.Trace("FSM: executeTick - program completed, no action")
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
		log.Info("FSM: Shutting down - turning off all power")
		p.psuController.setPower(psuOven, 0)
		p.psuController.setPower(psuFan, 0)
		p.psuController.setPower(psuHumidifier, 0)
		log.Debug("FSM: Shutdown complete at %d", p.stopped)
	}
}

func (p *programFSMController) Start(program *types.Program) {
	p.program = program
	p.state = fsmStateStart
	p.numberOfSteps = len(program.ProgramSteps)
	log.Info("FSM: Starting program '%s' with %d steps", program.ProgramName, p.numberOfSteps)
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

	// Set current step name with bounds checking
	switch {
	case p.step >= 0 && p.step < p.numberOfSteps:
		status.CurrentStep = p.program.ProgramSteps[p.step].Name
	case p.step < 0:
		status.CurrentStep = "Initializing"
	default:
		status.CurrentStep = "Completed"
	}

	status.Temperatures.Material = p.temperatures.reading.Material
	status.Temperatures.Oven = p.temperatures.reading.Oven
	status.PowerStatus.Heater = int8(p.psuStatus.reading.Heater.Percent)
	status.PowerStatus.Fan = int8(p.psuStatus.reading.Fan.Percent)
	status.PowerStatus.Humidifier = int8(p.psuStatus.reading.Humidifier.Percent)
}
