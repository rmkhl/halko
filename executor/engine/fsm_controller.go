package engine

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
)

const (
	fsmStateIdle         fsmState = "idle"
	fsmStateWaiting      fsmState = "waiting"
	fsmStateCheckPreHeat fsmState = "check_preheat"
	fsmStatePreHeat      fsmState = "preheat"
	fsmStateGoingUp      fsmState = "going_up"
	fsmStateGoingDown    fsmState = "going_down"
	fsmStateFailed       fsmState = "failed"

	deltaDirectionUp   deltaDirection = "up"
	deltaDirectionDown deltaDirection = "down"
)

type (
	fsmState       string
	deltaDirection string

	CurrentProgram struct {
		mutex         sync.RWMutex
		program       *types.Program
		numberOfSteps int
		psuStatus     struct {
			updated int64
			reading psuReadings
		}
		temperatures struct {
			updated int64
			reading temperatureReadings
		}
	}

	programFSMController struct {
		step                        int
		stepStarted                 int64
		started                     int64
		stopped                     int64
		state                       fsmState
		previousMaterialTemperature float32
		previousDirection           deltaDirection
		program                     *CurrentProgram
		runner                      <-chan string
		psuController               *psuController
	}
)

func newCurrentProgram(program *types.Program) *CurrentProgram {
	return &CurrentProgram{program: program, numberOfSteps: len(program.ProgramSteps)}
}

func newProgramFSMController(psuController *psuController, runner <-chan string) *programFSMController {
	return &programFSMController{
		runner:        runner,
		psuController: psuController,
	}
}

// Determine the delta between the material and oven temperatures.
// Also determine the direction of the delta.
func (p *programFSMController) determineDelta() (float32, deltaDirection) {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	delta := p.program.temperatures.reading.Material - p.program.temperatures.reading.Oven
	materialDelta := p.program.temperatures.reading.Material - p.previousMaterialTemperature

	// Unless the temperature is changing by more than 1 degree, we consider it unchanged.
	if math.Abs(float64(materialDelta)) < 1.0 {
		return delta, p.previousDirection
	}
	p.previousMaterialTemperature = p.program.temperatures.reading.Material

	// If the material temperature is going down, we consider the delta to be going down.
	if materialDelta < 0 {
		p.previousDirection = deltaDirectionDown
		return delta, deltaDirectionDown
	}
	p.previousDirection = deltaDirectionUp
	return delta, deltaDirectionUp
}

func (p *programFSMController) startPhase(state fsmState, n int) {
	// If the program has been stopped, we don't want to start any new phases.
	if state == fsmStateIdle || state == fsmStateFailed {
		p.state = state
		return
	}

	// If the step is out of bounds, program has completed normally.
	if n >= p.program.numberOfSteps {
		p.state = fsmStateIdle
		return
	}

	p.state = state
	p.step = n
	p.stepStarted = time.Now().Unix()
}

// Waiting to get sensor readings.
func (p *programFSMController) stateWaiting() {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	// Only proceed if we have sensor readings.
	if p.program.temperatures.updated != 0 && p.program.psuStatus.updated != 0 {
		p.state = fsmStateCheckPreHeat
		return
	}
}

// Check if preheating is needed.
func (p *programFSMController) stateCheckPreHeat() {
	delta, _ := p.determineDelta()

	// if the initial oven temperature is higher than the material temperature, we can start the program.
	if delta < 0 {
		p.startPhase(fsmStateGoingUp, 0)
		return
	}
	// if the initial oven temperature is lower than the material temperature, we need to preheat.
	p.psuController.setPower(psuOven, 100)
	p.psuController.setPower(psuFan, 100)
	p.startPhase(fsmStatePreHeat, 0)
}

// Preheat the oven. (Oven temperature needs to be higher than the material temperature.)
func (p *programFSMController) statePreHeat() {
	delta, _ := p.determineDelta()

	// if the oven temperature is higher than the material temperature, we can start the program.
	if delta < 0 {
		p.startPhase(fsmStateGoingUp, 0)
		return
	}
}

