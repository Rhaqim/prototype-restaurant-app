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

func SendNotification(user_ID primitive.ObjectID, message []byte) {
	userID := user_ID.Hex()
	config.Logs("info",
		"Sending notification to user: "+user_ID.Hex()+"\n"+
			"Message: "+string(message)+"\n"+
			"Number of connections: "+fmt.Sprintf("%d", len(Connections[userID]))+"\n",
		"pkg/notifications/main.go")

	ConnectionsLock.RLock()
	conns, ok := Connections[userID]
	ConnectionsLock.RUnlock()
	if !ok {
		config.Logs("info", "No connections for user: "+userID+"", "pkg/notifications/main.go")
		return
	}

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// Remove the connection if it is no longer usable
			conn.Close()
		}
	}
}

func WsHandler(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Handle error
		config.Logs("error", "Error upgrading HTTP connection to WebSocket connection: "+err.Error()+"", "pkg/notifications/main.go")
		return
	}
	defer conn.Close()

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
			break
		}
	}
}
