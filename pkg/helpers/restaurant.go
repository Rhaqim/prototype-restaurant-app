package helpers

import (
	"context"
	"errors"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var restaurantCollection = config.RestaurantCollection

type Restaurant struct {
	ID        primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	OwnerID   primitive.ObjectID  `json:"owner_id,omitempty" bson:"owner_id,omitempty"`
	Name      string              `json:"name,omitempty" bson:"name" binding:"required"`
	Address   string              `json:"address,omitempty" bson:"address" binding:"required"`
	Phone     string              `json:"phone,omitempty" bson:"phone" binding:"required"`
	CreatedAt primitive.Timestamp `json:"created_at,omitempty" bson:"created_at" default:"now()"`
	UpdatedAt primitive.Timestamp `json:"updated_at,omitempty" bson:"updated_at" default:"now()"`
}

// Check if the Restaurant Belongs to the Signin User
func CheckRestaurantBelongsToUser(c context.Context, restaurantID primitive.ObjectID, user UserResponse) (bool, error) {
	var funcName = ut.GetFunctionName()
	var restaurant Restaurant

	restaurant, err := GetRestaurantByID(c, restaurantID)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return false, err
	}

	if restaurant.OwnerID != user.ID {
		return false, errors.New("Restaurant does not belong to the user")
	}

	return true, nil
}

// Get Restaurant By ID
func GetRestaurantByID(c context.Context, restaurantID primitive.ObjectID) (Restaurant, error) {

	var funcName = ut.GetFunctionName()
	var restaurant Restaurant

	filter := bson.M{"_id": restaurantID}

	err := restaurantCollection.FindOne(c, filter).Decode(&restaurant)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return restaurant, err
	}

	return restaurant, nil
}
