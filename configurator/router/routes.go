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
	programs.GET(":id", program(db.Programs))
	programs.POST("", createProgram(db.Programs))
	programs.PUT(":id", updateProgram(db.Programs))

	phases := v1.Group("phases")
	phases.GET("", allPhases(db.Phases))
	phases.GET(":id", phase(db.Phases))
	phases.POST("", createPhase(db.Phases))
	phases.PUT(":id", updatePhase(db.Phases))
}
