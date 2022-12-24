package handlers

import (
	views "github.com/Rhaqim/thedutchapp/pkg/controllers"
	"github.com/gin-gonic/gin"
)

func GinRouter() *gin.Engine {
	router := gin.Default()

	/* Auth Routes */
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

	/* User Routes */
	user := router.Group("/user")
	user.GET("/getUserById", views.GetUserByID)
	user.GET("/getUserByEmail", views.GetUserByEmail)
	user.Use(TokenGuardMiddleware())
	{
		user.POST("/createUser", views.CreatNewUser)
		user.PUT("/updateUser", views.UpdateAvatar)
		user.PUT("/updateUserKyc", views.UpdateUsersKYC)
		user.DELETE("/deleteUser", views.DeleteUser)

		/* Transaction Routes */
		transactions := user.Group("/transactions")
		{
			transactions.POST("/createTransaction", views.CreateTransaction)
			transactions.PUT("/updateTransaction", views.UpdateTransactionStatus)
			transactions.GET("/getTransactions", views.GetTransactions)
		}

		/* Social Routes */
		social := user.Group("/social")
		{
			social.POST("/sendFriendRequest", views.SendFriendRequest)
			social.POST("/acceptFriendRequest", views.AcceptFriendRequest)
			social.POST("/declineFriendRequest", views.DeclineFriendRequest)
			social.POST("/block", views.BlockUser)
			social.POST("/unblock", views.UnblockUser)
			social.GET("/getFriends", views.GetFriends)
			social.GET("/getFriendRequests", views.GetFriendRequests)
			social.GET("/getBlockedUsers", views.GetBlockedUsers)
		}
	}

	/* Event Routes */
	event := router.Group("/event")
	// event.GET("/getAllEvents", views.GetEventByID)
	event.Use(TokenGuardMiddleware())
	{
		event.POST("/createEvent", views.CreateEvent)
		event.PUT("/updateEvent", views.UpdateEvent)
		event.DELETE("/deleteEvent/:id", views.DeleteEvent)
		event.GET("/getEventByHost", views.GetUserEventsByHost)
		event.POST("order", views.CreateOrder)
	}

	return router
}
