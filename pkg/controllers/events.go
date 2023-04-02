package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	eventCollection = config.EventCollection

	CreateEvent         = AbstractConnection(createEvent)
	GetEvent            = AbstractConnection(getEvent)
	GetEvents           = AbstractConnection(getEvents)
	GetUserEvents       = AbstractConnection(getUserEvents)
	GetRestaurantEvents = AbstractConnection(getRestaurantEvents)
	UpdateEvent         = AbstractConnection(updateEvent)
	DeleteEvent         = AbstractConnection(deleteEvent)
	CancelEvent         = AbstractConnection(cancelEvent)
)

// CreateEvent creates an event
// It accests the Title of the event, the Restaurant for the event,
// a group of invited friends and also the event time
// it stores the event in the database with the User as the host and a status of upcoming
// It sends a notification to the invited users and also the restaurant for the event.
func createEvent(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	var request hp.Event

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

	// VERIFICATION

	// Check if DateTime is after time.Now()
	// okDate := hp.VeryifyDateTimeAfterNow(request.Date, request.Time)
	// if !okDate {
	// 	response := hp.SetError(err, "Date and Time must be after current Date and Time", funcName)
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, response)
	// 	return
	// }

	// Check if User has a Wallet already
	exists, err := hp.CheckifWalletExists(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error checking if wallet exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	if !exists {
		response := hp.SetError(err, "User does not have a wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if Wallet balance is sufficient for Budget
	wallet, err := hp.GetWallet(ctx, bson.M{"_id": user.Wallet})
	if err != nil {
		response := hp.SetError(err, "Error getting wallet balance", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	if wallet.Balance < request.Budget {
		response := hp.SetError(err, "Insufficient balance to create Event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.ID = primitive.NewObjectID()
	request.HostID = user.ID
	request.EventType = hp.EventType(hp.EventType(request.EventType).String())
	request.EventStatus = hp.Upcoming
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()
	// Add Host to Attendees
	request.Attendees = append(request.Attendees, user.ID)

	_, err = eventCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error inserting into database", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get Venue
	venue, err := hp.GetRestaurant(ctx, bson.M{"_id": request.RestaurantID})
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// LOCK BUDGET
	err = hp.LockBudget(ctx, wallet, request.Budget, venue.OwnerID)
	if err != nil {
		response := hp.SetError(err, "Error locking budget", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Unlock Budget and return to wallet after 24 hours
	go func() {
		time.Sleep(24 * time.Hour)

		c := context.Background()

		err = hp.BudgetoWallet(c, venue.OwnerID, user)
		if err != nil {
			hp.SetDebug("Error returning budget to wallet: "+err.Error(), funcName)
		}

		msg := "Your budget of " + fmt.Sprintf("%.2f", request.Budget) + " has been returned to your wallet."

		err = nf.AlertUser(config.BudgetReturned, msg, user.ID)
		if err != nil {
			hp.SetDebug("Error sending notification: "+err.Error(), funcName)
		}
	}()

	// send invite to invited users
	err = hp.SendInviteToEvent(ctx, request.ID, request.Invited, user)
	if err != nil {
		response := hp.SetError(err, "Error sending invite to event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// NOTIFICATION
	// send notification to invited users
	msgInvited := []byte(config.Invite_ +
		user.Username +
		" invited you to " + request.Title +
		" at " + venue.Name +
		" on " + request.Date.Format("02-01-2006") +
		" at " + request.Time.Format("15:04"),
	)

	notifyInvited := nf.NewNotification(
		request.Invited,
		msgInvited,
	)
	notifyInvited.Send()

	// send notification to venue owner
	msgVenue := []byte(config.Reservation_ +
		user.Username +
		" has created an event at " + venue.Name +
		" on " + request.Date.Format("02-01-2006") +
		" at " + request.Time.Format("15:04") +
		" capacity: " + fmt.Sprint(len(request.Invited)+1) +
		" Special request: " + request.SpecialRequest,
	)
	venueList := []primitive.ObjectID{venue.OwnerID}
	notifyVenue := nf.NewNotification(
		venueList,
		msgVenue,
	)
	notifyVenue.Send()

	// Send notification to host 5 minutes before event and change event status to ongoing
	go func() {
		// Get the time difference in minutes
		minutes := request.GetTimeDifference()
		hp.SetInfo("Minutes to event: "+fmt.Sprint(minutes), funcName)

		// Set the duration to sleep
		duration := time.Duration(minutes-5) * time.Minute
		time.Sleep(duration)

		ctx := context.Background()

		// Change event status to ongoing
		filter := bson.M{"_id": request.ID}
		update := bson.M{"$set": bson.M{"event_status": hp.Ongoing}}

		event, err := hp.UpdateEvent(ctx, filter, update)
		if err != nil {
			hp.SetError(err, "Error updating event status", funcName)
		}

		// Send Notification to Attendees
		msgAttendees := []byte(
			"Your event " + request.Title +
				" at " + venue.Name +
				" on " + request.Date.Format("02-01-2006") +
				" at " + request.Time.Format("15:04") +
				" is about to start",
		)
		attendees := event.Attendees

		NotifyAttendees := nf.NewNotification(
			attendees,
			msgAttendees,
		)

		NotifyAttendees.Send()
	}()

	response := hp.SetSuccess("Event created", request, funcName)
	c.JSON(http.StatusOK, response)
}

// GetEvent fetches an event by either the ID or the title
func getEvent(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	// Get event id from url
	id := c.Query("id")

	// Get event title from url
	title := c.Query("title")

	var filter bson.M

	switch {
	case id != "":
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			response := hp.SetError(err, "Invalid event id", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"_id": id}
	case title != "":
		filter = bson.M{"title": title}
	default:
		response := hp.SetError(nil, "No query parameters provided", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	event, err := hp.GetEvent(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event found", event, funcName)
	c.JSON(http.StatusOK, response)
}

func getUserEvents(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	// Get user from context
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "Error getting user from context", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get user events, where user is host and/or attendee
	filter := bson.M{
		"$or": []bson.M{
			{"host_id": user.ID},
			{"attendees": bson.M{
				"$in": []primitive.ObjectID{user.ID},
			}},
		},
	}

	events, err := hp.GetEvents(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting user events", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("User events found", events, funcName)
	c.JSON(http.StatusOK, response)
}

func getRestaurantEvents(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	// Get restaurants that belong to the user
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "Error getting user from context", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	id := c.Query("id")

	// Get venue id from url
	// get restaurant by id
	venue, err := hp.GetRestaurants(ctx, bson.M{"owner_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var venue_id string

	for _, v := range venue {
		if v.ID.Hex() == id {
			venue_id = v.ID.Hex()
			break
		}
	}

	// Get venue events
	filter := bson.M{"venue": venue_id}
	events, err := hp.GetEvents(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting venue events", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Venue events found", events, funcName)
	c.JSON(http.StatusOK, response)
}

// GetEvents fetches a list of events by either the type, venue, host,
// date or attended
func getEvents(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	// Get event type from url
	eventType := c.Query("event_type")

	// Get event Venue from url
	venue := c.Query("venue")

	// Get event by host id from url
	hostID := c.Query("host_id")

	// Get event by date from url
	date := c.Query("date")

	// Get events attended by user from url
	attended := c.Query("attended")

	var filter bson.M

	switch {
	case eventType != "":
		filter = bson.M{"event_type": eventType}
	case venue != "":
		filter = bson.M{"venue": venue}
	case hostID != "":
		hostID, err := primitive.ObjectIDFromHex(hostID)
		if err != nil {
			response := hp.SetError(err, "Invalid host id", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"host_id": hostID}
	case date != "":
		filter = bson.M{"created_at": date}
	case attended != "":
		user, err := hp.GetUserFromToken(c)
		if err != nil {
			response := hp.SetError(err, "User not logged in", funcName)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}
		filter = bson.M{"attendees": bson.M{"$in": []string{user.ID.Hex()}}}
	default:
		response := hp.SetError(nil, "No query parameters provided", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	events, err := hp.GetEvents(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting events", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	if len(events) == 0 {
		response := hp.SetError(nil, "No events found", funcName)
		c.AbortWithStatusJSON(http.StatusNotFound, response)
		return
	}

	response := hp.SetSuccess(" events found", events, funcName)
	c.JSON(http.StatusOK, response)
}

// UpdateEvent updates the event with the request sent.
func updateEvent(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	var request hp.Event

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

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	update := bson.M{
		"$set": request,
	}

	updateResult, err := eventCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		response := hp.SetError(err, "Error updating hosted event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event updated", updateResult, funcName)
	c.JSON(http.StatusOK, response)
}

// DeleteEvent deletes the event from the database
func deleteEvent(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	deleteResult, err := eventCollection.DeleteOne(ctx, filter)

	if err != nil {
		response := hp.SetError(err, "Error deleting hosted event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event deleted", deleteResult, funcName)
	c.JSON(http.StatusOK, response)
}

// CancelEvent cancels an event
func cancelEvent(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	// Get Event
	event, err := hp.GetEvent(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// NOTE: Might seperate the finished and ongoing checks
	// Check that event is not currently ongoing or finished
	if event.EventStatus == hp.Ongoing || event.EventStatus == hp.Finished {
		response := hp.SetError(err, "Event is currently ongoing or finished", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get Venue
	filter = bson.M{"_id": event.RestaurantID}
	venue, err := hp.GetRestaurant(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Return the budget to the wallet
	err = hp.BudgetoWallet(ctx, venue.OwnerID, user)
	if err != nil {
		response := hp.SetError(err, "Error returning budget to wallet", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	update := bson.M{
		"$set": bson.M{"event_status": hp.Cancelled},
	}

	_, err = eventCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error cancelling event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event cancelled", nil, funcName)
	c.JSON(http.StatusOK, response)
}
