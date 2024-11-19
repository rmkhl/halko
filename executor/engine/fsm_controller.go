package engine

import (
	"time"

	"github.com/rmkhl/halko/types"
)

const (
	fsmStateIdle      fsmState = "idle"
	fsmStateWaiting   fsmState = "waiting"
	fsmStatePreHeat   fsmState = "preheat"
	fsmStateHeatUp    fsmState = "heat_up"
	fsmStateAcclimate fsmState = "acclimate"
	fsmStateCoolDown  fsmState = "cool_down"
	fsmStateFailed    fsmState = "failed"
)

type (
	fsmState string

	fsmPSUStatus struct {
		updated int64
		reading psuReadings
	}

	fsmTemperatures struct {
		updated int64
		reading temperatureReadings
	}

	programFSMController struct {
		step        int
		stepStarted int64
		started     int64
		stopped     int64
		state       fsmState

		psuController *psuController
		psuStatus     fsmPSUStatus
		temperatures  fsmTemperatures

		// These are updated by the runner on regular bases.
		// No lock needed as the runner is the only one updating them
		// and will not update them while executeStep is running.
		currentPSUStatus    *fsmPSUStatus
		currentTemperatures *fsmTemperatures

		program       *types.Program
		numberOfSteps int
	}
)

func newProgramFSMController(psuController *psuController, psuStatus *fsmPSUStatus, temperatures *fsmTemperatures) *programFSMController {
	return &programFSMController{
		psuController:       psuController,
		currentPSUStatus:    psuStatus,
		currentTemperatures: temperatures,
	}
}

func (p *programFSMController) executeStep() {
	// Last thing record the previous psu and temperature status
	p.psuStatus = *p.currentPSUStatus
	p.temperatures = *p.currentTemperatures
}

// Shutdown the program. If the program has not completed normally we need to turn off all power.
func (p *programFSMController) shutdown() {
	if p.stopped != 0 {
		p.stopped = time.Now().Unix()
		p.psuController.setPower(psuOven, 0)
		p.psuController.setPower(psuFan, 0)
		p.psuController.setPower(psuHumidifier, 0)
	}
}

func (p *programFSMController) Reset(program *types.Program) {
	p.psuController.setPower(psuOven, 0)
	p.psuController.setPower(psuFan, 0)
	p.psuController.setPower(psuHumidifier, 0)
	p.program = program
	p.state = fsmStateWaiting
	p.started = time.Now().Unix()
	p.numberOfSteps = len(program.ProgramSteps)
	p.stopped = 0
	p.started = 0
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
