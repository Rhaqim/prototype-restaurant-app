package helpers

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create a bank account struct for external API
type BankAccount struct {
	AccountNumber string             `json:"account_number"`
	BankCode      string             `json:"bank_code"`
	BankName      string             `json:"bank_name"`
	CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt" default:"Now()"`
	UpdatedAt     primitive.DateTime `bson:"updatedAt" json:"updatedAt"  default:"Now()"`
}

type BankTransfer struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

func APICallTransferFunds(ctx context.Context, txn Transactions) (int, []byte, error) {
	var fromUser UserResponse
	var toUser UserResponse

	fromUser = GetUserByID(ctx, txn.FromID)
	toUser = GetUserByID(ctx, txn.ToID)

	trans := BankTransfer{
		From:   fromUser.Account.AccountNumber,
		To:     toUser.Account.AccountNumber,
		Amount: txn.Amount,
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(trans)
	if err != nil {
		config.Logs("error", "Error encoding JSON", err)
	}

	// Api call to transfer funds
	var api utils.BankAPI = utils.NewBankAPI(&buf)

	// Check if the transaction was successful
	status, body, err := api.TransferFunds()
	if err != nil {
		txn.Status = TxnFail
	} else {
		txn.Status = TxnSuccess
	}

	return status, body, err
}
