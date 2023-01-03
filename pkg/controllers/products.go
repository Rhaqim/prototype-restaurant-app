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

var productCollection = config.ProductCollection

func AddProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Product

	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding json", "AddProduct")
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check if User is Owner of the Restaurant
	_, err = hp.CheckRestaurantBelongsToUser(ctx, request.RestaurantID, user)
	if err != nil {
		response := hp.SetError(err, "Restaurant does not belong to the user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Check that product name is unique
	filter := bson.M{"name": request.Name, "restaurant_id": request.RestaurantID}
	_, err = hp.GetProduct(ctx, filter)
	if err == nil {
		response := hp.SetError(err, "Product name already exists", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Modify the request
	request.ID = primitive.NewObjectID()
	request.SuppliedID = user.ID
	request.Category = hp.Categories(hp.Categories(request.Category).String())
	request.CreatedAt, request.UpdatedAt = hp.CreatedAtUpdatedAt()

	insertResult, err := productCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error creating product", "AddProduct")
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Product created successfully", insertResult.InsertedID, funcName)
	c.JSON(http.StatusOK, response)
}

func GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var products []hp.Product

	// Get category from query params
	category := c.Query("category")

	// Get restaurant_id from query params
	restaurantID := c.Query("restaurant_id")

	// Get user_id from query params
	userID := c.Query("user_id")

	var filter bson.M

	switch {
	case category != "":
		filter = bson.M{"category": category}
	case restaurantID != "":
		restaurantID, err := primitive.ObjectIDFromHex(restaurantID)
		if err != nil {
			response := hp.SetError(err, "Invalid restaurant ID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"restaurant_id": restaurantID}
	case userID != "":
		userID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			response := hp.SetError(err, "Invalid user ID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"user_id": userID}
	default:
		filter = bson.M{}
	}

	products, err := hp.GetProducts(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting products", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Products retrieved successfully", products, funcName)
	c.JSON(http.StatusOK, response)
}

func GetProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	// Get name from query params
	name := c.Query("name")

	// Get ID from query params
	id := c.Query("id")

	var filter bson.M

	switch {
	case name != "":
		filter = bson.M{"name": name}
	case id != "":
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			response := hp.SetError(err, "Invalid Product ID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"_id": id}
	case name != "" && id != "":
		id, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			response := hp.SetError(err, "Invalid Product ID", funcName)
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		filter = bson.M{"name": name, "_id": id}
	default:
		filter = bson.M{}
	}

	product, err := hp.GetProduct(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting products", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Product retrieved successfully", product, funcName)
	c.JSON(http.StatusOK, response)
}

func GetProductsForRestaurant(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var products []hp.Product

	restaurantID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Invalid restaurant ID", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"restaurant_id": restaurantID}
	cursor, err := productCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting products", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	if err = cursor.All(ctx, &products); err != nil {
		response := hp.SetError(err, "Error getting products", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Products retrieved successfully", products, funcName)
	c.JSON(http.StatusOK, response)
}

func DeleteProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	productID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Invalid product ID", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	product, err := hp.GetProductbyID(ctx, productID)
	if err != nil {
		response := hp.SetError(err, "Error getting product", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check if User is Owner of the Restaurant
	_, err = hp.CheckRestaurantBelongsToUser(ctx, product.RestaurantID, user)
	if err != nil {
		response := hp.SetError(err, "Restaurant does not belong to the user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// Delete Product
	deleteResult, err := productCollection.DeleteOne(ctx, bson.M{"_id": productID})
	if err != nil {
		response := hp.SetError(err, "Error deleting product", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Product deleted successfully", deleteResult.DeletedCount, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateProduct(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Product
	if err := c.ShouldBindJSON(&request); err != nil {
		response := hp.SetError(err, "Error binding request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Check if User is Owner of the Restaurant
	_, err = hp.CheckRestaurantBelongsToUser(ctx, request.RestaurantID, user)
	if err != nil {
		response := hp.SetError(err, "Restaurant does not belong to the user", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	updateResult, err := productCollection.UpdateOne(ctx, bson.M{"_id": request.ID}, request)
	if err != nil {
		response := hp.SetError(err, "Error updating product", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var response = hp.SetSuccess("Product updated successfully", updateResult.ModifiedCount, funcName)
	c.JSON(http.StatusOK, response)
}
