package helpers

import (
	"context"
	"log"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Roles string

const (
	Admin Roles = "admin"
	User  Roles = "user"
)

type UserStruct struct {
	ID            primitive.ObjectID   `bson:"_id" json:"_id,omitempty"`
	Fullname      string               `bson:"fullname" json:"fullname"`
	Username      string               `bson:"username" json:"username"`
	Avatar        interface{}          `bson:"avatar" json:"avatar"`
	Email         string               `bson:"email" json:"email"`
	Password      string               `bson:"password" json:"password"`
	Social        interface{}          `bson:"social" json:"social"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends"`
	Location      primitive.ObjectID   `bson:"location" json:"location"`
	Wallet        float32              `bson:"wallet" json:"wallet"`
	Transactions  []primitive.ObjectID `bson:"transactions" json:"transactions"`
	RefreshToken  string               `bson:"refreshToken,omitempty" json:"refreshToken,omitempty"`
	EmailVerified bool                 `bson:"emailConfirmed,omitempty" json:"emailConfirmed,omitempty" default:"false"`
	Role          Roles                `bson:"role" json:"role"`
	CreatedAt     primitive.DateTime   `bson:"createdAt" json:"createdAt"`
	UpdatedAt     primitive.DateTime   `bson:"updatedAt" json:"updatedAt"`
}

type UserResponse struct {
	ID            primitive.ObjectID   `bson:"_id" json:"_id,omitempty"`
	Fullname      string               `bson:"fullname" json:"fullname"`
	Username      string               `bson:"username" json:"username"`
	Avatar        interface{}          `bson:"avatar" json:"avatar"`
	Email         string               `bson:"email" json:"email"`
	Social        interface{}          `bson:"social" json:"social"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends"`
	Location      primitive.ObjectID   `bson:"location" json:"location"`
	Wallet        float64              `bson:"wallet" json:"wallet"`
	Transactions  []primitive.ObjectID `bson:"transactions" json:"transactions"`
	RefreshToken  string               `bson:"refreshToken,omitempty" json:"refreshToken,omitempty"`
	EmailVerified bool                 `bson:"emailConfirmed,omitempty" json:"emailConfirmed,omitempty" default:"false"`
	Role          Roles                `bson:"role" json:"role"`
	CreatedAt     primitive.DateTime   `bson:"createdAt" json:"createdAt"`
	UpdatedAt     primitive.DateTime   `bson:"updatedAt" json:"updatedAt"`
}

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

type GetUserByEmailStruct struct {
	Email string `json:"email"`
}

type UpdateUserAvatar struct {
	ID        primitive.ObjectID `json:"id"`
	Avatar    string             `json:"avatar"`
	CreatedAt primitive.DateTime `json:"createdAt"`
	UpdatedAt primitive.DateTime `json:"updatedAt"`
}

func GetUserByID(userID primitive.ObjectID) UserResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user UserResponse
	config.Logs("info", "User ID: "+userID.Hex(), ut.GetFunctionName())

	filter := bson.M{"_id": userID}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error(), ut.GetFunctionName())
		return UserResponse{}
	}

	return user
}

func GetUserByEmail(email string) UserResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user UserResponse
	log.Print("Request ID sent by client:", email)

	config.Logs("info", "Email: "+email, ut.GetFunctionName())

	filter := bson.M{"email": email}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error(), ut.GetFunctionName())
		return UserResponse{}
	}

	return user
}
