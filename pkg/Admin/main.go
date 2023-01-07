package admin

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

func Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()

	var admin AdminModel

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := admin.CreateAdmin(ctx); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Admin created"})
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

	hp.SetDebug("userIDs"+ut.ToJSON(userIDs), funcName)

	//convert message to []byte
	msg := []byte(request.Message)

	notifyUsers := nf.NewNotification(userIDs, msg)
	notifyUsers.ToIntendedUsers(ctx, role)

	response := hp.SetSuccess("Notification sent", nil, funcName)
	c.JSON(http.StatusOK, response)
}
