package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type (
	PowerResponse struct {
		Percent int `json:"percent"`
	}

	PowerStatusResponse struct {
		Data map[string]PowerResponse `json:"data"`
	}

	psuReadings struct {
		Fan    PowerResponse
		Heater PowerResponse
		Steam  PowerResponse
	}

	temperatureResponse struct {
		Data map[string]float32 `json:"data"`
	}

	temperatureReadings struct {
		Material float32
		Kiln     float32
	}

	sensorReader struct {
		client    *http.Client
		sensorURL string
		commands  <-chan string
		// Closed by the runner once it has stopped receiving responses. A
		// reader must watch it everywhere it can block, or it outlives the
		// run and keeps the runner's WaitGroup up forever.
		shutdown <-chan struct{}
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

	request, err := http.NewRequest("GET", controller.sensorURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := controller.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

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

	return &temperatureReadings{
		Material: readingOrInvalid(dataResponse.Data, "material"),
		Kiln:     readingOrInvalid(dataResponse.Data, "kiln"),
	}, nil
}

// readingOrInvalid returns the named temperature, or the invalid sentinel
// when the sensor unit did not report it at all. A missing key would
// otherwise decode to a plausible looking 0 degrees.
func readingOrInvalid(data map[string]float32, name string) float32 {
	value, ok := data[name]
	if !ok {
		return types.InvalidTemperatureReading
	}
	return value
}

func newTemperatureSensorReader(url string, commands <-chan string, responses chan<- temperatureReadings, shutdown <-chan struct{}) (*temperatureSensorReader, error) {
	controller := temperatureSensorReader{
		sensorReader: sensorReader{
			client:    &http.Client{},
			sensorURL: url,
			commands:  commands,
			shutdown:  shutdown,
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

	request, err := http.NewRequest("GET", controller.sensorURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := controller.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

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

	return &psuReadings{Fan: dataResponse.Data["fan"], Heater: dataResponse.Data["heater"], Steam: dataResponse.Data["steam"]}, nil
}

func newPSUSensorReader(url string, commands <-chan string, responses chan<- psuReadings, shutdown <-chan struct{}) (*psuSensorReader, error) {
	controller := psuSensorReader{
		sensorReader: sensorReader{
			client:    &http.Client{},
			sensorURL: url,
			commands:  commands,
			shutdown:  shutdown,
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

// nextCommand blocks until the runner asks for something or shuts down. The
// second return value is false once the run is over and the reader must stop.
func (controller *sensorReader) nextCommand() (string, bool) {
	select {
	case <-controller.shutdown:
		return "", false
	case engineMessage := <-controller.commands:
		return engineMessage, true
	}
}

// publish hands a reading to the runner, giving up if the run ended while the
// sensor was being read. Nobody receives from the response channel after that,
// so a plain send would block for the lifetime of the process.
func publish[T any](responses chan<- T, shutdown <-chan struct{}, values T) bool {
	select {
	case responses <- values:
		return true
	case <-shutdown:
		return false
	}
}

func (controller *psuSensorReader) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		engineMessage, ok := controller.nextCommand()
		if !ok {
			return
		}
		switch engineMessage {
		case sensorRead:
			values, err := controller.readSensors()
			if err != nil {
				log.Error("Failed to read psu sensors: %v", err)
			} else if !publish(controller.runner, controller.shutdown, *values) {
				return
			}
		default:
			log.Warning("Unknown controller message: %s", engineMessage)
		}
	}
}

func (controller *temperatureSensorReader) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		engineMessage, ok := controller.nextCommand()
		if !ok {
			return
		}
		switch engineMessage {
		case sensorRead:
			values, err := controller.readTemperatures()
			if err != nil {
				log.Error("Failed to read temperature sensors: %v", err)
			} else if !publish(controller.runner, controller.shutdown, *values) {
				return
			}
		default:
			log.Warning("Unknown controller message: %s", engineMessage)
		}
	}
}
