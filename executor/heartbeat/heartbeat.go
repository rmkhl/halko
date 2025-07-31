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
		config     *types.ExecutorConfig
		ctx        context.Context
		cancel     context.CancelFunc
		wg         *sync.WaitGroup
		executorIP string
	}
)

var (
	ErrHeartbeatNotRunning = errors.New("heartbeat not running")
)

func NewManager(config *types.ExecutorConfig) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Get the IP address once during initialization
	executorIP, err := GetNetworkInterfaceIPv4(config.NetworkInterface)
	if err != nil {
		cancel() // Clean up the context since we're returning an error
		return nil, errors.New("failed to get IP address for network interface " + config.NetworkInterface + ": " + err.Error())
	}

	return &Manager{
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
		wg:         new(sync.WaitGroup),
		executorIP: executorIP,
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

	// Send initial heartbeat
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
	// Prepare the heartbeat payload using StatusRequest with IP as message
	payload := types.StatusRequest{
		Message: hm.executorIP,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the heartbeat to the sensor unit's status endpoint
	statusURL := hm.config.SensorUnitURL + "/status"
	resp, err := http.Post(statusURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send heartbeat, status: " + resp.Status)
	}

	return nil
}

// GetNetworkInterfaceIPv4 returns the IPv4 address of the specified network interface
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
