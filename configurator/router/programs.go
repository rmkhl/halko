package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

func current(programs database.Programs) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curr, err := programs.Current()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, fmt.Errorf("error fetching current program: %w", err))
			return
		}

		ctx.JSON(http.StatusOK, curr)
	}
}
