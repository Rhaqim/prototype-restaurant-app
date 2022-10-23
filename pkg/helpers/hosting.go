package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type HostingCreate struct {
	HostID    primitive.ObjectID   `json:"host_id" bson:"host_id"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" bson:"hosted_ids"`
	Bill      int                  `json:"bill" bson:"bill"`
}

type HostingUpdate struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	HostedIDs []primitive.ObjectID `json:"hosted_ids" bson:"hosted_ids"`
	Bill      int                  `json:"bill" bson:"bill"`
}
