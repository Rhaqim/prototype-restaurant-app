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

func CreateTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.Transactions
	var response = hp.MongoJsonResponse{
		Date: time.Now(),
	}

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

	insert := bson.M{
		"txn_uuid": request.Txn_uuid,
		"from_id":  user.ID,
		"to_id":    request.ToID,
		"amount":   request.Amount,
		"type":     request.Type,
		"status":   "pending",
	}

	insertResult, err := config.TransactionsCollection.InsertOne(ctx, insert)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("insertResult: ", insertResult)

	transactionResponse := hp.Transactions{
		ID:       insertResult.InsertedID.(primitive.ObjectID),
		Txn_uuid: request.Txn_uuid,
		FromID:   user.ID,
		ToID:     request.ToID,
		Amount:   request.Amount,
		Type:     request.Type,
		Status:   request.Status,
	}

	response.Type = "success"
	response.Message = "Transaction created successfully"
	response.Data = transactionResponse
	c.JSON(http.StatusOK, response)
}

func UpdateTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.Transactions
	var response = hp.MongoJsonResponse{
		Date: time.Now(),
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{
		"_id": request.ID,
	}

	update := bson.M{
		"$set": bson.M{
			"status": request.Status,
		},
	}

	updateResult, err := config.TransactionsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		config.Logs("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println("updateResult: ", updateResult)

	transactionResponse := hp.Transactions{
		ID:     request.ID,
		Status: request.Status,
	}

	response.Type = "success"
	response.Message = "Transaction updated successfully"
	response.Data = transactionResponse
	c.JSON(http.StatusOK, response)
}
