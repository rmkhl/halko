package types

type ApiErrorResponse struct {
	Err string `json:"data"`
}

type ApiResponse[T any] struct {
	Data T `json:"data"`
}

type TemperatureResponse map[string]float32

type PowerStatus string

const (
	POWER_ON   PowerStatus = "On"
	POWER_OFF  PowerStatus = "Off"
	POWER_NEXT PowerStatus = "Switch"
)

type PowerResponse struct {
	Status PowerStatus `json:"status"`
	Cycle  string      `json:"cycle,omitempty"`
}

type PowerStatusResponse map[string]PowerResponse

type PowerCycle struct {
	Name  string   `json:"name"`
	Ticks [10]bool `json:"ticks"`
}

type PowerCommand struct {
	Command PowerStatus `json:"command"`
	Cycle   PowerCycle  `json:"cycle"`
}

type PowerOperationResponse struct {
	Message string `json:"message"`
}
