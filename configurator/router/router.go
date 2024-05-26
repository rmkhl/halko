package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

var ErrNoName = errors.New("no name provided")

type Router struct {
	*gin.Engine
	db *database.Interface
}

func New(db *database.Interface) *Router {
	ginRouter := gin.Default()
	ginRouter.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"http://localhost:1234"},
		AllowMethods:  []string{"GET", "POST", "PUT"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	r := &Router{ginRouter, db}
	setupRoutes(ginRouter, db)

	return r
}

func errorJSON(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func statusAndError(err error) (int, error) {
	switch {
	case err == nil:
		return http.StatusOK, nil
	case errors.Is(err, database.ErrNotFound):
		return http.StatusNotFound, err
	default:
		return http.StatusInternalServerError, err
	}
}
