package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
)

func SetupRoutes(r *gin.Engine, storage *storage.ProgramStorage, engine *engine.ControlEngine) {
	engineAPI := r.Group("engine/api")
	engineAPIV1 := engineAPI.Group("v1")

	programStorage := engineAPIV1.Group("programs")
	programStorage.GET("", listAllPrograms(storage))
	programStorage.GET(":name", getProgram(storage))
	programStorage.DELETE(":name", deleteProgram(storage))

	engineControl := engineAPIV1.Group("running")
	engineControl.GET("", getCurrentProgram(engine))
	engineControl.POST("", startNewProgram(engine))
	engineControl.DELETE("", cancelRunningProgram(engine))
}
