package helpers

import (
	"context"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var reviewCollect = config.ReviewCollection

type Review struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	RestauranID primitive.ObjectID `json:"restaurant_id" bson:"restaurant_id"`
	Author      primitive.ObjectID `json:"author,omitempty" bson:"author,omitempty"`
	Message     string             `json:"message,omitempty" bson:"message,omitempty"`
	Stars       int                `json:"stars,omitempty" bson:"stars,omitempty" default:"1" min:"1" max:"5"`
	CreatedAt   primitive.DateTime `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt   primitive.DateTime `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

func (r *Review) CreateReview(ctx context.Context) error {
	_, err := reviewCollect.InsertOne(ctx, r)
	if err != nil {
		return err
	}
	return nil
}
