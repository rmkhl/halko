package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/powerunit/power"
)

type Router struct {
	*gin.Engine
}

func New(p *power.Controller) *Router {
	ginRouter := gin.Default()
	ginRouter.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	r := &Router{ginRouter}
	setupRoutes(ginRouter, p)

	return r
}

func errorJSON(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func statusAndError(err error) (int, error) {
	switch {
	case err == nil:
		return http.StatusOK, nil
	default:
		return http.StatusInternalServerError, err
	}
}
