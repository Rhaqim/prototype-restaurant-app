package controllers

import (
	"context"
	"log"
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

var eventCollection = config.EventCollection
var orderCollection = config.OrderCollection
var productCollection = config.ProductCollection

func CreateEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Event

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

	//  Ensure that hostedIDs are not empty
	if len(request.Invited) < 1 {
		response := hp.SetError(nil, "IDs cannot be empty", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	request.ID = primitive.NewObjectID()

	insertResult, err := eventCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error inserting into database", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}
	log.Println("insertResult: ", insertResult)

	hostingResponse := hp.Event{
		ID:        insertResult.InsertedID.(primitive.ObjectID),
		Title:     request.Title,
		HostID:    user.ID,
		Invited:   request.Invited,
		Attendees: request.Attendees,
		Venue:     request.Venue,
		Type:      request.Type,
		Bill:      request.Bill,
		Budget:    request.Budget,
		CreatedAt: request.CreatedAt,
	}

	response := hp.SetSuccess("Event created", hostingResponse, funcName)
	c.JSON(http.StatusOK, response)
}

func GetUserEventsByHost(c *gin.Context) {
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

	filter := bson.M{"host_id": user.ID}
	cursor, err := eventCollection.Find(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error finding hosted events", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	var hosting []hp.Event
	if err = cursor.All(ctx, &hosting); err != nil {
		response := hp.SetError(err, "Error decoding hosted events", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" events found", hosting, funcName)
	c.JSON(http.StatusOK, response)
}

func UpdateEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var funcName = ut.GetFunctionName()

	var request hp.Event

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

	id, err := primitive.ObjectIDFromHex(request.ID.Hex())
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	update := bson.M{
		"$set": bson.M{
			"title":     request.Title,
			"invited":   request.Invited,
			"attendees": request.Attendees,
			"orders":    request.Orders,
			"venue":     request.Venue,
			"type":      request.Type,
			"bill":      request.Bill,
			"updatedAt": request.UpdatedAt,
		},
	}

	updateResult, err := eventCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		response := hp.SetError(err, "Error updating hosted event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event updated", updateResult, funcName)
	c.JSON(http.StatusOK, response)
}

func DeleteEvent(c *gin.Context) {
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

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": id, "host_id": user.ID}

	deleteResult, err := eventCollection.DeleteOne(ctx, filter)

	if err != nil {
		response := hp.SetError(err, "Error deleting hosted event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	response := hp.SetSuccess(" event deleted", deleteResult, funcName)
	c.JSON(http.StatusOK, response)
}

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

	insert := bson.M{
		"_id":       primitive.NewObjectID(),
		"event":     request.Event,
		"customer":  user,
		"product":   request.Product,
		"quantity":  request.Quantity,
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
	}

	insertResult, err := orderCollection.InsertOne(ctx, insert)
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
		product_id, err := primitive.ObjectIDFromHex(request.Product.ID.Hex())
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
		event_id, err := primitive.ObjectIDFromHex(request.Event.ID.Hex())
		if err != nil {
			response := hp.SetError(err, "Error converting id to object id", funcName)
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
				"bill": float64(request.Product.Price * float64(request.Quantity)),
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

	filter := bson.M{"customer.id": user.ID}

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

	filter := bson.M{"_id": order_id, "customer.id": user.ID}

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

func InviteFriendToEvent(c *gin.Context) {
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

	var request hp.InviteFriendToEventRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		response := hp.SetError(err, "Error binding request", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	// check if event exists
	event_id, err := primitive.ObjectIDFromHex(request.Event.ID.Hex())
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"_id": event_id}
	var event hp.Event
	err = eventCollection.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// check if friendship exists
	if !hp.VerifyFriends(user, request.Friend.ID) {
		response := hp.SetError(err, "Friendship does not exist", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	response := hp.SetSuccess("Friend invited", request.Friend.Username, funcName)
	c.JSON(http.StatusOK, response)
}
