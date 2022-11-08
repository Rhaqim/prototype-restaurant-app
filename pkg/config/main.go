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
	TRANSACTION = "transactions"
)

var (
	UserCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, USERS)
	SessionCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, SESSION)
	RestaurantCollection   = database.OpenCollection(database.ConnectMongoDB(), DB, RESTAURAUNT)
	HostCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, HOSTING)
	TransactionsCollection = database.OpenCollection(database.ConnectMongoDB(), DB, TRANSACTION)
)

var (
	JWTSecret        = os.Getenv("SECRET")
	JWTRefreshSecret = os.Getenv("REFRESH_SECRET")
)

// Log Messages
type LogType string

const (
	Error   LogType = "error"
	Info    LogType = "info"
	Warning LogType = "warning"
	Debug   LogType = "debug"
)

func Logs(level LogType, message interface{}) {
	switch level {
	case Info:
		log.Printf("INFO: \n %s ---> %s", time.Now(), message)
	case Error:
		log.Printf("ERROR: \n %s ---> %s", time.Now(), message)
	case Warning:
		log.Printf("WARNING: \n %s ---> %s", time.Now(), message)
	case Debug:
		log.Printf("DEBUG: \n %s ---> %s", time.Now(), message)
	default:
		log.Printf("INFO: \n %s ---> %s", time.Now(), message)
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
