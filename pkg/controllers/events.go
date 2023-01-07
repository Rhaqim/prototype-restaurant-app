package controllers

import (
	"context"
	"fmt"
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

var eventCollection = config.EventCollection

func CreateEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

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
	request.Type = hp.EventType(hp.EventType(request.Type).String())
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

	_, err = eventCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error inserting into database", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// send invite to invited users
	err = hp.SendInviteToEvent(ctx, request.ID, request.Invited, user)
	if err != nil {
		response := hp.SetError(err, "Error sending invite to event", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// NOTIFICATION
	venue, err := hp.GetRestaurant(ctx, bson.M{"_id": request.Venue})
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

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
	notifyInvited.Create(ctx)

	// send notification to venue owner
	msgVenue := []byte(config.Reservation_ +
		user.Username +
		" has created an event at " + venue.Name +
		" on " + request.Date.Format("02-01-2006") +
		" at " + request.Time.Format("15:04") +
		" capacity: " + fmt.Sprint(len(request.Invited)),
	)
	go nf.SendNotification(venue.OwnerID, msgVenue)

	response := hp.SetSuccess("Event created", request, funcName)
	c.JSON(http.StatusOK, response)
}

func GetEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

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

func GetEvents(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	// Get event type from url
	eventType := c.Query("type")

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
		filter = bson.M{"type": eventType}
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

	response := hp.SetSuccess(" events found", events, funcName)
	c.JSON(http.StatusOK, response)
}

func GetUserEventsByHost(c *gin.Context) {
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

	filter := bson.M{"host_id": user.ID}
	cursor, err := eventCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error finding hosted events", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var hosting []hp.Event
	if err = cursor.All(ctx, &hosting); err != nil {
		response := hp.SetError(err, "Error decoding hosted events", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" events found", hosting, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

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

func DeleteEvent(c *gin.Context) {
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
