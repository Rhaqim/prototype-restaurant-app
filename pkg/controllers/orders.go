package controllers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var orderCollection = config.OrderCollection
var productCollection = config.ProductCollection

func CreateOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Order

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

	request.ID = primitive.NewObjectID()
	request.CustomerID = user.ID

	insertResult, err := orderCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error creating order", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// define wait group for concurrency
	wg := sync.WaitGroup{}

	wg.Add(2)

	//  Update the Product with the new order and decrement stock by quantity also update the event with the new order using concurrency
	go func() {
		defer wg.Done()
		// Update the Product with the new order
		product_id, err := primitive.ObjectIDFromHex(request.ProductID.Hex())
		if err != nil {
			response := hp.SetError(err, "Error converting id to object id", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		product_filter := bson.M{"_id": product_id}
		product_update := bson.M{
			// decrement stock by quantity
			"$inc": bson.M{
				"stock": -request.Quantity,
			},
		}

		_, err = productCollection.UpdateOne(ctx, product_filter, product_update)
		if err != nil {
			response := hp.SetError(err, "Error updating hosted event", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()
	go func() {
		defer wg.Done()
		// Update the Event with the new order
		event_id, err := primitive.ObjectIDFromHex(request.EventID.Hex())
		if err != nil {
			response := hp.SetError(err, "Error converting id to object id", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		// Fetch the product to get the price
		var product hp.Product
		err = productCollection.FindOne(ctx, bson.M{"_id": request.ProductID}).Decode(&product)
		if err != nil {
			response := hp.SetError(err, "Error finding product", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		event_filter := bson.M{"_id": event_id}
		event_update := bson.M{
			"$push": bson.M{
				"orders": insertResult.InsertedID,
			},
			// update bill with new order
			"$inc": bson.M{
				"bill": float64(product.Price * float64(request.Quantity)),
			},
		}

		_, err = eventCollection.UpdateOne(ctx, event_filter, event_update)
		if err != nil {
			response := hp.SetError(err, "Error updating hosted event", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}
	}()

	wg.Wait()

	response := hp.SetSuccess("Order created", insertResult, funcName)
	c.JSON(http.StatusOK, response)
}

func GetOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	filter := bson.M{"customer_id": user.ID}

	cursor, err := orderCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting orders", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	var orders []hp.Order
	if err = cursor.All(ctx, &orders); err != nil {
		response := hp.SetError(err, "Error getting orders", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Orders retrieved", orders, funcName)
	c.JSON(http.StatusOK, response)
}

func GetOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	order_id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": order_id, "customer_id": user.ID}

	var order hp.Order
	err = orderCollection.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		response := hp.SetError(err, "Error getting order", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess("Order retrieved", order, funcName)
	c.JSON(http.StatusOK, response)
}
