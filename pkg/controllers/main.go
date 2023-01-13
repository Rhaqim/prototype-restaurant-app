package controllers

import (
	"context"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	"github.com/gin-gonic/gin"
)

type handlerFunc func(*gin.Context, context.Context)

func AbstractConnection(fn handlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
		defer cancel()
		defer database.DisconnectMongoDB()
		fn(c, ctx)
	}
}

var AuthViews = map[string]handlerFunc{
	"signin":  signin,
	"signout": signout,
}

type Route struct {
	Method      string
	Path        string
	HandlerFunc handlerFunc
}

var AuthRoutes = []Route{
	{
		Method:      "POST",
		Path:        "/signin",
		HandlerFunc: signin,
	},
	{
		Method:      "GET",
		Path:        "/signout",
		HandlerFunc: signout,
	},
}
