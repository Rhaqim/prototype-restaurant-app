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

/* EVENT TRANSACTION */

// PayBill sends money to the venue of the event
// it takes the event id and the pin of the user paying for the event
// Verifies the pin and confirms user has suffiecient amount in wallet
// Sends the money to the venue's owner's wallet
// Updates the Event Status to Finished and the bills for the Users to Paid
// Sends a notification to the venue about the payment
// it returns the transaction details
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
		hp.SetError(err, "Error updating event status", funcName)
	case err := <-orderErrChan:
		hp.SetError(err, "Error updating order status", funcName)
	default:
	}

	// Send Notification to the Venue
	billAmount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)

	msgVenue := []byte(config.Transaction_ +
		user.Username + " has paid the bill for " + event.Title +
		" of " + billAmount + " money has been sent to your wallet",
	)

	venue, err := hp.GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		hp.SetError(err, "Error fetching venue", funcName)
	}

	venueList := []primitive.ObjectID{venue.OwnerID}

	notifyVenue := nf.NewNotification(
		venueList,
		msgVenue,
	)

	notifyVenue.Send()

	response := hp.SetSuccess("Money sent to venue successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

// SendMoneytoHost sends money for Orders to the host of the event
// it takes the event id and the pin of the user sending the money
// Verifies the pin and confirms user has suffiecient amount in wallet
// It sends the money directly to the host's wallet
// Sends a notification to the host about the payment
// it returns the transaction details
func SendMoneytoHost(c *gin.Context) {
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

	notifyHost.Send()

	response := hp.SetSuccess("Money sent to host successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

// PayOwnBill pays the bill for Orders made by the user
// it takes the event id and the pin of the user sending the money
// Verifies the pin and confirms user has suffiecient amount in wallet
// It sends the money directly to the venue owner's wallet
// Sends a notification to the venue owner about the payment
// Sends a notification to the host about the payment
// it returns the transaction details
func PayOwnBill(c *gin.Context) {
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

	// Verify Payment for Event
	err = hp.VerificationforEventPayment(ctx, request, event, user)
	if err != nil {
		response := hp.SetError(err, "Error verifying event payment", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Pay Own Bill
	txn, err := hp.PayOwnBillforEvent(ctx, event, user)
	if err != nil {
		response := hp.SetError(err, "Error paying own bill", funcName)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update All Customer Orders for the Event to paid
	orderErrChan := make(chan error)
	eventErrChan := make(chan error)

	go hp.UpdateCustomerOrders(ctx, event, user, txn, orderErrChan, eventErrChan)

	select {
	case err := <-orderErrChan:
		if err != nil {
			hp.SetError(err, "Error updating customer orders", funcName)
		}
	case err := <-eventErrChan:
		if err != nil {
			hp.SetError(err, "Error updating event", funcName)
		}
	}

	// Send Notification to the Host
	billAmount := strconv.FormatFloat(txn.Amount, 'f', 2, 64)

	venue, err := hp.GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		hp.SetError(err, "Error fetching venue", funcName)
	}

	msgHost := []byte(config.Transaction_ +
		user.Username + " has paid their bill for " + event.Title +
		" amount of " + billAmount + " has been sent to " + venue.Name,
	)

	hostList := []primitive.ObjectID{event.HostID, venue.OwnerID}

	notifyHost := nf.NewNotification(
		hostList,
		msgHost,
	)

	notifyHost.Send()

	response := hp.SetSuccess("Bill paid successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}

// SendToOtherUsers sends money to another user
// it takes the username of the user and the amount to be sent
// Sends the money to the user's wallet
// Sends a notification to the user about the payment
// it returns the transaction details
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

	nf.AlertUser(config.Transaction_, msg, user2.ID)

	response := hp.SetSuccess("Money sent to users successfully", txn, funcName)
	c.JSON(http.StatusOK, response)
}
