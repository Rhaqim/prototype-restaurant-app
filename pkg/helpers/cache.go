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

	SetDebug("Users being stored to cache: "+ut.ToJSON(userIDs), funcName)

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

	// check if users is empty
	// if len(users) == 0 {
	// 	SetInfo("No users in cache", funcName)

	// 	// Set users in cache
	// 	err = SetUserIDsCache(ctx, filter, key)
	// 	if err != nil {
	// 		SetError(err, "Error setting users in cache", funcName)
	// 		return nil, err
	// 	}

	// 	// Get users from cache
	// 	users, err = GetUserIDsFromCache(ctx, filter, key)
	// 	if err != nil {
	// 		SetError(err, "Error getting users from cache", funcName)
	// 		return nil, err
	// 	}

	// 	return users, nil
	// }

	return users, nil
}

/* Store all users in cache with username as key and other details as value, store under key: users */
func SetUsersCache(ctx context.Context) error {
	funcName := ut.GetFunctionName()

	SetInfo("Setting Users in cache", funcName)

	filter := bson.M{}

	key := "users"

	// loop through users and set each user in cache with username as key and other details as value
	users, err := GetUsers(ctx, filter)
	if err != nil {
		return err
	}

	for _, user := range users {
		redis := db.NewCache(
			user.Username,
			user,
		)

		// Clear cache before setting new data
		SetDebug("Clearing cache", funcName)
		err = redis.Delete()
		if err != nil {
			return err
		}

		SetDebug("Setting users in cache", funcName)
		err = redis.HMSet()
		if err != nil {
			return err
		}

		redis2 := db.NewCache(
			key,
			user.Username,
		)

		err = redis2.SAdd()
		if err != nil {
			return err
		}
	}

	return nil
}
