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

// decodeSwitchResponse reads an RPC reply, rejecting anything that is not a
// successful response. The HTTP status has to be checked before the body is
// trusted: a failed request whose body still decodes would otherwise look like
// a switch that is off (GetState) or a command that landed (SetState).
func decodeSwitchResponse(resp *http.Response) (*getStatusResponse, error) {
	var statusResp getStatusResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&statusResp)

	if resp.StatusCode != http.StatusOK {
		// Gen2 devices report RPC failures as JSON even on a non-200, so
		// prefer that message when present and fall back to the status line.
		if decodeErr == nil && len(statusResp.Message) != 0 {
			return nil, fmt.Errorf("API error: status '%s', code '%d', message '%s'",
				resp.Status, statusResp.Code, statusResp.Message)
		}
		return nil, fmt.Errorf("unexpected HTTP status '%s'", resp.Status)
	}

	if decodeErr != nil {
		return nil, decodeErr
	}

	if statusResp.Code != 0 || len(statusResp.Message) != 0 {
		return nil, fmt.Errorf("API error: code '%d', message '%s'", statusResp.Code, statusResp.Message)
	}

	return &statusResp, nil
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

	statusResp, err := decodeSwitchResponse(resp)
	if err != nil {
		log.Warning("Failed to read state for device %d: %v", id, err)
		return Unknown, err
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

	if _, err := decodeSwitchResponse(resp); err != nil {
		log.Warning("Failed to set device %d to %s: %v", id, state, err)
		return Unknown, err
	}

	log.Debug("Successfully set device %d to %s", id, state)
	return state, nil
}

func (s *Shelly) Shutdown() error {
	log.Info("Shutting down all Shelly devices")
	for id := range NumberOfDevices {
		if _, err := s.SetState(Off, id); err != nil {
			log.Error("Failed to shut down device %d: %v", id, err)
			return fmt.Errorf("failed to shut down device %d: %w", id, err)
		}
	}
	log.Debug("All Shelly devices shut down successfully")
	return nil
}
