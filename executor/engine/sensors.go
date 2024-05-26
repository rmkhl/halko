package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rmkhl/halko/executor/types"
)

const (
	PowerOn  PowerStatus = "On"
	PowerOff PowerStatus = "Off"
)

type (
	PowerStatus string

	PowerResponse struct {
		Status  PowerStatus `json:"status"`
		Percent int         `json:"percent,omitempty"`
	}

	PowerStatusResponse struct {
		Data map[string]PowerResponse `json:"data"`
	}

	psuReadings struct {
		Fan        PowerResponse
		Heater     PowerResponse
		Humidifier PowerResponse
	}

	temperatureResponse struct {
		Data map[string]float32 `json:"data"`
	}

	temperatureReadings struct {
		Material float32
		Oven     float32
	}

	sensorReader struct {
		client    *http.Client
		sensorURl string
		commands  <-chan string
	}

	temperatureSensorReader struct {
		sensorReader
		runner chan<- temperatureReadings
	}

	psuSensorReader struct {
		sensorReader
		runner chan<- psuReadings
	}
)

func (controller *temperatureSensorReader) readTemperatures() (*temperatureReadings, error) {
	var dataResponse temperatureResponse

	request, err := http.NewRequest("GET", controller.sensorURl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := controller.client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse types.APIErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil && errorResponse.Err != "" {
			return nil, fmt.Errorf("cannot read sensors (%s)", errorResponse.Err)
		}
		return nil, fmt.Errorf("cannot read sensors (%s)", response.Status)
	}

	err = json.Unmarshal(body, &dataResponse)
	if err != nil {
		return nil, err
	}

	return &temperatureReadings{Material: dataResponse.Data["material"], Oven: dataResponse.Data["oven"]}, nil
}

func newTemperatureSensorReader(url string, commands <-chan string, responses chan<- temperatureReadings) (*temperatureSensorReader, error) {
	controller := temperatureSensorReader{
		sensorReader: sensorReader{
			client:    &http.Client{},
			sensorURl: url,
			commands:  commands,
		},
		runner: responses,
	}

	// verify we can read from the sensors
	_, err := controller.readTemperatures()
	if err != nil {
		return nil, err
	}

	return &controller, nil
}

func (controller *psuSensorReader) readSensors() (*psuReadings, error) {
	var dataResponse PowerStatusResponse

	request, err := http.NewRequest("GET", controller.sensorURl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := controller.client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse types.APIErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil && errorResponse.Err != "" {
			return nil, fmt.Errorf("cannot read sensors (%s)", errorResponse.Err)
		}
		return nil, fmt.Errorf("cannot read sensors (%s)", response.Status)
	}

	err = json.Unmarshal(body, &dataResponse)
	if err != nil {
		return nil, err
	}

	return &psuReadings{Fan: dataResponse.Data["fan"], Heater: dataResponse.Data["heater"], Humidifier: dataResponse.Data["heater"]}, nil
}

func newPSUSensorReader(url string, commands <-chan string, responses chan<- psuReadings) (*psuSensorReader, error) {
	controller := psuSensorReader{
		sensorReader: sensorReader{
			client:    &http.Client{},
			sensorURl: url,
			commands:  commands,
		},
		runner: responses,
	}

	// verify we can read from the sensors
	_, err := controller.readSensors()
	if err != nil {
		return nil, err
	}

	return &controller, nil
}

func (controller *psuSensorReader) Run() {
	for {
		engineMessage := <-controller.commands
		switch engineMessage {
		case controllerDone:
			return
		case sensorRead:
			values, err := controller.readSensors()
			if err != nil {
				log.Printf("Failed to read psu sensors (%s)", err.Error())
			} else {
				controller.runner <- *values
			}
		default:
			log.Printf("Unknown controller message (%s)", engineMessage)
		}
	}
}

func (controller *temperatureSensorReader) Run() {
	for {
		engineMessage := <-controller.commands
		switch engineMessage {
		case controllerDone:
			return
		case sensorRead:
			values, err := controller.readTemperatures()
			if err != nil {
				log.Printf("Failed to read temperature sensors (%s)", err.Error())
			} else {
				controller.runner <- *values
			}
		default:
			log.Printf("Unknown controller message (%s)", engineMessage)
		}
	}
}
