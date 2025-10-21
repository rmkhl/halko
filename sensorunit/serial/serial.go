package serial

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
	"github.com/tarm/serial"
)

const (
	HeloCommand  = "helo;"
	ReadCommand  = "read;"
	ShowCommand  = "show"
	HeloResponse = "helo"
)

type SensorUnit struct {
	port      *serial.Port
	mutex     sync.Mutex
	connected bool
	config    *serial.Config
	// Device monitoring goroutine management
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
	monitoring bool
}

type Temperature struct {
	Name  string  `json:"name"`
	Value float32 `json:"value"`
	Unit  string  `json:"unit"`
}

func NewSensorUnit(device string, baudRate int) (*SensorUnit, error) {
	// Calculate timeout based on character transmission time:
	// 8 data bits + 2 stop bits = 10 bits per character
	// Time per character = 10 bits / baud_rate seconds
	// Cover time for 4 characters with 20% safety margin
	charTimeMs := float64(10*1000) / float64(baudRate)
	timeoutMs := 4 * charTimeMs * 1.2
	timeout := time.Duration(timeoutMs) * time.Millisecond

	config := &serial.Config{
		Name:        device,
		Baud:        baudRate,
		ReadTimeout: timeout,
	}

	ctx, cancel := context.WithCancel(context.Background())

	sensorUnit := &SensorUnit{
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
		wg:         new(sync.WaitGroup),
		monitoring: false,
	}

	// Start device monitoring automatically
	sensorUnit.StartDeviceMonitoring()

	return sensorUnit, nil
}

func (s *SensorUnit) Connect() error {
	log.Trace("Attempting to connect to sensor unit on device %s at %d baud", s.config.Name, s.config.Baud)
	start := time.Now()
	s.mutex.Lock()

	if s.connected {
		s.mutex.Unlock()
		return nil
	}

	log.Debug("Opening serial connection to %s at %d baud", s.config.Name, s.config.Baud)
	port, err := serial.OpenPort(s.config)
	if err != nil {
		log.Error("Failed to open serial port %s: %v", s.config.Name, err)
		s.mutex.Unlock()
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	log.Trace("Serial port opened successfully")
	s.port = port
	s.connected = true

	log.Trace("Clearing any initialization garbage")
	s.clearInputBuffer()
	s.mutex.Unlock()

	// Try handshake with retries to handle Arduino serial initialization timing
	var response string
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, err = s.sendCommand(HeloCommand)
		if err == nil && response == HeloResponse {
			break
		}

		if attempt < maxRetries {
			backoffDuration := time.Duration(200*attempt) * time.Millisecond
			log.Warning("Handshake attempt %d failed (err=%v, response=%q), Arduino may still be initializing, retrying in %v",
				attempt, err, response, backoffDuration)
			time.Sleep(backoffDuration)
		}
	}

	if err != nil || response != HeloResponse {
		log.Warning("Sensor unit handshake failed: err=%v, response=%q", err, response)
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

	log.Info("Sensor unit connection established on %s", s.config.Name)
	log.Debug("Serial connection established in %v", time.Since(start))
	return nil
}

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
		log.Error("Error closing serial port: %v", err)
	} else {
		log.Info("Sensor unit connection closed")
	}
	return err
}

// Shutdown completely shuts down the sensor unit including stopping device monitoring
// This should be called when the sensorunit service itself is shutting down
func (s *SensorUnit) Shutdown() error {
	log.Info("Shutting down sensor unit completely")

	// Stop monitoring first
	s.StopDeviceMonitoring()

	// Then close the connection if still connected
	return s.Close()
}

