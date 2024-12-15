package shelly

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	powerResp  struct {
		WasOn powerBool `json:"was_on"`
	}
)

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

func (i *Interface) SetState(state PowerState, id ID) (PowerState, error) {
	i.m.Lock()
	defer i.m.Unlock()
	return getPowerState(func() (*http.Response, error) {
		return i.client.Get(i.switchSetURI(state, id))
	})
}

func (i *Interface) GetState(id ID) (PowerState, error) {
	i.m.Lock()
	defer i.m.Unlock()
	return getPowerState(func() (*http.Response, error) {
		return i.client.Get(fmt.Sprintf("http://%s/rpc/Switch.GetStatus?id=%d", i.addr, id))
	})
}

func (i *Interface) switchSetURI(state PowerState, id ID) string {
	on := false
	if state == On {
		on = true
	}
	return fmt.Sprintf("http://%s/rpc/Switch.Set?id=%d&on=%v", i.addr, id, on)
}

func getPowerState(powerCall func() (*http.Response, error)) (PowerState, error) {
	resp, err := powerCall()
	if err != nil {
		return Unknown, err
	}
	defer resp.Body.Close()
	powerResp := powerResp{}
	if err = json.NewDecoder(resp.Body).Decode(&powerResp); err != nil {
		return Unknown, err
	}
	return powerResp.WasOn.PowerState(), nil
}
