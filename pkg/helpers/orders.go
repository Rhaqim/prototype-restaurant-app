package helpers

import (
	"context"
	"sync"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var orderCollection = config.OrderCollection

type Order struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EventID    primitive.ObjectID `json:"event_id" bson:"event_id" binding:"required"`
	CustomerID primitive.ObjectID `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	Products   []OrderRequest     `json:"products,omitempty" bson:"products" min:"1" binding:"required"`
	Bill       float64            `json:"bill,omitempty" bson:"bill" binding:"number" default:"0"`
	Paid       bool               `json:"paid,omitempty" bson:"paid" default:"false"`
	CreatedAt  primitive.DateTime `json:"created_at" bson:"created_at" default:"time.Now()"`
	UpdatedAt  primitive.DateTime `json:"updated_at" bson:"updated_at" default:"time.Now()"`
}

type Orders []Order

type OrderRequest struct {
	ProductID primitive.ObjectID `json:"product_id," bson:"product_id" binding:"required,len=24,notblank"`
	Quantity  int                `json:"quantity," bson:"quantity" binding:"required,number,gt=0"`
}

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

func GetOrders(c context.Context, filter bson.M) (Orders, error) {
	var orders Orders

	cur, err := orderCollection.Find(c, filter)
	if err != nil {
		return orders, err
	}

	for cur.Next(c) {
		var order Order
		err := cur.Decode(&order)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func GetOrder(c context.Context, filter bson.M) (Order, error) {
	var order Order

	err := orderCollection.FindOne(c, filter).Decode(&order)
	if err != nil {
		return order, err
	}

	return order, nil
}

func GetOrdersbyEventID(c context.Context, id primitive.ObjectID) (Orders, error) {
	var orders Orders

	filter := bson.M{"event_id": id}
	orders, err := GetOrders(c, filter)
	if err != nil {
		return orders, err
	}

	return orders, nil
}

func GetOrderbyID(c context.Context, id primitive.ObjectID) (Order, error) {
	var order Order

	filter := bson.M{"_id": id}
	order, err := GetOrder(c, filter)
	if err != nil {
		return order, err
	}

	return order, nil
}

func UpdateOrder(ctx context.Context, filter bson.M, update bson.M) (bool, error) {
	_, err := orderCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateManyOrders(ctx context.Context, filter bson.M, update bson.M) (bool, error) {
	_, err := orderCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateCustomerOrders(ctx context.Context, event Event, user UserResponse, orderErrChan chan error) {
	var wg sync.WaitGroup

	// get orders by event id
	filter := bson.M{"event_id": event.ID, "customer_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"paid": true,
		},
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// update orders
		_, err := UpdateManyOrders(ctx, filter, update)
		if err != nil {
			orderErrChan <- err
			return
		}
	}()

	go func() {
		wg.Wait()
		close(orderErrChan)
	}()

}
