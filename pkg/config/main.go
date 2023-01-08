package config

import (
	"log"
	"os"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/database"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
)

// Time format
var (
	TimeFormat = time.Now().Format("15:04:05 02-01-2006")
)

// Server Port
var (
	ServerPort = os.Getenv("PORT")
)

// token expiration time
var (
	AccessTokenExpireTime  = time.Now().Add(time.Hour * 24)
	RefreshTokenExpireTime = time.Now().Add(time.Hour * 24 * 7)
)

// JWT Secret
var (
	JWTSecret        = os.Getenv("SECRET")
	JWTRefreshSecret = os.Getenv("REFRESH_SECRET")
)

// Database Collections
const (
	DB           = "thedutchapp"
	ADMIN        = "admins"
	ATTENDEE     = "attendees"
	BUDGET       = "budgets"
	CITY         = "city"
	COUNTRY      = "country"
	EVENT        = "events"
	FRIENDSHIP   = "friendship"
	NOTIFICATION = "notifications"
	ORDER        = "orders"
	PRODUCT      = "products"
	RESTAURAUNT  = "restaurants"
	SESSION      = "sessions"
	STATE        = "state"
	TRANSACTION  = "transactions"
	USERS        = "users"
	WALLET       = "wallets"
)

// Open Database Collections
var (
	AdminCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, ADMIN)
	AttendeeCollection     = database.OpenCollection(database.ConnectMongoDB(), DB, ATTENDEE)
	BudgetCollection       = database.OpenCollection(database.ConnectMongoDB(), DB, BUDGET)
	CityCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, CITY)
	CountryCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, COUNTRY)
	EventCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, EVENT)
	FriendshipCollection   = database.OpenCollection(database.ConnectMongoDB(), DB, FRIENDSHIP)
	NotificationCollection = database.OpenCollection(database.ConnectMongoDB(), DB, NOTIFICATION)
	OrderCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, ORDER)
	ProductCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, PRODUCT)
	RestaurantCollection   = database.OpenCollection(database.ConnectMongoDB(), DB, RESTAURAUNT)
	SessionCollection      = database.OpenCollection(database.ConnectMongoDB(), DB, SESSION)
	StateCollection        = database.OpenCollection(database.ConnectMongoDB(), DB, STATE)
	TransactionCollection  = database.OpenCollection(database.ConnectMongoDB(), DB, TRANSACTION)
	UserCollection         = database.OpenCollection(database.ConnectMongoDB(), DB, USERS)
	WalletCollection       = database.OpenCollection(database.ConnectMongoDB(), DB, WALLET)
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
	var timeFMT = TimeFormat
	var strFuncNmae = funcName.(string)

	var clrInfoTime = coloriseTime(timeFMT)
	var clrInfoFunc = coloriseFunc(strFuncNmae)

	var infoStr = coloriseInfo("[INFO]")
	var errorStr = coloriseError("[ERROR]")
	var warningStr = coloriseWarning("[WARNING]")
	var debugStr = coloriseDebug("[DEBUG]")

	var baseStr = " \n \n %s \n TIME:%s \n FUNC:%s \n MSG:%s \n \n"

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

func Colorise(message string, log LogType) string {
	switch log {
	case Info:
		return ut.Colorise("green", message)
	case Error:
		return ut.Colorise("red", message)
	case Warning:
		return ut.Colorise("yellow", message)
	case Debug:
		return ut.Colorise("blue", message)
	default:
		return ut.Colorise("green", message)
	}
}

// Types of messages sent over Websocket
const (
	Chat_         = "chat: "
	Invite_       = "invite: "
	Notification_ = "notification: "
	Order_        = "order: "
	Product_      = "product: "
	Reservation_  = "reservation: "
	Transaction_  = "transaction: "
)

// Context Timeout
const (
	ContextTimeout = 15 * time.Second
)

// Redis Keys
type CacheKey string

const (
	UserRole     CacheKey = "user_role:user"
	BusinessRole CacheKey = "user_role:business"
)

func (ck CacheKey) String() string {
	return string(ck)
}
