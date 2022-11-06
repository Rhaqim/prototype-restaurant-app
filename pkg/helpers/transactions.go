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
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Txn_uuid string             `json:"txn_uuid" bson:"txn_uuid"`
	FromID   primitive.ObjectID `json:"from_id" binding:"required" bson:"from_id"`
	ToID     primitive.ObjectID `json:"to_id" binding:"required" bson:"to_id"`
	Amount   float64            `json:"amount" bson:"amount"`
	Type     TxnType            `json:"type" bson:"type"`
	Status   TxnStatus          `json:"status" bson:"status"`
	Date     time.Time          `json:"date" bson:"date"`
}

type TransactionStatus struct {
	ID       primitive.ObjectID `json:"id" bson:"_id" binding:"required"`
	Txn_uuid string             `json:"txn_uuid" bson:"txn_uuid" binding:"required"`
	Status   TxnStatus          `json:"status" bson:"status" binding:"required"`
}

// func VerifyTransaction(user UserResponse, txn Transactions) bool {
// 	//  Check Reciver in Friends List
// 	for _, friend := range user.Friends {
// 		if friend == request.Receiver {
// 			//  Check if Sender has enough money
// 			if user.Balance >= request.Amount {
// 				//  Check if Receiver exists
// 				var receiver hp.UserResponse
// 				err := database.ConnectMongoDB().Database(config.DB).Collection(config.USERS).FindOne(ctx, bson.M{"_id": request.Receiver}).Decode(&receiver)
// 				if err != nil {
// 					response := hp.SetError(err, "Error finding receiver")
// 					c.JSON(http.StatusBadRequest, response)
// 					return
// 				}

// 				//  Update Sender Balance
// 				_, err = database.ConnectMongoDB().Database(config.DB).Collection(config.USERS).UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$inc": bson.M{"balance": -request.Amount}})
// 				if err != nil {
// 					response := hp.SetError(err, "Error updating sender balance")
// 					c.JSON(http.StatusBadRequest, response)
// 					return
// 				}

// 				//  Update Receiver Balance
// 				_, err = database.ConnectMongoDB().Database(config.DB).Collection(config.USERS).UpdateOne(ctx, bson.M{"_id": receiver.ID}, bson.M{"$inc": bson.M{"balance": request.Amount}})
// 				if err != nil {
// 					response := hp.SetError(err, "Error updating receiver balance")
// 					c.JSON(http.StatusBadRequest, response)
// 					return
// 				}

// 				//  Create Transaction
// 				transaction := hp.Transactions{
// 					ID:       primitive.NewObjectID(),
// 					FromID:   user.ID,
// 					ToID:     receiver.ID,
// 					Amount:   request.Amount,
// 					Type:     hp.Debit,
// 					Status:   hp.TxnSuccess,
// 					Date:     time.Now(),
// 					Txn_uuid: ut.GenerateUUID(),
// 				}

// 				_, err = database.ConnectMongoDB().Database(config.DB).Collection(config.TRANSACTIONS).InsertOne(ctx, transaction)
// 				if err != nil {
// 					response := hp.SetError(err, "Error creating transaction")
// 					c.JSON(http.StatusBadRequest, response)
// 					return
// 				}
// }
