package helpers

import (
	"time"

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
	FromID    primitive.ObjectID `json:"from_id" binding:"required" bson:"from_id"`
	ToID      primitive.ObjectID `json:"to_id" binding:"required" bson:"to_id"`
	Amount    float64            `json:"amount" bson:"amount"`
	Type      TxnType            `json:"type" bson:"type"`
	Status    TxnStatus          `json:"status" bson:"status"`
	Date      time.Time          `json:"date,omitempty" bson:"date"`
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

func UpdateSenderTransaction(user UserResponse, amount float64) bool {
	return true
}

func UpdateReceiverTransaction(user UserResponse, amount float64) bool {
	return true
}
