package helpers

import (
	"context"
	"errors"
	"sync"

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
	if txn.Status != TxnSuccess {
		return false
	}

	filter := bson.M{"user_id": user.ID}
	update := bson.M{"$set": bson.M{
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
	update := bson.M{"$set": bson.M{
		"balance": +amount,
	}}

	updateResult, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false
	}
	return updateResult.ModifiedCount == 1
}

func UpdateWalletBalance(ctx context.Context, txn Transactions) error {
	var fromUser UserResponse
	var toUser UserResponse

	fromUser = GetUserByID(ctx, txn.FromID)
	toUser = GetUserByID(ctx, txn.ToID)

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		if !UpdateSenderTransaction(ctx, fromUser, txn.Amount, txn) {
			errChan <- errors.New("error updating sender transaction")
			return
		}
	}()

	go func() {
		defer wg.Done()
		if !UpdateReceiverTransaction(ctx, toUser, txn.Amount, txn) {
			errChan <- errors.New("error updating receiver transaction")
			return
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	return nil
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

	// TODO: Determine if money goes from user to venue or
	// TODO: from event wallet to venue
	// TODO: if from event wallet to venue, then check if event has sufficient balance
	// TODO: or from user to venue owner

	// create transaction
	txn = Transactions{
		TransactionUID: TransactionUID,
		FromID:         user.ID,
		ToID:           event.Venue,
		Amount:         event.Bill,
		Type:           Debit,
		Status:         TxnSuccess,
	}

	// insert transaction
	txn, err := InsertTransaction(ctx, txn)
	if err != nil {
		return txn, err
	}

	// update wallet balance
	if err := UpdateWalletBalance(ctx, txn); err != nil {
		return txn, err
	}

	return txn, nil
}
