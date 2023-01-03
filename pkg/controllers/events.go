package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var eventCollection = config.EventCollection

func CreateEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
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

	request.ID = primitive.NewObjectID()
	request.HostID = user.ID
	request.Type = hp.EventType(hp.EventType(request.Type).String())
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

	// validate open hours
	err = hp.OpenHours(request.Time).Validate()
	if err != nil {
		response := hp.SetError(err, "Invalid open hours", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	_, err = eventCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error inserting into database", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Event created", request, funcName)
	c.JSON(http.StatusOK, response)
}

func GetUserEventsByHost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
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

	// update := bson.M{
	// 	"$set": bson.M{
	// 		"title":     request.Title,
	// 		"invited":   request.Invited,
	// 		"attendees": request.Attendees,
	// 		"declined":  request.Declined,
	// 		"venue":     request.Venue,
	// 		"type":      request.Type,
	// 		"bill":      request.Bill,
	// 		"updated_at": request.UpdatedAt,
	// 	},
	// }

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
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
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
