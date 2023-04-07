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

// JSON response types for the user
type JsonResponseType string

const (
	Error   JsonResponseType = "error"
	Success JsonResponseType = "success"
)

// JSON RESPONSE STRUCT
// All responses to the user will be in this format
// The type is either error or success
// The data is the data that is returned to the user
// The message is the message that is returned to the user
type MongoJsonResponse struct {
	Type    JsonResponseType `json:"type"`
	Data    interface{}      `json:"data"`
	Message string           `json:"message"`
	Date    time.Time        `json:"date"`
}

// Error codes for edge cases
type Codes string // ERROR CODES

const (
	NotFound          Codes = "not_found"
	InsufficientFunds Codes = "insufficient_funds"
	SuccessCode       Codes = "success"
	AlreadyCompleted  Codes = "already_completed"
)

type Address struct {
	HouseNumber string `json:"house_number"`
	Street      string `json:"street" binding:"required"`
	City        string `json:"city"`
	State       string `json:"state"`
	Zipcode     string `json:"zip_code"`
	// CountryCode string `json:"country_code" binding:"required,iso_3166_1_alpha_2"`
	CountryCode string `json:"country_code" binding:"required"`
}

type MapInfo struct {
	Lat     float64 `json:"lat,omitempty"`
	Long    float64 `json:"long,omitempty"`
	PlaceID string  `json:"place_id,omitempty"`
}

// SET ERROR
// Uses the struct MongoJsonResponse
// Accepts an error, a message and the function name
// Returns the struct MongoJsonResponse
func SetError(err error, message string, funcName string) MongoJsonResponse {
	if err != nil {
		config.Logs("error", err.Error()+" "+message, funcName)
		return MongoJsonResponse{
			Type:    Error,
			Message: message + ", " + err.Error(),
			Date:    time.Now(),
		}
	}

	config.Logs("error", message, funcName)
	return MongoJsonResponse{
		Type:    Error,
		Message: message,
		Date:    time.Now(),
	}
}

// SET INFO
func SetInfo(message interface{}, funcName string) {
	config.Logs("info", message, funcName)
}

// SET SUCCESS
// Uses the struct MongoJsonResponse
// Accepts a message, the data and the function name
// Returns the struct MongoJsonResponse
func SetSuccess(message string, data interface{}, funcName string) MongoJsonResponse {
	config.Logs("info", message, funcName)
	if data == nil {
		return MongoJsonResponse{
			Type:    Success,
			Message: message,
			Date:    time.Now(),
		}
	}
	return MongoJsonResponse{
		Type:    Success,
		Data:    data,
		Message: message,
		Date:    time.Now(),
	}
}

// SET DEBUG
func SetDebug(message string, funcName string) {
	config.Logs("debug", message, funcName)
	// panic("debug")
}

// SET WARNING
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
	_, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// String returns the string representation of the Roles
func (r Roles) String() string {
	switch r {
	case Admin:
		return "admin"
	case User:
		return "user"
	case Business:
		return "business"
	default:
		return "user"
	}
}

func TxnTypeIsValid(TT TxnType) bool {
	switch TT {
	case Debit, Credit:
		return true
	}
	return false
}

func (tt TxnType) String() string {
	switch tt {
	case Debit:
		return "debit"
	case Credit:
		return "credit"
	default:
		return "debit"
	}
}

func TxnStatusIsValid(TS TxnStatus) bool {
	switch TS {
	case TxnSuccess, TxnPending, TxnFail:
		return true
	}
	return false
}

func (ts TxnStatus) String() string {
	switch ts {
	case TxnSuccess:
		return "success"
	case TxnPending:
		return "pending"
	case TxnFail:
		return "fail"
	default:
		return "pending"
	}
}

// Get from DB
func GetFromDB(ctx context.Context, filter bson.M, collection *mongo.Collection, result interface{}) (interface{}, error) {
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

// implement search feature in mongodb using golang, use contains function in mongodb
func Search(ctx context.Context, collection *mongo.Collection, query, search string) ([]*interface{}, error) {
	filter := bson.M{
		search: bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var obj []*interface{}
	if err = cursor.All(ctx, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}
