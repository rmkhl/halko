package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	return &Shelly{
		address: address,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

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

func (s *Shelly) Shutdown() error {
	for id := range 3 {
		if _, err := s.SetState(Off, id); err != nil {
			return fmt.Errorf("failed to shut down device %d: %w", id, err)
		}
	}
	return nil
}
