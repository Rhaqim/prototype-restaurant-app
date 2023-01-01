package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func FundWallet(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	request := hp.FundWalletRequest{}

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Await a response from the Paystack API
	// paystackResponse, err := hp.FundWalletPaystack(request, user)
	// if err != nil {
	// 	response := hp.SetError(err, "Error funding wallet", funcName)
	// 	c.AbortWithStatusJSON(http.StatusBadRequest, response)
	// 	return
	// }

	user.Wallet += request.Amount

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"wallet": user.Wallet}}

	_, err = authCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error updating user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Wallet funded", user, funcName)
	c.JSON(http.StatusOK, response)
}

func CreateWallet(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.CreateWalletRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if User has a pin already
	if user.TxnPin != "" || len(user.TxnPin) > 0 {
		response := hp.SetError(err, "Pin already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if user has a wallet
	if user.Wallet != 0 {
		response := hp.SetError(err, "Wallet already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user.Wallet = 0
	request.TxnPin, err = auth.HashPassword(request.TxnPin)
	if err != nil {
		response := hp.SetError(err, "Error hashing pin", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"wallet":  user.Wallet,
			"txn_pin": request.TxnPin,
		}}

	_, err = authCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error creating wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Wallet created", user, funcName)
	c.JSON(http.StatusOK, response)
}

func ChangePin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.ChangePinRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if User has a pin already
	if user.TxnPin == "" || len(user.TxnPin) == 0 {
		response := hp.SetError(err, "Pin does not exist", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if user has a wallet
	if user.Wallet == 0 {
		response := hp.SetError(err, "Wallet does not exist", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if old pin is correct
	ok := auth.CheckPasswordHash(request.OldPin, user.TxnPin)
	if !ok {
		response := hp.SetError(err, "Old pin is incorrect", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.NewPin, err = auth.HashPassword(request.NewPin)
	if err != nil {
		response := hp.SetError(err, "Error hashing pin", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"txn_pin": request.NewPin,
		}}

	_, err = authCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error creating wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Pin changed", user, funcName)
	c.JSON(http.StatusOK, response)
}
