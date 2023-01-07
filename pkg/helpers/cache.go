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

	SetDebug("Users: "+ut.ToJSON(userIDs), funcName)

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

	// Set users in cache
	err := SetUserIDsCache(ctx, filter, key)
	if err != nil {
		SetError(err, "Error setting users in cache", funcName)
		return nil, err
	}

	users, err := redis.GetList()
	if err != nil {
		SetError(err, "Error getting users from cache", funcName)
		return nil, err
	}

	// check if users is empty
	if len(users) == 0 {
		SetInfo("No users in cache", funcName)
	}

	return users, nil
}
