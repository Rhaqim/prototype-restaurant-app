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
		auth.GET("/signin", views.SignIn)
		auth.POST("/signout", views.Signout)
		auth.POST("/refreshToken", views.RefreshToken)
		// auth.POST("/forgotPassword", views.ForgotPassword)
		// auth.POST("/changePassword", views.ChangePassword)
	}

	user := router.Group("/user")
	{
		user.POST("/createUser", views.CreatNewUser)
		user.PUT("/updateUser", views.UpdateAvatar)
		user.DELETE("/deleteUser", views.DeleteUser)
		user.GET("/getUserById", views.GetUserByID)
		user.GET("/getUserByEmail", views.GetUserByID)
		// user.GET("/getAllUsers", views.GetAllUsers)
	}

	hosting := router.Group("/hosting")
	{
		hosting.POST("/createHosting", views.CreateHostedEvent)
		hosting.PUT("/updateHosting", views.UpdateHosting)
	}

	return router
}
