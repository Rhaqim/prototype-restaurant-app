package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TxnType string

type TxnStatus string

type TxnPurpose string

const (
	Debit  TxnType = "debit"
	Credit TxnType = "credit"
)

const (
	Event    TxnPurpose = "event"
	Personal TxnPurpose = "personal"
)

const (
	TxnSuccess TxnStatus = "success"
	TxnPending TxnStatus = "pending"
	TxnFail    TxnStatus = "fail"
)

type Transactions struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Txn_uuid string             `json:"txn_uuid" bson:"txn_uuid"`
	FromID   primitive.ObjectID `json:"from_id" bson:"from_id"`
	ToID     primitive.ObjectID `json:"to_id" bson:"to_id"`
	Amount   float64            `json:"amount" bson:"amount"`
	Type     TxnType            `json:"type" bson:"type"`
	Status   TxnStatus          `json:"status" bson:"status"`
	Date     time.Time          `json:"date" bson:"date"`
}

type TransactionStatus struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Status TxnStatus          `json:"status" bson:"status"`
}
