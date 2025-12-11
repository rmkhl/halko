package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/rmkhl/halko/types"
)

type (
	Manager struct {
		networkInterface string
		apiEndpoints     *types.APIEndpoints
		ctx              context.Context
		cancel           context.CancelFunc
		wg               *sync.WaitGroup
		executorIP       string
	}
)

var (
	ErrHeartbeatNotRunning = errors.New("heartbeat not running")
)

func NewManager(networkInterface string, apiEndpoints *types.APIEndpoints) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	executorIP, err := GetNetworkInterfaceIPv4(networkInterface)
	if err != nil {
		cancel() // Clean up the context since we're returning an error
		return nil, errors.New("failed to get IP address for network interface " + networkInterface + ": " + err.Error())
	}

	return &Manager{
		networkInterface: networkInterface,
		apiEndpoints:     apiEndpoints,
		ctx:              ctx,
		cancel:           cancel,
		wg:               new(sync.WaitGroup),
		executorIP:       executorIP,
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

func (hm *Manager) run() {
	defer hm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	if err := hm.sendHeartbeat(); err != nil {
		log.Printf("Error sending initial heartbeat: %v", err)
	}

	for {
		select {
		case <-hm.ctx.Done():
			log.Println("Heartbeat manager stopped")
			return
		case <-ticker.C:
			if err := hm.sendHeartbeat(); err != nil {
				log.Printf("Error sending heartbeat: %v", err)
			}
		}
	}
}

func (hm *Manager) sendHeartbeat() error {
	payload := types.DisplayRequest{
		Message: hm.executorIP,
	}

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
