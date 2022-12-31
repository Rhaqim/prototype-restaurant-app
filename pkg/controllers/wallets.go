package controllers

import (
	"context"
	"net/http"
	"time"

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
