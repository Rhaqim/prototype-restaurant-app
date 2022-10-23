package helpers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStruct struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username"`
	Email    string             `bson:"email"`
	Password string             `bson:"password"`
}

type Roles string

const (
	Admin Roles = "admin"
	User  Roles = "user"
)

type CreatUser struct {
	Fullname      string             `json:"fullname"`
	Username      string             `json:"username"`
	Avatar        interface{}        `json:"avatar"`
	Email         string             `json:"email"`
	Password      string             `json:"password"`
	Social        interface{}        `json:"social"`
	Role          Roles              `json:"role"`
	RefreshToken  string             `json:"refreshToken,omitempty"`
	EmailVerified bool               `json:"emailConfirmed,omitempty"`
	CreatedAt     primitive.DateTime `json:"createdAt"`
	UpdatedAt     primitive.DateTime `json:"updatedAt"`
}

type GetUserById struct {
	ID primitive.ObjectID `json:"id"`
}

type GetUserByEmail struct {
	Email string `json:"email"`
}

type UpdateUserAvatar struct {
	ID        primitive.ObjectID `json:"id"`
	Avatar    string             `json:"avatar"`
	CreatedAt primitive.DateTime `json:"createdAt"`
	UpdatedAt primitive.DateTime `json:"updatedAt"`
}
