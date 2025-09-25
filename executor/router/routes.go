package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
)

func SetupRoutes(r *gin.Engine, storage *storage.FileStorage, engine *engine.ControlEngine) {
	engineGroup := r.Group("engine")

	runStorage := engineGroup.Group("programs")
	runStorage.GET("", listAllRuns(storage))
	runStorage.GET(":name", getRun(storage))
	runStorage.DELETE(":name", deleteRun(storage))

	engineControl := engineGroup.Group("running")
	engineControl.GET("", getCurrentProgram(engine))
	engineControl.POST("", startNewProgram(engine))
	engineControl.DELETE("", cancelRunningProgram(engine))

	storageGroup := r.Group("storage")
	storageGroup.GET("programs", listAllPrograms(storage))
	storageGroup.GET("programs/:name", getProgram(storage))
	storageGroup.POST("programs", createProgram(storage, engine))
	storageGroup.POST("programs/:name", updateProgram(storage, engine))
	storageGroup.DELETE("programs/:name", deleteProgram(storage))
}
