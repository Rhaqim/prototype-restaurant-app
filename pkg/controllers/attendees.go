package controllers

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var attendeeCollection = config.AttendeeCollection

// SendEventInvites sends invites to friends for an event
// Gets the event from the database
// Checks if the user is the host of the event
// Checks if user is friends with the friends
// Checks if the friends are already invited
// Updates the event with the new invites
// Sends the invites to the friends
// Sends a notification to the friends
// Sends a notification to Venue with updated event details
func SendEventInvites(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
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
		if !hp.VerifyFriends(ctx, user, friend) {
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

	// Send invites to the friends
	err = hp.SendInviteToEvent(ctx, request.EventID, request.Friends, user)
	if err != nil {
		response := hp.SetError(err, "Error sending invites", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// NOTIFICATION
	venue, err := hp.GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// send notification to invited users and venue
	msgInvite := []byte(config.Invite_ +
		user.Username +
		" invited you to " + event.Title +
		" at " + venue.Name +
		" on " + event.Date.Format("02-01-2006") +
		" at " + event.Time.Format("15:04"),
	)

	notifyInvite := nf.NewNotification(
		request.Friends,
		msgInvite,
	)
	go notifyInvite.Send()

	// for _, invited := range request.Friends {
	// 	// send notification to invited users
	// 	msg := []byte(
	// 		user.Username +
	// 			" invited you to " + event.Title +
	// 			" at " + venue.Name +
	// 			" on " + event.Date.Format("02-01-2006") +
	// 			" at " + event.Time.Format("15:04"),
	// 	)
	// 	go nf.SendNotification(invited, msg)
	// }

	msg := []byte(config.Reservation_ +
		user.Username +
		" has updated the event " + event.Title +
		" at " + venue.Name +
		" on " + event.Date.Format("02-01-2006") +
		" at " + event.Time.Format("15:04") +
		" with new invites" +
		" Capacity: " + strconv.Itoa(len(event.Invited)),
	)
	go nf.SendNotification(venue.OwnerID, msg)

	response := hp.SetSuccess("Successfully invited friends to event", nil, funcName)
	c.JSON(http.StatusOK, response)
}

// AcceptInvite accepts an invite to an event
// Checks if the user is logged in
// Checks if the user is invited to the event
// Checks if the user is already attending the event
// Check if User has enough money to match budget set
// Updates the event with the new attendee
// Updates the attendee Collection with status attending
func AcceptInvite(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
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
		c.AbortWithStatusJSON(http.StatusNotAcceptable, response)
		return
	}

	// Check if the user has already accepted the invite
	filter := bson.M{"event_id": request.EventID, "user_id": user.ID}
	attendee, err := hp.GetAttendee(ctx, filter)
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

	// Get User Wallet
	wallet, err := hp.GetWallet(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error getting wallet", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Check User has enough budget in wallet
	if wallet.Balance < request.Budget {
		response := hp.SetError(err, "User does not have enough budget in wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Update the attendee and event with go routines
	var wg sync.WaitGroup
	wg.Add(3)

	errChan := make(chan error, 3)

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
			errChan <- err
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
			errChan <- err
			return
		}
	}()

	// Lock budget from the wallet
	go func() {
		defer wg.Done()

		venue, err := hp.GetRestaurant(ctx, bson.M{"_id": event.Venue})
		if err != nil {
			errChan <- err
			return
		}

		err = hp.LockBudget(ctx, wallet, request.Budget, venue.OwnerID)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		response := hp.SetError(err, "Error updating attendee or event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Send notification to event owner
	// Get the event owner
	eventOwner, err := hp.GetUser(ctx, bson.M{"_id": event.HostID})
	if err != nil {
		response := hp.SetError(err, "Error getting event owner", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Send notification to event owner
	msg := []byte(config.Notification_ + user.Username + " has accepted your invite to " + event.Title)
	go nf.SendNotification(eventOwner.ID, msg)

	response := hp.SetSuccess("Successfully accepted invite", nil, funcName)
	c.JSON(http.StatusOK, response)
}

// DeclineInvite declines an invite to an event
// Checks if the user is logged in
// Checks if the user is invited to the event
// Checks if the user has already declined the invite
// Updates the attendee and event with go routines
func DeclineInvite(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
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

	// Send notification to event owner
	// Get the event owner
	eventOwner, err := hp.GetUser(ctx, bson.M{"_id": event.HostID})
	if err != nil {
		response := hp.SetError(err, "Error getting event owner", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Send notification to event owner
	msg := []byte(config.Notification_ + user.Username + " has declined your invite to " + event.Title)
	go nf.SendNotification(eventOwner.ID, msg)

	response := hp.SetSuccess("Successfully declined invite", nil, funcName)
	c.JSON(http.StatusOK, response)
}
