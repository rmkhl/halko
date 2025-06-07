package serial

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/tarm/serial"
)

// Command constants
const (
	HeloCommand  = "helo;"
	ReadCommand  = "read;"
	ShowCommand  = "show"
	HeloResponse = "helo"
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

	if s.connected {
		s.mutex.Unlock()
		return nil
	}

	port, err := serial.OpenPort(s.config)
	if err != nil {
		s.mutex.Unlock()
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	s.port = port
	s.connected = true

	// We need to unlock the mutex before calling sendCommand to avoid deadlock
	s.mutex.Unlock()

	// Send helo command to verify connection
	response, err := s.sendCommand(HeloCommand)
	if err != nil || response != HeloResponse {
		// If communication fails, we need to close the port and mark as disconnected
		s.mutex.Lock()
		s.port.Close()
		s.connected = false
		s.mutex.Unlock()
		if err != nil {
			return fmt.Errorf("failed to connect to sensor unit: %w", err)
		}
		return fmt.Errorf("failed to connect to sensor unit: unexpected response %q", response)
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
	// First check if we're already connected
	s.mutex.Lock()
	isConnected := s.connected
	s.mutex.Unlock()

	if !isConnected {
		// If we're not connected, try to connect
		err := s.Connect()
		return err == nil
	}

	// If we're already connected, verify by sending a command
	response, err := s.sendCommand(HeloCommand)
	if err != nil || response != HeloResponse {
		// If we can't communicate, mark as disconnected
		s.mutex.Lock()
		s.connected = false
		s.mutex.Unlock()
		return false
	}

	return true
}

// GetTemperatures reads the current temperature values from the sensor unit
func (s *SensorUnit) GetTemperatures() ([]Temperature, error) {
	if err := s.Connect(); err != nil {
		return nil, err
	}

	response, err := s.sendCommand(ReadCommand)
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
			value = types.InvalidTemperatureReading
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

	_, err := s.sendCommand(fmt.Sprintf("%s %s;", ShowCommand, text))
	return err
}

// sendCommand sends a command to the Arduino and returns the response
func (s *SensorUnit) sendCommand(cmd string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.port == nil {
		return "", errors.New("not connected to sensor unit")
	}

	// Clear any pending data that might be initialization garbage
	s.clearInputBuffer()

	// Send command
	_, err := s.port.Write([]byte(cmd))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Special handling for show commands - they don't return a meaningful response
	if strings.HasPrefix(cmd, ShowCommand) {
		return "", nil
	}

	// Read response
	scanner := bufio.NewScanner(s.port)
	var response string

	// Wait for a relevant response based on the command type
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// For read commands, we're looking for the temperature data format
		if strings.HasPrefix(cmd, ReadCommand) && strings.Contains(line, "=") {
			response = line
			break
		}

		// For helo commands, we're looking for the helo response
		if strings.HasPrefix(cmd, HeloCommand) && line == HeloResponse {
			response = line
			break
		}

		// Log unexpected lines for debugging
		log.Printf("Skipping unexpected line for command %q: %q", cmd, line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}

// clearInputBuffer reads and discards any data in the input buffer
// This helps handle any initialization garbage from the Arduino
func (s *SensorUnit) clearInputBuffer() {
	// Create a temporary buffer to read and discard any pending data
	tempBuf := make([]byte, 1024)

	// Store the original timeout
	origTimeout := s.config.ReadTimeout

	// Set a short timeout for clearing the buffer
	s.config.ReadTimeout = time.Millisecond * 100

	// First flush any outgoing data
	s.port.Flush()

	// Then try to read with a short timeout to clear any garbage
	for {
		n, err := s.port.Read(tempBuf)
		if err != nil || n == 0 {
			break
		}
		log.Printf("Cleared %d bytes of garbage from serial buffer: %q", n, string(tempBuf[:n]))
	}

	// Restore the original timeout
	s.config.ReadTimeout = origTimeout
}
