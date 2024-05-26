package domain

type Program struct {
	HasName
	MaximumRuntime uint64         `json:"maximum_runtime"`
	Phases         []ProgramPhase `json:"phases"`
}

type ProgramPhase struct {
	TimeConstraint        uint64                `json:"time_constraint"`
	TemperatureConstraint TemperatureConstraint `json:"temperature_constraint"`
	Heater                Phase                 `json:"heater"`
	Fan                   Phase                 `json:"fan"`
	Humidifier            Phase                 `json:"humidifier"`
}

type TemperatureConstraint struct {
	Minimum float64 `json:"miminum"`
	Maximum float64 `json:"maximum"`
}
