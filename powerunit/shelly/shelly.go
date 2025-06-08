package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PowerState represents the power state of a Shelly device
type PowerState string

const (
	// Power states
	Off     PowerState = "off"
	On      PowerState = "on"
	Unknown PowerState = "unknown"

	NumberOfDevices = 3 // Number of devices controlled by Shelly
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
func (s *Shelly) GetState(id int) (PowerState, error) {
	url := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", s.address, id)
	resp, err := s.client.Get(url)
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()

	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return Unknown, err
	}

	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	if statusResp.Output {
		return On, nil
	}
	return Off, nil
}

func (s *Shelly) SetState(state PowerState, id int) (PowerState, error) {
	on := state == On

	url := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=%v", s.address, id, on)
	resp, err := s.client.Get(url)
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()

	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return Unknown, err
	}

	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	return state, nil
}

// Shutdown all powers
func (s *Shelly) Shutdown() error {
	for id := range 3 {
		if _, err := s.SetState(Off, id); err != nil {
			return fmt.Errorf("failed to shut down device %d: %w", id, err)
		}
	}
	return nil
}
