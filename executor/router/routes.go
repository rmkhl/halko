package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/executor/engine"
	"github.com/rmkhl/halko/executor/storage"
)

func SetupRoutes(r *gin.Engine, storage *storage.FileStorage, engine *engine.ControlEngine) {
	engineAPI := r.Group("engine/api")
	engineAPIV1 := engineAPI.Group("v1")

	runStorage := engineAPIV1.Group("programs")
	runStorage.GET("", listAllRuns(storage))
	runStorage.GET(":name", getRun(storage))
	runStorage.DELETE(":name", deleteRun(storage))

	engineControl := engineAPIV1.Group("running")
	engineControl.GET("", getCurrentProgram(engine))
	engineControl.POST("", startNewProgram(engine))
	engineControl.DELETE("", cancelRunningProgram(engine))

	storageAPI := r.Group("storage/api")
	storageAPIV1 := storageAPI.Group("v1")
	storageAPIV1.GET("programs", listAllPrograms(storage))
	storageAPIV1.GET("programs/:name", getProgram(storage))
	storageAPIV1.POST("programs", createProgram(storage))
	storageAPIV1.POST("programs/:name", updateProgram(storage))
	storageAPIV1.DELETE("programs/:name", deleteProgram(storage))
}
