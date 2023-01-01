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

type Friendship struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID  primitive.ObjectID `json:"friendId" bson:"friendId"`
	Status    FriendshipStatus   `json:"status" bson:"status" default:"0"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" default:"now()"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt" default:"now()"`
}

type FriendshipRequest struct {
	FriendID  primitive.ObjectID `json:"friendId" bson:"friendId"`
	Status    FriendshipStatus   `json:"status" bson:"status" default:"0"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt" default:"now()"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt" default:"now()"`
}

type FriendshipAcceptRequest struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`
	FriendID primitive.ObjectID `json:"friendId" bson:"friendId"`
}

func checkIfFriendExists(user, friend UserResponse) bool {
	for _, person := range user.Friends {
		if person == friend.ID {
			return true
		}
	}
	return false
}

func VerifyFriends(user UserResponse, friendID primitive.ObjectID) bool {
	var friend UserResponse
	err := config.UserCollection.FindOne(context.TODO(), bson.M{"_id": friendID}).Decode(&friend)
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
	friendship := VerifyFriends(userID, friendID)
	// If the friendship already exists, return a message.
	if friendship {
		return Friendship{}, errors.New("Friendship already exists")
	}

	// Create a new friendship request.
	friendshipRequest := bson.M{
		"user_id":   userID.ID,
		"friendId":  friendID,
		"status":    FriendshipStatusPending,
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
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
	friendship := VerifyFriends(TO, FROM.ID)
	// If the friendship already exists, return a message.
	if friendship {
		return errors.New("Friendship already exists")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"_id": friendshipID, "friendId": TO.ID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusAccepted, "updatedAt": time.Now()}}
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
	friendship := VerifyFriends(TO, FROM.ID)
	// If the friendship already exists, return a message.
	if friendship {
		return errors.New("Friendship already exists")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"_id": friendshipID, "friendId": TO.ID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusDeclined, "updatedAt": time.Now()}}
	_, err := config.FriendshipCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GetSocial(ctx context.Context, userID primitive.ObjectID, status FriendshipStatus) ([]FriendshipRequest, error) {
	var friendRequests []FriendshipRequest

	// Find all the friendship requests for the user.
	cursor, err := config.FriendshipCollection.Find(ctx, bson.M{"friendId": userID, "status": status})
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

func BlockUser(ctx context.Context, user UserResponse, friendID primitive.ObjectID) error {
	// Check if the friendship already exists.
	friendship := VerifyFriends(user, friendID)
	// If the friendship already exists, return a message.
	if !friendship {
		return errors.New("Friendship does not exist")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"user_id": user.ID, "friendId": friendID}
	var update = bson.M{"$set": bson.M{"status": FriendshipStatusBlocked, "updatedAt": time.Now()}}
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

func UnblockUser(ctx context.Context, user UserResponse, friendID primitive.ObjectID) error {
	// Check if the friendship already exists.
	friendship := VerifyFriends(user, friendID)
	// If the friendship already exists, return a message.
	if !friendship {
		return errors.New("Friendship does not exist")
	}

	// Update the friendship request in the database.
	var filter = bson.M{"user_id": user.ID, "friendId": friendID}
	// delete the friendship from the database
	_, err := config.FriendshipCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
