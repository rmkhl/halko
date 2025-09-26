package serial

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
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
	// Calculate timeout based on character transmission time:
	// 8 data bits + 2 stop bits = 10 bits per character
	// Time per character = 10 bits / baud_rate seconds
	// Cover time for 2 characters with 20% safety margin
	charTimeMs := float64(10*1000) / float64(baudRate) // milliseconds per character
	timeoutMs := 2 * charTimeMs * 1.2                  // 2 characters + 20% safety margin
	timeout := time.Duration(timeoutMs) * time.Millisecond

	config := &serial.Config{
		Name:        device,
		Baud:        baudRate,
		ReadTimeout: timeout,
	}

	return &SensorUnit{
		config: config,
	}, nil
}

// Connect establishes a connection to the sensor unit
func (s *SensorUnit) Connect() error {
	log.Trace("Attempting to connect to sensor unit on device %s at %d baud", s.config.Name, s.config.Baud)
	s.mutex.Lock()

	if s.connected {
		log.Trace("Already connected to sensor unit")
		s.mutex.Unlock()
		return nil
	}

	log.Trace("Opening serial port %s", s.config.Name)
	port, err := serial.OpenPort(s.config)
	if err != nil {
		log.Trace("Failed to open serial port %s: %v", s.config.Name, err)
		s.mutex.Unlock()
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	s.port = port
	s.connected = true
	log.Trace("Serial port opened successfully, clearing any initialization garbage")

	// Clear any initialization garbage from the Arduino before sending commands
	s.clearInputBuffer()

	log.Trace("Buffer cleared, sending hello command")

	// We need to unlock the mutex before calling sendCommand to avoid deadlock
	s.mutex.Unlock()

	// Send helo command to verify connection
	response, err := s.sendCommand(HeloCommand)
	if err != nil || response != HeloResponse {
		log.Trace("Hello command failed: err=%v, response=%q", err, response)
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

	log.Trace("Successfully connected to sensor unit")
	return nil
}

// Close closes the connection to the sensor unit
func (s *SensorUnit) Close() error {
	log.Trace("Closing connection to sensor unit")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected {
		log.Trace("Sensor unit already disconnected")
		return nil
	}

	err := s.port.Close()
	s.connected = false
	if err != nil {
		log.Trace("Error closing serial port: %v", err)
	} else {
		log.Trace("Successfully closed connection to sensor unit")
	}
	return err
}

// IsConnected returns true if the connection is established and the Arduino is responding
func (s *SensorUnit) IsConnected() bool {
	log.Trace("Checking sensor unit connection status")
	// First check if we're already connected
	s.mutex.Lock()
	isConnected := s.connected
	s.mutex.Unlock()

	if !isConnected {
		log.Trace("Not currently connected, attempting to connect")
		// If we're not connected, try to connect
		err := s.Connect()
		result := err == nil
		log.Trace("Connection attempt result: %t", result)
		return result
	}

	log.Trace("Already connected, verifying with hello command")
	// If we're already connected, verify by sending a command
	response, err := s.sendCommand(HeloCommand)
	if err != nil || response != HeloResponse {
		log.Trace("Verification failed: err=%v, response=%q, marking as disconnected", err, response)
		// If we can't communicate, mark as disconnected
		s.mutex.Lock()
		s.connected = false
		s.mutex.Unlock()
		return false
	}

	log.Trace("Connection verified successfully")
	return true
}

// GetTemperatures reads the current temperature values from the sensor unit
func (s *SensorUnit) GetTemperatures() ([]Temperature, error) {
	log.Trace("Getting temperatures from sensor unit")
	if err := s.Connect(); err != nil {
		log.Trace("Failed to connect for temperature reading: %v", err)
		return nil, err
	}

	response, err := s.sendCommand(ReadCommand)
	if err != nil {
		log.Trace("Failed to read temperatures: %v", err)
		return nil, err
	}

	log.Trace("Parsing temperature response: %q", response)
	// Parse response format: OvenPrimary=XX.XC,OvenSecondary=XX.XC,Wood=XX.XC
	readings := strings.Split(response, ",")
	if len(readings) != 3 {
		log.Trace("Invalid temperature reading format, expected 3 readings but got %d", len(readings))
		return nil, fmt.Errorf("invalid temperature reading format: %s", response)
	}

	temperatures := make([]Temperature, 0, 3)
	for i, reading := range readings {
		log.Trace("Processing temperature reading %d: %q", i+1, reading)
		parts := strings.Split(reading, "=")
		if len(parts) != 2 {
			log.Trace("Skipping malformed reading: %q", reading)
			continue
		}

		name := parts[0]
		valueStr := parts[1]
		log.Trace("Parsing sensor %q with value %q", name, valueStr)

		var value float32
		var unit string

		if valueStr == "NaN" {
			log.Trace("Sensor %q has invalid reading (NaN)", name)
			value = types.InvalidTemperatureReading
			unit = "C"
		} else {
			unit = string(valueStr[len(valueStr)-1])
			valueStr = valueStr[:len(valueStr)-1]
			_, err := fmt.Sscanf(valueStr, "%f", &value)
			if err != nil {
				log.Info("Failed to parse temperature value: %s", valueStr)
				log.Trace("Parse error for sensor %q value %q: %v", name, valueStr, err)
				continue
			}
			log.Trace("Parsed sensor %q: %.1f%s", name, value, unit)
		}

		temperatures = append(temperatures, Temperature{
			Name:  name,
			Value: value,
			Unit:  unit,
		})
	}

	log.Trace("Successfully parsed %d temperature readings", len(temperatures))
	return temperatures, nil
}

// SetStatusText sets the status text on the LCD display
func (s *SensorUnit) SetStatusText(text string) error {
	log.Trace("Setting status text on LCD: %q", text)
	if err := s.Connect(); err != nil {
		log.Trace("Failed to connect for status text update: %v", err)
		return err
	}

	// Truncate text to fit on the LCD (15 characters max)
	if len(text) > 15 {
		log.Trace("Truncating status text from %d to 15 characters", len(text))
		text = text[:15]
	}

	command := fmt.Sprintf("%s %s;", ShowCommand, text)
	log.Trace("Sending status command: %q", command)
	_, err := s.sendCommand(command)
	if err != nil {
		log.Trace("Failed to set status text: %v", err)
	} else {
		log.Trace("Successfully set status text to: %q", text)
	}
	return err
}

// sendCommand sends a command to the Arduino and returns the response
func (s *SensorUnit) sendCommand(cmd string) (string, error) {
	log.Trace("Sending command to sensor unit: %q", cmd)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.port == nil {
		log.Trace("Cannot send command: not connected to sensor unit")
		return "", errors.New("not connected to sensor unit")
	}

	// Send command
	log.Trace("Writing command bytes to serial port")
	_, err := s.port.Write([]byte(cmd))
	if err != nil {
		log.Trace("Failed to write command to serial port: %v", err)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Special handling for show commands - they don't return a meaningful response
	if strings.HasPrefix(cmd, ShowCommand) {
		log.Trace("Show command sent, no response expected")
		return "", nil
	}

	// Read response
	log.Trace("Reading response from sensor unit")
	scanner := bufio.NewScanner(s.port)
	var response string

	// Wait for a relevant response based on the command type
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		log.Trace("Received line from sensor unit: %q", line)

		// Skip empty lines
		if line == "" {
			log.Trace("Skipping empty line")
			continue
		}

		// For read commands, we're looking for the temperature data format
		if strings.HasPrefix(cmd, ReadCommand) && strings.Contains(line, "=") {
			log.Trace("Found temperature data response: %q", line)
			response = line
			break
		}

		// For helo commands, we're looking for the helo response
		if strings.HasPrefix(cmd, HeloCommand) && line == HeloResponse {
			log.Trace("Found hello response: %q", line)
			response = line
			break
		}

		// Log unexpected lines for debugging
		log.Info("Skipping unexpected line for command %q: %q", cmd, line)
	}

	if err := scanner.Err(); err != nil {
		log.Trace("Scanner error while reading response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Trace("Command completed with response: %q", response)
	return response, nil
}

// clearInputBuffer reads and discards any data in the input buffer
// This helps handle any initialization garbage from the Arduino
func (s *SensorUnit) clearInputBuffer() {
	log.Trace("Clearing serial input buffer")
	// Create a temporary buffer to read and discard any pending data
	tempBuf := make([]byte, 1024)

	// Store the original timeout
	origTimeout := s.config.ReadTimeout
	log.Trace("Storing original timeout: %v, setting short timeout for buffer clear", origTimeout)

	// Set a short timeout for clearing the buffer
	s.config.ReadTimeout = time.Millisecond * 100

	// First flush any outgoing data
	log.Trace("Flushing outgoing data")
	s.port.Flush()

	// Then try to read with a short timeout to clear any garbage
	totalCleared := 0
	for {
		n, err := s.port.Read(tempBuf)
		if err != nil || n == 0 {
			break
		}
		totalCleared += n
		log.Info("Cleared %d bytes of garbage from serial buffer: %q", n, string(tempBuf[:n]))
	}

	log.Trace("Cleared %d total bytes from buffer, restoring original timeout", totalCleared)
	// Restore the original timeout
	s.config.ReadTimeout = origTimeout
}
