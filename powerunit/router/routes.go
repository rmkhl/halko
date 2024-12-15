package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
)

var powers = []string{"fan", "heater", "humidifier"}

func setupRoutes(r *gin.Engine, p *power.Controller) {
	api := r.Group("api")
	v1 := api.Group("v1")

	powers := v1.Group("powers")

	powers.GET("", statusAllPowers(p))
	powers.GET(":power", statusPower(p))

	powers.POST(":power", operatePower(p))
	powers.PUT(":power", operatePower(p))
	powers.PATCH(":power", operatePower(p))
}
