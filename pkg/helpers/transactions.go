package helpers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	ut "github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var transactionCollection = config.TransactionCollection

// unique id for transactions
var TransactionUID = "TC-" + ut.GenerateReferenceNumber()

type TxnType string

type TxnStatus string

const (
	Debit  TxnType = "debit"
	Credit TxnType = "credit"
)

const (
	TxnStart   TxnStatus = "start"
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

type SendMoneyOtherUser struct {
	Username string  `json:"username" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	TxnPin   string  `json:"txn_pin" binding:"required"`
}

type SenfMoneyOtherBank struct {
	AccountNumber string  `json:"account_number" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	TxnPin        string  `json:"txn_pin" binding:"required"`
	BankCode      string  `json:"bank_code" binding:"required"`
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
// TODO: have it return the transaction and an error
func UpdateSenderTransaction(ctx context.Context, user UserResponse, amount float64, txn Transactions) (Transactions, error) {
	funcName := ut.GetFunctionName()

	if txn.Status != TxnStart {
		return txn, errors.New("transaction status is not start")
	}

	// Get Money from the budget
	budgetAmount := UnlockBudget(ctx, txn.ToID, user)
	SetInfo(fmt.Sprintf("Unlocked budget amount: %f", budgetAmount), funcName)

	amount = amount - budgetAmount

	// Update the wallet to get the rest from the wallet or return the rest to the wallet
	filter := bson.M{"user_id": user.ID}
	update := bson.M{"$inc": bson.M{
		"balance": -amount,
	}}

	_, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		SetDebug("error updating sender wallet balance: "+err.Error(), funcName)
		return txn, err
	}

	// Update transaction status to pending
	txn, err = UpdateAndReturnTransaction(ctx, txn, TxnPending)
	if err != nil {
		SetDebug("error updating transaction status: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

// UpdateReceiverTransaction updates the receiver wallet balance
// It adds the amount to the receiver wallet balance
// It returns true if successful
func UpdateReceiverTransaction(ctx context.Context, to_id primitive.ObjectID, amount float64, txn Transactions) bool {
	funcName := ut.GetFunctionName()

	if txn.Status != TxnPending {
		SetDebug("transaction status is not pending", funcName)
		return false
	}

	filter := bson.M{"user_id": to_id}
	update := bson.M{"$inc": bson.M{
		"balance": +amount,
	}}

	updateResult, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		SetDebug("error updating receiver wallet balance: "+err.Error(), funcName)
		return false
	}
	return updateResult.ModifiedCount == 1
}

// UpdateAndReturnTransaction updates the transaction status and returns the updated transaction
func UpdateAndReturnTransaction(ctx context.Context, txn Transactions, status TxnStatus) (Transactions, error) {
	funcName := ut.GetFunctionName()

	// Update transaction status to success
	filter := bson.M{"transaction_uid": txn.TransactionUID, "_id": txn.ID}
	update := bson.M{"$set": bson.M{
		"status": status,
	}}
	// update and return new document
	updateResult, errs := transactionCollection.UpdateOne(ctx, filter, update)
	if errs != nil {
		SetDebug("error updating transaction for success: "+errs.Error(), funcName)
		return txn, errs
	}
	if updateResult.ModifiedCount != 1 {
		return txn, errors.New("error updating transaction for success")
	}

	txn, errs = GetTransaction(ctx, filter)
	if errs != nil {
		SetDebug("error getting transaction for success: "+errs.Error(), funcName)
		return txn, errs
	}

	return txn, nil
}

// UpdateWalletBalance updates the wallet balance of the sender and receiver
// If an error occurs, it rolls back the transaction and updates the transaction status to fail
// If the transaction is successful, it updates the transaction status to success
func UpdateWalletBalance(ctx context.Context, txn Transactions) (Transactions, error) {
	funcName := ut.GetFunctionName()

	fromUser := GetUserByID(ctx, txn.FromID)

	txn, err := UpdateSenderTransaction(ctx, fromUser, txn.Amount, txn)
	if err != nil {

		SetDebug("error updating sender transaction: "+err.Error(), funcName)

		_, _ = UpdateAndReturnTransaction(ctx, txn, TxnFail)

		return txn, err
	}

	if !UpdateReceiverTransaction(ctx, txn.ToID, txn.Amount, txn) {
		// Rollback to Sender
		SetDebug("error updating receiver transaction", funcName)

		_, _ = UpdateSenderTransaction(ctx, fromUser, -txn.Amount, txn)

		txn, err := UpdateAndReturnTransaction(ctx, txn, TxnFail)
		if err != nil {
			return txn, err
		}

		return txn, errors.New("error updating receiver transaction")
	}

	txn, err = UpdateAndReturnTransaction(ctx, txn, TxnSuccess)
	if err != nil {
		SetDebug("error updating transaction for success: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

// StartDebitTransaction starts a debit transaction
// It creates a transaction with the sender, receiver and amount
// It stores the transaction in the database with a status of start
// Then it updates the wallet balance of the sender and receiver
func startDebitTransaction(from, to primitive.ObjectID, amount float64) (Transactions, error) {
	funcName := ut.GetFunctionName()

	ctx := context.Background()

	// create transaction
	txn := Transactions{
		ID:             primitive.NewObjectID(),
		TransactionUID: TransactionUID,
		FromID:         from,
		ToID:           to,
		Amount:         amount,
		Type:           Debit,
		Status:         TxnStart,
		CreatedAt:      primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:      primitive.NewDateTimeFromTime(time.Now()),
	}

	// insert transaction
	txn, err := InsertTransaction(ctx, txn)
	if err != nil {
		SetDebug("error inserting transaction: "+err.Error(), funcName)
		return txn, err
	}

	// update wallet balance
	txn, err = UpdateWalletBalance(ctx, txn)
	if err != nil {
		SetDebug("error updating wallet balance: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

/* Event Transaction */
type EventBillPayment struct {
	EventID primitive.ObjectID `json:"event_id" bson:"event_id" binding:"required"`
	TxnPin  string             `json:"txn_pin" bson:"txn_pin" binding:"required"`
}

// SendtoVenuePayforEvent sends money to venues
// Anyone can pay for the total bill of an event
// It returns the transaction and error if any
// It returns error if the user has insufficient balance
// It returns error if the event is not found
func SendtoVenuePayforEvent(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	funcName := ut.GetFunctionName()

	var txn Transactions

	// Get Venue Owner
	restaurant, err := GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		SetDebug("error getting restaurant: "+err.Error(), funcName)
		return txn, err
	}

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, event.Bill) {
		return txn, errors.New("insufficient balance")
	}

	// start debit transaction
	// Send Money to Venue Owner
	txn, err = startDebitTransaction(user.ID, restaurant.OwnerID, event.Bill)
	if err != nil {
		SetDebug("error starting debit transaction: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

// SendToHost sends money to the host of the event
// It takes a context and an event and a user
// It gets the total bill from orders made for the event
// and sends the money to the host
// Get Unlocks the budget for the user and adds it to the wallet balance
// It returns a transaction and an error
func SendToHost(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	funcName := ut.GetFunctionName()

	// Get Venue Owner
	restaurant, err := GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		SetDebug("error getting restaurant: "+err.Error(), funcName)
		return Transactions{}, err
	}

	var orders []Order

	// Get total bill from orders
	orders, err = GetOrders(ctx, bson.M{"event_id": event.ID, "customer_id": user.ID, "paid": false})
	if err != nil {
		SetDebug("error getting orders: "+err.Error(), funcName)
		return Transactions{}, err
	}

	var totalBill float64
	for _, order := range orders {
		totalBill += order.Bill
	}

	SetInfo(fmt.Sprintf("total bill: %f", totalBill), funcName)

	err = BudgetoWallet(ctx, restaurant.OwnerID, user)
	if err != nil {
		SetDebug("error returning budget: "+err.Error(), funcName)
		return Transactions{}, err
	}

	// Check if totalBill is greater than 0
	if totalBill <= 0 {
		SetDebug("total bill is less than or equal to 0", funcName)
		return Transactions{}, errors.New("there are no orders to pay for")
	}

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, totalBill) {
		SetDebug("insufficient balance", funcName)
		return Transactions{}, errors.New("insufficient balance")
	}

	// Start Debit Transaction
	// Send money to Host of Event
	txn, err := startDebitTransaction(user.ID, event.HostID, totalBill)
	if err != nil {
		SetDebug("error starting debit transaction: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

// SendMoneyPayOwnBill sends money to venue to pay for own bill
// It takes a context, the user and the event
// It returns a transaction and an error
func PayOwnBillforEvent(ctx context.Context, event Event, user UserResponse) (Transactions, error) {
	funcName := ut.GetFunctionName()

	var txn Transactions

	// Get Venue Owner
	restaurant, err := GetRestaurant(ctx, bson.M{"_id": event.Venue})
	if err != nil {
		SetDebug("error getting restaurant: "+err.Error(), funcName)
		return txn, err
	}

	var orders []Order

	// Get total bill from orders
	orders, err = GetOrders(ctx, bson.M{"event_id": event.ID, "customer_id": user.ID, "paid": false})
	if err != nil {
		SetDebug("error getting orders: "+err.Error(), funcName)
		return Transactions{}, err
	}

	var totalBill float64
	for _, order := range orders {
		totalBill += order.Bill
	}

	SetInfo(fmt.Sprintf("total bill: %f", totalBill), funcName)

	err = BudgetoWallet(ctx, restaurant.OwnerID, user)
	if err != nil {
		SetDebug("error returning budget: "+err.Error(), funcName)
		return Transactions{}, err
	}

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, user, totalBill) {
		SetDebug("insufficient balance", funcName)
		return Transactions{}, errors.New("insufficient balance")
	}

	// start debit transaction
	// Send Money to Venue Owner
	txn, err = startDebitTransaction(user.ID, restaurant.OwnerID, totalBill)
	if err != nil {
		SetDebug("error starting debit transaction: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

// SendToOtherUsers sends money to other users
// It takes a context, the user the money is being sent to and the user sending the money
// It returns a transaction and an error
func SendToOtherUsers(ctx context.Context, toUser UserResponse, fromUser UserResponse, amount float64) (Transactions, error) {
	funcName := ut.GetFunctionName()

	// check if user has sufficient balance
	if !VerifyWalletSufficientBalance(ctx, fromUser, amount) {
		return Transactions{}, errors.New("insufficient balance")
	}

	// start debit transaction
	//Send Money to User
	txn, err := startDebitTransaction(fromUser.ID, toUser.ID, amount)
	if err != nil {
		SetDebug("error starting debit transaction: "+err.Error(), funcName)
		return txn, err
	}

	return txn, nil
}

func VerificationforEventPayment(ctx context.Context, request EventBillPayment, event Event, user UserResponse) error {
	funcName := ut.GetFunctionName()

	// Check if Event is ongoing
	if event.EventStatus != Ongoing {
		SetError(nil, "Event is not ongoing", funcName)

		return errors.New("event is not ongoing")
	}

	//verify pin
	if !VeryfyPin(ctx, user, request.TxnPin) {
		SetError(nil, "Invalid Pin", funcName)

		return errors.New("invalid pin")
	}

	return nil
}