func (s *SensorUnit) IsConnected() bool {
	log.Trace("Checking sensor unit connection status")
	s.mutex.Lock()
	isConnected := s.connected
	s.mutex.Unlock()

	if !isConnected {
		log.Debug("Sensor unit not connected, attempting connection")
		// If we're not connected, try to connect
		err := s.Connect()
		result := err == nil
		log.Trace("Connection attempt result: %t", result)
		return result
	}

	log.Trace("Already connected, verifying with hello command")
	response, err := s.sendCommand(HeloCommand)
	if err != nil || response != HeloResponse {
		log.Warning("Sensor unit connection lost - verification failed: %v", err)
		// If we can't communicate, mark as disconnected
		s.mutex.Lock()
		s.connected = false
		s.mutex.Unlock()
		return false
	}

	log.Trace("Connection verified successfully")
	return true
}

func (s *SensorUnit) GetTemperatures() ([]Temperature, error) {
	log.Debug("Reading temperature values from sensor unit")
	if err := s.Connect(); err != nil {
		log.Error("Failed to connect for temperature reading: %v", err)
		return nil, err
	}

	response, err := s.sendCommand(ReadCommand)
	if err != nil {
		log.Error("Failed to read temperatures from sensor unit: %v", err)
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
			log.Debug("Sensor %q has invalid reading (NaN)", name)
			value = types.InvalidTemperatureReading
			unit = "C"
		} else {
			unit = string(valueStr[len(valueStr)-1])
			valueStr = valueStr[:len(valueStr)-1]
			_, err := fmt.Sscanf(valueStr, "%f", &value)
			if err != nil {
				log.Warning("Failed to parse temperature value for sensor %s: %s", name, valueStr)
				continue
			}
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

func (s *SensorUnit) SetStatusText(text string) error {
	log.Debug("Updating LCD display text: %q", text)
	if err := s.Connect(); err != nil {
		log.Error("Failed to connect for status text update: %v", err)
		return err
	}

	originalText := text
	if len(text) > 15 {
		log.Debug("Truncating LCD text from %d to 15 characters", len(text))
		text = text[:15]
	}

	command := fmt.Sprintf("%s %s;", ShowCommand, text)
	log.Trace("Sending status command: %q", command)
	_, err := s.sendCommand(command)
	if err != nil {
		log.Error("Failed to set LCD status text: %v", err)
	} else {
		if originalText != text {
			log.Info("LCD display updated (truncated): %q", text)
		} else {
			log.Info("LCD display updated: %q", text)
		}
	}
	return err
}

func (s *SensorUnit) sendCommand(cmd string) (string, error) {
	log.Debug("Sending serial command: %q", cmd)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.port == nil {
		log.Debug("Cannot send command - sensor unit not connected")
		return "", errors.New("not connected to sensor unit")
	}

	_, err := s.port.Write([]byte(cmd))
	if err != nil {
		// On write failure, mark as disconnected and release the device
		// In case of disconnect / reconnect scenarios for the USB serial device
		// it will get a new tty assignment if we hold on to the old one
		s.Close()
		log.Error("Failed to write command to serial port: %v", err)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Special handling for show command - it does not return a response
	if strings.HasPrefix(cmd, ShowCommand) {
		return "", nil
	}
	scanner := bufio.NewScanner(s.port)
	var response string

	// Wait for a relevant response based on the command type
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		log.Debug("Received line from sensor unit: %q", line)

		if line == "" {
			continue
		}

		// For read commands, we're looking for the temperature data format
		if strings.HasPrefix(cmd, ReadCommand) && strings.Contains(line, "=") {
			log.Debug("Received temperature data from sensor unit")
			response = line
			break
		}

		// For helo commands, we're looking for the helo response
		if strings.HasPrefix(cmd, HeloCommand) && line == HeloResponse {
			log.Debug("Received handshake response from sensor unit")
			response = line
			break
		}

		log.Warning("Received unexpected line from sensor unit for command %q: %q", cmd, line)
	}

	if err := scanner.Err(); err != nil {
		log.Error("Serial scanner error while reading response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug("Serial command completed successfully")
	return response, nil
}

// clearInputBuffer reads and discards any data in the input buffer
// This helps handle any initialization garbage from the Arduino
func (s *SensorUnit) clearInputBuffer() {
	log.Trace("Clearing serial input buffer")
	tempBuf := make([]byte, 1024)

	// Clear any garbage by reading it off from the line
	totalCleared := 0
	for {
		n, err := s.port.Read(tempBuf)
		log.Debug("Read %d bytes, (%v)", n, err)
		if err != nil || n == 0 {
			break
		}
		totalCleared += n
		displayLen := n
		if displayLen > 50 {
			displayLen = 50
		}
		log.Debug("Cleared %d bytes of serial initialization data: %q", n, string(tempBuf[:displayLen]))
	}

	if totalCleared > 0 {
		log.Debug("Cleared %d total bytes from serial buffer", totalCleared)
	}
	s.port.Flush()
}

// CloseIfUnavailable verifies that the serial device file still exists
// if the device doesn't exist but the sensor unit is marked as connected,
// it will automatically close the connection.
// This should handle cases where the device is unplugged or becomes unavailable
// and is then plugged in again, if we hold the original tty device open when this
// happens it will get a new tty device that will not match the configuration rendering
// the sensorunit inoperable.
func (s *SensorUnit) CloseIfUnavailable() error {
	log.Trace("Checking if serial device %s exists", s.config.Name)

	if !s.connected {
		log.Trace("Not connected to device %s, skipping existence check", s.config.Name)
		return nil
	}

	_, err := os.Stat(s.config.Name)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warning("Serial device %s no longer exists", s.config.Name)

			log.Info("Device %s disappeared but connection marked as active, closing connection", s.config.Name)
			if closeErr := s.Close(); closeErr != nil {
				log.Error("Failed to close connection after device disappeared: %v", closeErr)
				return fmt.Errorf("device %s not found and failed to close connection: %w", s.config.Name, closeErr)
			}

			log.Debug("Successfully closed connection for unavailable device %s", s.config.Name)
			return nil
		}

		// Other stat errors (permission denied, etc.)
		log.Error("Failed to check device %s status: %v", s.config.Name, err)
		return fmt.Errorf("failed to check device status: %w", err)
	}

	log.Trace("Serial device %s exists", s.config.Name)
	return nil
}

// StartDeviceMonitoring starts a goroutine that monitors the serial device existence
// four times per second (every 250ms) and automatically closes the connection
// if the device disappears while marked as connected. Monitoring continues
// as long as the sensorunit is running.
func (s *SensorUnit) StartDeviceMonitoring() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.monitoring {
		log.Debug("Device monitoring already running for %s", s.config.Name)
		return
	}

	log.Info("Starting device monitoring for %s", s.config.Name)
	s.monitoring = true
	s.wg.Add(1)
	go s.runDeviceMonitoring()
}

// StopDeviceMonitoring stops the device monitoring goroutine
func (s *SensorUnit) StopDeviceMonitoring() {
	s.mutex.Lock()
	wasMonitoring := s.monitoring
	s.monitoring = false
	s.mutex.Unlock()

	if !wasMonitoring {
		log.Debug("Device monitoring not running for %s", s.config.Name)
		return
	}

	log.Info("Stopping device monitoring for %s", s.config.Name)
	s.cancel()
	s.wg.Wait()
}

// IsDeviceMonitoringActive returns true if device monitoring is currently running
func (s *SensorUnit) IsDeviceMonitoringActive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.monitoring
}

// runDeviceMonitoring is the main monitoring loop that runs in a separate goroutine
func (s *SensorUnit) runDeviceMonitoring() {
	defer s.wg.Done()

	// Check device existence 4 times per second (250ms interval)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	log.Debug("Device monitoring started for %s", s.config.Name)

	for {
		select {
		case <-s.ctx.Done():
			log.Debug("Device monitoring stopped for %s", s.config.Name)
			return
		case <-ticker.C:
			if err := s.CloseIfUnavailable(); err != nil {
				log.Debug("Device check failed during monitoring: %v", err)
			}
		}
	}
}
