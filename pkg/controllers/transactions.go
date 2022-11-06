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

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding JSON")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	check, ok := c.Get("user") //check if user is logged in
	if !ok {
		response := hp.SetError(nil, "User not logged in")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	request.Status = hp.TxnPending

	insert := bson.M{
		"txn_uuid": request.Txn_uuid,
		"from_id":  user.ID,
		"to_id":    request.ToID,
		"amount":   request.Amount,
		"type":     request.Type,
		"status":   request.Status,
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

	response := hp.SetSuccess("Transaction created successfully", transactionResponse)
	c.JSON(http.StatusOK, response)
}

func UpdateTransactionStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.Transactions

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Check Transaction Type Validity
	ok := hp.TxnStatusIsValid(request.Status)

	if !ok {
		response := hp.SetError(nil, "Invalid Transaction Type")
		c.JSON(http.StatusBadRequest, response)
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
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		response := hp.SetError(err, "Error updating transaction status")
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("updateResult: ", updateResult)

	transactionResponse := hp.Transactions{
		ID:     request.ID,
		Status: request.Status,
	}

	response := hp.SetSuccess("Transaction updated successfully", transactionResponse)
	c.JSON(http.StatusOK, response)
}
