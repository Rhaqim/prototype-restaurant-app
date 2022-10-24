package helpers

import (
	"context"
	"log"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoJsonResponse struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Date    time.Time   `json:"date"`
}

var PasswordOpts = options.FindOne().SetProjection(bson.M{"password": 0})

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

func UpdateRefreshToken(ctx context.Context, id primitive.ObjectID, refreshToken string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"refreshToken": refreshToken,
			"updatedAt":    primitive.NewDateTimeFromTime(time.Now()),
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
	case Admin, User:
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
	case Success, Pending, Fail:
		return true
	}
	return false
}
