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

// USER STRUCT for Signing Out
type SignOut struct {
	Username string `json:"username"`
}

// USER STRUCT for sending refresh token
type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

// USER STRUCT for sending email to reset password
type ForgotPassword struct {
	Email string `json:"email"`
}

// USER STRUCT for resetting password
type ResetPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// USER STRUCT for updating password
type UpdatePassword struct {
	NewPassword string `json:"new_password"`
}

// USER VALIDATION EMAIL AND USERNAME

// CheckIfEmailExists checks if email exists in the database
// Accepts email string
// returns true if email exists
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

// CheckIfUsernameExists checks if username exists in the database
// Accepts username string
// returns true if username exists
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
