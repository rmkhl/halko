package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rmkhl/halko/types/log"
)

type PowerState string

const (
	Off     PowerState = "off"
	On      PowerState = "on"
	Unknown PowerState = "unknown"

	NumberOfDevices = 3 // Number of devices controlled by Shelly
)

type Shelly struct {
	address string
	client  *http.Client
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type getStatusResponse struct {
	apiError
	Output bool `json:"output"`
}

func New(address string) *Shelly {
	log.Debug("Creating Shelly client for address: %s", address)
	return &Shelly{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *Shelly) GetState(id int) (PowerState, error) {
	url := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", s.address, id)
	log.Trace("Getting state for device %d: %s", id, url)

	resp, err := s.client.Get(url)
	if err != nil {
		log.Error("HTTP request failed for device %d: %v", id, err)
		return Unknown, err
	}
	defer resp.Body.Close()

	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		log.Error("Failed to decode response for device %d: %v", id, err)
		return Unknown, err
	}

	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		log.Warning("Shelly API error for device %d: code=%d, message=%s", id, statusResp.Code, statusResp.Message)
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	state := Off
	if statusResp.Output {
		state = On
	}
	log.Trace("Device %d state: %s", id, state)
	return state, nil
}

func (s *Shelly) SetState(state PowerState, id int) (PowerState, error) {
	on := state == On
	url := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=%v", s.address, id, on)
	log.Trace("Setting device %d to %s: %s", id, state, url)

	resp, err := s.client.Get(url)
	if err != nil {
		log.Error("HTTP request failed when setting device %d to %s: %v", id, state, err)
		return Unknown, err
	}
	defer resp.Body.Close()

	var statusResp getStatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		log.Error("Failed to decode response when setting device %d to %s: %v", id, state, err)
		return Unknown, err
	}

	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		log.Warning("Shelly API error when setting device %d to %s: code=%d, message=%s", id, state, statusResp.Code, statusResp.Message)
		return Unknown, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	log.Debug("Successfully set device %d to %s", id, state)
	return state, nil
}

func (s *Shelly) Shutdown() error {
	log.Info("Shutting down all Shelly devices")
	for id := range 3 {
		if _, err := s.SetState(Off, id); err != nil {
			log.Error("Failed to shut down device %d: %v", id, err)
			return fmt.Errorf("failed to shut down device %d: %w", id, err)
		}
	}
	log.Debug("All Shelly devices shut down successfully")
	return nil
}
