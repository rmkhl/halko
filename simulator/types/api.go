package types

type APIErrorResponse struct {
	Err string `json:"data"`
}

type APIResponse[T any] struct {
	Data T `json:"data"`
}

type TemperatureResponse map[string]float32

type PowerStatus string

const (
	PowerOn  PowerStatus = "On"
	PowerOff PowerStatus = "Off"
)

type PowerResponse struct {
	Status  PowerStatus `json:"status"`
	Percent int         `json:"percent,omitempty"`
}

type PowerStatusResponse map[string]PowerResponse

type PowerCycle struct {
	Name  string   `json:"name"`
	Ticks [10]bool `json:"ticks"`
}

type PowerCommand struct {
	Command PowerStatus `json:"command"`
	Percent int         `json:"percent,omitempty"`
}

type PowerOperationResponse struct {
	Message string `json:"message"`
}
