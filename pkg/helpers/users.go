package helpers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/database"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Roles string

const (
	Admin    Roles = "admin"
	User     Roles = "user"
	Business Roles = "business"
)

type UserStruct struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`

	FirstName     string               `json:"firstName" bson:"firstName" binding:"required"`
	LastName      string               `json:"lastName" bson:"lastName" binding:"required"`
	Email         string               `bson:"email" json:"email" binding:"required,email"`
	Username      string               `bson:"username" json:"username" binding:"required"`
	Password      string               `bson:"password" json:"password" binding:"required,min=8,max=32,alphanum"`
	Avatar        interface{}          `bson:"avatar" json:"avatar"`
	Social        interface{}          `bson:"social" json:"social"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends"`
	Location      string               `bson:"location" json:"location"`
	Wallet        float64              `bson:"wallet" json:"wallet"`
	Account       BankAccount          `bson:"account" json:"account"`
	Transactions  []Transactions       `bson:"transactions" json:"transactions"`
	RefreshToken  string               `bson:"refreshToken,omitempty" json:"refreshToken,omitempty"`
	EmailVerified bool                 `bson:"emailConfirmed,omitempty" json:"emailConfirmed,omitempty" default:"false"`
	Role          Roles                `bson:"role" json:"role" default:"user"`
	CreatedAt     primitive.DateTime   `bson:"createdAt" json:"createdAt" default:"Now()"`
	UpdatedAt     primitive.DateTime   `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}

type UserResponse struct {
	ID            primitive.ObjectID   `bson:"_id" json:"_id,omitempty"`
	FirstName     string               `json:"firstName"`
	LastName      string               `json:"lastName"`
	Email         string               `bson:"email" json:"email"`
	Username      string               `bson:"username" json:"username"`
	Avatar        interface{}          `bson:"avatar" json:"avatar"`
	Social        interface{}          `bson:"social" json:"social"`
	Friends       []primitive.ObjectID `bson:"friends" json:"friends"`
	Location      string               `bson:"location" json:"location"`
	Wallet        float64              `bson:"wallet" json:"wallet"`
	Account       BankAccount          `bson:"account" json:"account"`
	Transactions  []Transactions       `bson:"transactions" json:"transactions"`
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

/*Get user data by:
- ID
- Email
- Username
- From token
*/
// Get user by ID
func GetUserByID(ctx context.Context, userID primitive.ObjectID) UserResponse {
	var user UserResponse
	config.Logs("info", "User ID: "+userID.Hex(), ut.GetFunctionName())

	filter := bson.M{"_id": userID}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error(), ut.GetFunctionName())
		return UserResponse{}
	}

	return user
}

// Get user by Email
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

// Get user by Username
func GetUserByUsername(username string) UserResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer database.ConnectMongoDB().Disconnect(context.TODO())

	var user UserResponse
	log.Print("Request ID sent by client:", username)

	config.Logs("info", "Username: "+username, ut.GetFunctionName())

	filter := bson.M{"username": username}
	if err := usersCollection.FindOne(ctx, filter).Decode(&user); err != nil {
		config.Logs("error", err.Error(), ut.GetFunctionName())
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
