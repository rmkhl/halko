package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rmkhl/halko/types"
)

const (
	psuOven       = "heater"
	psuFan        = "fan"
	psuHumidifier = "humidifier"
)

type (
	PowerCommand struct {
		Command PowerStatus `json:"command"`
		Percent uint8       `json:"percent,omitempty"`
	}

	psuController struct {
		client          *http.Client
		powerControlURL string
	}
)

func newPSUController(config *types.ExecutorConfig) (*psuController, error) {
	return &psuController{
		client:          &http.Client{},
		powerControlURL: config.PowerControlURL,
	}, nil
}

func newPSUCommand(percentage uint8) *PowerCommand {
	cmd := PowerCommand{
		Percent: percentage,
	}
	if percentage == 0 {
		cmd.Command = PowerOff
	} else {
		cmd.Command = PowerOn
	}
	return &cmd
}

func (p *psuController) setPower(psu string, percentage uint8) {
	cmd, err := json.Marshal(newPSUCommand(percentage))
	if err != nil {
		log.Printf("Error marshalling power command: %v\n", err)
		return
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", p.powerControlURL, psu), bytes.NewBuffer(cmd))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading response: %v\n", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse types.APIErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil && errorResponse.Err != "" {
			log.Printf("Cannot set power %s: %s\n", psu, errorResponse.Err)
		}
		log.Printf("Cannot set power %s: %s\n", psu, response.Status)
	}
}
