package types

import (
	"encoding/json"
	"io"
	"os"
)

type (
	ExecutorConfig struct {
		BasePath             string                    `json:"base_path"`
		TickLength           int                       `json:"tick_length"`
		TemperatureSensorURL string                    `json:"temperature_sensor_url"`
		PowerSensorURL       string                    `json:"power_sensor_url"`
		PowerControlURL      string                    `json:"power_control_url"`
		PidSettings          map[StepType]*PidSettings `json:"pid_settings"`
	}

	HalkoConfig struct {
		ExecutorConfig *ExecutorConfig `json:"executor"`
	}
)

func ReadHalkoConfig(path string) (*HalkoConfig, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	content, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var config HalkoConfig
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
