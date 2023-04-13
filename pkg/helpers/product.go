package helpers

import (
	"context"
	"fmt"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productCollection = config.ProductCollection

// unique id for products
var ProductUID = "PC-" + ut.GenerateUUID()

type Product struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	ProductUID   string             `json:"product_uid,omitempty" bson:"product_uid"`
	SupplierID   primitive.ObjectID `json:"supplier_id,omitempty" bson:"supplier_id,omitempty"`
	RestaurantID primitive.ObjectID `json:"restaurant_id,omitempty" bson:"restaurant_id,omitempty"`
	Name         string             `json:"name" bson:"name" binding:"required,min=3,max=50,lowercase"`
	ProductImage Avatar             `json:"product_image,omitempty" bson:"product_image,omitempty"`
	Category     Categories         `json:"category,omitempty" bson:"category"`
	Price        float64            `json:"price" bson:"price" binding:"required"`
	Stock        uint64             `json:"stock" bson:"stock" binding:"required"`
	CreatedAt    primitive.DateTime `json:"created_at,omitempty" bson:"created_at" default:"time.Now()"`
	UpdatedAt    primitive.DateTime `json:"updated_at,omitempty" bson:"updated_at" default:"time.Now()"`
}

type Products []Product

type Categories string

const (
	Beverage  Categories = "beverage"
	Appetizer Categories = "appetizer"
	Main      Categories = "main"
	Dessert   Categories = "dessert"
	Snack     Categories = "snack"
	Others    Categories = "others"
)

func (c Categories) String() string {
	return string(c)
}

func GetProduct(c context.Context, filter bson.M) (Product, error) {
	var product Product

	funcName := ut.GetFunctionName()

	err := productCollection.FindOne(c, filter).Decode(&product)
	if err != nil {
		SetDebug(err.Error(), funcName)
		fmt.Printf("Error: %v", filter)
		return product, err
	}

	return product, nil
}

func GetProductbyID(c context.Context, productID primitive.ObjectID) (Product, error) {
	var product Product

	funcName := ut.GetFunctionName()

	filter := bson.M{"_id": productID}

	product, err := GetProduct(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return product, err
	}

	SetInfo(fmt.Sprintf("Product %v found", product.Name), funcName)
	return product, nil
}

func GetProductbyName(c context.Context, productName string) (Product, error) {
	var product Product

	funcName := ut.GetFunctionName()

	filter := bson.M{"name": productName}

	product, err := GetProduct(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return product, err
	}

	SetInfo(fmt.Sprintf("Product %v found", product.Name), funcName)
	return product, nil
}

func GetAllProductByRestaurant(c context.Context, restaurantID primitive.ObjectID) (Products, error) {
	var products Products

	funcName := ut.GetFunctionName()

	filter := bson.M{"restaurant_id": restaurantID}

	products, err := GetProducts(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return products, err
	}

	SetInfo(fmt.Sprintf("Found %v products", len(products)), funcName)
	return products, nil
}

func GetProducts(c context.Context, filter bson.M) (Products, error) {
	var products Products

	funcName := ut.GetFunctionName()

	cursor, err := productCollection.Find(c, filter)
	if err != nil {
		SetDebug(err.Error(), funcName)
		return products, err
	}

	for cursor.Next(c) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	SetInfo(fmt.Sprintf("Found %v products", len(products)), funcName)
	return products, nil
}
