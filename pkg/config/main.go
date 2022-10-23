package config

import (
	"log"
	"os"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/database"
)

const (
	DB          = "thedutchapp"
	USERS       = "users"
	SESSION     = "sessions"
	RESTAURAUNT = "restauraunts"
	HOSTING     = "hosting"
)

var (
	UserCollection       = database.OpenCollection(database.ConnectMongoDB(), DB, USERS)
	SessionCollection    = database.OpenCollection(database.ConnectMongoDB(), DB, SESSION)
	RestaurantCollection = database.OpenCollection(database.ConnectMongoDB(), DB, RESTAURAUNT)
	HostCollection       = database.OpenCollection(database.ConnectMongoDB(), DB, HOSTING)
)

var (
	JWTSecret = os.Getenv("SECRET")
)

// Log Messages
func Logs(level string, message interface{}) {
	switch level {
	case "info":
		log.Printf("INFO: %s --> %s", time.Now(), message)
	case "error":
		log.Printf("ERROR: %s --> %s", time.Now(), message)
	default:
		log.Printf("INFO: %s --> %s", time.Now(), message)
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