// Regardless of end state we need keep setting the power off as long as the program is not stopped.
func (p *programFSMController) stateIdleOrFailed() {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	if p.stopped != 0 {
		return
	}
	// Check that all power is off.
	if p.program.psuStatus.reading.Heater.Status == PowerOff && p.program.psuStatus.reading.Fan.Status == PowerOff && p.program.psuStatus.reading.Humidifier.Status == PowerOff {
		p.stopped = time.Now().Unix()
		return
	}
	// If not, turn off all power.
	p.psuController.setPower(psuOven, 0)
	p.psuController.setPower(psuFan, 0)
	p.psuController.setPower(psuHumidifier, 0)

}

// Check if the current program step has completed.
// End conditions are:
// - The step has run for the maximum allowed time.
// - The material temperature has reached the target temperature.
//
// Program is considered failed if:
// - The material temperature has not reached the target temperature before the maximum allowed time.
// - The material temperature is outside the valid range when acclimating.
func (p *programFSMController) stepCompleted(currentState fsmState) (bool, fsmState) {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	maximumRuntime := p.program.program.ProgramSteps[p.step].MaximumRuntime
	if maximumRuntime == 0 {
		maximumRuntime = p.program.program.DefaultStepTime
	}

	currentTime := time.Now().Unix()
	stepTimeExceeded := p.stepStarted+int64(maximumRuntime) >= currentTime

	switch p.program.program.ProgramSteps[p.step].StepType {
	case types.StepTypeHeating:
		targetTemperature := p.program.program.ProgramSteps[p.step].ValidRange.MaximumTemperature
		if p.program.temperatures.reading.Material >= targetTemperature {
			return true, currentState
		}
		if stepTimeExceeded {
			return true, fsmStateFailed
		}
	case types.StepTypeCooling:
		targetTemperature := p.program.program.ProgramSteps[p.step].ValidRange.MinimumTemperature
		if p.program.temperatures.reading.Material <= targetTemperature {
			return true, currentState
		}
		if stepTimeExceeded {
			return true, fsmStateFailed
		}
	case types.StepTypeAcclimate:
		minimumTemperature := p.program.program.ProgramSteps[p.step].ValidRange.MinimumTemperature
		maximumTemperature := p.program.program.ProgramSteps[p.step].ValidRange.MaximumTemperature
		if p.program.temperatures.reading.Material < minimumTemperature && p.program.temperatures.reading.Material > maximumTemperature {
			return false, fsmStateFailed
		}
		if stepTimeExceeded {
			return true, currentState
		}
	case types.StepTypeWaiting:
		if stepTimeExceeded {
			return true, currentState
		}
	default: // Unknown step type
		log.Printf("Unknown step type: %s\n", p.program.program.ProgramSteps[p.step].StepType)
		return true, fsmStateFailed
	}
	return false, currentState
}

// Find the power settings that match the delta.
// To find the correct cycle if we are going up, we need to find the cycle with the lowest delta that is
// higher than the current delta and return the power setting for below (we crossed the threshold from below).
// If we are going down, we need to find the cycle with the highest delta that is lower than the current delta
// and return the power setting for above (we crossed the threshold from above).
//
// cycles are sorted by delta and are assumed to have both below and above power settings.
func findMatchingPowerSettings(cycles []types.DeltaCycle, delta float32, direction deltaDirection) int {
	if delta < cycles[0].TemperatureDelta {
		if direction == deltaDirectionUp {
			return cycles[0].EnteredBelow
		}
		return cycles[0].EnteredAbove
	}
	nCycles := len(cycles)
	for i := 1; i < nCycles; i++ {
		if delta < cycles[i].TemperatureDelta {
			if direction == deltaDirectionUp {
				return cycles[i-1].EnteredBelow
			}
			return cycles[i-1].EnteredAbove
		}
	}
	if direction == deltaDirectionUp {
		return cycles[nCycles-1].EnteredBelow
	}
	return cycles[nCycles-1].EnteredAbove
}

