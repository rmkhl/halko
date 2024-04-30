package domain

type Phase struct {
	HasID
	Name          string             `json:"name"`
	ConstantCycle *Cycle             `json:"constant_cycle,omitempty"`
	DeltaCycles   []DeltaCycle       `json:"delta_cycles,omitempty"`
	ValidRange    []ValidSensorRange `json:"validRange"`
	CycleMode     string             `json:"cycleMode"`
}

type ValidSensorRange struct {
	Sensor string  `json:"sensor"`
	Above  float64 `json:"above"`
	Below  float64 `json:"below"`
}

type DeltaCycle struct {
	Delta float64 `json:"delta"`
	Above *Cycle  `json:"above"`
	Below *Cycle  `json:"below"`
}
