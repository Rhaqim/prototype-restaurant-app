package helpers

import (
	"context"

	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// USER STRUCT for Signing In
type SignIn struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" binding:"required"`
}

type SignOut struct {
	Username string `json:"username"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPassword struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	RefreshToken string `json:"refresh_token"`
	OldPassword  string `json:"old_password"`
	NewPassword  string `json:"new_password"`
}

// USER VALIDATION EMAIL AND USERNAME
func CheckIfEmailExists(email string) (bool, error) {
	var user UserStruct
	filter := bson.M{"email": email}
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		SetDebug(err.Error(), ut.GetFunctionName())
		return false, err
	}
	return true, nil
}

func CheckIfUsernameExists(username string) (bool, error) {
	var user UserStruct
	err := usersCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		SetDebug(err.Error(), ut.GetFunctionName())
		return false, err
	}
	if user.Username == username {
		return true, nil
	}
	return false, nil
}
