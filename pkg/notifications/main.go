package notifications

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Handle error
		config.Logs("error", "Error upgrading HTTP connection to WebSocket connection: "+err.Error()+"", "pkg/notifications/main.go")
		return
	}
	defer conn.Close()

	// Get the user from the token
	user, err := helpers.GetUserFromToken(c)
	if err != nil {
		config.Logs("error", "Error getting user from token: "+err.Error()+"", "pkg/notifications/main.go")
		return
	}

	userID := user.ID.Hex()

	config.Logs("info", "New connection for user: "+user.Username+"", "pkg/notifications/main.go")

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
			config.Logs("error", "Error reading message from client: "+err.Error()+"", "pkg/notifications/main.go")
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
	userID := user_ID.Hex()
	config.Logs("info",
		"\n Sending notification to user: "+user_ID.Hex()+"\n"+
			"Message: "+string(message)+"\n"+
			"Number of connections: "+fmt.Sprintf("%d", len(Connections[userID]))+"\n",
		"pkg/notifications/main.go")

	// Get the connections for the user
	ConnectionsLock.RLock()
	conns, ok := Connections[userID]
	ConnectionsLock.RUnlock()
	if !ok {
		config.Logs("info", "No connections for user: "+userID+"", "pkg/notifications/main.go")
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
	config.Logs("info",
		"\n Sending notification to all users\n"+
			"Message: "+string(message)+"\n",
		"pkg/notifications/main.go")

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
