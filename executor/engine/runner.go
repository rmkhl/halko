package engine

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rmkhl/halko/executor/storagefs"
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
		statusWriter               *storagefs.StateWriter
		logWriter                  *storagefs.ExecutionLogWriter
		// Note, we rely on the fact that the runner is the only one updating these and fsmController
		// relies on the fact that they will not be updated while executeStep() or updateStatus() is running.
		psuStatus         fsmPSUStatus
		temperatureStatus fsmTemperatures

		defaults *types.Defaults
	}
)

func newProgramRunner(halkoConfig *types.HalkoConfig, programStorage *storagefs.ExecutorFileStorage, program *types.Program, endpoints *types.APIEndpoints) (*programRunner, error) {
	runner := programRunner{
		wg:                         new(sync.WaitGroup),
		active:                     false,
		temperatureSensorCommands:  make(chan string),
		temperatureSensorResponses: make(chan temperatureReadings),
		psuSensorCommands:          make(chan string),
		psuSensorResponses:         make(chan psuReadings),
		currentProgram:             program,
		programStatus:              &types.ExecutionStatus{Program: *program},
		defaults:                   halkoConfig.ExecutorConfig.Defaults,
	}

	if halkoConfig.APIEndpoints == nil {
		return nil, errors.New("API endpoints not configured")
	}

	psuSensorReader, err := newPSUSensorReader(endpoints.PowerUnit.GetPowerURL(), runner.psuSensorCommands, runner.psuSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.psuSensorReader = psuSensorReader

	temperatureSensorReader, err := newTemperatureSensorReader(endpoints.SensorUnit.GetTemperaturesURL(), runner.temperatureSensorCommands, runner.temperatureSensorResponses)
	if err != nil {
		return nil, err
	}
	runner.temperatureSensorReader = temperatureSensorReader
	psuController, err := newPSUController(halkoConfig, endpoints)
	if err != nil {
		return nil, err
	}

	runner.fsmController = newProgramFSMController(psuController, &runner.psuStatus, &runner.temperatureStatus, runner.defaults)

	programName := fmt.Sprintf("%s@%s", program.ProgramName, time.Now().Format(time.RFC3339))
	err = programStorage.CreateExecutedProgram(programName, program)
	if err != nil {
		return nil, err
	}
	runner.statusWriter = storagefs.NewStateWriter(programStorage, programName)
	err = runner.statusWriter.UpdateState(types.ProgramStatePending)
	if err != nil {
		return nil, err
	}
	runner.logWriter = storagefs.NewExecutionLogWriter(programStorage, programName, 60)
	return &runner, nil
}

func (runner *programRunner) Run() {
	ticker := time.NewTicker(6000 * time.Millisecond)
	defer runner.wg.Done()

	_ = runner.statusWriter.UpdateState(types.ProgramStateRunning)
	runner.fsmController.Start(runner.currentProgram)
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
			runner.fsmController.executeTick()
		case psuState := <-runner.psuSensorResponses:
			runner.psuStatus.updated = time.Now().Unix()
			runner.psuStatus.reading = psuState
		case temperatures := <-runner.temperatureSensorResponses:
			runner.temperatureStatus.updated = time.Now().Unix()
			runner.temperatureStatus.reading = temperatures
		}
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

	runner.fsmController.Start(runner.currentProgram)
	go runner.Run()
}
