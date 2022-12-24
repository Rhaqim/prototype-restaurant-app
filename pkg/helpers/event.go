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
	HostedIDs []primitive.ObjectID `json:"hostedIds" binding:"required" bson:"hostedIds"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      EventType            `json:"type" bson:"type" default:"close"`
	Bill      int                  `json:"bill" bson:"bill" binding:"required,number" default:"0"`
	CreatedAt primitive.DateTime   `bson:"createdAt" json:"createdAt" default:"Now()"`
	UpdatedAt primitive.DateTime   `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}

type Event struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string               `json:"title" binding:"required" bson:"title"`
	HostID    primitive.ObjectID   `json:"hostId" bson:"hostId"`
	HostedIDs []primitive.ObjectID `json:"hostedIds" bson:"hostedIds"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      EventType            `json:"type" bson:"type"`
	Bill      int                  `json:"bill" bson:"bill"`
	CreatedAt primitive.DateTime   `bson:"createdAt" json:"createdAt" omitEmpty:"true"`
	UpdatedAt primitive.DateTime   `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}
