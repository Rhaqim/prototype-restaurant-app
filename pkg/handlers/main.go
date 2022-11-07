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
		auth.POST("/signin", views.SignIn)
	}
	tokenProtected := auth.Group("/protected")
	tokenProtected.Use(TokenGuardMiddleware())
	{
		tokenProtected.GET("/signout", views.Signout)

	}

	refreshTokenProtected := auth.Group("/protected")
	refreshTokenProtected.Use(RefreshTokenGuardMiddleware())
	{
		refreshTokenProtected.POST("/refreshToken", views.RefreshToken)
		refreshTokenProtected.POST("/forgotPassword", views.ForgotPassword)
		refreshTokenProtected.POST("/changePassword", views.ResetPassword)
	}

	user := router.Group("/user")
	user.GET("/getUserById", views.GetUserByID)
	user.GET("/getUserByEmail", views.GetUserByEmail)
	user.Use(TokenGuardMiddleware())
	{
		user.POST("/createUser", views.CreatNewUser)
		user.PUT("/updateUser", views.UpdateAvatar)
		user.DELETE("/deleteUser", views.DeleteUser)
		// user.GET("/getAllUsers", views.GetAllUsers)

		transactions := user.Group("/transactions")
		{
			transactions.POST("/createTransaction", views.CreateTransaction)
			transactions.PUT("/updateTransaction", views.UpdateTransactionStatus)
		}
	}

	hosting := router.Group("/hosting")
	// hosting.GET("/getAllHostedEvents", views.GetHostingByID)
	hosting.Use(TokenGuardMiddleware())
	{
		hosting.POST("/createHosting", views.CreateHostedEvent)
		hosting.PUT("/updateHosting", views.UpdateHostedEvent)
		hosting.DELETE("/deleteHosting/:id", views.DeleteHostedEvent)
		hosting.GET("/getHostingByHost", views.GetUserHostedEventsByHost)
	}

	return router
}
