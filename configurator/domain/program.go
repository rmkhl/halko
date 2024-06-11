package domain

type Program struct {
	HasName
	DefaultStepRuntime uint64 `json:"defaultStepRuntime"`
	PreheatTo          uint64 `json:"preheatTo"`
	Steps              []Step `json:"steps"`
}

type Step struct {
	TimeConstraint        uint64                `json:"timeConstraint"`
	TemperatureConstraint TemperatureConstraint `json:"temperatureConstraint"`
	Heater                Phase                 `json:"heater"`
	Fan                   Phase                 `json:"fan"`
	Humidifier            Phase                 `json:"humidifier"`
}

type TemperatureConstraint struct {
	Minimum float64 `json:"minimum"`
	Maximum float64 `json:"maximum"`
}
