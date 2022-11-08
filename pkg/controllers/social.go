package controllers

import (
	"context"
	"net/http"
	"time"

	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
)

func SendFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	check, ok := c.Get("user")

	if !ok {
		c.JSON(http.StatusBadRequest, hp.SetError(nil, "Invalid token", funcName))
		return
	}

	// Get the user ID from the token.
	user := check.(hp.UserResponse)

	request.UserID = user

	// Send Friend Request
	// Send a friend request to another user.
	err := hp.SendFriendRequest(ctx, request.UserID, request.FriendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Failed to send friend request", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request sent", nil, funcName))
}

// Accept Friend Request
// Accept a friend request from another user.
func AcceptFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	check, ok := c.Get("user")

	if !ok {
		c.JSON(http.StatusBadRequest, hp.SetError(nil, "Invalid token", funcName))
		return
	}

	// Get the user ID from the token.
	user := check.(hp.UserResponse)

	// Check that friendID is User ID
	if request.FriendID != user.ID {
		c.JSON(http.StatusBadRequest, hp.SetError(nil, "Invalid Friend reqeust", funcName))
		return
	}

	// Accept Friend Request
	// Accept a friend request from another user.
	err := hp.AcceptFriendRequest(ctx, request.UserID, request.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Failed to accept friend request", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request accepted", nil, funcName))
}
