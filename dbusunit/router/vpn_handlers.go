package router

import (
	"encoding/json"
	"net/http"

	"github.com/rmkhl/halko/dbusunit/dbus"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// listVPNs returns a handler that lists all OpenVPN client services
func listVPNs(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request: GET /vpn")

		vpns, err := manager.ListVPNs()
		if err != nil {
			log.Error("Failed to list VPNs: %v", err)
			http.Error(w, `{"error": "failed to list VPNs"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[[]dbus.VPNStatus]{
			Data: vpns,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode VPN list response: %v", err)
		}
	}
}

// getVPNStatus returns a handler that gets status for a specific VPN
func getVPNStatus(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		log.Debug("Received request: GET /vpn/%s", name)

		if name == "" {
			http.Error(w, `{"error": "VPN name is required"}`, http.StatusBadRequest)
			return
		}

		status, err := manager.GetVPNStatus(name)
		if err != nil {
			log.Error("Failed to get VPN status for %s: %v", name, err)
			http.Error(w, `{"error": "failed to get VPN status"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[*dbus.VPNStatus]{
			Data: status,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode VPN status response: %v", err)
		}
	}
}

// startVPN returns a handler that starts a VPN connection
func startVPN(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		log.Debug("Received request: POST /vpn/%s/start", name)

		if name == "" {
			http.Error(w, `{"error": "VPN name is required"}`, http.StatusBadRequest)
			return
		}

		if err := manager.StartVPN(name); err != nil {
			log.Error("Failed to start VPN %s: %v", name, err)
			http.Error(w, `{"error": "failed to start VPN"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[map[string]string]{
			Data: map[string]string{
				"message": "VPN started successfully",
				"name":    name,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode start VPN response: %v", err)
		}
	}
}

// stopVPN returns a handler that stops a VPN connection
func stopVPN(manager *dbus.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		log.Debug("Received request: POST /vpn/%s/stop", name)

		if name == "" {
			http.Error(w, `{"error": "VPN name is required"}`, http.StatusBadRequest)
			return
		}

		if err := manager.StopVPN(name); err != nil {
			log.Error("Failed to stop VPN %s: %v", name, err)
			http.Error(w, `{"error": "failed to stop VPN"}`, http.StatusInternalServerError)
			return
		}

		response := types.APIResponse[map[string]string]{
			Data: map[string]string{
				"message": "VPN stopped successfully",
				"name":    name,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("Failed to encode stop VPN response: %v", err)
		}
	}
}
