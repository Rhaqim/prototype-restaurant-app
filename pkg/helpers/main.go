package helpers

import (
	"context"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoJsonResponse struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Date    time.Time   `json:"date"`
}

var usersCollection = database.OpenCollection(database.ConnectMongoDB(), config.DB, config.USERS)

func CheckIfEmailExists(email string) (bool, error) {
	var user UserStruct
	filter := bson.M{"email": email}
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CheckIfUsernameExists(username string) (bool, error) {
	var user UserStruct
	err := usersCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		return false, err
	}
	return true, nil
}
