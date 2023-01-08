package helpers

import (
	"context"

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
// Create a new Paystack client
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
// }

func GetWallet(ctx context.Context, filter bson.M) (Wallet, error) {
	var wallet Wallet

	err := walletCollection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		return Wallet{}, err
	}

	return wallet, nil
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

/* Budget */
var budgetCollection = config.BudgetCollection

type Budget struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id" binding:"required"`
	PurposeID primitive.ObjectID `json:"purpose_id" bson:"purpose_id" binding:"required"`
	Amount    float64            `json:"amount" bson:"amount" binding:"required"`
}

func GetBudget(ctx context.Context, filter bson.M) (Budget, error) {
	var budget Budget

	err := budgetCollection.FindOne(ctx, filter).Decode(&budget)
	if err != nil {
		return Budget{}, err
	}

	return budget, nil
}

// TODO: Logic to lock amount soecified as budget for an expense
// LockBudget and UnlockBudget
// LockBudget locks the amount specified as budget for an expense
// UnlockBudget unlocks the amount to be either transferred to wallet or refunded to user
func LockBudget(ctx context.Context, wallet Wallet, amount float64, PurposeID primitive.ObjectID) error {
	// Subtract amount from wallet balance
	wallet.Balance = wallet.Balance - amount

	_, err := walletCollection.UpdateOne(ctx, bson.M{"_id": wallet.ID}, bson.M{"$set": bson.M{"balance": wallet.Balance}})
	if err != nil {
		return err
	}

	// Create a new budget
	budget := Budget{
		ID:        primitive.NewObjectID(),
		UserID:    wallet.UserID,
		PurposeID: PurposeID,
		Amount:    amount,
	}

	_, err = budgetCollection.InsertOne(ctx, budget)
	if err != nil {
		return err
	}

	return nil
}

// UnlockBudget unlocks the amount to be either transferred to wallet or refunded to user
func UnlockBudget(ctx context.Context, event Event, user UserResponse) float64 {
	// Get budget for the event
	budget, err := GetBudget(ctx, bson.M{"purpose_id": event.ID, "user_id": user.ID})
	if err != nil {
		return 0
	}

	// Delete budget
	_, err = budgetCollection.DeleteOne(ctx, bson.M{"_id": budget.ID})
	if err != nil {
		return 0
	}

	return budget.Amount
}
