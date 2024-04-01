package domain

type Program struct {
	Name           string         `json:"name"`
	MaximumRuntime uint64         `json:"maximum_runtime"`
	Phases         []ProgramPhase `json:"phases"`
}

type ProgramPhase struct {
	TimeConstraint        TimeConstraint        `json:"time_constraint"`
	TemperatureConstraint TemperatureConstraint `json:"temperature_constraint"`
	Heater                Phase                 `json:"heater"`
	Fan                   Phase                 `json:"fan"`
	Humidifier            Phase                 `json:"humidifier"`
}

type TimeConstraint struct {
	Runtime uint64 `json:"runtime"`
}

type TemperatureConstraint struct {
	Minimum float64 `json:"miminum"`
	Maximum float64 `json:"maximum"`
}
