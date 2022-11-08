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

func Logs(level LogType, message, funcName interface{}) {
	var timeFMT = time.Now().Format("2006-01-02 15:04:05")
	var strFuncNmae = funcName.(string)
	// var strMessage = message.(string)

	var clrInfoTime = coloriseInfo(timeFMT)
	var clrInfoFunc = coloriseInfo(strFuncNmae)
	// var clrInfoMessage = coloriseInfo(strMessage)

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

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
