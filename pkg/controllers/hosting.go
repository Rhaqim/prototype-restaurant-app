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
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	insert := bson.M{
		"host_id":    request.HostID,
		"hosted_ids": request.HostedIDs,
		"bill":       request.Bill,
	}

	insertResult, err := hostCollection.InsertOne(ctx, insert)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("insertResult: ", insertResult)
	response.Type = "success"
	response.Message = "User created"
	c.JSON(http.StatusOK, response)
}

func UpdateHosting(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.HostingUpdate
	response := hp.MongoJsonResponse{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"hosted_ids": request.HostedIDs,
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
	response.Message = "User created"
	c.JSON(http.StatusOK, response)
}
