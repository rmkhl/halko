package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
)

func setupRoutes(r *gin.Engine, p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string) {
	powers := r.Group("powers")
	api := powers.Group("api")
	v1 := api.Group("v1")

	v1.GET("", getAllPercentages(p, idMapping))
	v1.POST("", setAllPercentages(p, powerMapping))
}
