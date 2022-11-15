package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

// Create a bank account struct for external API
type BankAccount struct {
	AccountNumber string             `json:"account_number"`
	BankCode      string             `json:"bank_code"`
	BankName      string             `json:"bank_name"`
	CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt" default:"Now()"`
	UpdatedAt     primitive.DateTime `bson:"updatedAt" json:"updatedAt"  default:"Now()"`
}
