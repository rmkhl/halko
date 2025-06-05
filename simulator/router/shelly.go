package router

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/engine"
	"github.com/rmkhl/halko/types"
)

func readSwitchStatus(powers map[int8]engine.PowerManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract the switch ID from the query parameters
		switchID := ctx.Query("id")
		if switchID == "" {
			ctx.String(http.StatusBadRequest, "Switch ID is required")
			return
		}

		id, err := strconv.Atoi(switchID)
		if err != nil {
			ctx.String(http.StatusBadRequest, "Invalid Switch ID %s", switchID)
			return
		}

		if power, exists := powers[int8(id)]; exists {
			// Log the switch status
			_, turnedOn := power.Info()

			output := "off"
			if turnedOn {
				output = "on"
			}

			ctx.JSON(http.StatusOK, types.ShellySwitchGetStatusResponse{
				ID:     strconv.Itoa(id),
				Source: "HTTP_in",
				Output: output,
				Temperature: struct {
					TC float32 `json:"tC"`
					TF float32 `json:"tF"`
				}{
					TC: 20.0,
					TF: 68.0,
				},
			})
		} else {
			ctx.String(http.StatusNotFound, "Switch %d not found", id)
		}
	}
}

func setSwitchState(powers map[int8]engine.PowerManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Log all query parameters
		queryParams := ctx.Request.URL.Query()
		for key, values := range queryParams {
			for _, value := range values {
				log.Printf("Query Parameter: %s = %s", key, value)
			}
		}

		ctx.String(http.StatusOK, "%s", "unknown")
	}
}
