package notifications

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	hp "github.com/Rhaqim/thedutchapp/pkg/helpers"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var notificationCollection = config.NotificationCollection

// Websocket Notification Handler
// This handler is used to upgrade the HTTP connection to a WebSocket connection
// and add the connection to the Connections map
var (
	Upgrader = websocket.Upgrader{
		// Allow Connections from any origin
		CheckOrigin: func(r *http.Request) bool {
			// if r.Header.Get("Origin") == "http://localhost:3000" {
			// 	return true
			// }
			// return false
			return true
		},
	}

	// Connections is a map of WebSocket Connections keyed by user ID
	Connections     = make(map[string][]*websocket.Conn)
	ConnectionsLock sync.RWMutex
)

// WsHandler is the WebSocket handler
// It upgrades the HTTP connection to a WebSocket connection
// It adds the connection to the Connections map
// It removes the connection when it is closed
// It uses the connection to receive messages from the client
// It is called by the router when a client connects to the WebSocket endpoint
// It takes the Gin context as an argument
// It reads the message from SendNotification and sends it to the client
func WsHandler(c *gin.Context) {
	funcName := ut.GetFunctionName()

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Handle error
		hp.SetError(err, "Error upgrading HTTP connection to WebSocket connection", funcName)
		return
	}
	defer conn.Close()

	// Get the user from the token
	user, err := hp.GetUserFromToken(c)
	if err != nil {
		hp.SetError(err, "Error getting user from token", funcName)
		return
	}

	userID := user.ID.Hex()

	hp.SetInfo("New connection for user: "+user.Username+"", funcName)

	// Add the connection to the list of Connections for the user
	ConnectionsLock.Lock()
	Connections[userID] = append(Connections[userID], conn)
	ConnectionsLock.Unlock()

	// Remove the connection when it is closed
	defer func() {
		ConnectionsLock.Lock()
		defer ConnectionsLock.Unlock()
		for i, c := range Connections[userID] {
			if c == conn {
				Connections[userID] = append(Connections[userID][:i], Connections[userID][i+1:]...)
				break
			}
		}
	}()

	// Use the connection to receive messages from the client
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			hp.SetError(err, "Error reading message from client", funcName)
			break
		}
	}
}

// SendNotification sends a notification to a user
// This function is called by the event handlers to send a notification to a user
// It takes the user ID and the message to send
// It gets the connections for the user from the Connections map
// It loops over the connections and sends the message to each connection
// If the connection is no longer usable, it is removed from the Connections map
func SendNotification(user_ID primitive.ObjectID, message []byte) {
	funcName := ut.GetFunctionName()

	// convert the user ID to a string
	userID := user_ID.Hex()

	hp.SetInfo(
		"\n Sending notification to user: "+user_ID.Hex()+"\n"+
			"Message: "+string(message)+"\n"+
			"Number of connections: "+fmt.Sprintf("%d", len(Connections[userID]))+"\n",
		funcName)

	// Get the connections for the user
	ConnectionsLock.RLock()
	conns, ok := Connections[userID]
	ConnectionsLock.RUnlock()
	if !ok {
		hp.SetInfo("No connections for user: "+userID+"", funcName)
		return
	}

	// Loop over the connections and send the message
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// Remove the connection if it is no longer usable
			conn.Close()
		}
	}
}

// BroadcastNotification sends a notification to all users
// This function is called by the event handlers to send a notification to all users
// It takes the message to send
// It gets the connections for all users from the Connections map
// It loops over the connections and sends the message to each connection
// If the connection is no longer usable, it is removed from the Connections map
func BroadcastNotification(message []byte) {
	funcName := ut.GetFunctionName()

	hp.SetInfo(
		"\n Sending notification to all users\n"+
			"Message: "+string(message)+"\n",
		funcName)

	// Get the connections for all users
	ConnectionsLock.RLock()
	conns := Connections
	ConnectionsLock.RUnlock()

	// Loop over the connections and send the message
	for _, conns := range conns {
		for _, conn := range conns {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				// Remove the connection if it is no longer usable
				conn.Close()
			}
		}
	}
}

