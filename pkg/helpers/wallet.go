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
	Balance   float64            `json:"balance" bson:"balance"`
	TxnPin    string             `json:"txn_pin" bson:"txn_pin"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
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
