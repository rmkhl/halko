package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	log.Debug("HTTP Response: %d", statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("Failed to encode JSON response: %v", err)
		_ = err
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	log.Debug("HTTP Error Response: %d - %s", statusCode, message)
	writeJSON(w, statusCode, types.APIErrorResponse{Err: message})
}

func (api *API) getTemperatures(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing temperature request from %s", r.RemoteAddr)

	// Attempt to get temperatures, with retry if all readings are invalid
	var temperatures []serial.Temperature
	var err error
	const maxAttempts = 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		temperatures, err = api.sensorUnit.GetTemperatures()
		if err != nil {
			log.Error("Failed to get temperatures from sensor unit (attempt %d/%d): %v", attempt, maxAttempts, err)
			if attempt == maxAttempts {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			log.Warning("Retrying temperature read in 500ms...")
			time.Sleep(500 * time.Millisecond)
			continue
		}
		log.Debug("Retrieved %d temperature readings from sensor unit (attempt %d/%d)", len(temperatures), attempt, maxAttempts)

		response, dies := temperatureResponseFrom(temperatures)
		api.storeDieReadings(dies)
		kilnPrimary := response["kiln_primary"]
		kilnSecondary := response["kiln_secondary"]

		log.Debug("Temperature readings processed (attempt %d/%d): KilnPrimary=%.2f°C, KilnSecondary=%.2f°C, Material=%.2f°C, dies %.2f/%.2f/%.2f°C",
			attempt, maxAttempts, kilnPrimary, kilnSecondary, response["material"],
			dies["kiln_primary_die"], dies["kiln_secondary_die"], dies["material_die"])

		// Check if all readings are invalid
		allInvalid := (kilnPrimary == types.InvalidTemperatureReading &&
			kilnSecondary == types.InvalidTemperatureReading &&
			response["material"] == types.InvalidTemperatureReading)

		if allInvalid && attempt < maxAttempts {
			log.Warning("All temperature readings are invalid on attempt %d/%d, retrying in 500ms...", attempt, maxAttempts)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Process temperature readings (either some are valid or this is the final attempt)
		switch {
		case kilnPrimary != types.InvalidTemperatureReading && kilnSecondary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorBothOK)
		case kilnPrimary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorPrimaryOnly)
		case kilnSecondary != types.InvalidTemperatureReading:
			api.updateKilnStatus(kilnSensorSecondaryOnly)
		default:
			api.updateKilnStatus(kilnSensorBothInvalid)
		}
		api.updateMaterialStatus(response["material"] != types.InvalidTemperatureReading)

		log.Debug("Returning temperature data: kiln primary=%.1f°C, secondary=%.1f°C, material=%.1f°C",
			response["kiln_primary"], response["kiln_secondary"], response["material"])
		writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
			Data: response,
		})
		return
	}
}

// The probe names as the device reports them. They are the wire contract with
// the firmware, so they live in one place rather than being spelled out at
// each use.
const (
	probeKilnPrimary      = "KilnPrimary"
	probeKilnSecondary    = "KilnSecondary"
	probeWood             = "Wood"
	probeKilnPrimaryDie   = "KilnPrimaryDie"
	probeKilnSecondaryDie = "KilnSecondaryDie"
	probeWoodDie          = "WoodDie"
)

// temperatureResponseFrom turns one device read into the two maps this service
// serves: the temperatures, and the cold junctions the die endpoint answers
// from. A probe the device did not report reads as the invalid sentinel, never
// as zero degrees.
//
// Both kiln readings are reported unresolved. Which one the controller acts on
// is configured in the control unit; this service's job is to say what the
// probes read.
//
// The cold junctions stay out of the temperature response: they are diagnostics
// about the measurement, not temperatures the system controls on. They are kept
// per chip rather than folded into one value because whether they moved
// together is what says a shift is the board and not the kiln.
func temperatureResponseFrom(readings []serial.Temperature) (types.TemperatureResponse, types.TemperatureResponse) {
	response := types.TemperatureResponse{
		"kiln_primary":   types.InvalidTemperatureReading,
		"kiln_secondary": types.InvalidTemperatureReading,
		"material":       types.InvalidTemperatureReading,
	}
	dies := make(types.TemperatureResponse, dieSensorCount)

	for _, reading := range readings {
		switch reading.Name {
		case probeKilnPrimary:
			response["kiln_primary"] = reading.Value
		case probeKilnSecondary:
			response["kiln_secondary"] = reading.Value
		case probeWood:
			response["material"] = reading.Value
		case probeKilnPrimaryDie:
			dies["kiln_primary_die"] = reading.Value
		case probeKilnSecondaryDie:
			dies["kiln_secondary_die"] = reading.Value
		case probeWoodDie:
			dies["material_die"] = reading.Value
		}
	}

	return response, dies
}

// dieSensorCount is how many cold junctions the unit reports, one per chip.
const dieSensorCount = 3

// getDieTemperatures serves the cold junction readings recorded by the last
// temperature read. It deliberately does not trigger a read of its own: the
// values move with the sensor board, not the kiln, so a poll-old value says
// the same thing as a fresh one and the serial link stays free for the
// readings the run depends on.
func (api *API) getDieTemperatures(w http.ResponseWriter, r *http.Request) {
	log.Debug("Processing die temperature request from %s", r.RemoteAddr)

	writeJSON(w, http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: api.dieReadings(),
	})
}
