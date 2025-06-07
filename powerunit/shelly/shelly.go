package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PowerState represents the power state of a Shelly device
type PowerState string

// ID represents the identifier for a specific power device
type ID int

const (
	// Power states
	Off     PowerState = "off"
	On      PowerState = "on"
	Unknown PowerState = "unknown"

	// Power device IDs
	UnknownID ID = iota - 1
	Fan
	Heater
	Humidifier
)

// Shelly represents a Shelly device controller
type Shelly struct {
	address string
	client  *http.Client
}

// Response structures for API calls
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type getStatusResponse struct {
	apiError
	Output bool `json:"output"`
}

// String returns the string representation of a power device ID
func (id ID) String() string {
	switch id {
	case Fan:
		return "fan"
	case Heater:
		return "heater"
	case Humidifier:
		return "humidifier"
	default:
		return "unknown"
	}
}

// New creates a new Shelly controller with the specified address
func New(address string) *Shelly {
	return &Shelly{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetState retrieves the current power state of a specified device
func (s *Shelly) GetState(id ID) (PowerState, error) {
	// Make API request
	url := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", s.address, id)
	resp, err := s.client.Get(url)
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()

	// Parse response
	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return Unknown, err
	}

	// Check for API error
	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	// Return power state
	if statusResp.Output {
		return On, nil
	}
	return Off, nil
}

// SetState sets the power state of a specified device
func (s *Shelly) SetState(state PowerState, id ID) (PowerState, error) {
	// Convert state to boolean for API request
	on := state == On

	// Make API request to set state
	url := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=%v", s.address, id, on)
	resp, err := s.client.Get(url)
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()

	// Parse response
	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return Unknown, err
	}

	// Check for API error
	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	// Return the new state
	return state, nil
}

// Shutdown turns off all devices
func (s *Shelly) Shutdown() error {
	for _, id := range []ID{Fan, Heater, Humidifier} {
		if _, err := s.SetState(Off, id); err != nil {
			return fmt.Errorf("failed to shut down %s: %w", id.String(), err)
		}
	}
	return nil
}
