package types

type ProgramState string

const (
	ProgramStateCanceled  ProgramState = "canceled"
	ProgramStateCompleted ProgramState = "completed"
	ProgramStateFailed    ProgramState = "failed"
	ProgramStatePending   ProgramState = "pending"
	ProgramStateRunning   ProgramState = "running"
	ProgramStateUnknown   ProgramState = "unknown"
)
