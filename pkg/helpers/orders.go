package helpers

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EventID    primitive.ObjectID `json:"event_id" bson:"event_id" binding:"required"`
	CustomerID primitive.ObjectID `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	Products   []OrderRequest     `json:"products,omitempty" bson:"products" min:"1" binding:"required"`
	Bill       float64            `json:"bill,omitempty" bson:"bill" binding:"number" default:"0"`
	CreatedAt  primitive.DateTime `json:"created_at" bson:"created_at" default:"now()"`
	UpdatedAt  primitive.DateTime `json:"updated_at" bson:"updated_at" default:"now()"`
}

type Orders []Order

type OrderRequest struct {
	ProductID primitive.ObjectID `json:"product_id," bson:"product_id" binding:"required,len=24,notblank"`
	Quantity  int                `json:"quantity," bson:"quantity" binding:"required,number,gt=0"`
}

type OrderRequest2 map[*primitive.ObjectID]int

func UpdateBill(ctx context.Context, request Order, billErrChan chan error, totalchan chan float64) {
	billChan := make(chan float64)

	billWg := sync.WaitGroup{}

	for i := range request.Products {
		billWg.Add(1)
		go func(i int) {
			defer billWg.Done()

			// get product from db go routine
			product_filter := bson.M{"_id": request.Products[i].ProductID}
			product_fetched, err := GetProduct(ctx, product_filter)
			if err != nil {
				billErrChan <- err
				return
			}

			bill := float64(float64(request.Products[i].Quantity) * product_fetched.Price)

			// send bill value through the channel
			billChan <- bill

			// update event bill
			event_filter := bson.M{"_id": request.EventID}
			event_update := bson.M{
				// update bill with new order
				"$inc": bson.M{
					"bill": +bill,
				},
			}

			_, err = eventCollection.UpdateOne(ctx, event_filter, event_update)
			if err != nil {
				billErrChan <- err
				return
			}
		}(i)
	}

	go func() {
		billWg.Wait()
		close(billChan)
		close(billErrChan)
	}()

	// calculate total bill
	var totalBill float64
	for bill := range billChan {
		totalBill += bill
	}

	// send total bill through the channel
	totalchan <- totalBill
}

func UpdateStock(ctx context.Context, request Order, errChan chan error) {

	var wg sync.WaitGroup
	for i := range request.Products {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// get product from db go routine
			product_filter := bson.M{"_id": request.Products[i].ProductID}
			product_update := bson.M{
				// decrement stock by quantity
				"$inc": bson.M{
					"stock": -request.Products[i].Quantity,
				},
			}
			_, err := productCollection.UpdateOne(ctx, product_filter, product_update)
			if err != nil {
				errChan <- err
				return
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()
}
