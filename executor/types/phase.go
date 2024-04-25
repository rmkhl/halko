package types

type (
	DeltaCycle struct {
		TemperatureDelta float32 `json:"delta"`
		EnteredBelow     int     `json:"below"`
		EnteredAbove     int     `json:"above"`
	}

	SensorRange struct {
		MinimumTemperature float32 `json:"minimum"`
		MaximumTemperature float32 `json:"maximum"`
	}

	PSUPhase struct {
		Name          string       `json:"name"`
		ConstantCycle int          `json:"constant_cycle"`
		DeltaCycles   []DeltaCycle `json:"delta_cycles"`
	}
)
