package handlers

import (
	ad "github.com/Rhaqim/thedutchapp/pkg/admin"
	views "github.com/Rhaqim/thedutchapp/pkg/controllers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	"github.com/gin-gonic/gin"
)

func GinRouter() *gin.Engine {
	router := gin.Default()

	// Websocket Notification Handler
	// router.GET("/ws", nf.WsHandler)

	/* Auth Routes */
	auth := router.Group("/auth")
	{
		auth.POST("/signUp", views.Signup)
		auth.GET("/verifyEmail", views.VerifyEmail)
		auth.POST("/signIn", views.SignIn)
	}
	tokenProtected := auth.Group("/protected", TokenGuardMiddleware())
	{
		tokenProtected.GET("/signOut", views.Signout)
		tokenProtected.GET("/ws", nf.WsHandler)

	}

	auth.POST("/forgotPassword", views.ForgotPassword)
	refreshTokenProtected := auth.Group("/protected", RefreshTokenGuardMiddleware())
	{
		refreshTokenProtected.POST("/refreshToken", views.RefreshToken)
		refreshTokenProtected.POST("/changePassword", views.ResetPassword)
	}

	/* User Routes */
	user := router.Group("/user")
	user.GET("/getUser", views.GetUser)
	user.Use(TokenGuardMiddleware())
	{
		user.POST("/createUser", views.CreatNewUser)
		user.PUT("/updateUser", views.UpdateUser)
		user.PUT("/updateKyc", views.UpdateUsersKYC)
		user.DELETE("/deleteUser", views.DeleteUser)

		/* Transaction Routes */
		transactions := user.Group("/transactions")
		{
			transactions.POST("/createTransaction", views.CreateTransaction)
			transactions.PUT("/updateTransaction", views.UpdateTransactionStatus)
			transactions.POST("/paybill", views.PayBillforEvent)
			transactions.POST("/sendMoneytoHost", views.SendMoneytoHost)
			transactions.POST("/payOwnBill", views.PayOwnBill)
			transactions.POST("/sendmoney", views.SendToOtherUsers)
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

		/* Wallet Routes */
		wallet := user.Group("/wallet")
		{
			wallet.POST("/create", views.CreateWallet)
			wallet.POST("/fund", views.FundWallet)
			wallet.POST("/pinChange", views.ChangePin)
			wallet.GET("/balance", views.GetWalletBalance)
		}

		/* Notification Routes */
		notification := user.Group("/notification")
		{
			notification.GET("/get", views.GetNotifications)
			notification.POST("/markAs", views.UpdateNotificationStatus)
		}
	}

	/* Event Routes */
	event := router.Group("/event")
	event.GET("/getEvent", views.GetEvent)
	event.GET("/getEvents", views.GetEvents)
	event.Use(TokenGuardMiddleware())
	{
		event.POST("/createEvent", views.CreateEvent)
		event.PUT("/updateEvent", views.UpdateEvent)
		event.DELETE("/deleteEvent/:id", views.DeleteEvent)
		event.GET("/cancel/:id", views.CancelEvent)

		/* Order Routes */
		order := event.Group("/order")
		{
			order.POST("create", views.CreateOrder)
			order.GET("getOrders", views.GetOrders)
			order.GET("getEventOrders/:id", views.GetEventOrders)
			order.GET("getUserEventOrders/:id", views.GetUserEventOrders)
		}

		attend := event.Group("/attend")
		{
			attend.POST("/sendInvites", views.SendEventInvites)
			attend.GET("/acceptInvite", views.AcceptInvite)
			attend.POST("/declineInvite", views.DeclineInvite)
			// attend.GET("/getInvites", views.GetInvites)
			// attend.GET("/getAttendees", views.GetAttendees)
		}
	}

	/* Restaurant Routes */
	restaurant := router.Group("/restaurant")
	restaurant.GET("/getRestaurant", views.GetRestaurant)
	restaurant.GET("/getRestaurants", views.GetRestaurants)
	restaurant.Use(TokenGuardMiddleware())
	{
		restaurant.POST("/create", views.CreateRestaurant)
		restaurant.PUT("/update", views.UpdateRestaurant)
		restaurant.DELETE("/delete", views.DeleteRestaurant)
	}

	/* Product Routes */
	product := router.Group("/product")
	product.GET("/getProduct", views.GetProduct)
	product.GET("/getProducts", views.GetProducts)
	product.Use(TokenGuardMiddleware())
	{
		product.POST("/add", views.AddProduct)
		product.PUT("/update", views.UpdateProduct)
		product.DELETE("/delete", views.DeleteProduct)
	}

	/* Admin Routes */
	admin := router.Group("/admin")
	admin.Use(TokenGuardMiddleware())
	admin.POST("/createAdmin", ad.Create)

	protected := admin.Group("/protected")
	protected.Use(AdminGuardMiddleware())
	{
		protected.POST("/sendNotification", ad.SendNotificationtoUsers)
	}

	return router
}
