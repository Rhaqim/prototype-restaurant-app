package helpers

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
