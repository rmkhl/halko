package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func readSwitchStatus() gin.HandlerFunc {
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

func setSwitchState() gin.HandlerFunc {
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
