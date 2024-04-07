package domain

type Phase struct {
	HasID
	Name          string       `json:"name"`
	ConstantCycle *Cycle       `json:"constant_cycle,omitempty"`
	DeltaCycles   []DeltaCycle `json:"delta_cycles,omitempty"`
}

type DeltaCycle struct {
	Delta float64 `json:"delta"`
	Above *Cycle  `json:"above"`
	Below *Cycle  `json:"below"`
}
