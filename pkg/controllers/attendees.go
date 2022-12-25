package controllers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var attendeeCollection = config.AttendeeCollection

func SendEventInvites(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.InviteFriendsToEventRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get the event
	var event hp.Event
	err = eventCollection.FindOne(ctx, bson.M{"_id": request.EventID}).Decode(&event)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Check if the user is the host of the event
	if event.HostID != user.ID {
		response := hp.SetError(err, "User is not the host of the event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Verify friendship
	for _, friend := range request.Friends {
		if !hp.VerifyFriends(user, friend) {
			response := hp.SetError(err, "Friendship not verified: "+friend.Hex(), funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Check if the Friends are already invited
	for _, friend := range request.Friends {
		for _, attendee := range event.Invited {
			if friend == attendee {
				response := hp.SetError(err, "Friend is already invited", funcName)
				c.AbortWithStatusJSON(http.StatusBadRequest, response)
				return
			}
		}
	}

	// Add the friends to the event
	event.Invited = append(event.Invited, request.Friends...)

	// Update the event
	_, err = eventCollection.UpdateOne(ctx, bson.M{"_id": request.EventID}, bson.M{"$set": event})
	if err != nil {
		response := hp.SetError(err, "Error updating event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Add to the attendees collection using go routines
	var wg sync.WaitGroup
	wg.Add(len(request.Friends))

	for _, friend := range request.Friends {
		go func(friend primitive.ObjectID) {
			defer wg.Done()

			var attendee = hp.EventAttendee{
				EventID:   request.EventID,
				Status:    hp.Invited,
				InvitedBy: user.ID,
				InvitedAt: primitive.NewDateTimeFromTime(time.Now()),
			}
			attendee.UserID = friend

			_, err = attendeeCollection.InsertOne(ctx, attendee)
			if err != nil {
				response := hp.SetError(err, "Error inserting attendee", funcName)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response)
				return
			}
		}(friend)
	}

	wg.Wait()

	response := hp.SetSuccess("Successfully invited friends to event", nil, funcName)
	c.JSON(http.StatusOK, response)
}

func AcceptInvite(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.AcceptInviteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get the event
	var event hp.Event
	err = eventCollection.FindOne(ctx, bson.M{"_id": request.EventID}).Decode(&event)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Check if the user is invited to the event
	var invited = false
	for _, attendee := range event.Invited {
		if attendee == user.ID {
			invited = true
			break
		}
	}

	if !invited {
		response := hp.SetError(err, "User is not invited to the event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if the user has already accepted the invite
	var attendee hp.EventAttendee
	err = attendeeCollection.FindOne(ctx, bson.M{"event_id": request.EventID, "user_id": user.ID}).Decode(&attendee)
	if err != nil {
		response := hp.SetError(err, "Error getting attendee", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	if attendee.Status == hp.Attending {
		response := hp.SetError(err, "User has already accepted the invite", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Update the attendee and event with go routines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		// Update the attendee
		filter := bson.M{"event_id": request.EventID, "user_id": user.ID}
		update := bson.M{
			"$set": bson.M{
				"status":      hp.Attending,
				"accepted_at": primitive.NewDateTimeFromTime(time.Now()),
			},
			"$inc": bson.M{
				"budget": +request.Budget,
			}}

		_, err = attendeeCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			response := hp.SetError(err, "Error updating attendee", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()

	go func() {
		defer wg.Done()

		// Update the event
		filter := bson.M{"_id": request.EventID}
		update := bson.M{
			"$pull": bson.M{"invited": user.ID},
			"$push": bson.M{"attendees": user.ID},
			"$inc": bson.M{
				"attendee_count": 1,
				"budget":         +request.Budget,
			}}

		_, err = eventCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			response := hp.SetError(err, "Error updating event", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()

	wg.Wait()

	response := hp.SetSuccess("Successfully accepted invite", nil, funcName)
	c.JSON(http.StatusOK, response)
}

func DeclineInvite(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.DeclineInviteRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get the event
	var event hp.Event
	err = eventCollection.FindOne(ctx, bson.M{"_id": request.EventID}).Decode(&event)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Check if the user is invited to the event
	var invited = false
	for _, attendee := range event.Invited {
		if attendee == user.ID {
			invited = true
			break
		}
	}

	if !invited {
		response := hp.SetError(err, "User is not invited to the event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if the user has already declined the invite
	var attendee hp.EventAttendee
	err = attendeeCollection.FindOne(ctx, bson.M{"event_id": request.EventID, "user_id": user.ID}).Decode(&attendee)
	if err != nil {
		response := hp.SetError(err, "Error getting attendee", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	if attendee.Status == hp.NotAttending {
		response := hp.SetError(err, "User has already declined the invite", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Update the attendee and event with go routines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {

		defer wg.Done()

		// Update the attendee
		filter := bson.M{"event_id": request.EventID, "user_id": user.ID}
		update := bson.M{
			"$set": bson.M{
				"status":     hp.NotAttending,
				"updated_at": primitive.NewDateTimeFromTime(time.Now()),
			}}

		_, err = attendeeCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			response := hp.SetError(err, "Error updating attendee", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()

	go func() {

		defer wg.Done()

		// Update the event
		filter := bson.M{"_id": request.EventID}
		update := bson.M{
			"$pull": bson.M{"invited": user.ID},
			"$push": bson.M{"declined": user.ID},
			"$inc": bson.M{
				"declined_count": 1,
			}}

		_, err = eventCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			response := hp.SetError(err, "Error updating event", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()

	wg.Wait()

	response := hp.SetSuccess("Successfully declined invite", nil, funcName)
	c.JSON(http.StatusOK, response)
}
