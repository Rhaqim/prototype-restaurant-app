package helpers

import (
	"context"
	"log"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// JSON RESPONSE TO USERS
type JsonResponseType string

const (
	Error   JsonResponseType = "error"
	Success JsonResponseType = "success"
)

type MongoJsonResponse struct {
	Type    JsonResponseType `json:"type"`
	Data    interface{}      `json:"data"`
	Message string           `json:"message"`
	Date    time.Time        `json:"date"`
}

type Codes string // ERROR CODES

const (
	NotFound          Codes = "not_found"
	InsufficientFunds Codes = "insufficient_funds"
	SuccessCode       Codes = "success"
	AlreadyCompleted  Codes = "already_completed"
)

func SetError(err error, message string, funcName string) *MongoJsonResponse {
	if err != nil {
		config.Logs("error", err.Error()+" "+message, funcName)
		return &MongoJsonResponse{
			Type:    Error,
			Message: message + ", " + err.Error(),
			Date:    time.Now(),
		}
	}

	config.Logs("error", message, funcName)
	return &MongoJsonResponse{
		Type:    Error,
		Message: message,
		Date:    time.Now(),
	}
}

func SetInfo(message interface{}, funcName string) {
	config.Logs("info", message, funcName)
}

func SetSuccess(message string, data interface{}, funcName string) *MongoJsonResponse {
	config.Logs("info", message, funcName)
	if data == nil {
		return &MongoJsonResponse{
			Type:    Success,
			Message: message,
			Date:    time.Now(),
		}
	}
	return &MongoJsonResponse{
		Type:    Success,
		Data:    data,
		Message: message,
		Date:    time.Now(),
	}
}

func SetDebug(message string, funcName string) {
	config.Logs("debug", message, funcName)
	// panic("debug")
}

func SetWarning(message string, funcName string) {
	config.Logs("warning", message, funcName)
	log.Fatal(message)
}

// REMOVE PASSWORD FROM USER STRUCT
var PasswordOpts = options.FindOne().SetProjection(bson.M{"password": 0})

var usersCollection = database.OpenCollection(database.ConnectMongoDB(), config.DB, config.USERS)

// UPDATE REFRESH TOKEN
func UpdateRefreshToken(ctx context.Context, id primitive.ObjectID, refreshToken string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"refreshToken": refreshToken,
			"updated_at":   primitive.NewDateTimeFromTime(time.Now()),
		},
	}
	updateResult, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Println("updateResult: ", updateResult)
	return nil
}

// Validate Role, TxnType, TxnStatus enums
func RoleIsValid(role Roles) bool {
	switch role {
	case Admin, User, Business:
		return true
	}
	return false
}

func TxnTypeIsValid(TT TxnType) bool {
	switch TT {
	case Debit, Credit:
		return true
	}
	return false
}

func TxnStatusIsValid(TS TxnStatus) bool {
	switch TS {
	case TxnSuccess, TxnPending, TxnFail:
		return true
	}
	return false
}

// Get from DB
func GetFromDB(ctx context.Context, filter bson.M, collection *mongo.Collection) (bson.M, error) {
	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update in DB
func UpdateInDB(ctx context.Context, filter bson.M, update bson.M, collection *mongo.Collection) error {
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

// CreatedAt and UpdatedAt Helper
func CreatedAtUpdatedAt() (primitive.DateTime, primitive.DateTime) {
	now := primitive.NewDateTimeFromTime(time.Now())
	return now, now
}

// Work in progress
func (mr *MongoJsonResponse) Error(funcName string) *MongoJsonResponse {
	config.Logs("info", mr.Message, funcName)
	return &MongoJsonResponse{
		Type:    Error,
		Message: mr.Message,
		Date:    time.Now(),
	}
}
