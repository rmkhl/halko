package engine

import (
	"errors"
	"sync"

	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/types"
)

type (
	ControlEngine struct {
		wg      *sync.WaitGroup
		config  *types.ExecutorConfig
		storage *storage.FileStorage
		runner  *programRunner
	}
)

var (
	ErrProgramAlreadyRunning = errors.New("program already running")
	ErrNoProgramRunning      = errors.New("no program running")
)

func NewEngine(config *types.ExecutorConfig, storage *storage.FileStorage) *ControlEngine {
	engine := ControlEngine{
		config:  config,
		runner:  nil,
		storage: storage,
		wg:      new(sync.WaitGroup),
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

	runner, err := newProgramRunner(engine.config, engine.storage, program)
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
