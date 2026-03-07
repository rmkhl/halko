package types

// ExecutionStorage defines the interface for managing execution history and logs
type ExecutionStorage interface {
	// Program execution history operations
	ListExecutedPrograms() ([]string, error)
	LoadExecutedProgram(programName string) (*Program, error)
	DeleteExecutedProgram(programName string) error
	LoadState(programName string) (ProgramState, int64, error)
	GetLogPath(programName string) string
	GetRunningLogPath(programName string) string

	// System resource operations
	GetAvailableSpaceMB() int64
}

// ProgramStorage defines the interface for managing stored program templates
type ProgramStorage interface {
	// Program template operations
	ListStoredProgramsWithInfo() ([]StoredProgramInfo, error)
	LoadStoredProgram(programName string) (*Program, error)
	CreateStoredProgram(programName string, program *Program) error
	UpdateStoredProgram(programName string, program *Program) error
	DeleteStoredProgram(programName string) error
}
