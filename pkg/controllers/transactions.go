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

func CreateTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Transactions

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, ", Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Transaction Verification
	friends := hp.VerifyFriends(user, request.ToID)

	if !friends {
		response := hp.SetError(nil, "You are not friends with this user", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	sufficientBalance := hp.VerifyWalletSufficientBalance(user, request.Amount)

	if !sufficientBalance {
		response := hp.SetError(nil, "Insufficient balance", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	request.Status = hp.TxnPending
	request.Txn_uuid = ut.GenerateUUID()

	insertResult, err := config.TransactionsCollection.InsertOne(ctx, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionResponse := hp.Transactions{
		ID:       insertResult.InsertedID.(primitive.ObjectID),
		Txn_uuid: request.Txn_uuid,
		FromID:   user.ID,
		ToID:     request.ToID,
		Amount:   request.Amount,
		Type:     request.Type,
		Status:   request.Status,
	}

	response := hp.SetSuccess("Transaction created successfully", transactionResponse, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateTransactionStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.TransactionStatus

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Check Transaction Type Validity
	ok := hp.TxnStatusIsValid(request.Status)

	if !ok {
		response := hp.SetError(nil, "Invalid Transaction Type", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{
		"_id":      request.ID,
		"fromId":   user.ID,
		"txn_uuid": request.Txn_uuid,
	}

	update := bson.M{
		"$set": bson.M{
			"status": request.Status,
		},
	}

	updateResult, err := config.TransactionsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error updating transaction status", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}
	log.Println("updateResult: ", updateResult)

	// Update Wallet Balance
	var transaction hp.Transactions
	err = config.TransactionsCollection.FindOne(ctx, filter).Decode(&transaction)
	if err != nil {
		response := hp.SetError(err, "Error fetching transaction", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if transaction.Status != hp.TxnSuccess {
		response := hp.SetError(nil, "Transaction not successful", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update Wallet Balance
	err = hp.UpdateWalletBalance(ctx, transaction)
	if err != nil {
		response := hp.SetError(err, "Error updating wallet balance", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Transaction status updated successfully", updateResult, funcName)
	c.JSON(http.StatusOK, response)
}
