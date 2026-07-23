package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

type (
	Manager struct {
		networkInterface string
		apiEndpoints     *types.APIEndpoints
		ctx              context.Context
		cancel           context.CancelFunc
		wg               *sync.WaitGroup
		executorIP       string
		displayMessage   string
		displayMutex     sync.RWMutex
	}
)

var (
	ErrHeartbeatNotRunning = errors.New("heartbeat not running")
)

func NewManager(networkInterface string, apiEndpoints *types.APIEndpoints) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// The network interface may not have an address yet at boot (e.g. WiFi
	// still associating). Don't fail startup for this - sendHeartbeat retries
	// the lookup until an address becomes available.
	executorIP, err := GetNetworkInterfaceIPv4(networkInterface)
	if err != nil {
		log.Warning("Heartbeat: no IP address for interface %s yet, will retry: %v", networkInterface, err)
	}

	return &Manager{
		networkInterface: networkInterface,
		apiEndpoints:     apiEndpoints,
		ctx:              ctx,
		cancel:           cancel,
		wg:               new(sync.WaitGroup),
		executorIP:       executorIP,
		displayMessage:   "idle",
	}, nil
}

func (hm *Manager) Start() error {
	hm.wg.Add(1)
	go hm.run()

	return nil
}

func (hm *Manager) Stop() error {
	select {
	case <-hm.ctx.Done():
		return ErrHeartbeatNotRunning
	default:
	}

	hm.cancel()
	hm.wg.Wait()

	return nil
}

// SetDisplayMessage sets a custom message for the display's status line
func (hm *Manager) SetDisplayMessage(message string) {
	hm.displayMutex.Lock()
	defer hm.displayMutex.Unlock()
	hm.displayMessage = message
	log.Debug("Heartbeat: Display message set to: %s", message)
}

func (hm *Manager) run() {
	defer hm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	if err := hm.sendHeartbeat(); err != nil {
		log.Warning("Error sending initial heartbeat: %v", err)
	}

	for {
		select {
		case <-hm.ctx.Done():
			log.Info("Heartbeat manager stopped")
			return
		case <-ticker.C:
			if err := hm.sendHeartbeat(); err != nil {
				log.Warning("Error sending heartbeat: %v", err)
			}
		}
	}
}

// buildDisplayRequest builds the per-beat display update: the status
// message (defaulting to "idle") plus the executor IP for the display's
// dedicated address line. An empty IP (interface not up yet) leaves the
// address unset so the sensorunit skips the addr command.
func buildDisplayRequest(message, ip string) types.DisplayRequest {
	if message == "" {
		message = "idle"
	}
	return types.DisplayRequest{Message: message, Address: ip}
}

func (hm *Manager) sendHeartbeat() error {
	if hm.executorIP == "" {
		if ip, err := GetNetworkInterfaceIPv4(hm.networkInterface); err == nil {
			hm.executorIP = ip
		}
	}

	hm.displayMutex.RLock()
	customMessage := hm.displayMessage
	hm.displayMutex.RUnlock()

	payload := buildDisplayRequest(customMessage, hm.executorIP)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	displayURL := hm.apiEndpoints.SensorUnit.URL + hm.apiEndpoints.SensorUnit.Display
	resp, err := http.Post(displayURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send heartbeat, status: " + resp.Status)
	}

	return nil
}

func GetNetworkInterfaceIPv4(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if v, ok := addr.(*net.IPNet); ok && v.IP.To4() != nil {
			return v.IP.String(), nil
		}
	}

	return "", errors.New("no IPv4 address found for interface " + interfaceName)
}
