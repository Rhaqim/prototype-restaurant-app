package helpers

import (
	"context"
	"errors"

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

func UpdateSenderTransaction(user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnSuccess {
		return false
	}
	user.Wallet -= amount
	return true
}

func UpdateReceiverTransaction(user UserResponse, amount float64, txn Transactions) bool {
	if txn.Status != TxnSuccess {
		return false
	}
	user.Wallet += amount
	return true
}

func UpdateWalletBalance(ctx context.Context, txn Transactions) error {
	var fromUser UserResponse
	var toUser UserResponse

	fromUser = GetUserByID(ctx, txn.FromID)
	toUser = GetUserByID(ctx, txn.ToID)

	if !UpdateSenderTransaction(fromUser, txn.Amount, txn) {
		return errors.New("error updating sender transaction")
	}

	if !UpdateReceiverTransaction(toUser, txn.Amount, txn) {
		return errors.New("error updating receiver transaction")
	}

	return nil
}
