package controllers

import (
	"context"
	"net/http"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetNotifications returns all notifications for a user
func GetNotifications(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	notifs, err := nf.GetNotificationsByUser(ctx, user)
	if err != nil {
		response := hp.SetError(err, "Error getting notifications", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// convert the []byte to string
	notifications := make([]nf.NotificationResponse, len(notifs))
	for i, notif := range notifs {
		notifications[i] = nf.NotificationResponse{
			ID:           notif.ID,
			Notification: string(notif.Notification),
			Seen:         notif.Seen,
			Time:         notif.Time,
		}
	}

	c.JSON(http.StatusOK, notifications)
}

// UpdateNotificationStatus updates the status of a notification
// or group of notifications to seen or unseen
func UpdateNotificationStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	_, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	var notification = struct {
		NotificationIDs []primitive.ObjectID `json:"notification_ids"`
	}{
		NotificationIDs: []primitive.ObjectID{},
	}
	err = c.ShouldBindJSON(&notification)
	if err != nil {
		response := hp.SetError(err, "Error binding JSON", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	for _, id := range notification.NotificationIDs {
		err = nf.UpdateNotificationStatus(ctx, id)
		if err != nil {
			response := hp.SetError(err, "Error updating notification status", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	response := hp.SetSuccess("Notification status updated", nil, funcName)
	c.JSON(http.StatusOK, response)
}

// SendNotificationtoUsers sends a notification to a list of users with role user
func SendNotificationtoUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check if user is Admin
	if user.Role != hp.Admin {
		response := hp.SetError(err, "User is not an admin", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	var request = struct {
		Message string `json:"message"`
		Inteded string `json:"intended"`
	}{
		Message: "",
		Inteded: "",
	}
	err = c.ShouldBindJSON(&request)
	if err != nil {
		response := hp.SetError(err, "Error binding JSON", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// convert inteded to role
	var role hp.Roles
	switch request.Inteded {
	case "user":
		role = hp.User
	case "admin":
		role = hp.Admin
	default:
		response := hp.SetError(err, "Invalid intended role", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// get all users with role user from cache
	filter := bson.M{"role": role}
	users, err := hp.GetUserIDsFromCache(ctx, filter, config.UserRole)
	if err != nil {
		response := hp.SetError(err, "Error getting users from cache", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// convert []string to []primitive.ObjectID
	var userIDs []primitive.ObjectID
	for _, id := range users {
		userID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			response := hp.SetError(err, "Error converting user id to ObjectID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		userIDs = append(userIDs, userID)
	}

	//convert message to []byte
	msg := []byte(request.Message)

	notifyUsers := nf.NewNotification(userIDs, msg)
	notifyUsers.ToIntendedUsers(ctx, role)

	response := hp.SetSuccess("Notification sent", nil, funcName)
	c.JSON(http.StatusOK, response)
}
