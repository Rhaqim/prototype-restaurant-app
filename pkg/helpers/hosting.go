package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type HostingType string

const (
	Open  HostingType = "open"
	Close HostingType = "close"
)

func (h HostingType) String() string {
	switch h {
	case Open:
		return "open"
	case Close:
		return "close"
	default:
		return "close"
	}
}

type HostingCreate struct {
	Title     string               `json:"title" binding:"required"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" binding:"required" bson:"hosted_ids"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      HostingType          `json:"type" bson:"type"`
	Bill      int                  `json:"bill" bson:"bill"`
}

type Hosting struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string               `json:"title" binding:"required" bson:"title"`
	HostID    primitive.ObjectID   `json:"host_id" bson:"host_id"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" bson:"hosted_ids"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      HostingType          `json:"type" bson:"type"`
	Bill      int                  `json:"bill" bson:"bill"`
}
