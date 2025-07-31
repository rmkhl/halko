package types

import (
	"encoding/json"
	"io"
	"os"
)

type (
	ExecutorConfig struct {
		BasePath          string                    `json:"base_path"`
		Port              int                       `json:"port"`
		TickLength        int                       `json:"tick_length"`
		PowerUnitURL      string                    `json:"power_unit_url"`
		SensorUnitURL     string                    `json:"sensor_unit_url"`
		StatusMessageURL  string                    `json:"status_message_url"`
		PidSettings       map[StepType]*PidSettings `json:"pid_settings"`
		MaxDeltaHeating   float32                   `json:"max_delta_heating"`
		MinDeltaHeating   float32                   `json:"min_delta_heating"`
		NetworkInterface  string                    `json:"network_interface"`
	}

	PowerUnit struct {
		ShellyAddress string         `json:"shelly_address"`
		CycleLength   int            `json:"cycle_length"` // Duration of a power cycle in seconds
		PowerMapping  map[string]int `json:"power_mapping"`
		MaxIdleTime   int            `json:"max_idle_time"` // Maximum idle time in seconds before a executor is considered idle
	}

	SensorUnitConfig struct {
		SerialDevice string `json:"serial_device"`
		BaudRate     int    `json:"baud_rate"`
	}

	HalkoConfig struct {
		ExecutorConfig *ExecutorConfig   `json:"executor"`
		PowerUnit      *PowerUnit        `json:"power_unit"`
		SensorUnit     *SensorUnitConfig `json:"sensorunit"`
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
