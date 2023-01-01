package helpers

import (
	"context"
	"errors"
	"sync"

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
	Phone     string              `json:"phone,omitempty" bson:"phone" binding:"required"`
	Email     string              `json:"email,omitempty" bson:"email" binding:"required"`
	Address   string              `json:"address,omitempty" bson:"address" binding:"required"`
	OpenHours [7]OpenHours        `json:"open_hours,omitempty" bson:"open_hours" binding:"required"`
	Currency  string              `json:"currency,omitempty" bson:"currency" binding:"required"`
	Verified  bool                `json:"verified,omitempty" bson:"verified"`
	CreatedAt primitive.Timestamp `json:"created_at,omitempty" bson:"created_at" default:"time.Now()"`
	UpdatedAt primitive.Timestamp `json:"updated_at,omitempty" bson:"updated_at" default:"time.Now()"`
}

type RestaurantCategory string

const (
	// Restaurant Categories
	Bar    RestaurantCategory = "Bar"
	Lounge RestaurantCategory = "Lounge"
	Cafe   RestaurantCategory = "Cafe"
)

func (rc RestaurantCategory) String() string {
	switch rc {
	case Bar:
		return "Bar"
	case Lounge:
		return "Lounge"
	case Cafe:
		return "Cafe"
	default:
		return "Bar"
	}
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

func GetRestaurant(c context.Context, filter bson.M) (Restaurant, error) {

	var funcName = ut.GetFunctionName()
	var restaurant Restaurant

	err := restaurantCollection.FindOne(c, filter).Decode(&restaurant)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return restaurant, err
	}

	return restaurant, nil

}

// Get Restaurant By ID
func GetRestaurantByID(c context.Context, restaurantID primitive.ObjectID) (Restaurant, error) {

	var funcName = ut.GetFunctionName()

	filter := bson.M{"_id": restaurantID}

	restaurant, err := GetRestaurant(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return restaurant, err
	}

	return restaurant, nil
}

// Check if Restaurant Exists
func CheckRestaurantExists(c context.Context, filter bson.M) (bool, error) {

	var funcName = ut.GetFunctionName()

	_, err := GetRestaurant(c, filter)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		SetDebug(err.Error(), funcName)
		return false, err
	}

	return true, nil
}

// Validate create Restaurant request
func ValidateCreateRestaurantRequest(c context.Context, request Restaurant) (bool, error) {

	var funcName = ut.GetFunctionName()

	// create a wait group
	var wg sync.WaitGroup

	errChan := make(chan error, 3)

	// Check if Restaurant Exists
	nameFilter := bson.M{"name": request.Name}
	emailFilter := bson.M{"email": request.Email}
	phoneFilter := bson.M{"phone": request.Phone}

	wg.Add(3)
	go func() {
		defer wg.Done()
		exists, err := CheckRestaurantExists(c, nameFilter)
		if err != nil {
			errChan <- err
			return
		}
		if exists {
			errChan <- errors.New("Restaurant name already exists")
			return
		}
	}()

	go func() {
		defer wg.Done()
		exists, err := CheckRestaurantExists(c, emailFilter)
		if err != nil {
			errChan <- err
			return
		}
		if exists {
			errChan <- errors.New("Restaurant email already exists")
			return
		}
	}()

	go func() {
		defer wg.Done()
		exists, err := CheckRestaurantExists(c, phoneFilter)
		if err != nil {
			errChan <- err
			return
		}
		if exists {
			errChan <- errors.New("Restaurant phone already exists")
			return
		}
	}()

	// wait for all go routines to finish
	wg.Wait()

	// close the channel
	close(errChan)

	// check if there are any errors
	for err := range errChan {
		SetDebug(err.Error(), funcName)
		if err.Error() == "mongo: no documents in result" {
			continue
		}
		return false, err
	}

	return true, nil
}
