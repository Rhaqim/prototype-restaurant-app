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
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Send Friend Request
	// Send a friend request to another user.
	friendRequest, err := hp.SendFriendRequest(ctx, user, request.FriendID)
	if err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Failed to send friend request", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request sent", friendRequest, funcName))
}

// Accept Friend Request
// Accept a friend request from another user.
func AcceptFriendRequest(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var funcName = ut.GetFunctionName()

	var request = hp.FriendshipAcceptRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Invalid JSON", funcName))
		return
	}

	// Get the user ID from the token.
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Check that friendID is User ID
	if request.FriendID != user.ID {
		c.JSON(http.StatusBadRequest, hp.SetError(nil, "Invalid Friend request", funcName))
		return
	}

	from := hp.GetUserByID(ctx, request.UserID)

	// Accept Friend Request
	// Accept a friend request from another user.
	err = hp.AcceptFriendRequest(ctx, from, user, request.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, hp.SetError(err, "Failed to accept friend request", funcName))
		return
	}

	c.JSON(http.StatusOK, hp.SetSuccess("Friend request accepted", nil, funcName))
}
