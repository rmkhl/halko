package types

type (
	APIErrorResponse struct {
		Err string `json:"error"`
	}

	APIResponse[T any] struct {
		Data T `json:"data"`
	}

	ProgramListing struct {
		Programs []string `json:"programs"`
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
		Program              ProgramListing    `json:"program"`
		StartedAt            int64             `json:"started_at,omitempty"`
		CurrentStep          string            `json:"current_step,omitempty"`
		CurrentStepStartedAt int64             `json:"current_step_started_at,omitempty"`
		Temperatures         TemperatureStatus `json:"temperatures,omitempty"`
		PowerStatus          PSUStatus         `json:"power_status,omitempty"`
	}
)
