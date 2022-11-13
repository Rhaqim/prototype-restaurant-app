package controllers

import (
	"context"
	"log"
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

var hostCollection = config.HostCollection

func CreateHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.HostingCreate

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	//  Ensure that hostedIDs are not empty
	if len(request.HostedIDs) < 1 {
		response := hp.SetError(nil, "HostedIDs cannot be empty", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	insert := bson.M{
		"host_id":    user.ID,
		"title":      request.Title,
		"hosted_ids": request.HostedIDs,
		"venue":      request.Venue,
		"type":       request.Type,
		"bill":       request.Bill,
	}

	insertResult, err := hostCollection.InsertOne(ctx, insert)
	if err != nil {
		response := hp.SetError(err, "Error inserting into database", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("insertResult: ", insertResult)

	hostingResponse := hp.Hosting{
		ID:        insertResult.InsertedID.(primitive.ObjectID),
		Title:     request.Title,
		HostID:    user.ID,
		HostedIDs: request.HostedIDs,
		Venue:     request.Venue,
		Type:      request.Type,
		Bill:      request.Bill,
	}

	response := hp.SetSuccess("Event created", hostingResponse, funcName)
	c.JSON(http.StatusOK, response)
}

func GetUserHostedEventsByHost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"host_id": user.ID}
	cursor, err := hostCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error finding hosted events", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var hosting []hp.Hosting
	if err = cursor.All(ctx, &hosting); err != nil {
		response := hp.SetError(err, "Error decoding hosted events", funcName)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Hosted events found", hosting, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Hosting

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	update := bson.M{
		"$set": bson.M{
			"title":      request.Title,
			"hosted_ids": request.HostedIDs,
			"venue":      request.Venue,
			"type":       request.Type,
			"bill":       request.Bill,
		},
	}

	updateResult, err := hostCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		response := hp.SetError(err, "Error updating hosted event", funcName)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Hosted event updated", updateResult, funcName)
	c.JSON(http.StatusOK, response)
}

func DeleteHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	deleteResult, err := hostCollection.DeleteOne(ctx, filter)

	if err != nil {
		response := hp.SetError(err, "Error deleting hosted event", funcName)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Hosted event deleted", deleteResult, funcName)
	c.JSON(http.StatusOK, response)
}