// Notifications is a model struct for notifications
// It is used to store notifications in the database
// It is used to send notifications to the client
// It contains the user ID, the notification message, the time the notification was created and whether it has been seen
type Notifications struct {
	ID           primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	UserIDs      []primitive.ObjectID `json:"user_ids,omitempty" bson:"user_ids,omitempty"`
	Notification []byte               `json:"notification,omitempty" bson:"notification,omitempty"`
	Seen         bool                 `json:"seen,omitempty" bson:"seen,omitempty"`
	Time         time.Time            `json:"time,omitempty" bson:"time,omitempty"`
}

// NewNotification creates a new notification
// It takes the user ID and the notification message
// It returns a pointer to the notification
func NewNotification(userIDs []primitive.ObjectID, notification []byte) *Notifications {
	return &Notifications{
		ID:           primitive.NewObjectID(),
		UserIDs:      userIDs,
		Notification: notification,
		Seen:         false,
		Time:         time.Now(),
	}
}

// Create inserts the notification into the database
// It takes the context
// It returns an error if there is one
func (n *Notifications) Create(ctx context.Context) error {
	funcName := ut.GetFunctionName()

	// Send the notification to the users
	n.SendNotification()

	// Insert the notification into the database
	_, err := notificationCollection.InsertOne(ctx, n)
	if err != nil {
		config.Logs("error", "Error inserting notification: "+err.Error()+"", funcName)
		return err
	}

	return nil
}

// SendNotification sends the notification to all the users listed in the UserIDs field
// It Uses the SendNotification function to send the notification to the users
// it is called by the event handlers
// It takes the context
// It saves the notification to the database
func (n *Notifications) SendNotification() {
	funcName := ut.GetFunctionName()

	hp.SetInfo("Sending notification to users", funcName)

	for _, userID := range n.UserIDs {
		// Get the user's WebSocket connections
		go SendNotification(userID, []byte(n.Notification))
	}
}

// BroadcastNotification sends the notification to all users
// It Uses the BroadcastNotification function to send the notification to all users
// it is called by the event handlers
// It takes the context
// It saves the notification to the database
func (n *Notifications) BroadcastNotification() {
	funcName := ut.GetFunctionName()

	hp.SetInfo("Sending notification to all users", funcName)

	go BroadcastNotification([]byte(n.Notification))
}

// GetNotifications returns all notifications that match the filter
// It takes a context and a filter as arguments
// It returns a slice of Notifications and an error
func GetNotifications(ctx context.Context, filter bson.M) ([]Notifications, error) {
	funcName := ut.GetFunctionName()

	var notifications []Notifications

	cur, err := notificationCollection.Find(ctx, filter)
	if err != nil {
		hp.SetError(err, "Error finding notifications", funcName)
		return nil, err
	}

	for cur.Next(ctx) {
		var notification Notifications
		err := cur.Decode(&notification)
		if err != nil {
			hp.SetError(err, "Error decoding notification", funcName)
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetNotificationsByUser returns all notifications for a user
// Checks if userID present in the UserIDs field
// It takes a context and a user ID as arguments
// It returns a slice of Notifications and an error
func GetNotificationsByUser(ctx context.Context, userID primitive.ObjectID) ([]Notifications, error) {
	funcName := ut.GetFunctionName()

	hp.SetInfo("Getting notifications for user: "+userID.Hex()+"", funcName)

	// Check if id present is userIDs field
	filter := bson.M{
		"user_ids": bson.M{
			"$in": []primitive.ObjectID{userID},
		},
	}

	notifications, err := GetNotifications(ctx, filter)
	if err != nil {
		hp.SetError(err, "Error getting notifications for user: "+userID.Hex()+"", funcName)
		return nil, err
	}

	return notifications, nil
}

func UpdateNotification(ctx context.Context, filter bson.M, update bson.M) error {
	funcName := ut.GetFunctionName()

	_, err := notificationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		hp.SetError(err, "Error updating notification", funcName)
		return err
	}

	return nil
}
