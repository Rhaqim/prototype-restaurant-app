package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var hostCollection = config.HostCollection

func CreateHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.HostingCreate
	response := hp.MongoJsonResponse{}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Type = "error"
		response.Message = err.Error() + " Request body is not valid"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response.Type = "error"
		response.Message = "User not logged in"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	//  Ensure that hostedIDs are not empty
	if len(request.HostedIDs) < 1 {
		response.Type = "error"
		response.Message = "HostedIDs cannot be empty"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	insert := bson.M{
		"host_id":    user.ID,
		"title":      request.Title,
		"hosted_ids": request.HostedIDs,
		"venue":      request.Venue,
		"bill":       request.Bill,
	}

	insertResult, err := hostCollection.InsertOne(ctx, insert)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("insertResult: ", insertResult)

	hostingResponse := hp.Hosting{
		ID:        insertResult.InsertedID.(primitive.ObjectID),
		Title:     request.Title,
		HostID:    user.ID,
		HostedIDs: request.HostedIDs,
		Venue:     request.Venue,
		Bill:      request.Bill,
	}

	response.Type = "success"
	response.Message = "Event created"
	response.Data = hostingResponse
	c.JSON(http.StatusOK, response)
}

func GetUserHostedEventsByHost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	response := hp.MongoJsonResponse{}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response.Type = "error"
		response.Message = "User not logged in"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	filter := bson.M{"host_id": user.ID}
	cursor, err := hostCollection.Find(ctx, filter)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hosting []hp.Hosting
	if err = cursor.All(ctx, &hosting); err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response.Type = "success"
	response.Message = "Hosted Events by:" + user.Username
	response.Data = hosting
	c.JSON(http.StatusOK, response)
}

func UpdateHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.Hosting
	response := hp.MongoJsonResponse{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response.Type = "error"
		response.Message = "User not logged in"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	update := bson.M{
		"$set": bson.M{
			"title":      request.Title,
			"hosted_ids": request.HostedIDs,
			"venue":      request.Venue,
			"bill":       request.Bill,
		},
	}

	updateResult, err := hostCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("insertResult: ", updateResult)
	response.Type = "success"
	response.Message = "Event updated"
	c.JSON(http.StatusOK, response)
}

func DeleteHostedEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	response := hp.MongoJsonResponse{}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response.Type = "error"
		response.Message = "User not logged in"
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	deleteResult, err := hostCollection.DeleteOne(ctx, filter)

	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Println("insertResult: ", deleteResult)
	response.Type = "success"
	response.Message = "Event deleted"
	c.JSON(http.StatusOK, response)
}
