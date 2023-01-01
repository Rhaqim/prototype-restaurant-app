package helpers

import (
	"context"
	"errors"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Txn_uuid  string             `json:"txn_uuid" bson:"txn_uuid"`
	FromID    primitive.ObjectID `json:"fromId" binding:"required" bson:"fromId"`
	ToID      primitive.ObjectID `json:"toId" binding:"required" bson:"toId"`
	Amount    float64            `json:"amount" bson:"amount"`
	Type      TxnType            `json:"type" bson:"type"`
	Status    TxnStatus          `json:"status" bson:"status"`
	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt" default:"Now()"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"updatedAt" default:"Now()"`
}

type TransactionStatus struct {
	ID       primitive.ObjectID `json:"id" bson:"_id" binding:"required"`
	Txn_uuid string             `json:"txn_uuid" bson:"txn_uuid" binding:"required"`
	Status   TxnStatus          `json:"status" bson:"status" binding:"required"`
}

func VerifyWalletSufficientBalance(user UserResponse, amount float64) bool {
	return user.Wallet >= amount
}

func UpdateSenderTransaction(ctx context.Context, user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnSuccess {
		return false
	}
	user.Wallet -= amount

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"wallet": user.Wallet}}

	updateResult, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false
	}
	return updateResult.ModifiedCount == 1
}

func UpdateReceiverTransaction(ctx context.Context, user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnSuccess {
		return false
	}
	user.Wallet += amount

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": bson.M{"wallet": user.Wallet}}

	updateResult, err := usersCollection.UpdateOne(ctx, filter, update)
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
