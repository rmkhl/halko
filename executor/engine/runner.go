package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/executor/types"
)

const (
	controllerDone = "done"
	sensorRead     = "read"
	programDone    = "done"
	programStep    = "step"
)

type (
	programRunner struct {
		active                     bool
		mutex                      sync.RWMutex
		wg                         *sync.WaitGroup
		currentProgram             *CurrentProgram
		fsmCommands                chan string
		fsmController              *programFSMController
		psuSensorCommands          chan string
		psuSensorResponses         chan psuReadings
		psuSensorReader            *psuSensorReader
		temperatureSensorCommands  chan string
		temperatureSensorResponses chan temperatureReadings
		temperatureSensorReader    *temperatureSensorReader
		programStatus              *types.ProgramStatus
		statusWriter               *storage.StateWriter
	}
)

func newRunner(config *types.ExecutorConfig, programStorage *storage.ProgramStorage, program *types.Program) (*programRunner, error) {
	runner := programRunner{
		wg:                         new(sync.WaitGroup),
		active:                     false,
		currentProgram:             newCurrentProgram(program),
		temperatureSensorCommands:  make(chan string),
		temperatureSensorResponses: make(chan temperatureReadings),
		psuSensorCommands:          make(chan string),
		psuSensorResponses:         make(chan psuReadings),
		fsmCommands:                make(chan string),
		programStatus:              &types.ProgramStatus{Program: *program},
	}

	psuSensorReader, err := newPSUSensorReader(config.PowerSensorURl, runner.psuSensorCommands, runner.psuSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.psuSensorReader = psuSensorReader

	temperatureSensorReader, err := newTemperatureSensorReader(config.TemperatureSensorURl, runner.temperatureSensorCommands, runner.temperatureSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.temperatureSensorReader = temperatureSensorReader
	psuController, err := newPSUController(config)
	if err != nil {
		return nil, err
	}

	runner.fsmController = newProgramFSMController(psuController, runner.fsmCommands)

	programName := fmt.Sprintf("%s@%s", program.ProgramName, time.Now().Format(time.RFC3339))
	err = programStorage.CreateProgram(programName, program)
	if err != nil {
		return nil, err
	}
	runner.statusWriter = storage.NewStateWriter(programStorage, programName)
	err = runner.statusWriter.UpdateState(types.ProgramStatePending)
	if err != nil {
		return nil, err
	}
	return &runner, nil
}

func (runner *programRunner) Run() {
	ticker := time.NewTicker(6000 * time.Millisecond)
	defer ticker.Stop()

	runner.mutex.RLock()
	runner.statusWriter.UpdateState(types.ProgramStateRunning)
	for runner.active {
		runner.mutex.RUnlock()
		select {
		case <-ticker.C:
			runner.mutex.RLock()
			runner.temperatureSensorCommands <- sensorRead
			runner.psuSensorCommands <- sensorRead
			runner.fsmCommands <- programStep
			if runner.fsmController.Completed() {
				runner.active = false
			}
			runner.mutex.RUnlock()
		case psuState := <-runner.psuSensorResponses:
			runner.mutex.RLock()
			runner.currentProgram.mutex.Lock()
			runner.currentProgram.psuStatus.updated = time.Now().Unix()
			runner.currentProgram.psuStatus.reading = psuState
			runner.currentProgram.mutex.Unlock()
			runner.mutex.RUnlock()
		case temperatures := <-runner.temperatureSensorResponses:
			runner.mutex.RLock()
			runner.currentProgram.mutex.Lock()
			runner.currentProgram.temperatures.updated = time.Now().Unix()
			runner.currentProgram.temperatures.reading = temperatures
			runner.currentProgram.mutex.Unlock()
			runner.mutex.RUnlock()
		}
		// Update program status
		runner.mutex.Lock()
		runner.fsmController.UpdateStatus(runner.programStatus)
		runner.mutex.Unlock()
		runner.mutex.RLock()
	}
	if runner.fsmController.Completed() {
		if runner.fsmController.Failed() {
			runner.statusWriter.UpdateState(types.ProgramStateFailed)
		} else {
			runner.statusWriter.UpdateState(types.ProgramStateCompleted)
		}
	} else {
		runner.statusWriter.UpdateState(types.ProgramStateCanceled)
	}
	runner.fsmCommands <- programDone
	runner.psuSensorCommands <- controllerDone
	runner.temperatureSensorCommands <- controllerDone
	runner.mutex.RUnlock()
}

func (runner *programRunner) Stop() {
	runner.mutex.Lock()
	runner.active = false
	runner.mutex.Unlock()

	runner.wg.Wait()
}

func (runner *programRunner) Start() {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	runner.active = true
	runner.wg.Add(4)
	go runner.psuSensorReader.Run()
	go runner.temperatureSensorReader.Run()
	go runner.fsmController.Run(runner.currentProgram)
	go runner.Run()
}
