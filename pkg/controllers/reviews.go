package controllers

import (
	"context"
	"net/http"

	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	AddReview    = AbstractConnection(addReview)
	DeleteReview = AbstractConnection(deleteReview)
)

func addReview(c *gin.Context, ctx context.Context) {
	// get restaurant id from request params
	// validate restaurant id
	// get review data from request body
	// validate review data
	// check if restaurant exists
	// add review

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "Error getting user from token", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var request hp.Review

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.ID = primitive.NewObjectID()
	request.Author = user.ID
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

	// insert review
	request.CreateReview(ctx)

	response := hp.SetSuccess("Review added successfully", request.ID.Hex(), funcName)
	c.JSON(http.StatusOK, response)
}

func deleteReview(c *gin.Context, ctx context.Context) {
	// get review id from request params
	// validate review id
	// check if review exists
	// delete review

	var funcName = ut.GetFunctionName()

	var request hp.Review

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// delete review
	request.DeleteReview(ctx)

	response := hp.SetSuccess("Review deleted successfully", request.ID.Hex(), funcName)
	c.JSON(http.StatusOK, response)
}

func getReviews(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	var user_id = c.Query("user_id")
	var restaurant_id = c.Query("restaurant_id")

	var filter = bson.M{}

	switch {
	case user_id != "":
		user_id, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			response := hp.SetError(err, "Error getting user id from query", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter["author"] = user_id
	case restaurant_id != "":
		// get reviews by restaurant id
		restaurant_id, err := primitive.ObjectIDFromHex(restaurant_id)
		if err != nil {
			response := hp.SetError(err, "Error getting restaurant id from query", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter["restaurant"] = restaurant_id
	default:
		// get all reviews
		response := hp.SetError(nil, "No query params provided", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	reviews, err := hp.GetReviews(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting reviews", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Reviews fetched successfully", reviews, funcName)
	c.JSON(http.StatusOK, response)
}
