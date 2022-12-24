package config

import (
	"log"
	"os"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/database"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

const (
	DB          = "thedutchapp"
	USERS       = "users"
	SESSION     = "sessions"
	RESTAURAUNT = "restauraunts"
	EVENT       = "events"
	ATTENDEE    = "attendees"
	ORDER       = "orders"
	PRODUCT     = "products"
	TRANSACTION = "transactions"
	FRIENDSHIP  = "friendship"
	COUNTRY     = "country"
	STATE       = "state"
	CITY        = "city"
)

var (
	UserCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, USERS)
	SessionCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, SESSION)
	RestaurantCollection   = database.OpenCollection(database.ConnectMongoDB(), DB, RESTAURAUNT)
	EventCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, EVENT)
	AttendeeCollection     = database.OpenCollection(database.ConnectMongoDB(), DB, ATTENDEE)
	OrderCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, ORDER)
	ProductCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, PRODUCT)
	TransactionsCollection = database.OpenCollection(database.ConnectMongoDB(), DB, TRANSACTION)
	FriendshipCollection   = database.OpenCollection(database.ConnectMongoDB(), DB, FRIENDSHIP)
	CountryCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, COUNTRY)
	StateCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, STATE)
	CityCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, CITY)
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

func Logs(level LogType, message, funcName interface{}) {
	var timeFMT = time.Now().Format("15:04:05 02-01-2006")
	var strFuncNmae = funcName.(string)

	var clrInfoTime = coloriseTime(timeFMT)
	var clrInfoFunc = coloriseFunc(strFuncNmae)

	var infoStr = coloriseInfo("[INFO]")
	var errorStr = coloriseError("[ERROR]")
	var warningStr = coloriseWarning("[WARNING]")
	var debugStr = coloriseDebug("[DEBUG]")

	var baseStr = " \n %s \n TIME:%s \n FUNC:%s \n MSG:%s \n"

	switch level {
	case Info:
		log.Printf(baseStr, infoStr, clrInfoTime, clrInfoFunc, coloriseInfo(message.(string)))
	case Error:
		log.Printf(baseStr, errorStr, clrInfoTime, clrInfoFunc, coloriseError(message.(string)))
	case Warning:
		log.Printf(baseStr, warningStr, clrInfoTime, clrInfoFunc, coloriseWarning(message.(string)))
	case Debug:
		log.Printf(baseStr, debugStr, clrInfoTime, clrInfoFunc, coloriseDebug(message.(string)))
	default:
		log.Printf(baseStr, infoStr, clrInfoTime, clrInfoFunc, coloriseInfo(message.(string)))

	}
}

func coloriseInfo(message string) string {
	return ut.Colorise("green", message)
}

func coloriseError(message string) string {
	return ut.Colorise("red", message)
}

func coloriseWarning(message string) string {
	return ut.Colorise("yellow", message)
}

func coloriseDebug(message string) string {
	return ut.Colorise("blue", message)
}

func coloriseTime(message string) string {
	return ut.Colorise("cyan", message)
}

func coloriseFunc(message string) string {
	return ut.Colorise("magenta", message)
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
