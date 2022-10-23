package helpers

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SignIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
