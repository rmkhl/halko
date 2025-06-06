package types

// executor API
const (
	ProgramStateCanceled  ProgramState = "canceled"
	ProgramStateCompleted ProgramState = "completed"
	ProgramStateFailed    ProgramState = "failed"
	ProgramStatePending   ProgramState = "pending"
	ProgramStateRunning   ProgramState = "running"
	ProgramStateUnknown   ProgramState = "unknown"
)

type (
	ProgramState string

	APIErrorResponse struct {
		Err string `json:"error"`
	}

	APIResponse[T any] struct {
		Data T `json:"data"`
	}

	SavedProgram struct {
		Name        string       `json:"name"`
		State       ProgramState `json:"state"`
		StartedAt   int64        `json:"started_at,omitempty"`
		CompletedAt int64        `json:"completed_at,omitempty"`
	}

	ExecutedProgram struct {
		SavedProgram
		Program Program `json:"program"`
	}

	ProgramListing struct {
		Programs []SavedProgram `json:"programs"`
	}

	// TemperatureStatus represents the current temperature of the material and oven in Celsius.
	TemperatureStatus struct {
		Material float32 `json:"material"`
		Oven     float32 `json:"oven"`
	}

	// PSUStatus represents the power level (in percentage) of the heater, fan, and humidifier.
	PSUStatus struct {
		Heater     int8 `json:"heater"`
		Fan        int8 `json:"fan"`
		Humidifier int8 `json:"humidifier"`
	}

	ExecutionStatus struct {
		Program              Program           `json:"program"`
		StartedAt            int64             `json:"started_at,omitempty"`
		CurrentStep          string            `json:"current_step,omitempty"`
		CurrentStepStartedAt int64             `json:"current_step_started_at,omitempty"`
		Temperatures         TemperatureStatus `json:"temperatures,omitempty"`
		PowerStatus          PSUStatus         `json:"power_status,omitempty"`
	}
)

// Power controller API
const (
	PowerOn  PowerStatus = "On"
	PowerOff PowerStatus = "Off"
)

type (
	PowerStatus string

	PowerResponse struct {
		Status  PowerStatus `json:"status"`
		Percent uint8       `json:"percent,omitempty"`
	}

	PowerStatusResponse map[string]PowerResponse

	PowerCommand struct {
		Command PowerStatus `json:"command"`
		Percent uint8       `json:"percent,omitempty"`
	}

	PowerOperationResponse struct {
		Message string `json:"message"`
	}
)

// Temperature sensor API
type TemperatureResponse map[string]float32

// Shelly API responses
type (
	ShellySwitchGetStatusResponse struct {
		ID          string `json:"id"`
		Source      string `json:"source"`
		Output      bool   `json:"output"`
		Temperature struct {
			TC float32 `json:"tC"`
			TF float32 `json:"tF"`
		} `json:"temperature"`
	}
	ShellySwitchSetResponse struct {
		WasOn bool `json:"was_on"`
	}
)
