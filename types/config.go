package types

type (
	ExecutorConfig struct {
		BasePath             string `json:"base_path"`
		TickLength           int    `json:"tick_length"`
		TemperatureSensorURl string `json:"temperature_sensor_url"`
		PowerSensorURl       string `json:"power_sensor_url"`
		PowerControlURl      string `json:"power_control_url"`
	}

	HalkoConfig struct {
		ExecutorConfig *ExecutorConfig `json:"executor"`
	}
)
