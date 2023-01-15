package handlers

import (
	"context"
	"log"
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

// TokenGuardMiddleware is a middleware to check if the token is valid
// User must be signed in to access this resource
// This middleware is used to protect routes
// It will check if the token is valid and if the user exists
// If the user does not exist, it will return an error
// If the user exists, it will set the user in the context
// The user can then be accessed in the controller
// Example:
//
//	func GetProfile(c *gin.Context) {
//		user := c.MustGet("user").(hp.UserResponse)
//		c.JSON(http.StatusOK, user)
//	}
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
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		var user = hp.UserResponse{}
		filter := bson.M{"email": claims.Email}
		options := hp.PasswordOpts
		if err := usersCollection.FindOne(ctx, filter, options).Decode(&user); err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(nil, err.Error(), ut.GetFunctionName())
			c.JSON(http.StatusInternalServerError, response)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// RefreshTokenGuardMiddleware is a middleware to check if the refresh token is valid
// User does not have to be signed in to access this resource
func RefreshTokenGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer database.ConnectMongoDB().Disconnect(context.TODO())

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
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		var user = hp.UserResponse{}
		filter := bson.M{"email": claims.Email}
		options := hp.PasswordOpts
		if err := usersCollection.FindOne(ctx, filter, options).Decode(&user); err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(nil, err.Error(), ut.GetFunctionName())
			c.JSON(http.StatusInternalServerError, response)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// AdminGuardMiddleware is a middleware to check if the user is an admin
// User has to be signed in to access this resource
// User has to be an admin to access this resource
func AdminGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := hp.GetUserFromToken(c)

		hp.SetInfo("User: "+user.Email+
			"User Role: "+user.Role.String(),
			ut.GetFunctionName())

		if user.Role != hp.Admin {
			response := hp.SetError(nil, "You are not authorized to access this resource!", ut.GetFunctionName())
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}
		c.Next()
	}
}

// cookieGuardMiddleware is a middleware to check if the cookie is valid
func CookieGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer database.ConnectMongoDB().Disconnect(context.TODO())

		cookie, err := c.Cookie("token")
		log.Println("Cookie gotten is: ", cookie)
		if err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(err, "Invalid Token!", ut.GetFunctionName())
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		claims, err := auth.VerifyToken(cookie)
		if err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(err, "Invalid Token!", ut.GetFunctionName())
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		var user = hp.UserResponse{}
		filter := bson.M{"email": claims.Email}
		options := hp.PasswordOpts
		if err := usersCollection.FindOne(ctx, filter, options).Decode(&user); err != nil {
			hp.SetDebug(err.Error(), ut.GetFunctionName())
			response := hp.SetError(nil, err.Error(), ut.GetFunctionName())
			c.JSON(http.StatusInternalServerError, response)
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
