package helpers

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var PasswordOpts = options.FindOne().SetProjection(bson.M{"password": 0})

type SignIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignOut struct {
	Username string `json:"username"`
}

type RefreshToken struct {
	ID           primitive.ObjectID `json:"id"`
	RefreshToken string             `json:"refresh_token"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	ID           primitive.ObjectID `json:"id"`
	Email        string             `json:"email"`
	RefreshToken string             `json:"refresh_token"`
	OldPassword  string             `json:"old_password"`
	NewPassword  string             `json:"new_password"`
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
