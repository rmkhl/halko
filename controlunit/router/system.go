package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// getSystemStatus returns comprehensive system status including all services and system info
func getSystemStatus(storage types.ExecutionStorage, config *types.HalkoConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := types.SystemStatusResponse{
			Services: getServicesStatus(config.APIEndpoints),
			System:   getSystemInfo(storage),
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.SystemStatusResponse]{Data: response})
	}
}

// getServicesStatus queries all service status endpoints
func getServicesStatus(endpoints *types.APIEndpoints) types.SystemServicesStatus {
	services := types.SystemServicesStatus{
		ControlUnit: queryServiceStatus(endpoints.ControlUnit.GetStatusURL(), "controlunit"),
		PowerUnit:   queryServiceStatus(endpoints.PowerUnit.GetStatusURL(), "powerunit"),
		SensorUnit:  queryServiceStatus(endpoints.SensorUnit.GetStatusURL(), "sensorunit"),
	}
	return services
}

// queryServiceStatus queries a service's /status endpoint
func queryServiceStatus(url, serviceName string) types.ServiceStatusResponse {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Debug("Failed to query %s status: %v", serviceName, err)
		return types.ServiceStatusResponse{
			Status:  types.ServiceStatusUnavailable,
			Service: serviceName,
			Details: map[string]interface{}{"error": err.Error()},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.ServiceStatusResponse{
			Status:  types.ServiceStatusDegraded,
			Service: serviceName,
			Details: map[string]interface{}{"http_status": resp.StatusCode},
		}
	}

	var apiResponse types.APIResponse[types.ServiceStatusResponse]
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Debug("Failed to parse %s status response: %v", serviceName, err)
		return types.ServiceStatusResponse{
			Status:  types.ServiceStatusDegraded,
			Service: serviceName,
			Details: map[string]interface{}{"parse_error": err.Error()},
		}
	}

	return apiResponse.Data
}

// getSystemInfo returns general system information
func getSystemInfo(storage types.ExecutionStorage) types.SystemInfo {
	info := types.SystemInfo{}

	// Read memory information from /proc/meminfo
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			var value int64
			if _, err := fmt.Sscanf(fields[1], "%d", &value); err != nil {
				continue
			}
			// Convert from kB to MB
			valueMB := value / 1024

			switch fields[0] {
			case "MemTotal:":
				info.MemoryTotalMB = valueMB
			case "MemAvailable:":
				if info.MemoryTotalMB > 0 {
					info.MemoryUsedMB = info.MemoryTotalMB - valueMB
				}
			case "SwapTotal:":
				info.SwapTotalMB = valueMB
			case "SwapFree:":
				if info.SwapTotalMB > 0 {
					info.SwapUsedMB = info.SwapTotalMB - valueMB
				}
			}
		}
	}

	// Get disk space from storage abstraction
	info.DiskSpaceMB = storage.GetAvailableSpaceMB()

	// Read system uptime from /proc/uptime
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		var uptimeSeconds float64
		if _, err := fmt.Sscanf(string(data), "%f", &uptimeSeconds); err == nil {
			info.UptimeSeconds = int64(uptimeSeconds)
		}
	}

	return info
}

// getHardwareStatus returns current hardware status (Shelly connectivity)
func getHardwareStatus(endpoints *types.APIEndpoints) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		response := types.HardwareStatusResponse{
			Shelly: getShellyStatus(endpoints.PowerUnit.GetPowerURL()),
		}

		writeJSON(w, http.StatusOK, types.APIResponse[types.HardwareStatusResponse]{Data: response})
	}
}

// getShellyStatus checks if the Shelly device is reachable
func getShellyStatus(baseURL string) types.ShellyStatus {
	status := types.ShellyStatus{
		Reachable: false,
	}

	// Try to query power status as a proxy for Shelly reachability
	client := &http.Client{Timeout: 2 * time.Second}

	// Remove trailing /power if present
	baseURL = strings.TrimSuffix(baseURL, "/power")
	url := baseURL + "/power"
	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			status.Reachable = true
		}
	}

	return status
}
