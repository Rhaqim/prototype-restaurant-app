package handlers

import (
	ad "github.com/Rhaqim/thedutchapp/pkg/admin"
	views "github.com/Rhaqim/thedutchapp/pkg/controllers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GinRouter() *gin.Engine {
	router := gin.Default()

	// Setting the trusted proxies
	router.SetTrustedProxies([]string{
		"http://localhost:3000",
	})

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
	}))

	/* Auth Routes */
	auth := router.Group("/auth")
	{
		auth.POST("/signin", views.SignIn)
		auth.POST("/signup", views.Signup)
		auth.GET("/verify_email", views.VerifyEmail)
		auth.GET("/forgot_password", views.ForgotPassword)
		auth.GET("/verify_password_reset", views.VerifyPasswordResetCode)
	}
	auth.Use(TokenGuardMiddleware())
	{
		auth.GET("/signout", views.Signout)
		auth.GET("/ws", nf.WsHandler)

	}
	/* For frontend login */
	cookie := router.Group("/cookie", CookieGuardMiddleware())
	{
		cookie.GET("/ws", nf.WsHandler)
	}

	refreshTokenProtected := auth.Group("/protected", RefreshTokenGuardMiddleware())
	{
		refreshTokenProtected.POST("/refresh_token", views.RefreshToken)
		refreshTokenProtected.POST("/reset_password", views.ResetPassword)
		refreshTokenProtected.POST("/update_password", views.UpdatePassword)
	}

	/* User Routes */
	user := router.Group("/user")
	user.GET("/get_profile", views.GetUser)
	user.GET("/set_users_cache", views.SetUsersCache)
	user.GET("/get_users_cache", views.GetUsersCache)
	user.Use(TokenGuardMiddleware())
	{
		user.PUT("/update_profile", views.UpdateUser)
		user.PUT("/begin_kyc", views.BegingKycVerification)
		user.PUT("/update_kyc", views.UpdateUsersKYC)
		user.DELETE("/delete_profile", views.DeleteUser)

		/* Transaction Routes */
		transactions := user.Group("/transactions")
		{
			transactions.POST("/paybill", views.PayBillforEvent)
			transactions.POST("/pay_own_bill", views.PayOwnBill)
			transactions.POST("/send_money_to_host", views.SendMoneytoHost)
			transactions.POST("/send_money_to_user", views.SendToOtherUsers)
			transactions.GET("/get_transactions", views.GetTransactions)
		}

		/* Social Routes */
		social := user.Group("/social")
		{
			social.POST("/send_friend_request", views.SendFriendRequest)
			social.POST("/accept_friend_request", views.AcceptFriendRequest)
			social.POST("/decline_friend_request", views.DeclineFriendRequest)
			social.POST("/block", views.BlockUser)
			social.POST("/unblock", views.UnblockUser)
			social.GET("/get_friends", views.GetFriends)
			social.GET("/get_friend_requests", views.GetFriendRequests)
			social.GET("/get_blocked_friends", views.GetBlockedUsers)
		}

		/* Wallet Routes */
		wallet := user.Group("/wallet")
		{
			wallet.POST("/create", views.CreateWallet)
			wallet.POST("/fund", views.FundWallet)
			wallet.POST("/pin_change", views.ChangePin)
			wallet.GET("/balance", views.GetWalletBalance)
		}

		/* Notification Routes */
		notification := user.Group("/notification")
		{
			notification.GET("/get", views.GetNotifications)
			notification.POST("/mark_as", views.UpdateNotificationStatus)
			// notification.POST("/mark_all_as", views.UpdateAllNotificationsStatus)
		}
	}

	/* Event Routes */
	event := router.Group("/event")
	event.GET("/get", views.GetEvent)
	event.GET("/get_events", views.GetEvents)
	event.Use(TokenGuardMiddleware())
	{
		event.GET("/get_user_events", views.GetUserEvents)
		event.GET("/get_user_events_by_status", views.GetRestaurantEvents)
		event.POST("/create", views.CreateEvent)
		event.PUT("/update", views.UpdateEvent)
		event.DELETE("/delete/:id", views.DeleteEvent)
		event.GET("/cancel/:id", views.CancelEvent)

		/* Order Routes */
		order := event.Group("/order")
		{
			order.POST("create", views.CreateOrder)
			order.GET("get_orders", views.GetOrders)
			order.GET("get_order", views.GetOrder)
			order.GET("getEventOrders/:id", views.GetEventOrders)
			order.GET("getUserEventOrders/:id", views.GetUserEventOrders)
		}

		attend := event.Group("/attend")
		{
			attend.POST("/send_invites", views.SendEventInvites)
			attend.GET("/accept_invite", views.AcceptInvite)
			attend.POST("/decline_invite", views.DeclineInvite)
			// attend.GET("/getInvites", views.GetInvites)
			// attend.GET("/get_attendees", views.GetAttendees)
		}
	}

	/* Restaurant Routes */
	restaurant := router.Group("/restaurant")
	restaurant.GET("/get", views.GetRestaurant)
	restaurant.GET("/get_restaurants", views.GetRestaurants)
	restaurant.GET("/get_reviews", views.GetReviews)
	restaurant.Use(TokenGuardMiddleware())
	{
		restaurant.POST("/create", views.CreateRestaurant)
		restaurant.PUT("/update", views.UpdateRestaurant)
		restaurant.DELETE("/delete", views.DeleteRestaurant)
		restaurant.POST("/add_review", views.AddReview)
	}

	/* Product Routes */
	product := router.Group("/product")
	product.GET("/get", views.GetProduct)
	product.GET("/get_products", views.GetProducts)
	product.Use(TokenGuardMiddleware())
	{
		product.POST("/add", views.AddProduct)
		product.PUT("/update", views.UpdateProduct)
		product.DELETE("/delete", views.DeleteProduct)
	}

	/* Admin Routes */
	admin := router.Group("/admin")
	admin.Use(TokenGuardMiddleware())
	admin.POST("/create", ad.Create)

	protected := admin.Group("/protected")
	protected.Use(AdminGuardMiddleware())
	{
		protected.POST("/send_notification", ad.SendNotificationtoUsers)
	}

	return router
}
