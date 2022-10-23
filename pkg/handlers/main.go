package handlers

import (
	views "github.com/Rhaqim/thedutchapp/pkg/controllers"
	"github.com/gin-gonic/gin"
)

func GinRouter() *gin.Engine {
	router := gin.Default()

	auth := router.Group("/auth")
	{
		auth.POST("/signup", views.Signup)
		auth.GET("/signin", views.GetUserByID)
		auth.POST("/signout", views.UpdateAvatar)
		auth.POST("/refreshToken", views.SignIn)
		auth.POST("/forgotPassword", views.SignIn)
		auth.POST("/changePassword", views.SignIn)
	}

	user := router.Group("/user")
	{
		user.POST("/createUser", views.GetUserByID)
		user.PUT("/updateUser", views.UpdateAvatar)
		user.DELETE("/deleteUser", views.DeleteUser)
		user.GET("/getUserById", views.GetUserByID)
		user.GET("/getUserByEmail", views.GetUserByID)
		user.GET("/getAllUsers", views.GetUserByID)
	}

	return router
}
