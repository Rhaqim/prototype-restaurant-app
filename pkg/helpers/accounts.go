package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/Rhaqim/thedutchapp/pkg/config"
	"github.com/Rhaqim/thedutchapp/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create a bank account struct for external API
type BankAccount struct {
	AccountName   string             `json:"account_name"`
	AccountNumber string             `json:"account_number"`
	BankCode      string             `json:"bank_code"`
	BankName      string             `json:"bank_name"`
	CreatedAt     primitive.DateTime `bson:"created_at" json:"created_at" default:"Now()"`
	UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"  default:"Now()"`
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

var wg sync.WaitGroup

func ApiCall(ctx context.Context, c chan ExternalBankAPIResponse, method string, url string, body io.Reader) {
	defer wg.Done()
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c <- ExternalBankAPIResponse{error: err}
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+utils.GetEnv("BANK_API_KEY"))
	resp, err := client.Do(req)
	if err != nil {
		c <- ExternalBankAPIResponse{error: err}
	}
	defer resp.Body.Close()
	body = io.Reader(resp.Body)
	// convert body to []byte
	b, err := io.ReadAll(body)
	if err != nil {
		c <- ExternalBankAPIResponse{error: err}
	}
	c <- ExternalBankAPIResponse{StatusCode: resp.StatusCode, Body: b}
}

type IExternalBankAPI interface {
	NewCustomer() (int, []byte, error)
	GetCustomer(customerId string) (int, []byte, error)
	NewTransaction(customerId string, data io.Reader) (int, []byte, error)
	GetTransactions(customerId string) (int, []byte, error)
}

type APIMethods string

const (
	POST   APIMethods = "POST"
	GET    APIMethods = "GET"
	PUT    APIMethods = "PUT"
	DELETE APIMethods = "DELETE"
)

type ExternalBankAPI struct {
	Method APIMethods
	URL    string
	Body   io.Reader
}

type ExternalBankAPIResponse struct {
	StatusCode int
	Body       []byte
	error      error
}

func NewExternalBankAPI(body io.Reader) *ExternalBankAPI {
	return &ExternalBankAPI{
		Method: POST,
		URL:    "https://api.paystack.co/transaction/initialize",
		Body:   body,
	}
}

func (api *ExternalBankAPI) NewCustomer(ctx context.Context) ExternalBankAPIResponse {
	var response = make(chan ExternalBankAPIResponse, 20)
	var body = io.Reader(api.Body)
	wg.Add(1)
	go ApiCall(ctx, response, string(api.Method), api.URL, body)
	wg.Wait()
	close(response)
	select {
	case <-ctx.Done():
		return ExternalBankAPIResponse{error: ctx.Err()}
	case r := <-response:
		return r
	}
}

func (api *ExternalBankAPI) GetCustomer(ctx context.Context, customerId string) ExternalBankAPIResponse {
	var response = make(chan ExternalBankAPIResponse, 20)
	var body = io.Reader(nil)
	wg.Add(1)
	go ApiCall(ctx, response, string(api.Method), api.URL+customerId, body)
	wg.Wait()
	close(response)
	select {
	case <-ctx.Done():
		return ExternalBankAPIResponse{error: ctx.Err()}
	case r := <-response:
		return r
	}
}

func (api *ExternalBankAPI) NewTransaction(ctx context.Context, customerId string, data io.Reader) (int, []byte, error) {
	return 0, nil, nil
}

func (api *ExternalBankAPI) GetTransactions(ctx context.Context, customerId string) (int, []byte, error) {
	return 0, nil, nil
}
