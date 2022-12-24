package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
	ID         primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	EventID    primitive.ObjectID  `json:"event_id,omitempty" bson:"event_id,omitempty"`
	CustomerID primitive.ObjectID  `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	ProductID  primitive.ObjectID  `json:"product_id,omitempty" bson:"product_id,omitempty"`
	Quantity   int                 `json:"quantity,omitempty" bson:"quantity,omitempty" binding:"required,number,min=1"`
	CreatedAt  primitive.Timestamp `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  primitive.Timestamp `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Orders []Order
