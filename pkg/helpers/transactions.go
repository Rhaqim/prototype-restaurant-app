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

// func UpdateWalletBalance(ctx context.Context, txn Transactions) (Transactions, error) {
// 	var fromUser UserResponse
// 	var toUser UserResponse

// 	fromUser = GetUserByID(ctx, txn.FromID)
// 	toUser = GetUserByID(ctx, txn.ToID)

// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	errChan := make(chan error, 2)

// 	go func() {
// 		defer wg.Done()
// 		if !UpdateSenderTransaction(ctx, fromUser, txn.Amount, txn) {
// 			errChan <- errors.New("error updating sender transaction")
// 			return
// 		}
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		if !UpdateReceiverTransaction(ctx, toUser, txn.Amount, txn) {
// 			errChan <- errors.New("error updating receiver transaction")
// 			return
// 		}
// 	}()

// 	wg.Wait()
// 	close(errChan)

// 	for err := range errChan {
// 		return txn, err
// 	}

//		return txn, nil
//	}
func UpdateWalletBalance(ctx context.Context, txn Transactions) (Transactions, error) {
	var fromUser UserResponse
	var toUser UserResponse

	fromUser = GetUserByID(ctx, txn.FromID)
	toUser = GetUserByID(ctx, txn.ToID)

	if !UpdateSenderTransaction(ctx, fromUser, txn.Amount, txn) {
		return txn, errors.New("error updating sender transaction")
	}

	if !UpdateReceiverTransaction(ctx, toUser, txn.Amount, txn) {
		// Rollback to Sender
		_ = UpdateSenderTransaction(ctx, fromUser, -txn.Amount, txn)

		return txn, errors.New("error updating receiver transaction")
	}

	// Update transaction status to success
	filter := bson.M{"transaction_uid": txn.TransactionUID, "_id": txn.ID}
	update := bson.M{"$set": bson.M{
		"status": TxnSuccess,
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

// RollbackTransaction rolls back a transaction
func RollbackTransaction(err error, txn Transactions, to UserResponse, from UserResponse) (Transactions, error) {
	ctx := context.Background()

	// Update transaction status to fail
	filter := bson.M{"transaction_uid": txn.TransactionUID, "_id": txn.ID}
	update := bson.M{"$set": bson.M{
		"status": TxnFail,
	}}
	// update and return new document
	updateResult, errs := transactionCollection.UpdateOne(ctx, filter, update)
	if errs != nil {
		return txn, errs
	}
	if updateResult.ModifiedCount != 1 {
		return txn, errors.New("error updating transaction for rollback")
	}

	txn, errs = GetTransaction(ctx, filter)
	if errs != nil {
		return txn, errs
	}

	if err.Error() == "error updating sender transaction" {
		// Update sender wallet
		if !UpdateSenderTransaction(ctx, from, txn.Amount, txn) {
			return txn, errors.New("error updating sender wallet")
		}
	}

	if err.Error() == "error updating receiver transaction" {
		// Update receiver wallet
		if !UpdateReceiverTransaction(ctx, to, txn.Amount, txn) {
			return txn, errors.New("error updating receiver wallet")
		}
	}

	return txn, nil
}

// SendtoVenues sends money to venues
// It takes a context and an event and a user
// It returns a transaction and an error
func SendtoVenues(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	var txn Transactions

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, event.Bill) {
		return txn, errors.New("insufficient balance")
	}

	// Get Venue Owner
	_, err := GetRestaurant(ctx, bson.M{"_id": event.Venue})
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
		ToID:           event.Venue,
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
