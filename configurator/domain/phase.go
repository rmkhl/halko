package domain

type Phase struct {
	HasID
	Name          string       `json:"name"`
	ConstantCycle *uint8       `json:"constantCycle,omitempty"`
	DeltaCycles   []DeltaCycle `json:"deltaCycles,omitempty"`
	CycleMode     string       `json:"cycleMode"`
}

type DeltaCycle struct {
	Delta float64 `json:"delta"`
	Above uint8   `json:"above"`
	Below uint8   `json:"below"`
}
