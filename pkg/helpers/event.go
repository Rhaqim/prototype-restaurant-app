package helpers

import (
	"context"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var eventCollection = config.EventCollection

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

type EventStatus string

const (
	Upcoming EventStatus = "upcoming"
	Ongoing  EventStatus = "ongoing"
	Finished EventStatus = "finished"
)

func (h EventStatus) String() string {
	switch h {
	case Upcoming:
		return "upcoming"
	case Ongoing:
		return "ongoing"
	case Finished:
		return "finished"
	default:
		return "upcoming"
	}
}

type Event struct {
	ID             primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	HostID         primitive.ObjectID   `json:"host_id" bson:"host_id"`
	Title          string               `json:"title" binding:"required" bson:"title"`
	Venue          primitive.ObjectID   `json:"venue" bson:"venue" binding:"required"`
	Date           CustomDate           `json:"date" bson:"date" binding:"required" time_format:"2006-01-02"`
	Time           CustomTime           `json:"time" bson:"time" binding:"required" time_format:"15:04"`
	Invited        []primitive.ObjectID `json:"invited" bson:"invited" default:"[]"`
	Attendees      []primitive.ObjectID `json:"attendees" bson:"attendees" default:"[]"`
	Declined       []primitive.ObjectID `json:"declined" bson:"declined" default:"[]"`
	Type           EventType            `json:"type" bson:"type"`
	EventStatus    EventStatus          `json:"event_status" bson:"event_status"`
	SpecialRequest string               `json:"special_request,omitempty" bson:"special_request,omitempty"`
	Budget         float64              `json:"budget" bson:"budget" binding:"required,number"`
	Bill           float64              `json:"bill,omitempty" bson:"bill,omitempty" default:"0"`
	CreatedAt      primitive.DateTime   `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt      primitive.DateTime   `bson:"updated_at" json:"updated_at" default:"Now()"`
}

func GetEvent(ctx context.Context, filter bson.M) (Event, error) {
	var event Event
	err := eventCollection.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

func GetEvents(ctx context.Context, filter bson.M) ([]Event, error) {
	var events []Event
	cur, err := eventCollection.Find(ctx, filter)
	if err != nil {
		return events, err
	}

	for cur.Next(ctx) {
		var event Event
		err = cur.Decode(&event)
		if err != nil {
			return events, err
		}

		events = append(events, event)
	}

	return events, nil
}

func CreateEvent(ctx context.Context, event Event) (Event, error) {
	result, err := eventCollection.InsertOne(ctx, event)
	if err != nil {
		return event, err
	}

	event.ID = result.InsertedID.(primitive.ObjectID)

	return event, nil
}

func UpdateEvent(ctx context.Context, filter bson.M, update bson.M) (Event, error) {
	var event Event
	err := eventCollection.FindOneAndUpdate(ctx, filter, update).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

func DeleteEvent(ctx context.Context, filter bson.M) (Event, error) {
	var event Event
	err := eventCollection.FindOneAndDelete(ctx, filter).Decode(&event)
	if err != nil {
		return event, err
	}

	return event, nil
}

// Return the difference in minutes between current time and the event time
func (e *Event) GetTimeDifference() int {
	currentTime := time.Now().Minute()
	eventTime := e.Time.Time.Minute()

	return eventTime - currentTime
}
