package router

import (
	// "github.com/dutchapp/backend/internal/adapters/auth"
	// "github.com/dutchapp/backend/internal/adapters/email"
	// "github.com/dutchapp/backend/internal/adapters/notification"
	// "github.com/dutchapp/backend/pkg/authentication"
	// "github.com/dutchapp/backend/pkg/email"
	// "github.com/dutchapp/backend/pkg/notification"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// authService := auth.NewService()
	// notificationService := notification.NewService()
	// emailService := email.NewService()

	// authAPI := authentication.NewAPI(authService)
	// notificationAPI := notificationpkg.NewAPI(notificationService)
	// emailAPI := emailpkg.NewAPI(emailService)

	// v1 := r.Group("/api/v1")
	// {
	// 	v1.POST("/login", authAPI.Login)
	// 	v1.GET("/notifications", notificationAPI.ListNotifications)
	// 	v1.POST("/email", emailAPI.SendEmail)
	// }

	return r
}
