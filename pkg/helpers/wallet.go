package helpers

import (
	"context"

	"github.com/Rhaqim/thedutchapp/pkg/auth"
	"github.com/Rhaqim/thedutchapp/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var walletCollection = config.WalletCollection

type Wallet struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Balance   float64            `json:"balance" bson:"balance" default:"0"`
	TxnPin    string             `json:"txn_pin" bson:"txn_pin" binding:"required,min=4,max=4"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at" default:"time.Now()"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at" default:"time.Now()"`
}

type CreateWalletRequest struct {
	TxnPin string `json:"txn_pin" bson:"txn_pin" binding:"required,min=4,max=4"`
}

type ChangePinRequest struct {
	OldPin string `json:"old_pin" bson:"old_pin" binding:"required,min=4,max=4"`
	NewPin string `json:"new_pin" bson:"new_pin" binding:"required,min=4,max=4"`
}

type FundWalletRequest struct {
	Amount float64 `json:"amount" bson:"amount"`
}

type FundWalletResponse struct {
	Amount      int    `json:"amount" bson:"amount"`
	Reference   string `json:"reference" bson:"reference"`
	Transaction string `json:"transaction" bson:"transaction"`
}

// func FundWalletPaystack(request FundWalletRequest, user UserResponse) (FundWalletResponse, error) {
// // Create a new Paystack client
// 	client := paystack.NewClient(paystack.ClientOptions{
// 		SecretKey: os.Getenv("PAYSTACK_SECRET_KEY"),
// 	})

// 	// Create a new Paystack transaction
// 	transaction := paystack.Transaction{
// 		Amount:   int(request.Amount * 100),
// 		Email:    user.Email,
// 		Metadata: map[string]interface{}{"user_id": user.ID},
// 	}

// 	// Initialize the transaction
// 	transaction.Initialize(client)

// 	// Await a response from the Paystack API
// 	paystackResponse, err := transaction.Charge()
// 	if err != nil {
// 		return FundWalletResponse{}, err
// 	}

// 	var paystackResponse FundWalletResponse

// 	return paystackResponse, nil

// Insert new Credit transaction
// transaction := hp.Transaction{
// 	ID:        primitive.NewObjectID(),
// 	From:      user.ID,
// 	To:        user.ID,
// 	Amount:    request.Amount,
// 	TransactionUID: paystackResponse.Data.Reference,
// 	Status:    hp.Start,
// 	Type:      "credit",
// 	CreatedAt: time.Now(),
// 	UpdatedAt: time.Now(),
// }

// _, err = transactionCollection.InsertOne(ctx, transaction)
// if err != nil {
// 	response := hp.SetError(err, "Error creating transaction", funcName)
// 	c.AbortWithStatusJSON(http.StatusBadRequest, response)
// 	return
// }

// Await a response from the Paystack API
// paystackResponse, err := hp.FundWalletPaystack(request, user)
// if err != nil {
//  // Update transaction status to failed
//  filter := bson.M{"_id": transaction.ID}
//  update := bson.M{"$set": bson.M{
//   "status": hp.Failed,
//  }}

//  _, err = transactionCollection.UpdateOne(ctx, filter
//  if err != nil {
//   response := hp.SetError(err, "Error updating transaction", funcName)
//   c.AbortWithStatusJSON(http.StatusBadRequest, response)
//   return
//  }

// 	response := hp.SetError(err, "Error funding wallet", funcName)
// 	c.AbortWithStatusJSON(http.StatusBadRequest, response)
// 	return
// }

// user.Wallet += request.Amount

// Update Transaction status to success
// filter := bson.M{"_id": transaction.ID}
// update := bson.M{"$set": bson.M{
//  "status": hp.Success,
// }}

// _, err = transactionCollection.UpdateOne(ctx, filter, update)
// if err != nil {
//  response := hp.SetError(err, "Error updating transaction", funcName)
//  c.AbortWithStatusJSON(http.StatusBadRequest, response)
//  return
// }