// Material temperature is heating or cooling
func (p *programFSMController) stateGoingUpOrDown() {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	completed, nextState := p.stepCompleted(p.state)
	if completed {
		p.startPhase(nextState, p.step+1)
		return
	}
	// if the state changed, we need to start a new phase.
	if nextState != p.state {
		p.state = nextState
		return
	}
	// adjust power settings based on the delta
	delta, direction := p.determineDelta()
	heaterPowerSetting := p.program.program.ProgramSteps[p.step].Heater.ConstantCycle
	if p.program.program.ProgramSteps[p.step].Heater.DeltaCycles != nil {
		heaterPowerSetting = findMatchingPowerSettings(p.program.program.ProgramSteps[p.step].Heater.DeltaCycles, delta, direction)
	}
	fanPowerSetting := p.program.program.ProgramSteps[p.step].Fan.ConstantCycle
	if p.program.program.ProgramSteps[p.step].Fan.DeltaCycles != nil {
		fanPowerSetting = findMatchingPowerSettings(p.program.program.ProgramSteps[p.step].Fan.DeltaCycles, delta, direction)
	}
	humidityPowerSetting := p.program.program.ProgramSteps[p.step].Humidifier.ConstantCycle
	if p.program.program.ProgramSteps[p.step].Humidifier.DeltaCycles != nil {
		humidityPowerSetting = findMatchingPowerSettings(p.program.program.ProgramSteps[p.step].Humidifier.DeltaCycles, delta, direction)
	}
	p.psuController.setPower(psuOven, heaterPowerSetting)
	p.psuController.setPower(psuFan, fanPowerSetting)
	p.psuController.setPower(psuHumidifier, humidityPowerSetting)
}

func (p *programFSMController) executeStep() {
	switch p.state {
	case fsmStateFailed:
		p.stateIdleOrFailed()
	case fsmStateIdle:
		p.stateIdleOrFailed()
	case fsmStateWaiting:
		p.stateWaiting()
	case fsmStateCheckPreHeat:
		p.stateCheckPreHeat()
	case fsmStatePreHeat:
		p.statePreHeat()
	case fsmStateGoingUp:
		p.stateGoingUpOrDown()
	case fsmStateGoingDown:
		p.stateGoingUpOrDown()
	}
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

func (p *programFSMController) Run(wg *sync.WaitGroup, program *CurrentProgram) {
	defer wg.Done()

	p.program = program
	p.state = fsmStateWaiting
	p.started = time.Now().Unix()

	for {
		command := <-p.runner
		switch command {
		case programStep:
			p.executeStep()
		case programDone:
			p.shutdown()
			return
		}
	}
}

func (p *programFSMController) Completed() bool {
	return p.state == fsmStateFailed || p.state == fsmStateIdle
}

func (p *programFSMController) Failed() bool {
	return p.state == fsmStateFailed
}

func (p *programFSMController) UpdateStatus(status *types.ExecutionStatus) {
	p.program.mutex.RLock()
	defer p.program.mutex.RUnlock()

	status.StartedAt = p.started
	status.CurrentStepStartedAt = p.stepStarted
	status.CurrentStep = p.program.program.ProgramSteps[p.step].Name
	status.Temperatures.Material = p.program.temperatures.reading.Material
	status.Temperatures.Oven = p.program.temperatures.reading.Oven
	status.Temperatures.Delta = p.program.temperatures.reading.Material - p.program.temperatures.reading.Oven
	status.PowerStatus.Heater = p.program.psuStatus.reading.Heater.Percent
	status.PowerStatus.Fan = p.program.psuStatus.reading.Fan.Percent
	status.PowerStatus.Humidifier = p.program.psuStatus.reading.Humidifier.Percent
}
