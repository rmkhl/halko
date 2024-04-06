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
	programs.GET("current", currentProgram(db.Programs))
	programs.GET(":id", program(db.Programs))

	cycles := v1.Group("cycles")
	cycles.GET("", allCycles(db.Cycles))
	cycles.GET(":id", cycle(db.Cycles))

	phases := v1.Group("phases")
	phases.GET("", allPhases(db.Phases))
	phases.GET(":id", phase(db.Phases))
}
