package types

type (
	APIErrorResponse struct {
		Err string `json:"error"`
	}

	APIResponse[T any] struct {
		Data T `json:"data"`
	}

	SavedProgram struct {
		Name        string       `json:"name"`
		State       ProgramState `json:"state"`
		StartedAt   int64        `json:"started_at"`
		CompletedAt int64        `json:"completed_at"`
	}

	ProgramListing struct {
		Programs []SavedProgram `json:"programs"`
	}

	TemperatureStatus struct {
		Material float32 `json:"material"`
		Oven     float32 `json:"oven"`
		Delta    float32 `json:"delta"`
	}

	PSUStatus struct {
		Heater     int `json:"heater"`
		Fan        int `json:"fan"`
		Humidifier int `json:"humidifier"`
	}

	ProgramStatus struct {
		Program              Program           `json:"program"`
		StartedAt            int64             `json:"started_at,omitempty"`
		CurrentStep          string            `json:"current_step,omitempty"`
		CurrentStepStartedAt int64             `json:"current_step_started_at,omitempty"`
		Temperatures         TemperatureStatus `json:"temperatures,omitempty"`
		PowerStatus          PSUStatus         `json:"power_status,omitempty"`
	}

	ExecutedProgram struct {
		Program     Program      `json:"program"`
		State       ProgramState `json:"state"`
		StartedAt   int64        `json:"started_at"`
		CompletedAt int64        `json:"completed_at"`
	}
)
