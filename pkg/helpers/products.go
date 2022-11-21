package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	ID         primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	SuppliedID UserStruct          `json:"supplied_id,omitempty" bson:"supplied_id,omitempty"`
	Name       string              `json:"name,omitempty" bson:"name,omitempty"`
	Price      float64             `json:"price,omitempty" bson:"price,omitempty"`
	CreatedAt  primitive.Timestamp `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  primitive.Timestamp `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Products []Product
