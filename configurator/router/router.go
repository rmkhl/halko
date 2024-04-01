package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/configurator/database"
)

type Router struct {
	*gin.Engine
	db database.Interface
}

func New(db database.Interface) *Router {
	ginRouter := gin.Default()

	r := &Router{ginRouter, db}
	setupRoutes(ginRouter, db)

	return r
}
