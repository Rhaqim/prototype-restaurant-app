package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var usersCollection = config.UserCollection

func TokenGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer database.ConnectMongoDB().Disconnect(context.TODO())

		token := c.Request.Header.Get("Authorization")
		if token == "" {
			hp.SetDebug("Missing Token required!", ut.GetFunctionName())
			response := hp.SetError(nil, "Missing Token required!", ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		token = strings.Replace(token, "Bearer ", "", 1)

		claims, err := auth.VerifyToken(token)
		if err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(err, "Invalid Token!", ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		var user = hp.UserResponse{}
		filter := bson.M{"email": claims.Email}
		if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(nil, err.Error(), ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func RefreshTokenGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer database.ConnectMongoDB().Disconnect(context.TODO())

		// response := hp.MongoJsonResponse{
		// 	Type: "error",
		// 	Data: nil,
		// 	Date: time.Now(),
		// }

		token := c.Request.Header.Get("Authorization")
		if token == "" {
			response := hp.SetError(nil, "Missing Token required!", ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		token = strings.Replace(token, "Bearer ", "", 1)

		claims, err := auth.VerifyRefreshToken(token)
		if err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(err, "Invalid Token!", ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		var user = hp.UserResponse{}
		filter := bson.M{"email": claims.Email}
		if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(nil, err.Error(), ut.GetFunctionName())
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
