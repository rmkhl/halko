package router

import (
	"log"

	"net/http"

	"github.com/rmkhl/halko/types"

	"github.com/gin-gonic/gin"
)

// getTemperatures handles GET requests to fetch temperature data
func (api *API) getTemperatures(c *gin.Context) {
	temperatures, err := api.sensorUnit.GetTemperatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.APIErrorResponse{
			Err: err.Error(),
		})
		return
	}

	response := make(types.TemperatureResponse)

	// Map the sensor values to the expected keys
	var ovenPrimary float32
	var ovenSecondary float32

	for _, temp := range temperatures {
		log.Printf("Temperature %s: %.2f", temp.Name, temp.Value)
		switch temp.Name {
		case "OvenPrimary":
			ovenPrimary = temp.Value
		case "OvenSecondary":
			ovenSecondary = temp.Value
		case "Wood":
			response["material"] = temp.Value
		}
	}
	// in case the primary or secondary temperature is not available we only use the other one
	// if both are available we use the higher one
	switch {
	case ovenPrimary != types.InvalidTemperatureReading && ovenSecondary != types.InvalidTemperatureReading:
		if ovenPrimary > ovenSecondary {
			response["oven"] = ovenPrimary
		} else {
			response["oven"] = ovenSecondary
		}
	case ovenPrimary != types.InvalidTemperatureReading:
		log.Println("Secondary oven temperature reading is invalid.")
		response["oven"] = ovenPrimary
	case ovenSecondary != types.InvalidTemperatureReading:
		log.Println("Primary oven temperature reading is invalid.")
		response["oven"] = ovenSecondary
	default:
		log.Println("Oven temperature reading is invalid.")
		response["oven"] = types.InvalidTemperatureReading
	}
	if response["material"] == types.InvalidTemperatureReading {
		log.Println("Wood temperature reading is invalid.")
	}

	c.JSON(http.StatusOK, types.APIResponse[types.TemperatureResponse]{
		Data: response,
	})
}
