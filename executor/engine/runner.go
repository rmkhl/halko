package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
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
		currentProgram             *types.Program
		fsmController              *programFSMController
		psuSensorCommands          chan string
		psuSensorResponses         chan psuReadings
		psuSensorReader            *psuSensorReader
		temperatureSensorCommands  chan string
		temperatureSensorResponses chan temperatureReadings
		temperatureSensorReader    *temperatureSensorReader
		programStatus              *types.ExecutionStatus
		statusWriter               *storage.StateWriter
		logWriter                  *storage.ExecutionLogWriter
		// Note, we rely on the fact that the runner is the only one updating these and fsmController
		// relies on the fact that they will not be updated while executeStep() or updateStatus() is running.
		psuStatus         fsmPSUStatus
		temperatureStatus fsmTemperatures
	}
)

func newProgramRunner(config *types.ExecutorConfig, programStorage *storage.ProgramStorage, program *types.Program) (*programRunner, error) {
	runner := programRunner{
		wg:                         new(sync.WaitGroup),
		active:                     false,
		temperatureSensorCommands:  make(chan string),
		temperatureSensorResponses: make(chan temperatureReadings),
		psuSensorCommands:          make(chan string),
		psuSensorResponses:         make(chan psuReadings),
		currentProgram:             program,
		programStatus:              &types.ExecutionStatus{Program: *program},
	}

	psuSensorReader, err := newPSUSensorReader(config.PowerSensorURL, runner.psuSensorCommands, runner.psuSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.psuSensorReader = psuSensorReader

	temperatureSensorReader, err := newTemperatureSensorReader(config.TemperatureSensorURL, runner.temperatureSensorCommands, runner.temperatureSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.temperatureSensorReader = temperatureSensorReader
	psuController, err := newPSUController(config)
	if err != nil {
		return nil, err
	}

	runner.fsmController = newProgramFSMController(psuController, &runner.psuStatus, &runner.temperatureStatus)

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

	_ = runner.statusWriter.UpdateState(types.ProgramStateRunning)
	for runner.active && !runner.fsmController.Completed() {
		defer ticker.Stop()
		now := time.Now().Unix()
		select {
		case <-ticker.C:
			if now-runner.temperatureStatus.updated > 30 {
				runner.temperatureSensorCommands <- sensorRead
			}
			if now-runner.psuStatus.updated > 30 {
				runner.psuSensorCommands <- sensorRead
			}
			runner.fsmController.executeStep()
		case psuState := <-runner.psuSensorResponses:
			runner.psuStatus.updated = time.Now().Unix()
			runner.psuStatus.reading = psuState
		case temperatures := <-runner.temperatureSensorResponses:
			runner.temperatureStatus.updated = time.Now().Unix()
			runner.temperatureStatus.reading = temperatures
		}
		// Update program status
		runner.fsmController.UpdateStatus(runner.programStatus)
		runner.logWriter.AddLine(runner.programStatus)
	}
	if runner.fsmController.Completed() {
		if runner.fsmController.Failed() {
			_ = runner.statusWriter.UpdateState(types.ProgramStateFailed)
		} else {
			_ = runner.statusWriter.UpdateState(types.ProgramStateCompleted)
		}
	} else {
		_ = runner.statusWriter.UpdateState(types.ProgramStateCanceled)
	}
	runner.logWriter.Close()
	runner.fsmController.shutdown()
	runner.psuSensorCommands <- controllerDone
	runner.temperatureSensorCommands <- controllerDone
}

func (runner *programRunner) Stop() {
	runner.active = false

	runner.wg.Wait()
}

func (runner *programRunner) Start() {
	runner.active = true
	runner.wg.Add(3)
	go runner.psuSensorReader.Run(runner.wg)
	go runner.temperatureSensorReader.Run(runner.wg)
	runner.fsmController.Reset(runner.currentProgram)
	go runner.Run()
}
