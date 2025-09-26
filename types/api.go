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

// SensorStatus values
const (
	SensorStatusConnected    SensorStatus = "connected"
	SensorStatusDisconnected SensorStatus = "disconnected"
	SensorStatusOK           SensorStatus = "ok"
)

const (
	// signals invalid temperature reading
	InvalidTemperatureReading = -273.15 // Absolute zero in Celsius, used to indicate an invalid reading
)

// StatusRequest defines the structure for a set status request body
type StatusRequest struct {
	Message string `json:"message"`
}

// DisplayRequest defines the structure for a display update request body
type DisplayRequest struct {
	Message string `json:"message"`
}

type (
	ProgramState string
	SensorStatus string

	// APIErrorResponse is a generic error response
	APIErrorResponse struct {
		Err string `json:"error"`
	}

	// StatusResponse defines the structure for a status API response
	StatusResponse struct {
		Status SensorStatus `json:"status"`
	}

	APIResponse[T any] struct {
		Data T `json:"data"`
	}

	RunHistory struct {
		Name        string       `json:"name"`
		State       ProgramState `json:"state"`
		StartedAt   int64        `json:"started_at,omitempty"`
		CompletedAt int64        `json:"completed_at,omitempty"`
	}

	ExecutedProgram struct {
		RunHistory
		Program Program `json:"program"`
	}

	ProgramListing struct {
		Programs []RunHistory `json:"programs"`
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
type (
	PowerResponse struct {
		Percent uint8 `json:"percent"`
	}

	PowerStatusResponse map[string]PowerResponse

	PowerCommand struct {
		Percent uint8 `json:"percent"`
	}

	PowersCommand map[string]PowerCommand

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
