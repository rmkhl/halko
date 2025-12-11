package engine

import (
	"errors"
	"sync"

	"github.com/rmkhl/halko/controlunit/storagefs"
	"github.com/rmkhl/halko/types"
)

type (
	ControlEngine struct {
		wg          *sync.WaitGroup
		config      *types.ControlUnitConfig
		halkoConfig *types.HalkoConfig
		storage     *storagefs.ExecutorFileStorage
		runner      *programRunner
		endpoints   *types.APIEndpoints
	}
)

var (
	ErrProgramAlreadyRunning = errors.New("program already running")
	ErrNoProgramRunning      = errors.New("no program running")
)

func NewEngine(halkoConfig *types.HalkoConfig, storage *storagefs.ExecutorFileStorage, endpoints *types.APIEndpoints) *ControlEngine {
	engine := ControlEngine{
		halkoConfig: halkoConfig,
		config:      halkoConfig.ControlUnitConfig,
		runner:      nil,
		storage:     storage,
		endpoints:   endpoints,
		wg:          new(sync.WaitGroup),
	}

	return &engine
}

func (engine *ControlEngine) CurrentStatus() *types.ExecutionStatus {
	if engine.runner == nil {
		return nil
	}

	return engine.runner.programStatus
}

func (engine *ControlEngine) GetDefaults() *types.Defaults {
	return engine.config.Defaults
}

func (engine *ControlEngine) StartEngine(program *types.Program) error {
	if engine.runner != nil {
		return ErrProgramAlreadyRunning
	}

	runner, err := newProgramRunner(engine.halkoConfig, engine.storage, program, engine.endpoints)
	if err != nil {
		return err
	}

	engine.runner = runner
	engine.wg.Add(1)
	engine.runner.Start()

	return nil
}

func (engine *ControlEngine) StopEngine() error {
	if engine.runner != nil {
		engine.runner.Stop()
		engine.wg.Done()
		engine.runner = nil
		return nil
	}
	return ErrNoProgramRunning
}
