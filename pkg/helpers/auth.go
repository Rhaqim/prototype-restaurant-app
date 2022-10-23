package helpers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
