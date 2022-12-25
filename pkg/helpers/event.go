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

// type EventCreate struct {
// 	Title     string               `json:"title" binding:"required"`
// 	Invited   []primitive.ObjectID `json:"invited" bson:"invited" default:"[]"`
// 	Attendees []primitive.ObjectID `json:"attendees" bson:"attendees" default:"[]"`
// 	Orders    Orders               `json:"orders" bson:"orders" default:"[]"`
// 	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
// 	Type      EventType            `json:"type" bson:"type" default:"close"`
// 	Budget    float64              `json:"budget" bson:"budget" binding:"number" default:"0"`
// 	Bill      float64              `json:"bill" bson:"bill" binding:"number" default:"0"`
// 	CreatedAt primitive.DateTime   `bson:"created_at" json:"created_at" default:"Now()"`
// 	UpdatedAt primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
// }

type Event struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title     string               `json:"title" binding:"required" bson:"title"`
	HostID    primitive.ObjectID   `json:"host_id" bson:"host_id"`
	Invited   []primitive.ObjectID `json:"invited" bson:"invited" default:"[]"`
	Attendees []primitive.ObjectID `json:"attendees" bson:"attendees" default:"[]"`
	Orders    []primitive.ObjectID `json:"orders" bson:"orders" default:"[]"`
	Venue     primitive.ObjectID   `json:"venue" bson:"venue"`
	Type      EventType            `json:"type" bson:"type"`
	Budget    float64              `json:"budget" bson:"budget" binding:"number" default:"0"`
	Bill      float64              `json:"bill" bson:"bill"`
	CreatedAt primitive.DateTime   `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
}
