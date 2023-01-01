package helpers

import (
	"context"
	"errors"

	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Roles string

const (
	Admin    Roles = "admin"
	User     Roles = "user"
	Business Roles = "business"
)

type UserStruct struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName     string               `json:"first_name" bson:"first_name" binding:"required"`
	LastName      string               `json:"last_name" bson:"last_name" binding:"required"`
	Email         string               `bson:"email" json:"email" binding:"required,email"`
	Username      string               `bson:"username" json:"username" binding:"required"`
	Password      string               `bson:"password" json:"password" binding:"required,min=8,max=32,alphanum"`
	Avatar        Avatar               `bson:"avatar" json:"avatar" default:"{}"`
	Social        SocialNetwork        `bson:"social" json:"social" default:"{}"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends" default:"[]"`
	Location      string               `bson:"location" json:"location"`
	Wallet        primitive.ObjectID   `bson:"wallet" json:"wallet" default:"null"`
	Account       BankAccount          `bson:"account" json:"account" default:"{}"`
	Transactions  []Transactions       `bson:"transactions" json:"transactions" default:"[]"`
	RefreshToken  string               `bson:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	EmailVerified bool                 `bson:"email_confirmed,omitempty" json:"email_confirmed,omitempty" default:"false"`
	Role          Roles                `bson:"role" json:"role" default:"user"`
	CreatedAt     primitive.DateTime   `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt     primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
}

type Avatar struct {
	Alt string `json:"alt,omitempty" bson:"alt,omitempty"`
	URL string `json:"url,omitempty" bson:"url,omitempty"`
}

type SocialNetwork struct {
	Network string `json:"network,omitempty" bson:"network,omitempty"`
	Link    string `json:"link,omitempty" bson:"link,omitempty"`
}

type UserResponse struct {
	ID            primitive.ObjectID   `bson:"_id" json:"_id,omitempty"`
	FirstName     string               `json:"first_name"`
	LastName      string               `json:"last_name"`
	Email         string               `bson:"email" json:"email"`
	Username      string               `bson:"username" json:"username"`
	Avatar        interface{}          `bson:"avatar" json:"avatar"`
	Social        interface{}          `bson:"social" json:"social"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends"`
	Location      string               `bson:"location" json:"location"`
	Wallet        primitive.ObjectID   `bson:"wallet" json:"wallet"`
	Account       BankAccount          `bson:"account" json:"account"`
	Transactions  []Transactions       `bson:"transactions" json:"transactions"`
	RefreshToken  string               `bson:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	EmailVerified bool                 `bson:"email_confirmed,omitempty" json:"email_confirmed,omitempty" default:"false"`
	Role          Roles                `bson:"role" json:"role"`
	CreatedAt     primitive.DateTime   `bson:"created_at" json:"created_at"`
	UpdatedAt     primitive.DateTime   `bson:"updated_at" json:"updated_at"`
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
	CreatedAt     primitive.DateTime `json:"created_at"`
	UpdatedAt     primitive.DateTime `json:"updated_at"`
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
	CreatedAt primitive.DateTime `json:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at"`
}

/*
Get user data by:
- ID
- Email
- Username
- From token
*/
func GetUser(ctx context.Context, filter bson.M) (UserResponse, error) {
	var user UserResponse
	funcName := ut.GetFunctionName()

	opts := options.FindOne().SetProjection(bson.M{"password": 0, "refreshToken": 0})

	if err := usersCollection.FindOne(ctx, filter, opts).Decode(&user); err != nil {
		SetError(err, "error", funcName)
		return UserResponse{}, err
	}

	return user, nil
}

func GetUserAllInfo(ctx context.Context, filter bson.M) (UserResponse, error) {
	var user UserResponse
	funcName := ut.GetFunctionName()

	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		SetError(err, "error", funcName)
		return UserResponse{}, err
	}

	return user, nil
}

// Get user by ID
func GetUserByID(ctx context.Context, userID primitive.ObjectID) UserResponse {
	filter := bson.M{"_id": userID}

	user, err := GetUser(ctx, filter)
	if err != nil {
		return UserResponse{}
	}

	return user
}

// Get user by Email
func GetUserByEmail(ctx context.Context, email string) UserResponse {
	filter := bson.M{"email": email}

	user, err := GetUser(ctx, filter)
	if err != nil {
		return UserResponse{}
	}

	return user
}

// Get user by Username
func GetUserByUsername(ctx context.Context, username string) UserResponse {

	filter := bson.M{"username": username}

	user, err := GetUser(ctx, filter)
	if err != nil {
		return UserResponse{}
	}

	return user
}

// Get user from token
func GetUserFromToken(c *gin.Context) (UserResponse, error) {
	check, ok := c.Get("user") // Check if user is logged in
	if !ok {
		return UserResponse{}, errors.New("Unauthorized")
	}

	user := check.(UserResponse)

	return user, nil
}
