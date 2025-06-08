package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
	"github.com/rmkhl/halko/powerunit/shelly"
)

type Router struct {
	*gin.Engine
}

func New(p *power.Controller, powerMapping map[string]int, idMapping [shelly.NumberOfDevices]string) *Router {
	ginRouter := gin.Default()
	ginRouter.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	r := &Router{ginRouter}
	setupRoutes(ginRouter, p, powerMapping, idMapping)

	return r
}
