package helpers

import (
	"context"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productCollection = config.ProductCollection

type Product struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SuppliedID   primitive.ObjectID `json:"supplied_id,omitempty" bson:"supplied_id,omitempty"`
	RestaurantID primitive.ObjectID `json:"restaurant_id,omitempty" bson:"restaurant_id,omitempty"`
	Name         string             `json:"name,omitempty" bson:"name" binding:"required"`
	Category     Categories         `json:"category,omitempty" bson:"category" binding:"required"`
	Price        float64            `json:"price,omitempty" bson:"price" binding:"required"`
	Stock        int                `json:"stock,omitempty" bson:"stock" binding:"required"`
	CreatedAt    primitive.DateTime `json:"created_at,omitempty" bson:"created_at" default:"now()"`
	UpdatedAt    primitive.DateTime `json:"updated_at,omitempty" bson:"updated_at" default:"now()"`
}

type Products []Product

type Categories string

const (
	Drink  Categories = "drink"
	Food   Categories = "food"
	Others Categories = "others"
)

func (c Categories) String() string {
	switch c {
	case Drink:
		return "drink"
	case Food:
		return "food"
	case Others:
		return "others"
	default:
		return "others"
	}
}

func GetProductbyID(c context.Context, productID primitive.ObjectID) (Product, error) {
	var product Product

	funcName := ut.GetFunctionName()

	filter := bson.M{"_id": productID}

	err := productCollection.FindOne(c, filter).Decode(&product)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return product, err
	}

	return product, nil
}
