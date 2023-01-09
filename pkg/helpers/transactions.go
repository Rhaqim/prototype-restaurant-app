package helpers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var transactionCollection = config.TransactionCollection

// unique id for transactions
var TransactionUID = "TC-" + ut.GenerateUUID()

type TxnType string

type TxnStatus string

const (
	Debit  TxnType = "debit"
	Credit TxnType = "credit"
)

const (
	TxnSuccess TxnStatus = "success"
	TxnPending TxnStatus = "pending"
	TxnFail    TxnStatus = "fail"
)

type Transactions struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	TransactionUID string             `json:"transaction_uid,omitempty" bson:"transaction_uid"`
	FromID         primitive.ObjectID `json:"from_id" binding:"required" bson:"from_id"`
	ToID           primitive.ObjectID `json:"to_id" binding:"required" bson:"to_id"`
	Amount         float64            `json:"amount" bson:"amount"`
	Type           TxnType            `json:"type" bson:"type"`
	Status         TxnStatus          `json:"status" bson:"status"`
	CreatedAt      primitive.DateTime `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt      primitive.DateTime `bson:"updated_at" json:"updated_at" default:"Now()"`
}

type TransactionStatus struct {
	ID       primitive.ObjectID `json:"id" bson:"_id" binding:"required"`
	Txn_uuid string             `json:"txn_uuid" bson:"txn_uuid" binding:"required"`
	TxnPin   string             `json:"txn_pin" bson:"txn_pin" binding:"required"`
	Status   TxnStatus          `json:"status" bson:"status" binding:"required"`
}

func GetTransaction(ctx context.Context, filter bson.M) (Transactions, error) {
	var txn Transactions
	err := transactionCollection.FindOne(ctx, filter).Decode(&txn)
	return txn, err
}

func InsertTransaction(ctx context.Context, txn Transactions) (Transactions, error) {
	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err = transactionCollection.InsertOne(ctx, txn)
	}()
	wg.Wait()
	return txn, err
}

func VerifyWalletSufficientBalance(ctx context.Context, user UserResponse, amount float64) bool {
	wallet, err := GetWallet(ctx, bson.M{"user_id": user.ID})
	if err != nil {
		return false
	}
	return wallet.Balance >= amount
}

// UpdateSenderTransaction updates the sender wallet balance
// It tries to get the money from the budget first
// If the budget is not sufficient, it gets the rest from the wallet
// it updates the wallet balance and returns true if successful
func UpdateSenderTransaction(ctx context.Context, user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnPending {
		return false
	}

	// Get Money from the budget
	budgetAmount := UnlockBudget(ctx, txn.ToID, user)

	amount = amount - budgetAmount

	// Update the wallet to get the rest from the wallet or return the rest to the wallet
	filter := bson.M{"user_id": user.ID}
	update := bson.M{"$inc": bson.M{
		"balance": -amount,
	}}

	updateResult, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false
	}
	return updateResult.ModifiedCount == 1
}

// UpdateReceiverTransaction updates the receiver wallet balance
// It adds the amount to the receiver wallet balance
// It returns true if successful
func UpdateReceiverTransaction(ctx context.Context, user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnSuccess {
		return false
	}

	filter := bson.M{"user_id": user.ID}
	update := bson.M{"$inc": bson.M{
		"balance": +amount,
	}}

	updateResult, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false
	}
	return updateResult.ModifiedCount == 1
}

// UpdateAndReturnTransaction updates the transaction status and returns the updated transaction
func UpdateAndReturnTransaction(ctx context.Context, txn Transactions, status TxnStatus) (Transactions, error) {
	// Update transaction status to success
	filter := bson.M{"transaction_uid": txn.TransactionUID, "_id": txn.ID}
	update := bson.M{"$set": bson.M{
		"status": status,
	}}
	// update and return new document
	updateResult, errs := transactionCollection.UpdateOne(ctx, filter, update)
	if errs != nil {
		return txn, errs
	}
	if updateResult.ModifiedCount != 1 {
		return txn, errors.New("error updating transaction for success")
	}

	txn, errs = GetTransaction(ctx, filter)
	if errs != nil {
		return txn, errs
	}

	return txn, nil
}

// UpdateWalletBalance updates the wallet balance of the sender and receiver
// If an error occurs, it rolls back the transaction and updates the transaction status to fail
// If the transaction is successful, it updates the transaction status to success
func UpdateWalletBalance(ctx context.Context, txn Transactions) (Transactions, error) {
	var fromUser UserResponse
	var toUser UserResponse

	fromUser = GetUserByID(ctx, txn.FromID)
	toUser = GetUserByID(ctx, txn.ToID)

	if !UpdateSenderTransaction(ctx, fromUser, txn.Amount, txn) {
		txn, err := UpdateAndReturnTransaction(ctx, txn, TxnFail)
		if err != nil {
			return txn, err
		}
		return txn, errors.New("error updating sender transaction")
	}

	if !UpdateReceiverTransaction(ctx, toUser, txn.Amount, txn) {
		// Rollback to Sender
		_ = UpdateSenderTransaction(ctx, fromUser, -txn.Amount, txn)
		txn, err := UpdateAndReturnTransaction(ctx, txn, TxnFail)
		if err != nil {
			return txn, err
		}

		return txn, errors.New("error updating receiver transaction")
	}

	txn, err := UpdateAndReturnTransaction(ctx, txn, TxnSuccess)
	if err != nil {
		return txn, err
	}

	return txn, nil
}

// SendtoVenues sends money to venues
// It takes a context and an event and a user
// It returns a transaction and an error
func SendtoVenues(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	var txn Transactions

	// Get Budget
	budget, err := GetBudget(ctx, bson.M{"purpose_id": event.Venue, "user_id": user.ID})
	if err != nil {
		return Transactions{}, err
	}
	var totalAmount float64 = budget.Amount + event.Bill

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, totalAmount) {
		return txn, errors.New("insufficient balance")
	}

	// Get Venue Owner
	restaurant, err := GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		return txn, err
	}

	// TODO: Determine if money goes from user to venue or
	// TODO: from event wallet to venue
	// TODO: if from event wallet to venue, then check if event has sufficient balance
	// TODO: or from user to venue owner

	// create transaction
	txn = Transactions{
		ID:             primitive.NewObjectID(),
		TransactionUID: TransactionUID,
		FromID:         user.ID,
		ToID:           restaurant.OwnerID,
		Amount:         event.Bill,
		Type:           Debit,
		Status:         TxnPending,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now()),
	}

	// insert transaction
	txn, err = InsertTransaction(ctx, txn)
	if err != nil {
		return txn, err
	}

	// update wallet balance
	txn, err = UpdateWalletBalance(ctx, txn)
	if err != nil {
		return txn, err
	}

	return txn, nil
}

/* Event Transaction */
type EventBillPayment struct {
	EventID primitive.ObjectID `json:"event_id" bson:"event_id" binding:"required"`
	TxnPin  string             `json:"txn_pin" bson:"txn_pin" binding:"required"`
}

// SendToHost sends money to the host of the event
// It takes a context and an event and a user
// It gets the total bill from orders made for the event
// and sends the money to the host
// It returns a transaction and an error
func SendToHost(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	var orders []Order

	// Get total bill from orders
	orders, err := GetOrders(ctx, bson.M{"event_id": event.Venue, "user_id": user.ID})
	if err != nil {
		return Transactions{}, err
	}

	var totalBill float64
	for _, order := range orders {
		totalBill += order.Bill
	}

	// Get Budget
	budget, err := GetBudget(ctx, bson.M{"purpose_id": event.ID})
	if err != nil {
		return Transactions{}, err
	}
	var totalAmount float64 = budget.Amount + totalBill

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, totalAmount) {
		return Transactions{}, errors.New("insufficient balance")
	}

	// Get Host
	host, err := GetUser(ctx, bson.M{"_id": event.HostID})
	if err != nil {
		return Transactions{}, err
	}

	// create transaction
	txn := Transactions{
		ID:             primitive.NewObjectID(),
		TransactionUID: TransactionUID,
		FromID:         user.ID,
		ToID:           host.ID,
		Amount:         totalBill,
		Type:           Debit,
		Status:         TxnPending,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now()),
	}

	// insert transaction
	txn, err = InsertTransaction(ctx, txn)
	if err != nil {
		return txn, err
	}

	// update wallet balance
	txn, err = UpdateWalletBalance(ctx, txn)
	if err != nil {
		return txn, err
	}

	return txn, nil
}
