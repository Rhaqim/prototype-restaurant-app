package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type InviteFriendsToEventRequest struct {
	EventID primitive.ObjectID   `json:"event_id" bson:"event_id"`
	Friends []primitive.ObjectID `json:"friends" bson:"friends"`
}

type AttendingStatus string

const (
	Invited      AttendingStatus = "invited"
	Attending    AttendingStatus = "attending"
	NotAttending AttendingStatus = "not attending"
)

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

type EventAttendee struct {
	EventID    primitive.ObjectID `json:"event_id" bson:"event_id"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	Status     AttendingStatus    `json:"status" bson:"status"`
	Budget     float64            `json:"budget" bson:"budget"`
	InvitedBy  primitive.ObjectID `json:"invited_by" bson:"invited_by"`
	InvitedAt  primitive.DateTime `json:"invited_at" bson:"invited_at"`
	AttendedAt primitive.DateTime `json:"attended_at" bson:"attended_at"`
}

func SendInviteToEvent(event Event, friend UserStruct) {
	// Send invite to friend
}

func InviteFriendToEvent(event Event, friend UserStruct) {
	// Create invite
	// Send invite to friend
}

func AcceptInviteToEvent(event Event, friend UserStruct) {
	// Update invite
	// Send invite to friend
}

func DeclineInviteToEvent(event Event, friend UserStruct) {
	// Update invite
	// Send invite to friend
}
