package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type HostingCreate struct {
	HostedIDs []primitive.ObjectID `json:"hosted_ids" bson:"hosted_ids"`
	Bill      int                  `json:"bill" bson:"bill"`
}

type Hosting struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" bson:"hosted_ids"`
	Bill      int                  `json:"bill" bson:"bill"`
}
