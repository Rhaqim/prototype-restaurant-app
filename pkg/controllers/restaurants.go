package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var restaurantCollection = config.RestaurantCollection

func CreateRestaurant(c *gin.Context) {
	// get restaurant data from request body
	// validate restaurant data
	// check if restaurant already exists
	// create restaurant

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Restaurant

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", "CreateRestaurant")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Modify the request
	request.ID = primitive.NewObjectID()
	request.OwnerID = user.ID

	// Check if Restaurant already exists
	//Name is unique
	ok, err := hp.ValidateCreateRestaurantRequest(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error validating restaurant request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	if !ok {
		response := hp.SetError(err, "Restaurant already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Validate all OpenHours
	for _, openHour := range request.OpenHours {
		err := hp.OpenHours(openHour).Validate()
		if err != nil {
			response := hp.SetError(err, "Invalid OpenHours", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Create Restaurant
	_, err = restaurantCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error creating restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant created successfully", request.ID.Hex(), funcName)
	c.JSON(http.StatusOK, response)
}
