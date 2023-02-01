package controllers

import (
	"context"
	"net/http"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	restaurantCollection = config.RestaurantCollection

	CreateRestaurant = AbstractConnection(createRestaurant)
	GetRestaurant    = AbstractConnection(getRestaurant)
	GetRestaurants   = AbstractConnection(getRestaurants)
	UpdateRestaurant = AbstractConnection(updateRestaurant)
	DeleteRestaurant = AbstractConnection(deleteRestaurant)
)

func createRestaurant(c *gin.Context, ctx context.Context) {
	// get restaurant data from request body
	// validate restaurant data
	// check if restaurant already exists
	// create restaurant

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

	// Check if KYC is complete
	if !hp.CheckKYCStatus(user) {
		response := hp.SetError(err, "KYC not complete", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Modify the request
	request.ID = primitive.NewObjectID()
	request.RestaurantUID = hp.RestaurantUID
	request.Slug = ut.Slugify(request.Name)
	request.OwnerID = user.ID
	request.Category = hp.RestaurantCategory(hp.RestaurantCategory(request.Category).String())
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()
	request.Latitude, request.Longitude, _ = hp.GetLatLon(request.Address)

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

func getRestaurant(c *gin.Context, ctx context.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get restaurant from db
	// return restaurant

	var funcName = ut.GetFunctionName()

	// Get restaurant id from request query params
	restaurantID := c.Query("id")

	// Get restaurant name from request query params
	restaurantName := c.Query("name")

	// Get restaurant slug from request query params
	restaurantSlug := c.Query("slug")

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
	case restaurantSlug != "":
		filter = bson.M{"slug": restaurantSlug}
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

func getRestaurants(c *gin.Context, ctx context.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get restaurant from db
	// return restaurant

	var funcName = ut.GetFunctionName()

	// Get restaurant id from request query params
	category := c.Query("category")

	// Get restaurant name from request query params
	country := c.Query("country")

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
	restaurant, err := hp.GetRestaurants(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant found successfully", restaurant, funcName)
	c.JSON(http.StatusOK, response)
}

func updateRestaurant(c *gin.Context, ctx context.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get restaurant data from request body
	// validate restaurant data
	// check if restaurant already exists
	// update restaurant

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "Error getting user from token", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get restaurant data from request body
	var request hp.Restaurant

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", "UpdateRestaurant")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Modify the request
	request.UpdatedAt, _ = hp.CreatedAtUpdatedAt()

	// Validate all OpenHours
	for _, openHour := range request.OpenHours {
		err := hp.OpenHours(openHour).Validate()
		if err != nil {
			response := hp.SetError(err, "Invalid OpenHours", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}

	// Update Restaurant
	_, err = restaurantCollection.UpdateOne(ctx, bson.M{"_id": request.ID, "owner_id": user.ID}, bson.M{"$set": request})
	if err != nil {
		response := hp.SetError(err, "Error updating restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant updated successfully", request.ID.Hex(), funcName)
	c.JSON(http.StatusOK, response)
}

func deleteRestaurant(c *gin.Context, ctx context.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// delete restaurant

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "Error getting user from token", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Get restaurant id from request query params
	restaurantID := c.Query("id")

	// Validate restaurant id
	id, err := primitive.ObjectIDFromHex(restaurantID)
	if err != nil {
		response := hp.SetError(err, "Invalid restaurant id", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check Restaurant Belongs to User
	owner, err := hp.CheckRestaurantBelongsToUser(ctx, id, user)
	if err != nil {
		response := hp.SetError(err, "could not verify restaurant belongs to user", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
	}

	if !owner {
		response := hp.SetError(err, "restaurant does not belong to user", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
	}

	// Delete restaurant
	_, err = restaurantCollection.DeleteOne(ctx, bson.M{"_id": id, "owner_id": user.ID})
	if err != nil {
		response := hp.SetError(err, "Error deleting restaurant", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Restaurant deleted successfully", id.Hex(), funcName)
	c.JSON(http.StatusOK, response)
}
