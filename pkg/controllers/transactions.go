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

	friends := hp.VerifyFriends(user.ID, request.ToID)

	if !friends {
		response := hp.SetError(nil, "You are not friends with this user")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	request.Status = hp.TxnPending
	request.Txn_uuid = ut.GenerateUUID()

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

	response := hp.SetSuccess("Transaction created successfully", transactionResponse)
	c.JSON(http.StatusOK, response)
}

func UpdateTransactionStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var request hp.TransactionStatus

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	check, auth := c.Get("user") //check if user is logged in
	if !auth {
		response := hp.SetError(nil, "User not logged in")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user := check.(hp.UserResponse)

	// Check Transaction Type Validity
	ok := hp.TxnStatusIsValid(request.Status)

	if !ok {
		response := hp.SetError(nil, "Invalid Transaction Type")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{
		"_id":     request.ID,
		"from_id": user.ID,
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

// Deducts the amount from the user's wallet and adds it to the merchant's wallet
// func DeductFromUser(sender primitive.ObjectID, receiver primitive.ObjectID) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	defer database.ConnectMongoDB().Disconnect(context.TODO())

// 	// Get Transaction Details
// 	filter := bson.M{
// 		"_id": sender,
// 	}

// 	var transaction hp.Transactions

// 	err := config.TransactionsCollection.FindOne(ctx, filter).Decode(&transaction)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	// Get User Details
// 	filter = bson.M{
// 		"_id": transaction.FromID,
// 	}

// 	var user hp.UserResponse

// 	err = config.UsersCollection.FindOne(ctx, filter).Decode(&user)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	// Get Merchant Details
// 	filter = bson.M{
// 		"_id": transaction.ToID,
// 	}

// 	var merchant hp.UserResponse

// 	err = config.UsersCollection.FindOne(ctx, filter).Decode(&merchant)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	// Deduct from User
// 	filter = bson.M{
// 		"_id": transaction.FromID,
// 	}

// 	update := bson.M{
// 		"$set": bson.M{
// 			"wallet": user.Wallet - transaction.Amount,
// 		},

// 		"$push": bson.M{
// 			"transactions": transaction.ID,
// 		},

// 		"$inc": bson.M{
// 			"transactions_count": 1,
// 		},

// 		"$currentDate": bson.M{
// 			"last_updated": true,
// 		},

// 		"$setOnInsert": bson.M{
// 			"created_at": time.Now(),
// 		},

// 		"$addToSet": bson.M{
// 			"merchants": merchant.ID,
// 		},

// 		"$inc": bson.M{
// 			"merchants_count": 1,
// 		},
// 	}

// 	updateResult, err := config.UsersCollection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	log.Println("updateResult: ", updateResult)

// 	// Add to Merchant

// 	filter = bson.M{
// 		"_id": transaction.ToID,
// 	}

// 	update = bson.M{
// 		"$set": bson.M{
// 			"wallet": merchant.Wallet + transaction.Amount,
// 		},
// 		"$push": bson.M{
// 			"transactions": transaction.ID,
// 		},
// 	}

// 	updateResult, err = config.UsersCollection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	log.Println("updateResult: ", updateResult)

// 	// Update Transaction Status
// 	filter = bson.M{
// 		"_id": transaction.ID,
// 	}

// 	update = bson.M{
// 		"$set": bson.M{
// 			"status": "completed",
// 		},

// 		"$currentDate": bson.M{
// 			"last_updated": true,
// 		},
// 	}

// 	updateResult, err = config.TransactionsCollection.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		config.Logs("error", err.Error())
// 	}

// 	log.Println("updateResult: ", updateResult)
// }
