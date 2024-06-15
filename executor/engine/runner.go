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
		logWriter                  *storage.ExecutionLogWriter
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
	runner.logWriter = storage.NewExecutionLogWriter(programStorage, programName, 60)
	return &runner, nil
}

func (runner *programRunner) Run() {
	ticker := time.NewTicker(6000 * time.Millisecond)
	defer runner.wg.Done()

	runner.statusWriter.UpdateState(types.ProgramStateRunning)
	for runner.active && !runner.fsmController.Completed() {
		defer ticker.Stop()
		now := time.Now().Unix()
		select {
		case <-ticker.C:
			runner.currentProgram.mutex.RLock()
			if now-runner.currentProgram.temperatures.updated > 30 {
				runner.temperatureSensorCommands <- sensorRead
			}
			if now-runner.currentProgram.psuStatus.updated > 30 {
				runner.psuSensorCommands <- sensorRead
			}
			runner.currentProgram.mutex.RUnlock()
			runner.fsmCommands <- programStep
		case psuState := <-runner.psuSensorResponses:
			runner.currentProgram.mutex.Lock()
			runner.currentProgram.psuStatus.updated = time.Now().Unix()
			runner.currentProgram.psuStatus.reading = psuState
			runner.currentProgram.mutex.Unlock()
		case temperatures := <-runner.temperatureSensorResponses:
			runner.currentProgram.mutex.Lock()
			runner.currentProgram.temperatures.updated = time.Now().Unix()
			runner.currentProgram.temperatures.reading = temperatures
			runner.currentProgram.mutex.Unlock()
		}
		// Update program status
		runner.fsmController.UpdateStatus(runner.programStatus)
		runner.logWriter.AddLine(runner.programStatus)
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
	runner.logWriter.Close()
	runner.fsmCommands <- programDone
	runner.psuSensorCommands <- controllerDone
	runner.temperatureSensorCommands <- controllerDone
}

func (runner *programRunner) Stop() {
	runner.active = false

	runner.wg.Wait()
}

func (runner *programRunner) Start() {
	runner.active = true
	runner.wg.Add(4)
	go runner.psuSensorReader.Run(runner.wg)
	go runner.temperatureSensorReader.Run(runner.wg)
	go runner.fsmController.Run(runner.wg, runner.currentProgram)
	go runner.Run()
}
