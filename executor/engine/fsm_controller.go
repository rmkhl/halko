package engine

import (
	"time"

	"github.com/rmkhl/halko/types"
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
		defaultPids   *map[types.StepType]*types.PidSettings
	}
)

func (h *startStateHandler) executeState() fsmState {
	return fsmStateWaiting
}

func (h *startStateHandler) enterState() {
	h.fsm.started = time.Now().Unix()
	h.fsm.stopped = 0
}

func (h *waitingStateHandler) executeState() fsmState {
	// Make sure we have received updated temperature and psu status
	if h.fsm.currentPSUStatus.updated >= h.fsm.started && h.fsm.currentTemperatures.updated >= h.fsm.started {
		return fsmStatePreHeat
	}
	return fsmStateWaiting
}

func (h *waitingStateHandler) enterState() {
	h.fsm.step = -1
}

func (h *preHeatStateHandler) executeState() fsmState {
	// If the material is warmer than the oven, we start heating the oven immediately
	if h.fsm.currentTemperatures.reading.Oven < h.fsm.currentTemperatures.reading.Material {
		if h.fsm.psuStatus.reading.Heater.Percent == 0 {
			h.fsm.psuController.setPower(psuOven, 100)
		}
		return fsmStatePreHeat
	}
	// Once the oven is at the same or higher temperature than the material
	// we can start the program
	return fsmStateNextProgramStep
}

func (h *preHeatStateHandler) enterState() {
	// For preheat we turn on the fan
	h.fsm.psuController.setPower(psuFan, 50)
}

func (h *nextProgramStepHandler) executeState() fsmState {
	// Note this assumes that before first call fsm.steps is set to -1
	h.fsm.step++
	// End of the program reached
	if h.fsm.step == h.fsm.numberOfSteps-1 {
		return fsmStateIdle
	}
	h.fsm.stepStarted = time.Now().Unix()
	return h.fsm.stepToState[h.fsm.program.ProgramSteps[h.fsm.step].StepType]
}

func (h *nextProgramStepHandler) enterState() {
	h.fsm.stepStarted = time.Now().Unix()
}

func (h *heatUpStateHandler) executeState() fsmState {
	// If the target temperature is reached, we can move to the next step
	if h.fsm.temperatures.reading.Material >= float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature) {
		return fsmStateNextProgramStep
	}
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuHumidifier, h.humidifierPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
	}
	return fsmStateHeatUp
}

func (h *heatUpStateHandler) enterState() {
	h.fanPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power)
	h.heaterPower = NewPowerController(float64(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), &h.fsm.program.ProgramSteps[h.fsm.step].Heater, (*h.fsm.defaultPids)[types.StepTypeHeating])
	h.humidifierPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power)
}

func (h *acclimateStateHandler) executeState() fsmState {
	// Once we have been acclimating long enough, we can move to the next step
	if time.Now().Unix()-h.fsm.stepStarted >= int64(h.fsm.program.ProgramSteps[h.fsm.step].Duration.Seconds()) {
		return fsmStateNextProgramStep
	}
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuHumidifier, h.humidifierPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
	}
	return fsmStateAcclimate
}

func (h *acclimateStateHandler) enterState() {
	h.fanPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power)
	h.heaterPower = NewPowerController(float64(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), &h.fsm.program.ProgramSteps[h.fsm.step].Heater, (*h.fsm.defaultPids)[types.StepTypeHeating])
	h.humidifierPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power)
}

func (h *coolDownStateHandler) executeState() fsmState {
	// If we have been cooling down long enough, we can move to the next step
	if h.fsm.program.ProgramSteps[h.fsm.step].Duration != nil && time.Now().Unix()-h.fsm.stepStarted >= int64(h.fsm.program.ProgramSteps[h.fsm.step].Duration.Seconds()) {
		return fsmStateNextProgramStep
	}
	// If the wood has cooled enough (or we have reached the time limit), we can move to the next step
	if h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature != 0 && h.fsm.temperatures.reading.Material >= float32(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature) {
		return fsmStateNextProgramStep
	}
	// If we have new temperature readings, update the power settings
	if h.fsm.currentTemperatures.updated >= h.fsm.temperatures.updated {
		h.fsm.psuController.setPower(psuOven, h.heaterPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuFan, h.fanPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
		h.fsm.psuController.setPower(psuHumidifier, h.humidifierPower.Update(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power, h.fsm.temperatures.reading.Oven, h.fsm.temperatures.reading.Material))
	}
	return fsmStateCoolDown
}

func (h *coolDownStateHandler) enterState() {
	h.fanPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Fan.Power)
	h.heaterPower = NewPowerController(float64(h.fsm.program.ProgramSteps[h.fsm.step].TargetTemperature), &h.fsm.program.ProgramSteps[h.fsm.step].Heater, (*h.fsm.defaultPids)[types.StepTypeHeating])
	h.humidifierPower = newConstantPowerController(h.fsm.program.ProgramSteps[h.fsm.step].Humidifier.Power)
}

func (h *failedStateHandler) executeState() fsmState {
	// This is an end state, do not automatically transition from idle state
	return fsmStateFailed
}

func (h *failedStateHandler) enterState() {
	h.fsm.shutdown()
}

func (h *idleStateHandler) executeState() fsmState {
	// This is an end state, do not automatically transition from idle state
	return fsmStateIdle
}

func (h *idleStateHandler) enterState() {
	h.fsm.shutdown()
}

func newProgramFSMController(psuController *psuController, psuStatus *fsmPSUStatus, temperatures *fsmTemperatures, defaultPids *map[types.StepType]*types.PidSettings) *programFSMController {
	controller := &programFSMController{
		psuController:       psuController,
		currentPSUStatus:    psuStatus,
		currentTemperatures: temperatures,
		stepToState: map[types.StepType]fsmState{
			types.StepTypeHeating:   fsmStateHeatUp,
			types.StepTypeAcclimate: fsmStateAcclimate,
			types.StepTypeCooling:   fsmStateCoolDown,
		},
		defaultPids: defaultPids,
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
		return
	}

	previousState := p.state
	p.state = p.stateHandlers[p.state].executeState()
	if p.state != previousState {
		p.stepStarted = time.Now().Unix()
		p.stateHandlers[previousState].enterState()
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
		p.psuController.setPower(psuOven, 0)
		p.psuController.setPower(psuFan, 0)
		p.psuController.setPower(psuHumidifier, 0)
	}
}

func (p *programFSMController) Start(program *types.Program) {
	p.program = program
	p.state = fsmStateIdle
	p.numberOfSteps = len(program.ProgramSteps)
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
	status.CurrentStep = p.program.ProgramSteps[p.step].Name
	status.Temperatures.Material = p.temperatures.reading.Material
	status.Temperatures.Oven = p.temperatures.reading.Oven
	status.PowerStatus.Heater = int8(p.psuStatus.reading.Heater.Percent)
	status.PowerStatus.Fan = int8(p.psuStatus.reading.Fan.Percent)
	status.PowerStatus.Humidifier = int8(p.psuStatus.reading.Humidifier.Percent)
}
