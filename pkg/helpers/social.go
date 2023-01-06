package helpers

import (
	"context"
	"errors"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
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
	// FriendshipStatusDeclined means that the friendship is declined.
	FriendshipStatusDeclined
	// FriendshipStatusBlocked means that the friendship is blocked.
	FriendshipStatusBlocked
)

// Model for the friendship collection
type Friendship struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID  primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	Status    FriendshipStatus   `json:"status" bson:"status" default:"0"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at" default:"time.Now()"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at" default:"time.Now()"`
}

// FriendshipRequest is the request to send a friendship request
// to another user.
// Sets the status to pending by default.
type FriendshipRequest struct {
	FriendID  primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	Status    FriendshipStatus   `json:"status" bson:"status" default:"0"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at" default:"time.Now()"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at" default:"time.Now()"`
}

// FriendshipAcceptRequest is the request to accept a friendship request.
type FriendshipAcceptRequest struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID primitive.ObjectID `json:"friend_id" bson:"friend_id"`
}

// Get Friendship from database
func GetFriendship(ctx context.Context, filter bson.M) (Friendship, error) {
	var friendship Friendship
	err := config.FriendshipCollection.FindOne(ctx, filter).Decode(&friendship)
	if err != nil {
		return Friendship{}, err
	}
	return friendship, nil
}

// Check if request has been sent already
// Returns true if request has been sent
func CheckIfRequestExists(ctx context.Context, user UserResponse, friendID primitive.ObjectID) bool {
	filter := bson.M{"user_id": user.ID, "friend_id": friendID}

	friend, err := GetFriendship(ctx, filter)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false
		}

		SetDebug(err.Error(), ut.GetFunctionName())
		return false
	}

	if friend.Status == FriendshipStatusPending {
		return true
	}

	return false
}

// Check if User is friends with another user
// by checking if the friendID is in the user's friends array
// Returns true if friendship exists
func checkIfFriendExists(user, friend UserResponse) bool {
	for _, person := range user.Friends {
		if person == friend.ID {
			return true
		}
	}
	return false
}

// Verify if friendship exists
// Checks if friendship exists in both users
// by checking if the friendID is in the user's friends array
// and if the user's ID is in the friend's friends array.
// Returns true if friendship exists
func VerifyFriends(ctx context.Context, user UserResponse, friendID primitive.ObjectID) bool {

	filter := bson.M{"_id": friendID}

	friend, err := GetUser(ctx, filter)
	if err != nil {
		SetDebug(err.Error(), ut.GetFunctionName())
		return false
	}

	if checkIfFriendExists(user, friend) && checkIfFriendExists(friend, user) {
		return true
	}

	return false
}

func removeFriend(friends []primitive.ObjectID, friendID primitive.ObjectID) []primitive.ObjectID {
	for i, friend := range friends {
		if friend == friendID {
			return append(friends[:i], friends[i+1:]...)
		}
	}

	return friends
}

// Send Friend Request
// Send a friend request to another user.
func SendFriendRequest(ctx context.Context, userID UserResponse, friendID primitive.ObjectID) (Friendship, error) {
	// check that userID is not the same as friendID
	if userID.ID == friendID {
		return Friendship{}, errors.New("you cannot be friends with yourself, c'mon")
	}

	// Check if the friendship already exists.
	friendship := VerifyFriends(ctx, userID, friendID)
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
	// Check if the friendship already exists.
	friendship := VerifyFriends(ctx, TO, FROM.ID)
	// If the friendship already exists, return a message.
	if friendship {
		return errors.New("Friendship already exists")
	}

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

func DeclineFriendRequest(ctx context.Context, FROM, TO UserResponse, friendshipID primitive.ObjectID) error {
	// Check if the friendship already exists.
	friendship := VerifyFriends(ctx, TO, FROM.ID)
	// If the friendship already exists, return a message.
	if friendship {
		return errors.New("Friendship already exists")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"_id": friendshipID, "friend_id": TO.ID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusDeclined, "updated_at": time.Now()}}
	_, err := config.FriendshipCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GetSocial(ctx context.Context, userID primitive.ObjectID, status FriendshipStatus) ([]FriendshipRequest, error) {
	var friendRequests []FriendshipRequest

	// Find all the friendship requests for the user.
	cursor, err := config.FriendshipCollection.Find(ctx, bson.M{"friend_id": userID, "status": status})
	if err != nil {
		return friendRequests, err
	}

	// Iterate through the cursor and decode each document into a FriendshipRequest.
	for cursor.Next(ctx) {
		var friendRequest FriendshipRequest
		err := cursor.Decode(&friendRequest)
		if err != nil {
			return friendRequests, err
		}

		// Append the FriendshipRequest to the slice.
		friendRequests = append(friendRequests, friendRequest)
	}

	// Close the cursor once finished.
	cursor.Close(ctx)

	return friendRequests, nil
}

// Block User
// Updates the friendship status to blocked.
// Removes the friend from the user's friends list.
// Removes the user from the friend's friends list.
func BlockUser(ctx context.Context, user UserResponse, friendID primitive.ObjectID) error {
	// Check if the friendship already exists.
	friendship := VerifyFriends(ctx, user, friendID)

	// If the friendship already exists, return a message.
	if !friendship {
		return errors.New("Friendship does not exist")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"user_id": user.ID, "friend_id": friendID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusBlocked, "updated_at": time.Now()}}
	_, err := config.FriendshipCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Remove the friend from the user's friends list.
	var friend = GetUserByID(ctx, friendID)

	// Update Friendship list for user and friend.
	user.Friends = removeFriend(user.Friends, friend.ID)
	friend.Friends = removeFriend(friend.Friends, user.ID)

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

// Unblock User
// Checks if the friendship exists.
// If the friendship exists, delete the friendship from the database.
func UnblockUser(ctx context.Context, user UserResponse, friendID primitive.ObjectID) error {
	// Check if the friendship already exists.
	friendship := VerifyFriends(ctx, user, friendID)
	// If the friendship already exists, return a message.
	if !friendship {

		// Update the friendship request in the database.
		// Filter the friendship by the user's ID, friend's ID, and the status of blocked.
		var filter = bson.M{"user_id": user.ID, "friend_id": friendID, "status": FriendshipStatusBlocked}

		// delete the friendship from the database
		_, err := config.FriendshipCollection.DeleteOne(ctx, filter)
		if err != nil {
			return err
		}
	}

	return nil
}
