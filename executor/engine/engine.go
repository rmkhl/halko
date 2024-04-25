package engine

import (
	"errors"
	"sync"

	"github.com/rmkhl/halko/executor/storage"
	"github.com/rmkhl/halko/executor/types"
)

type (
	ControlEngine struct {
		mutex   sync.RWMutex
		wg      *sync.WaitGroup
		config  *types.ExecutorConfig
		storage *storage.ProgramStorage
		runner  *programRunner
	}
)

var (
	ErrProgramAlreadyRunning = errors.New("program already running")
	ErrNoProgramRunning      = errors.New("no program running")
)

func NewEngine(config *types.ExecutorConfig, storage *storage.ProgramStorage) *ControlEngine {
	engine := ControlEngine{
		config:  config,
		runner:  nil,
		storage: storage,
		wg:      new(sync.WaitGroup),
	}

	return &engine
}

func (engine *ControlEngine) CurrentlyRunning() *CurrentProgram {
	engine.mutex.RLock()
	defer engine.mutex.RUnlock()

	if engine.runner == nil {
		return nil
	}
	return engine.runner.getCurrentProgram()
}

func (engine *ControlEngine) StartEngine(program *types.Program) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if engine.runner != nil {
		return ErrProgramAlreadyRunning
	}

	runner, err := newRunner(engine.config, engine.storage, program)
	if err != nil {
		return err
	}

	engine.runner = runner
	engine.wg.Add(1)
	engine.runner.Start()

	return nil
}

func (engine *ControlEngine) StopEngine() error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if engine.runner != nil {
		engine.runner.Stop()
		return nil
	}
	return ErrNoProgramRunning
}
