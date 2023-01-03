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
	"go.mongodb.org/mongo-driver/bson"
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check that User role is Business
	if user.Role != hp.Business {
		response := hp.SetError(err, "User is not a Business", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Modify the request
	request.ID = primitive.NewObjectID()
	request.OwnerID = user.ID
	request.Category = hp.RestaurantCategory(hp.RestaurantCategory(request.Category).String())
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

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

func GetRestaurant(c *gin.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get restaurant from db
	// return restaurant

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	// Get restaurant id from request params
	restaurantID := c.Param("id")

	// Get restaurant name from request params
	restaurantName := c.Param("name")

	var filter bson.M

	switch {
	case restaurantID != "":
		// Validate restaurant id
		id, err := primitive.ObjectIDFromHex(restaurantID)
		if err != nil {
			response := hp.SetError(err, "Invalid restaurant id", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"_id": id}
	case restaurantName != "":
		filter = bson.M{"name": restaurantName}
	default:
		filter = bson.M{}
	}

	// Get restaurant from db
	restaurant, err := hp.GetRestaurant(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant found successfully", restaurant, funcName)
	c.JSON(http.StatusOK, response)
}

func GetRestaurants(c *gin.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get restaurant from db
	// return restaurant

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	// Get restaurant id from request params
	category := c.Param("category")

	// Get restaurant name from request params
	country := c.Param("country")

	var filter bson.M

	switch {
	case category != "":
		filter = bson.M{"category": category}
	case country != "":
		filter = bson.M{"country": country}
	default:
		filter = bson.M{}
	}

	// Get restaurant from db
	restaurant, err := hp.GetRestaurant(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant found successfully", restaurant, funcName)
	c.JSON(http.StatusOK, response)
}
