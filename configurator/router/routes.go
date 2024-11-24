package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

func setupRoutes(r *gin.Engine, db *database.Interface) {
	api := r.Group("api")
	v1 := api.Group("v1")

	programs := v1.Group("programs")
	programs.GET("", allPrograms(db.Programs))
	programs.GET(":name", program(db.Programs))
	programs.POST("", createProgram(db.Programs))
}
