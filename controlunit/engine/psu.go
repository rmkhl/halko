package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

const (
	psuOven       = "heater"
	psuFan        = "fan"
	psuHumidifier = "humidifier"
)

type (
	PowerCommand struct {
		Percent uint8 `json:"percent"`
	}

	psuController struct {
		client          *http.Client
		powerControlURL string
	}
)

func newPSUController(halkoConfig *types.HalkoConfig, endpoints *types.APIEndpoints) (*psuController, error) {
	if halkoConfig.APIEndpoints == nil {
		return nil, errors.New("API endpoints not configured")
	}

	return &psuController{
		client:          &http.Client{},
		powerControlURL: endpoints.PowerUnit.GetPowerURL(),
	}, nil
}

func newPSUCommand(percentage uint8) *PowerCommand {
	return &PowerCommand{
		Percent: percentage,
	}
}

func (p *psuController) setPower(psu string, percentage uint8) {
	cmd, err := json.Marshal(newPSUCommand(percentage))
	if err != nil {
		log.Error("Error marshalling power command: %v", err)
		return
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", p.powerControlURL, psu), bytes.NewBuffer(cmd))
	if err != nil {
		log.Error("Error creating request: %v", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		log.Error("Error sending request: %v", err)
		return
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Error reading response: %v", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse types.APIErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil && errorResponse.Err != "" {
			log.Error("Cannot set power %s: %s", psu, errorResponse.Err)
		}
		log.Error("Cannot set power %s: %s", psu, response.Status)
	}
}
