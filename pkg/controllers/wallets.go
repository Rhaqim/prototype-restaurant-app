package controllers

import (
	"context"
	"net/http"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	walletCollection = config.WalletCollection

	CreateWallet     = AbstractConnection(createWallet)
	ChangePin        = AbstractConnection(changePin)
	FundWallet       = AbstractConnection(fundWallet)
	GetWalletBalance = AbstractConnection(getWalletBalance)
)

func createWallet(c *gin.Context, ctx context.Context) {

	var funcName = ut.GetFunctionName()

	var walletPin struct {
		TxnPin string `json:"txn_pin" binding:"required"`
	}

	var request hp.Wallet

	if err := c.ShouldBindJSON(&walletPin); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check if KYC is complete
	if !hp.CheckKYCStatus(user) {
		response := hp.SetError(err, "KYC not complete", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if User has a Wallet already
	exists, err := hp.CheckifWalletExists(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error checking if wallet exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	if exists {
		response := hp.SetError(err, "Wallet already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// modify request
	request.ID = primitive.NewObjectID()
	request.UserID = user.ID
	request.TxnPin, err = auth.HashPassword(walletPin.TxnPin)
	if err != nil {
		response := hp.SetError(err, "Error hashing password", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

	// Create a new wallet
	insertResult, err := walletCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error creating wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// update user with wallet id
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{
		"wallet": insertResult.InsertedID,
	}}

	_, err = authCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error updating user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Wallet created", insertResult.InsertedID, funcName)
	c.JSON(http.StatusOK, response)
}

func changePin(c *gin.Context, ctx context.Context) {

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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get user wallet
	wallet, err := hp.GetWallet(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error getting wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check if old pin is correct
	ok := auth.CheckPasswordHash(request.OldPin, wallet.TxnPin)
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

	update_at, _ := hp.CreatedAtUpdatedAt()

	filter := bson.M{"_id": wallet.ID}
	update := bson.M{
		"$set": bson.M{
			"txn_pin":    request.NewPin,
			"updated_at": update_at,
		}}

	_, err = walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response := hp.SetError(err, "Error creating wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Pin changed", user, funcName)
	c.JSON(http.StatusOK, response)
}

func fundWallet(c *gin.Context, ctx context.Context) {

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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Perform Transaction Fund Wallet to Paystack
	amount, err := hp.FundWalletPaystack(request, user)
	if err != nil {
		response := hp.SetError(err, "Couldn't complete wallet funding", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Update user wallet
	err = hp.AddMoney(ctx, user, amount)
	if err != nil {
		response := hp.SetError(err, "Error updating user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Wallet funded", user, funcName)
	c.JSON(http.StatusOK, response)
}

func getWalletBalance(c *gin.Context, ctx context.Context) {

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	wallet, err := hp.GetWallet(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error getting wallet", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Wallet balance", wallet.Balance, funcName)
	c.JSON(http.StatusOK, response)
}