// }

func GetWallet(ctx context.Context, filter bson.M) (Wallet, error) {
	var wallet Wallet

	err := walletCollection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		return Wallet{}, err
	}

	return wallet, nil
}

func UpdateWallet(ctx context.Context, filter bson.M, update bson.M) error {
	_, err := walletCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func CheckifWalletExists(ctx context.Context, filter bson.M) (bool, error) {
	var wallet Wallet

	err := walletCollection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func VeryfyPin(ctx context.Context, user UserResponse, pin string) bool {
	funcName := "VeryfyPin"

	// Check that pin sent is correct
	// Get wallet
	filter := bson.M{
		"_id": user.Wallet,
	}
	wallet, err := GetWallet(ctx, filter)
	if err != nil {
		SetError(err, "Error fetching wallet", funcName)
		return false
	}

	// Check if pin is correct
	if !auth.CheckPasswordHash(pin, wallet.TxnPin) {
		SetError(err, "Incorrect pin", funcName)
		return false
	}

	return true
}

/* Budget */
var budgetCollection = config.BudgetCollection

type Budget struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id" binding:"required"`
	IntendedID primitive.ObjectID `json:"intended_id" bson:"intended_id" binding:"required"`
	Amount     float64            `json:"amount" bson:"amount" binding:"required"`
}

func GetBudget(ctx context.Context, filter bson.M) float64 {
	var budget Budget

	err := budgetCollection.FindOne(ctx, filter).Decode(&budget)
	if err != nil {
		return 0
	}

	return budget.Amount
}

// TODO: Logic to lock amount soecified as budget for an expense
// LockBudget and UnlockBudget
// LockBudget locks the amount specified as budget for an expense
// UnlockBudget unlocks the amount to be either transferred to wallet or refunded to user
func LockBudget(ctx context.Context, wallet Wallet, amount float64, intended_id primitive.ObjectID) error {
	// Subtract amount from wallet balance
	wallet.Balance = wallet.Balance - amount

	_, err := walletCollection.UpdateOne(ctx, bson.M{"_id": wallet.ID}, bson.M{"$set": bson.M{"balance": wallet.Balance}})
	if err != nil {
		return err
	}

	// Create a new budget
	budget := Budget{
		ID:         primitive.NewObjectID(),
		UserID:     wallet.UserID,
		IntendedID: intended_id,
		Amount:     amount,
	}

	_, err = budgetCollection.InsertOne(ctx, budget)
	if err != nil {
		return err
	}

	return nil
}

// UnlockBudget unlocks the amount to be either transferred to wallet or refunded to user
func UnlockBudget(ctx context.Context, intended_id primitive.ObjectID, user UserResponse) float64 {
	// Get budget for the event
	var amount float64

	filter := bson.M{"intended_id": intended_id, "user_id": user.ID}
	amount = GetBudget(ctx, filter)

	// Delete budget
	_, err := budgetCollection.DeleteOne(ctx, filter)
	if err != nil {
		return 0
	}

	return amount
}

func BudgetoWallet(ctx context.Context, intended_id primitive.ObjectID, user UserResponse) error {
	funcName := "BudgetoWaller"

	// Put Budget back in Wallet
	budgetAmount := UnlockBudget(ctx, intended_id, user)

	err := UpdateWallet(ctx, bson.M{"user_id": user.ID}, bson.M{"$inc": bson.M{"balance": +budgetAmount}})
	if err != nil {
		SetDebug("error updating wallet: "+err.Error(), funcName)
		return err
	}

	return nil
}

func AddMoney(ctx context.Context, user UserResponse, amount float64) error {
	funcName := "AddMoney"

	err := UpdateWallet(ctx, bson.M{"user_id": user.ID}, bson.M{"$inc": bson.M{"balance": +amount}})
	if err != nil {
		SetDebug("error updating wallet: "+err.Error(), funcName)
		return err
	}

	return nil
}
