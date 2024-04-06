package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

type Router struct {
	*gin.Engine
	db *database.Interface
}

func New(db *database.Interface) *Router {
	ginRouter := gin.Default()

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
