package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type (
	Interface struct {
		m      sync.Mutex
		addr   string
		client *http.Client
	}
	PowerState string
	ID         int
	IDString   string
	powerBool  bool
	apiError   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	powerResp interface {
		Error() error
		PowerState() PowerState
	}
	setPowerStateResp struct {
		apiError
		WasOn powerBool `json:"was_on"`
	}
	getStatusResp struct {
		apiError
		Output powerBool `json:"output"`
	}
	// powerCall returns the api response, the body it should be unmarshaled into
	// and the api error.
	powerCall func() (resp *http.Response, body powerResp, err error)
)

func (p *apiError) Error() error {
	if p.Code != 0 || len(p.Message) != 0 {
		return fmt.Errorf("error code '%d', message '%s'", p.Code, p.Message)
	}
	return nil
}

func (p *setPowerStateResp) PowerState() PowerState {
	return p.WasOn.PowerState()
}

func (p *setPowerStateResp) Error() error {
	return p.apiError.Error()
}

func (p *getStatusResp) PowerState() PowerState {
	return p.Output.PowerState()
}

func (p *getStatusResp) Error() error {
	return p.apiError.Error()
}

const (
	Off     PowerState = "off"
	On      PowerState = "on"
	Unknown PowerState = "unknown"
)

func (p powerBool) PowerState() PowerState {
	if p {
		return On
	}
	return Off
}

const (
	UnknownID ID = iota - 1
	Fan
	Heater
	Humidifier
)

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

func (id IDString) ID() (ID, bool) {
	switch id {
	case "fan":
		return Fan, true
	case "heater":
		return Heater, true
	case "humidifier":
		return Humidifier, true
	default:
		return UnknownID, false
	}
}

func New(addr string) *Interface {
	return &Interface{sync.Mutex{}, addr, http.DefaultClient}
}

func (i *Interface) setStateCall(state PowerState, id ID) powerCall {
	return func() (*http.Response, powerResp, error) {
		body := setPowerStateResp{}
		resp, err := i.client.Get(i.switchSetURI(state, id))
		if err != nil {
			return nil, &body, err
		}
		return resp, &body, nil
	}
}

func (i *Interface) getStatusCall(id ID) powerCall {
	return func() (*http.Response, powerResp, error) {
		body := getStatusResp{}
		apiResp, err := i.client.Get(fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", i.addr, id))
		if err != nil {
			return nil, &body, err
		}
		return apiResp, &body, nil
	}
}

func (i *Interface) SetState(state PowerState, id ID) (PowerState, error) {
	i.m.Lock()
	defer i.m.Unlock()
	// The response contains the state before this call.
	_, err := getPowerState(i.setStateCall(state, id))
	// The response contains the current state.
	pState, getStatusErr := getPowerState(i.getStatusCall(id))
	// err = getStatusErr || err || nil
	if getStatusErr != nil {
		err = getStatusErr
	}
	return pState, err
}

func (i *Interface) GetState(id ID) (PowerState, error) {
	i.m.Lock()
	defer i.m.Unlock()
	return getPowerState(i.getStatusCall(id))
}

func (i *Interface) switchSetURI(state PowerState, id ID) string {
	on := state == On
	return fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=%v", i.addr, id, on)
}

func (i *Interface) Shutdown() error {
	i.m.Lock()
	defer i.m.Unlock()
	failed := []string{}
	for _, id := range []ID{Fan, Heater, Humidifier} {
		if _, err := i.SetState(Off, id); err != nil {
			failed = append(failed, id.String())
		}
	}
	if len(failed) == 0 {
		return nil
	}
	return fmt.Errorf("failed to shut down Shelly power for %s", strings.Join(failed, ", "))
}

func getPowerState(powerCall powerCall) (PowerState, error) {
	resp, body, err := powerCall()
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(body); err != nil {
		return Unknown, err
	}
	return body.PowerState(), body.Error()
}
