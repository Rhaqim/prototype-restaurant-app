package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	nf "github.com/Rhaqim/thedutchapp/pkg/notifications"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	orderCollection = config.OrderCollection

	CreateOrder        = AbstractConnection(createOrder)
	GetOrders          = AbstractConnection(getOrders)
	GetOrder           = AbstractConnection(getOrder)
	GetUserEventOrders = AbstractConnection(getUserEventOrders)
	GetEventOrders     = AbstractConnection(getEventOrders)
)

func createOrder(c *gin.Context, ctx context.Context) {
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	// Get Event
	filter := bson.M{"_id": request.EventID}
	event, err := hp.GetEvent(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting event", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Check Event Status
	if event.EventStatus != hp.Ongoing {
		response := hp.SetError(err, "Event is not ongoing", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	request.ID = primitive.NewObjectID()
	request.CustomerID = user.ID
	request.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	request.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

	// Update Stock
	stockErrChan := make(chan error)
	go hp.UpdateStock(ctx, request, stockErrChan)

	// Update Bill
	billErrChan := make(chan error)
	totalChan := make(chan float64)
	go hp.UpdateBill(ctx, request, billErrChan, totalChan)

	// Check for errors in UpdateStock goroutine
	for err := range stockErrChan {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Error updating stock", funcName))
		return
	}

	// Check for errors in UpdateBill goroutine
	for err := range billErrChan {
		c.AbortWithStatusJSON(http.StatusInternalServerError, hp.SetError(err, "Error updating bill", funcName))
		return
	}

	// Get total bill for all products from UpdateBill goroutine
	// and add it to the order bill
	request.Bill = <-totalChan

	// Add order to database
	insertResult, err := orderCollection.InsertOne(ctx, request)
	if err != nil {
		response := hp.SetError(err, "Error creating order", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// NOTIFICATIONS
	// Get Venue Owner from venue id
	filter = bson.M{"_id": event.RestaurantID}
	venue, err := hp.GetRestaurant(ctx, filter)
	if err != nil {
		response := hp.SetError(err, "Error getting venue", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	// Get products from order
	productNames := make(map[string]int)
	for _, v := range request.Products {
		// Get product from database
		filter := bson.M{"_id": v.ProductID}
		product, err := hp.GetProduct(ctx, filter)
		if err != nil {
			response := hp.SetError(err, "Error getting product", funcName)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		productNames[product.Name] = v.Quantity
	}

	var message string

	for k, v := range productNames {
		message += fmt.Sprintf("%s x%d, ", k, v)
	}

	msg := []byte(config.Order_ +
		user.Username +
		" ordered " + message +
		"for " + event.Title +
		" on " + request.CreatedAt.Time().Format("02-01-2006 15:04:05"))

	// Send Notification to Venue regarding new order
	venueList := []primitive.ObjectID{venue.OwnerID}

	notifyVenue := nf.NewNotification(
		venueList,
		msg,
	)
	notifyVenue.Send()

	// Send Notification to Event group regarding new order
	// remove user from attendees so they don't get notified
	// about their own order
	for i, v := range event.Attendees {
		if v == user.ID {
			event.Attendees = append(event.Attendees[:i], event.Attendees[i+1:]...)
		}
	}
	notifyGroup := nf.NewNotification(
		event.Attendees,
		msg,
	)
	notifyGroup.Send()

	response := hp.SetSuccess("Order created", insertResult, funcName)
	c.JSON(http.StatusOK, response)
}

func getOrders(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
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

func getOrder(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
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

func getUserEventOrders(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	user, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
		return
	}

	event_id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"customer_id": user.ID, "event_id": event_id}

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

func getEventOrders(c *gin.Context, ctx context.Context) {
	var funcName = ut.GetFunctionName()

	_, err := hp.GetUserFromToken(c)
	if err != nil {
		response := hp.SetError(err, "User not logged in", funcName)
		c.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	event_id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		response := hp.SetError(err, "Error converting id to object id", funcName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
		return
	}

	filter := bson.M{"event_id": event_id}

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
