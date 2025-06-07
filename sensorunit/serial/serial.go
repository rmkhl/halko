package serial

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// SensorUnit represents a connection to the Arduino sensor unit
type SensorUnit struct {
	port      *serial.Port
	mutex     sync.Mutex
	connected bool
	config    *serial.Config
}

// Temperature represents a temperature reading from a sensor
type Temperature struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
	Unit  string  `json:"unit"`
}

// NewSensorUnit creates a new connection to the sensor unit
func NewSensorUnit(device string, baudRate int) (*SensorUnit, error) {
	config := &serial.Config{
		Name:        device,
		Baud:        baudRate,
		ReadTimeout: time.Second * 5,
	}

	return &SensorUnit{
		config: config,
	}, nil
}

// Connect establishes a connection to the sensor unit
func (s *SensorUnit) Connect() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connected {
		return nil
	}

	port, err := serial.OpenPort(s.config)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	s.port = port
	s.connected = true

	// Send helo command to verify connection
	_, err = s.sendCommand("helo;")
	if err != nil {
		s.port.Close()
		s.connected = false
		return fmt.Errorf("failed to connect to sensor unit: %w", err)
	}

	return nil
}

// Close closes the connection to the sensor unit
func (s *SensorUnit) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected {
		return nil
	}

	err := s.port.Close()
	s.connected = false
	return err
}

// IsConnected returns true if the connection is established and the Arduino is responding
func (s *SensorUnit) IsConnected() bool {
	// Use Connect which already verifies the Arduino responsiveness
	// Connect() will return nil if already connected, or try to establish a connection
	// and verify it with "helo;" command if not connected
	err := s.Connect()
	return err == nil
}

// GetTemperatures reads the current temperature values from the sensor unit
func (s *SensorUnit) GetTemperatures() ([]Temperature, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}

	response, err := s.sendCommand("read;")
	if err != nil {
		return nil, err
	}

	// Parse response format: OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC
	readings := strings.Split(response, ",")
	if len(readings) != 3 {
		return nil, fmt.Errorf("invalid temperature reading format: %s", response)
	}

	temperatures := make([]Temperature, 0, 3)
	for _, reading := range readings {
		parts := strings.Split(reading, "=")
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		valueStr := parts[1]

		var value float32
		var unit string

		if valueStr == "NaN" {
			value = 0
			unit = "C"
		} else {
			unit = string(valueStr[len(valueStr)-1])
			valueStr = valueStr[:len(valueStr)-1]
			_, err := fmt.Sscanf(valueStr, "%f", &value)
			if err != nil {
				log.Printf("Failed to parse temperature value: %s", valueStr)
				continue
			}
		}

		temperatures = append(temperatures, Temperature{
			Name:  name,
			Value: value,
			Unit:  unit,
		})
	}

	return temperatures, nil
}

// SetStatusText sets the status text on the LCD display
func (s *SensorUnit) SetStatusText(text string) error {
	if err := s.Connect(); err != nil {
		return err
	}

	// Truncate text to fit on the LCD (15 characters max)
	if len(text) > 15 {
		text = text[:15]
	}

	_, err := s.sendCommand(fmt.Sprintf("show %s;", text))
	return err
}

// sendCommand sends a command to the Arduino and returns the response
func (s *SensorUnit) sendCommand(cmd string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.port == nil {
		return "", errors.New("not connected to sensor unit")
	}

	// Clear any pending data
	s.port.Flush()

	// Send command
	_, err := s.port.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(s.port)
	var response string

	// First line is bufferIndex, second line is buffer content, third line is the actual response
	// from the Arduino code we can see it prints these lines when processing commands
	linesRead := 0
	for scanner.Scan() {
		line := scanner.Text()
		linesRead++

		// Skip bufferIndex and buffer content lines
		if linesRead >= 3 {
			response = line
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}
