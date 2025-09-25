package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
)

func setupRoutes(r *gin.Engine, p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string) {
	powers := r.Group("powers")

	powers.GET("", getAllPercentages(p, idMapping))
	powers.POST("", setAllPercentages(p, powerMapping))
}
