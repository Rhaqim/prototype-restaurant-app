package helpers

// Create a bank account struct for external API
type BankAccount struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

// Create a bank account struct for internal use
type BankAccountInternal struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	BankName      string `json:"bank_name"`
}

// Update a bank account struct for external API
type BankAccountUpdate struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}
