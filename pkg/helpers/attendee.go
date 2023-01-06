package helpers

import (
	"context"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var attendeeCollection = config.AttendeeCollection

// InviteFriendsToEventRequest is the request to invite friends to an event
type InviteFriendsToEventRequest struct {
	EventID primitive.ObjectID   `json:"event_id" bson:"event_id"`
	Friends []primitive.ObjectID `json:"friends" bson:"friends"`
}

// AcceptInviteRequest is the request to accept an invite
type AcceptInviteRequest struct {
	EventID primitive.ObjectID `json:"event_id" bson:"event_id"`
	Budget  float64            `json:"budget" bson:"budget"`
}

// DeclineInviteRequest is the request to decline an invite
type DeclineInviteRequest struct {
	EventID primitive.ObjectID `json:"event_id" bson:"event_id"`
}

// AttendingStatus is the status of the attendee
type AttendingStatus string

const (
	Invited      AttendingStatus = "invited"
	Attending    AttendingStatus = "attending"
	NotAttending AttendingStatus = "not attending"
)

// String returns the string representation of the attending status
func (h AttendingStatus) String() string {
	switch h {
	case Invited:
		return "invited"
	case Attending:
		return "attending"
	case NotAttending:
		return "not attending"
	default:
		return "invited"
	}
}

// EventAttendee is the model for the attendees collection
type EventAttendee struct {
	EventID    primitive.ObjectID `json:"event_id" bson:"event_id"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	Status     AttendingStatus    `json:"status" bson:"status"`
	Budget     float64            `json:"budget" bson:"budget"`
	InvitedBy  primitive.ObjectID `json:"invited_by" bson:"invited_by"`
	InvitedAt  primitive.DateTime `json:"invited_at" bson:"invited_at"`
	AttendedAt primitive.DateTime `json:"attended_at" bson:"attended_at"`
}

// SendInviteToEvent sends an invite to the friends
// It uses go routines to send the invites to the friends
// It returns an error if any of the invites fail
// It returns nil if all the invites are successful
// It accepts the context, event id, friends to invite and the user who is inviting
func SendInviteToEvent(ctx context.Context, event_id primitive.ObjectID, friends []primitive.ObjectID, user UserResponse) error {
	// Add to the attendees collection using go routines
	var wg sync.WaitGroup
	wg.Add(len(friends))

	// channel to send the error to the main thread
	errChan := make(chan error, len(friends))

	for _, friend := range friends {
		go func(friend primitive.ObjectID) {
			defer wg.Done()

			var attendee = EventAttendee{
				EventID:   event_id,
				Status:    Invited,
				InvitedBy: user.ID,
				InvitedAt: primitive.NewDateTimeFromTime(time.Now()),
			}
			attendee.UserID = friend

			_, err := attendeeCollection.InsertOne(ctx, attendee)
			if err != nil {
				errChan <- err
			}
		}(friend)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAttendee gets the attendee from the attendees collection
// It returns the attendee and an error if any
// It accepts the context and the filter to use
func GetAttendee(ctx context.Context, filter bson.M) (EventAttendee, error) {
	var attendee EventAttendee
	err := attendeeCollection.FindOne(ctx, filter).Decode(&attendee)
	if err != nil {
		return attendee, err
	}

	return attendee, nil
}
