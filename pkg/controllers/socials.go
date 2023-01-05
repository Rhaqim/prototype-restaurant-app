package controllers

import (
	"context"
	"net/http"
	"time"

	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
)

func SendFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check prior request not sent
	if hp.CheckIfRequestExists(user, request.FriendID) {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(nil, "Friend request already sent", funcName))
		return
	}

	// Send Friend Request
	// Send a friend request to another user.
	friendRequest, err := hp.SendFriendRequest(ctx, user, request.FriendID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to send friend request", funcName))
		return
	}

	// Notify the user that a friend request has been sent.
	var msg = []byte("You have a new friend request! from: " + user.Username)
	go nf.SendNotification(request.FriendID, msg)

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request sent", friendRequest, funcName))
}

// Accept Friend Request
// Accept a friend request from another user.
func AcceptFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipAcceptRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check that friendID is User ID
	if request.FriendID != user.ID {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(nil, "Invalid Friend request", funcName))
		return
	}

	from := hp.GetUserByID(ctx, request.UserID)

	// Accept Friend Request
	// Accept a friend request from another user.
	err = hp.AcceptFriendRequest(ctx, from, user, request.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to accept friend request", funcName))
		return
	}

	// Send Notification
	msg := []byte("Your friend request to: " + from.Username + " has been accepted!")
	go nf.SendNotification(from.ID, msg)

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request accepted", nil, funcName))
}

// Decline Friend Request
// Decline a friend request from another user.
func DeclineFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipAcceptRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check that friendID is User ID
	if request.FriendID != user.ID {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(nil, "Invalid Friend request", funcName))
		return
	}

	from := hp.GetUserByID(ctx, request.UserID)

	// Decline Friend Request
	// Decline a friend request from another user.
	err = hp.DeclineFriendRequest(ctx, from, user, request.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to decline friend request", funcName))
		return
	}

	// Send Notification
	msg := []byte("Your friend request to: " + from.Username + " has been declined!")
	go nf.SendNotification(from.ID, msg)

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request declined", nil, funcName))
}

// Get Friend Requests
// Get all friend requests for a user.
func GetFriendRequests(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Friend Requests
	// Get all friend requests for a user.
	friendRequests, err := hp.GetSocial(ctx, user.ID, hp.FriendshipStatusPending)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to get friend requests", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friend requests", friendRequests, funcName))
}

// Get Friends
// Get all friends for a user.
func GetFriends(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Friends
	// Get all friends for a user.
	friends, err := hp.GetSocial(ctx, user.ID, hp.FriendshipStatusAccepted)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to get friends", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friends", friends, funcName))
}

// Block User
// Block a user.
func BlockUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Block User
	// Block a user.
	err = hp.BlockUser(ctx, user, request.FriendID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to block user", funcName))
		return
	}

	blocked := hp.GetUserByID(ctx, request.FriendID)

	// Send Notification
	msg := []byte("You have blocked: " + blocked.Username)
	go nf.SendNotification(user.ID, msg)

	c.JSON(http.StatusOK, hp.SetSuccess("User blocked", nil, funcName))
}

// Unblock User
// Unblock a user.
func UnblockUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Unblock User
	// Unblock a user.
	err = hp.UnblockUser(ctx, user, request.FriendID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to unblock user", funcName))
		return
	}

	blocked := hp.GetUserByID(ctx, request.FriendID)

	// Send Notification
	msg := []byte("You have unblocked: " + blocked.Username)
	go nf.SendNotification(user.ID, msg)

	c.JSON(http.StatusOK, hp.SetSuccess("User unblocked", nil, funcName))
}

func GetBlockedUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Blocked Users
	// Get all blocked users for a user.
	blockedUsers, err := hp.GetSocial(ctx, user.ID, hp.FriendshipStatusBlocked)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Failed to get blocked users", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Blocked users", blockedUsers, funcName))
}
