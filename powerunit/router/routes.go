package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
)

func setupRoutes(r *gin.Engine, p *power.Controller) {
	powers := r.Group("powers")
	api := powers.Group("api")
	v1 := api.Group("v1")

	v1.GET("", statusAllPowers(p))
	v1.POST("", operateAllPowers(p))
}
