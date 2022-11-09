package helpers

import (
	"context"
	"errors"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	Friendship

This struct is used to represent a friendship between two users.
It is used to store the friendship in the database.
*/
type FriendshipStatus int

const (
	// FriendshipStatusPending means that the friendship is pending.
	FriendshipStatusPending FriendshipStatus = iota
	// FriendshipStatusAccepted means that the friendship is accepted.
	FriendshipStatusAccepted
	// FriendshipStatusBlocked means that the friendship is blocked.
	FriendshipStatusBlocked
)

type Friendship struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID  primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	Status    FriendshipStatus   `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at" default:"now()"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at" default:"now()"`
}

type FriendshipRequest struct {
	FriendID  primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	Status    FriendshipStatus   `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at" default:"now()"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at" default:"now()"`
}

type FriendshipAcceptRequest struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID primitive.ObjectID `json:"friend_id" bson:"friend_id"`
}

// Send Friend Request
// Send a friend request to another user.
func SendFriendRequest(ctx context.Context, userID UserResponse, friendID primitive.ObjectID) (Friendship, error) {
	// Check if the friendship already exists.
	friendship := VerifyFriends(userID, friendID)
	// If the friendship already exists, return a message.
	if friendship {
		return Friendship{}, errors.New("Friendship already exists")
	}

	// Create a new friendship request.
	friendshipRequest := bson.M{
		"user_id":    userID.ID,
		"friend_id":  friendID,
		"status":     FriendshipStatusPending,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	// Insert the friendship request into the database.
	insertUpdate, err := config.FriendshipCollection.InsertOne(ctx, friendshipRequest)
	if err != nil {
		return Friendship{}, err
	}

	return Friendship{
		ID:       insertUpdate.InsertedID.(primitive.ObjectID),
		UserID:   userID.ID,
		FriendID: friendID,
		Status:   FriendshipStatusPending,
	}, nil
}

// Accept Friend Request
// Accept a friend request from another user.
func AcceptFriendRequest(ctx context.Context, FROM, TO UserResponse, friendshipID primitive.ObjectID) error {
	// Update the friendship request in the database.
	var filter = bson.M{"_id": friendshipID, "friend_id": TO.ID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusAccepted, "updated_at": time.Now()}}
	_, err := config.FriendshipCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	var user = GetUserByID(ctx, FROM.ID)
	var friend = GetUserByID(ctx, TO.ID)

	// Update Friendship list for user and friend.
	user.Friends = append(user.Friends, friend.ID)
	friend.Friends = append(friend.Friends, user.ID)

	// Update the user and friend in the database.
	_, err = config.UserCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"friends": user.Friends}})
	if err != nil {
		return err
	}

	_, err = config.UserCollection.UpdateOne(ctx, bson.M{"_id": friend.ID}, bson.M{"$set": bson.M{"friends": friend.Friends}})
	if err != nil {
		return err
	}

	return nil
}
