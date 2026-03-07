package engine

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rmkhl/halko/controlunit/heartbeat"
	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
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
		previousStep               string
		heartbeatManager           *heartbeat.Manager
		programName                string
		programStorage             *storagefs.ExecutorFileStorage
		// Note, we rely on the fact that the runner is the only one updating these and fsmController
		// relies on the fact that they will not be updated while executeStep() or updateStatus() is running.
		psuStatus         fsmPSUStatus
		temperatureStatus fsmTemperatures

		defaults    *types.Defaults
		halkoConfig *types.HalkoConfig
	}
)

func newProgramRunner(halkoConfig *types.HalkoConfig, programStorage *storagefs.ExecutorFileStorage, program *types.Program, endpoints *types.APIEndpoints, heartbeatMgr *heartbeat.Manager) (*programRunner, error) {
	runner := programRunner{
		wg:                         new(sync.WaitGroup),
		active:                     false,
		temperatureSensorCommands:  make(chan string),
		temperatureSensorResponses: make(chan temperatureReadings),
		psuSensorCommands:          make(chan string),
		psuSensorResponses:         make(chan psuReadings),
		currentProgram:             program,
		programStatus:              &types.ExecutionStatus{Program: *program},
		defaults:                   halkoConfig.ControlUnitConfig.Defaults,
		heartbeatManager:           heartbeatMgr,
		programStorage:             programStorage,
		halkoConfig:                halkoConfig,
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
	runner.previousStep = ""

	programName := fmt.Sprintf("%s@%s", program.ProgramName, time.Now().Format(time.RFC3339))
	runner.programName = programName
	err = programStorage.CreateExecutedProgram(programName, program)
	if err != nil {
		return nil, err
	}
	runner.statusWriter = storagefs.NewStateWriter(programStorage, programName)
	err = runner.statusWriter.UpdateState(types.ProgramStatePending)
	if err != nil {
		return nil, err
	}

	// Capture start time once to ensure ExecutionLogWriter and FSM use the same timestamp
	startTime := time.Now().Unix()
	runner.logWriter = storagefs.NewExecutionLogWriter(programStorage, programName, 60, startTime)
	return &runner, nil
}

func (runner *programRunner) Run() {
	tickLength, err := time.ParseDuration(runner.halkoConfig.ControlUnitConfig.TickLength)
	if err != nil {
		log.Fatal("Invalid tick_length in configuration: %v", err)
	}
	log.Info("Program runner using tick length: %v", tickLength)
	ticker := time.NewTicker(tickLength)
	defer ticker.Stop()
	defer runner.wg.Done()

	_ = runner.statusWriter.UpdateState(types.ProgramStateRunning)
	// Note: fsmController.Start() was already called in Start() method
	for runner.active && !runner.fsmController.Completed() {
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

		// Update display if current step changed
		if runner.programStatus.CurrentStep != runner.previousStep {
			runner.updateDisplay(runner.programStatus.CurrentStep)
			runner.previousStep = runner.programStatus.CurrentStep
		}

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

	// Reset display to idle
	runner.heartbeatManager.SetDisplayMessage("idle")

	// Move files from running to history
	runner.statusWriter.MarkCompleted()
	if err := runner.programStorage.MoveToHistory(runner.programName); err != nil {
		log.Error("Failed to move program files to history: %v", err)
	}

	log.Debug("Runner: Signaling sensor readers to stop")

	// Drain any pending responses first to prevent deadlock
	// Sensor readers might be blocked trying to send responses
	for {
		select {
		case <-runner.psuSensorResponses:
			log.Trace("Runner: Drained pending PSU response")
		case <-runner.temperatureSensorResponses:
			log.Trace("Runner: Drained pending temperature response")
		default:
			// No more pending responses, proceed to send done signals
			goto sendDone
		}
	}

sendDone:
	// Now safe to send done signals with timeout
	select {
	case runner.psuSensorCommands <- controllerDone:
		log.Trace("Runner: Sent done signal to PSU sensor reader")
	case <-time.After(1 * time.Second):
		log.Warning("Runner: Timeout sending done signal to PSU sensor reader")
	}

	select {
	case runner.temperatureSensorCommands <- controllerDone:
		log.Trace("Runner: Sent done signal to temperature sensor reader")
	case <-time.After(1 * time.Second):
		log.Warning("Runner: Timeout sending done signal to temperature sensor reader")
	}

	log.Debug("Runner: Run() method completing")
}

// updateDisplay sets the display message via heartbeat manager
func (runner *programRunner) updateDisplay(stepName string) {
	if runner.heartbeatManager == nil {
		log.Trace("Runner: Heartbeat manager not available, skipping display update")
		return
	}

	runner.heartbeatManager.SetDisplayMessage(stepName)
	log.Debug("Runner: Display message updated to: %s", stepName)
}

func (runner *programRunner) Stop() {
	runner.active = false
	// Don't wait here - let the runner complete asynchronously
	// The monitoring goroutine in engine.go will clean up engine.runner after completion
}

func (runner *programRunner) Start() {
	runner.active = true
	runner.wg.Add(3)
	go runner.psuSensorReader.Run(runner.wg)
	go runner.temperatureSensorReader.Run(runner.wg)

	runner.fsmController.Start(runner.currentProgram, runner.logWriter.GetStartTime())
	go runner.Run()
}
