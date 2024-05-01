package domain

type Phase struct {
	HasID
	Name          string             `json:"name"`
	ConstantCycle *Cycle             `json:"constantCycle,omitempty"`
	DeltaCycles   []DeltaCycle       `json:"deltaCycles,omitempty"`
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
