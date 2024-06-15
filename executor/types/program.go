package types

import "sort"

const (
	StepTypeHeating   StepType = "heating"
	StepTypeCooling   StepType = "cooling"
	StepTypeAcclimate StepType = "acclimate"
	StepTypeWaiting   StepType = "waiting"
)

type (
	StepType string

	ProgramStep struct {
		Name           string      `json:"name"`
		StepType       StepType    `json:"type"`
		MaximumRuntime int         `json:"time_constraint"`
		ValidRange     SensorRange `json:"temperature_constraint"`
		Heater         PSUPhase    `json:"heater"`
		Fan            PSUPhase    `json:"fan"`
		Humidifier     PSUPhase    `json:"humidifier"`
	}

	Program struct {
		ProgramName     string        `json:"name"`
		DefaultStepTime int           `json:"default_step_time"`
		ProgramSteps    []ProgramStep `json:"steps"`
	}
)

func (p *Program) Validate() error {
	// Make sure all the cycles are sorted by delta.
	for _, step := range p.ProgramSteps {
		if step.Heater.DeltaCycles != nil {
			sort.Slice(step.Heater.DeltaCycles, func(i, j int) bool {
				return step.Heater.DeltaCycles[i].TemperatureDelta < step.Heater.DeltaCycles[j].TemperatureDelta
			})
		}
		if step.Fan.DeltaCycles != nil {
			sort.Slice(step.Heater.DeltaCycles, func(i, j int) bool {
				return step.Heater.DeltaCycles[i].TemperatureDelta < step.Heater.DeltaCycles[j].TemperatureDelta
			})
		}
		if step.Humidifier.DeltaCycles != nil {
			sort.Slice(step.Heater.DeltaCycles, func(i, j int) bool {
				return step.Heater.DeltaCycles[i].TemperatureDelta < step.Heater.DeltaCycles[j].TemperatureDelta
			})
		}
	}
	return nil
}
