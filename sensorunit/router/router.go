package router

import (
	"net/http"
	"sync"

	"github.com/rmkhl/halko/sensorunit/serial"
	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// kilnSensorStatus tracks which of the two kiln temperature sensors are
// currently providing valid readings, so status changes can be logged once
// instead of on every temperature poll.
type kilnSensorStatus int

const (
	kilnSensorBothOK kilnSensorStatus = iota
	kilnSensorPrimaryOnly
	kilnSensorSecondaryOnly
	kilnSensorBothInvalid
)

type API struct {
	sensorUnit *serial.SensorUnit

	statusMu      sync.Mutex
	kilnStatus    kilnSensorStatus
	materialValid bool
}

func NewAPI(sensorUnit *serial.SensorUnit) *API {
	return &API{
		sensorUnit:    sensorUnit,
		materialValid: true,
	}
}

// updateKilnStatus logs a message when the set of usable kiln temperature
// sensors changes, instead of on every poll. Losing one of the two sensors
// is informational (the other is still usable); losing both is an error.
func (api *API) updateKilnStatus(newStatus kilnSensorStatus) {
	api.statusMu.Lock()
	defer api.statusMu.Unlock()

	if newStatus == api.kilnStatus {
		return
	}

	switch newStatus {
	case kilnSensorPrimaryOnly:
		log.Info("Secondary kiln temperature reading is invalid, using primary only")
	case kilnSensorSecondaryOnly:
		log.Info("Primary kiln temperature reading is invalid, using secondary only")
	case kilnSensorBothInvalid:
		log.Error("Both kiln temperature readings are invalid")
	case kilnSensorBothOK:
		log.Info("Kiln temperature readings restored, both primary and secondary available again")
	}

	api.kilnStatus = newStatus
}

// updateMaterialStatus logs a message when the material (wood) temperature
// reading becomes invalid or becomes valid again, instead of on every poll.
func (api *API) updateMaterialStatus(valid bool) {
	api.statusMu.Lock()
	defer api.statusMu.Unlock()

	if valid == api.materialValid {
		return
	}

	if valid {
		log.Info("Material temperature reading restored")
	} else {
		log.Error("Material temperature reading is invalid")
	}

	api.materialValid = valid
}

func SetupRouter(api *API, endpoints *types.APIEndpoints) http.Handler {
	mux := http.NewServeMux()
	SetupRoutes(mux, api, endpoints)
	return addCORSHeaders(mux)
}

func addCORSHeaders(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("HTTP Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			log.Debug("HTTP Response: OPTIONS %s -> 204 No Content", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
