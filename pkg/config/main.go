package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/database"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

// Time format
var (
	TimeFormat = time.Now().Format("15:04:05 02-01-2006")
)

// Load Environment Variables
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Environment Variables
var (
	ServerPort       = os.Getenv("PORT")
	GoogleMapsAPIKey = os.Getenv("GOOGLE_MAPS_API_KEY")
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

// Open Collections
func OpenCollection(name string) *mongo.Collection {
	return database.OpenCollection(database.ConnectMongoDB(), DB, name)
}

// Open Database Collections
var (
	AdminCollection        = OpenCollection(ADMIN)
	AttendeeCollection     = OpenCollection(ATTENDEE)
	BudgetCollection       = OpenCollection(BUDGET)
	CityCollection         = OpenCollection(CITY)
	CountryCollection      = OpenCollection(COUNTRY)
	EventCollection        = OpenCollection(EVENT)
	FriendshipCollection   = OpenCollection(FRIENDSHIP)
	NotificationCollection = OpenCollection(NOTIFICATION)
	OrderCollection        = OpenCollection(ORDER)
	ProductCollection      = OpenCollection(PRODUCT)
	RestaurantCollection   = OpenCollection(RESTAURAUNT)
	SessionCollection      = OpenCollection(SESSION)
	StateCollection        = OpenCollection(STATE)
	TransactionCollection  = OpenCollection(TRANSACTION)
	UserCollection         = OpenCollection(USERS)
	WalletCollection       = OpenCollection(WALLET)
)

/* LOG MESSAGES */

type LogType string

const (
	Error   LogType = "error"
	Info    LogType = "info"
	Warning LogType = "warning"
	Debug   LogType = "debug"
)

var colorFuncs = map[LogType]func(string) string{
	Error:   coloriseError,
	Info:    coloriseInfo,
	Warning: coloriseWarning,
	Debug:   coloriseDebug,
}

func Logs(level LogType, message, funcName interface{}) {
	colorFunc := colorFuncs[level]
	if colorFunc == nil {
		colorFunc = colorFuncs[Info]
	}

	level = LogType(colorFunc(strings.ToUpper(string(level))))
	var timeFMT = TimeFormat
	var strFuncNmae = funcName.(string)

	var clrInfoTime = coloriseTime(timeFMT)
	var clrInfoFunc = coloriseFunc(strFuncNmae)

	var baseStr = "\n \n" +
		"######################################### \n" +
		" \n LEVEL:%s \n TIME:%s \n FUNC:%s \n MSG:%s \n" +
		"\n #########################################" +
		"\n \n"

	log.Printf(baseStr, level, clrInfoTime, clrInfoFunc, colorFunc(message.(string)))
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

type NotificationMessage string

const (
	BudgetReturned NotificationMessage = "Budget has been returned to your wallet"
	OrderCancelled NotificationMessage = "Order has been cancelled"
	OrderCreated   NotificationMessage = "Order has been created"
	OrderUpdated   NotificationMessage = "Order has been updated"
	OrderPaid      NotificationMessage = "Order has been paid"
	OrderRefunded  NotificationMessage = "Order has been refunded"
	EventCancelled NotificationMessage = "Event has been cancelled"
	EventCreated   NotificationMessage = "Event has been created"
	EventUpdated   NotificationMessage = "Event has been updated"
)

func (nm NotificationMessage) String() string {
	return string(nm)
}

// Context Timeout
const (
	ContextTimeout = 15 * time.Second
)

// Redis Keys
type CacheKey string

const (
	UserRole     CacheKey = "user_role_user"
	BusinessRole CacheKey = "user_role_business"
	AdminRole    CacheKey = "user_role_admin"
	Users        CacheKey = "users"
	Restaurants  CacheKey = "restaurants"
)

func (ck CacheKey) String() string {
	return string(ck)
}
