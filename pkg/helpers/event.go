package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type EventType string

const (
	Open  EventType = "open"
	Close EventType = "close"
)

func (h EventType) String() string {
	switch h {
	case Open:
		return "open"
	case Close:
		return "close"
	default:
		return "close"
	}
}

type EventCreate struct {
	Title     string               `json:"title" binding:"required"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" binding:"required" bson:"hosted_ids" default:"[]"`
	Orders    Orders               `json:"orders" bson:"orders" default:"[]"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      EventType            `json:"type" bson:"type" default:"close"`
	Budget    float64              `json:"budget" bson:"budget" binding:"number" default:"0"`
	Bill      float64              `json:"bill" bson:"bill" binding:"number" default:"0"`
	CreatedAt primitive.DateTime   `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
}

type Event struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string               `json:"title" binding:"required" bson:"title"`
	HostID    primitive.ObjectID   `json:"hostId" bson:"hostId"`
	HostedIDs []primitive.ObjectID `json:"hostedIds" bson:"hostedIds"`
	Orders    Orders               `json:"orders" bson:"orders" default:"[]"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      EventType            `json:"type" bson:"type"`
	Budget    float64              `json:"budget" bson:"budget" binding:"number" default:"0"`
	Bill      float64              `json:"bill" bson:"bill"`
	CreatedAt primitive.DateTime   `bson:"created_at" json:"created_at" omitEmpty:"true"`
	UpdatedAt primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
}

type InviteFriendToEventRequest struct {
	Event  Event      `json:"event_id" bson:"event_id"`
	Friend UserStruct `json:"user_id" bson:"user_id"`
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
