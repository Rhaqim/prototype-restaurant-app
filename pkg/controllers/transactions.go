package controllers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTransaction(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Transaction Verification
	friends := hp.VerifyFriends(ctx, user, request.ToID)

	if !friends {
		response := hp.SetError(nil, "You are not friends with this user", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	sufficientBalance := hp.VerifyWalletSufficientBalance(ctx, user, request.Amount)

	if !sufficientBalance {
		response := hp.SetError(nil, "Insufficient balance", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// modify the request
	request.ID = primitive.NewObjectID()
	request.TransactionUID = hp.TransactionUID
	request.FromID = user.ID
	request.Status = hp.TxnPending

	_, err = config.TransactionCollection.InsertOne(ctx, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := hp.SetSuccess("Transaction created successfully", request, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateTransactionStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get wallet
	var wallet hp.Wallet
	err = walletCollection.FindOne(ctx, bson.M{"user_id": user.ID}).Decode(&wallet)
	if err != nil {
		response := hp.SetError(err, "Error fetching wallet", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// check if pin is correct
	pin := auth.CheckPasswordHash(request.TxnPin, wallet.TxnPin)
	if !pin {
		response := hp.SetError(nil, "Incorrect pin", funcName)
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

	_, err = config.TransactionCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error updating transaction status", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update Wallet Balance
	var transaction hp.Transactions
	err = config.TransactionCollection.FindOne(ctx, filter).Decode(&transaction)
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
	txn, err := hp.UpdateWalletBalance(ctx, transaction)
	if err != nil {
		response := hp.SetError(err, "Error updating wallet balance", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Transaction status updated successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

func GetTransactions(c *gin.Context) {
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

	filter := bson.M{
		"$or": []bson.M{
			{"fromId": user.ID},
			{"toId": user.ID},
		},
	}

	var transactions []hp.Transactions

	cursor, err := config.TransactionCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error fetching transactions", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	if err = cursor.All(ctx, &transactions); err != nil {
		response := hp.SetError(err, "Error fetching transactions", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Transactions fetched successfully", transactions, funcName)
	c.JSON(http.StatusOK, response)
}

// PayBill sends money to the venue of the event
func PayBillforEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.EventBillPayment

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Event
	filter := bson.M{
		"_id": request.EventID,
	}
	event, err := hp.GetEvent(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error fetching event", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Verify Event for Payment
	err = hp.VerificationforEventPayment(ctx, request, event, user)
	if err != nil {
		response := hp.SetError(err, "Error verifying event payment", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Begin Transaction
	txn, err := hp.SendtoVenuePayforEvent(ctx, event, user)
	if err != nil {
		response := hp.SetError(err, "Error sending money to venue", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update Event and Orders
	// Event Status to Finished
	// Orders Paid to True
	eventErrChan := make(chan error, 1)
	orderErrChan := make(chan error, 1)

	go hp.UpdateEventandOrders(ctx, event, txn, eventErrChan, orderErrChan)

	select {
	case err := <-eventErrChan:
		response := hp.SetError(err, "Error updating event status", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	case err := <-orderErrChan:
		response := hp.SetError(err, "Error updating order status", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	default:
	}

	// Send Notification to the Venue
	billAmount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)

	msgVenue := []byte(config.Transaction_ +
		user.Username + " has paid the bill for " + event.Title +
		" of " + billAmount + " to your wallet",
	)

	venueList := []primitive.ObjectID{event.Venue}

	notifyVenue := nf.NewNotification(
		venueList,
		msgVenue,
	)

	err = notifyVenue.Send()
	if err != nil {
		response := hp.SetError(err, "Error sending notification to venue", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Money sent to venue successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

func SendOwnBillforEventToHost(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.EventBillPayment

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Event
	filter := bson.M{
		"_id": request.EventID,
	}
	event, err := hp.GetEvent(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error fetching event", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = hp.VerificationforEventPayment(ctx, request, event, user)
	if err != nil {
		response := hp.SetError(err, "Error verifying event payment", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	txn, err := hp.SendToHost(ctx, event, user)
	if err != nil {
		response := hp.SetError(err, "Error sending money to host", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Send Notification to the Host
	billAmount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)

	msgHost := []byte(config.Transaction_ +
		user.Username + " has sent you " + billAmount +
		" for " + event.Title,
	)

	hostList := []primitive.ObjectID{event.HostID}

	notifyHost := nf.NewNotification(
		hostList,
		msgHost,
	)

	err = notifyHost.Send()
	if err != nil {
		response := hp.SetError(err, "Error sending notification to host", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Money sent to host successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

func SendToOtherUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), config.ContextTimeout)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.SendMoneyOtherUser

	if err := c.ShouldBindJSON(&request); err != nil {
		var response = hp.SetError(err, "Error binding JSON", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Veryfy Pin of the User is correct
	if !hp.VeryfyPin(ctx, user, request.TxnPin) {
		response := hp.SetError(err, "Incorrect Pin", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get User
	filter := bson.M{
		"username": request.Username,
	}
	user2, err := hp.GetUser(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error fetching user", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	txn, err := hp.SendToOtherUsers(ctx, user2, user, request.Amount)
	if err != nil {
		response := hp.SetError(err, "Error sending money to other users", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Send Notification to the Users
	billAmount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)

	msg := user.Username + " has sent you " + billAmount

	if err := nf.AlertUser(config.Transaction_, msg, user2.ID); err != nil {
		response := hp.SetError(err, "Error sending notification to users", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Money sent to users successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}
