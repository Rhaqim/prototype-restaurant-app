package helpers

import (
	"context"

	db "github.com/Rhaqim/thedutchapp/pkg/cache"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// Fetches all users with filter and stores them in redis cache
// Accepts a context, filter and cache key
// Returns an error
func SetUserIDsCache(ctx context.Context, filter bson.M, key config.CacheKey) error {
	funcName := ut.GetFunctionName()

	SetInfo("Fetching users", funcName)

	// Get all users
	users, err := GetUsers(ctx, filter)
	if err != nil {
		SetError(err, "Error getting users", funcName)
		return err
	}

	var userIDs []string
	for _, user := range users {
		userIDs = append(userIDs, user.ID.Hex())
	}

	SetDebug("Users being stored to cache: "+ut.ToJsonString(userIDs), funcName)

	// Store userIDs in redis cache
	redis := db.NewCache(
		key.String(),
		userIDs,
	)

	// Clear cache before setting new data
	SetDebug("Clearing cache", funcName)
	err = redis.Delete()
	if err != nil {
		SetError(err, "Error clearing cache", funcName)
		return err
	}

	SetDebug("Setting users in cache", funcName)
	err = redis.SetList()
	if err != nil {
		SetError(err, "Error setting users in cache", funcName)
		return err
	}

	return nil
}

// Fetches from Redis cache and returns all users
// Accepts a context and cache key
// Returns a slice of users and an error
func GetUserIDsFromCache(ctx context.Context, filter bson.M, key config.CacheKey) ([]string, error) {
	funcName := ut.GetFunctionName()

	SetInfo("Fetching users from cache", funcName)

	// Get users from redis cache
	redis := db.NewCache(
		key.String(),
		nil,
	)

	// Update cache
	err := SetUserIDsCache(ctx, filter, key)
	if err != nil {
		SetError(err, "Error setting users in cache", funcName)
		return nil, err
	}

	users, err := redis.GetList()
	if err != nil {
		// If error, fetch from database
		SetError(err, "Error getting users from cache fetching from Database", funcName)

		users, err := GetUsers(ctx, filter)
		if err != nil {
			SetError(err, "Error getting users", funcName)
			return nil, err
		}

		var userIDs []string
		for _, user := range users {
			userIDs = append(userIDs, user.ID.Hex())
		}

		return userIDs, nil
	}

	return users, nil
}

/* Store all users in cache with username as key and other details as value, store under key: users */
func SetUsersCache(ctx context.Context) error {
	funcName := ut.GetFunctionName()

	// Get all users
	users, err := GetUsers(ctx, bson.M{})
	if err != nil {
		return err
	}

	// Store users in redis cache
	redis := db.NewCache(
		config.Users.String(),
		ut.ToJSON(users),
	)

	// Clear cache before setting new data
	err = redis.Delete()
	if err != nil {
		SetError(err, "Error clearing cache", funcName)
		return err
	}

	err = redis.Set()
	if err != nil {
		SetError(err, "Error setting users in cache", funcName)
		return err
	}

	return nil
}

// Fetches from Redis cache and returns all users
// Accepts a context and cache key
// Returns a slice of users and an error
func GetUsersFromCache(ctx context.Context) ([]UserResponse, error) {
	funcName := ut.GetFunctionName()

	// Get users from redis cache
	redis := db.NewCache(
		config.Users.String(),
		nil,
	)

	users, err := redis.Get()
	if err != nil {
		// If error, fetch from database
		SetError(err, "Error getting users from cache fetching from Database", funcName)

		users, err := GetUsers(ctx, bson.M{})
		if err != nil {
			SetError(err, "Error getting users", funcName)
			return nil, err
		}

		return users, nil
	}

	// convert users from json to slice of UserStruct
	var usersList []UserResponse
	ut.FromJSON(users, &usersList)

	return usersList, nil

}

// Set Restaurants in cache
func SetRestaurantsCache(ctx context.Context) error {

	funcName := ut.GetFunctionName()

	// Get all restaurants
	restaurants, err := GetRestaurants(ctx, bson.M{})
	if err != nil {
		return err
	}

	// Store restaurants in redis cache
	redis := db.NewCache(
		config.Restaurants.String(),
		ut.ToJSON(restaurants),
	)

	// Clear cache before setting new data
	err = redis.Delete()
	if err != nil {
		SetError(err, "Error clearing cache", funcName)
		return err
	}

	err = redis.Set()
	if err != nil {
		SetError(err, "Error setting restaurants in cache", funcName)
		return err
	}

	return nil
}

// Fetches from Redis cache and returns all restaurants
// Accepts a context and cache key
// Returns a slice of restaurants and an error
func GetRestaurantsFromCache(ctx context.Context) ([]Restaurant, error) {
	funcName := ut.GetFunctionName()

	// Get restaurants from redis cache
	redis := db.NewCache(
		config.Restaurants.String(),
		nil,
	)

	restaurants, err := redis.Get()
	if err != nil {
		// If error, fetch from database
		SetError(err, "Error getting restaurants from cache fetching from Database", funcName)

		restaurants, err := GetRestaurants(ctx, bson.M{})
		if err != nil {
			SetError(err, "Error getting restaurants", funcName)
			return nil, err
		}

		return restaurants, nil
	}

	// convert restaurants from json to slice of RestaurantStruct
	var restaurantsList []Restaurant
	ut.FromJSON(restaurants, &restaurantsList)

	return restaurantsList, nil
}
