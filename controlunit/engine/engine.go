package engine

import (
	"errors"
	"sync"

	"github.com/rmkhl/halko/controlunit/heartbeat"
	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type (
	ControlEngine struct {
		mu               sync.RWMutex
		wg               *sync.WaitGroup
		config           *types.ControlUnitConfig
		halkoConfig      *types.HalkoConfig
		storage          *storagefs.ExecutorFileStorage
		runner           *programRunner
		endpoints        *types.APIEndpoints
		heartbeatManager *heartbeat.Manager
	}
)

var (
	ErrProgramAlreadyRunning = errors.New("program already running")
	ErrNoProgramRunning      = errors.New("no program running")
)

func NewEngine(halkoConfig *types.HalkoConfig, storage *storagefs.ExecutorFileStorage, endpoints *types.APIEndpoints, heartbeatMgr *heartbeat.Manager) *ControlEngine {
	engine := ControlEngine{
		halkoConfig:      halkoConfig,
		config:           halkoConfig.ControlUnitConfig,
		runner:           nil,
		storage:          storage,
		endpoints:        endpoints,
		heartbeatManager: heartbeatMgr,
		wg:               new(sync.WaitGroup),
	}

	return &engine
}

func (engine *ControlEngine) CurrentStatus() *types.ExecutionStatus {
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	if engine.runner == nil {
		return nil
	}

	return engine.runner.programStatus
}

func (engine *ControlEngine) CurrentProgramName() string {
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	if engine.runner == nil {
		return ""
	}
	return engine.runner.programName
}

func (engine *ControlEngine) GetDefaults() *types.Defaults {
	return engine.config.Defaults
}

func (engine *ControlEngine) StartEngine(program *types.Program) error {
	engine.mu.Lock()
	if engine.runner != nil {
		engine.mu.Unlock()
		return ErrProgramAlreadyRunning
	}

	runner, err := newProgramRunner(engine.halkoConfig, engine.storage, program, engine.endpoints, engine.heartbeatManager)
	if err != nil {
		engine.mu.Unlock()
		return err
	}

	engine.runner = runner
	engine.mu.Unlock()

	engine.wg.Add(1)
	runner.Start()

	// Monitor runner completion to clean up
	go func() {
		log.Debug("Engine: Waiting for runner cleanup to complete")
		runner.wg.Wait()
		log.Info("Engine: Runner cleanup complete, clearing engine state")
		engine.mu.Lock()
		engine.runner = nil
		engine.mu.Unlock()
		log.Info("Engine: No program currently running")
		engine.wg.Done()
	}()

	return nil
}

func (engine *ControlEngine) StopEngine() error {
	engine.mu.Lock()
	runner := engine.runner
	engine.mu.Unlock()

	if runner != nil {
		runner.Stop()
		// Don't set engine.runner = nil here - let the monitoring goroutine handle it
		// after the runner fully completes its cleanup
		return nil
	}
	return ErrNoProgramRunning
}

func (engine *ControlEngine) Wait() {
	engine.wg.Wait()
}
