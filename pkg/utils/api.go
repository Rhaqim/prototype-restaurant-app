package utils

import (
	"io"
	"net/http"
)

// Call external API
type APIMethods string

const (
	POST   APIMethods = "POST"
	GET    APIMethods = "GET"
	PUT    APIMethods = "PUT"
	DELETE APIMethods = "DELETE"
)

type BankApiStruct struct {
	Method        APIMethods
	URL           string
	Body          io.Reader
	ContentType   string
	Authorization string
}

type BankAPI interface {
	Call() (int, []byte, error)
	TransferFunds() (int, []byte, error)
}

func InitBankApi(method APIMethods, endpoint string, body io.Reader, authorization string) *BankApiStruct {
	return &BankApiStruct{
		Method:        method,
		URL:           "https://api.finicity.com/" + endpoint,
		Body:          body,
		ContentType:   "application/json",
		Authorization: "Bearer " + authorization,
	}
}

func NewBankAPI(data io.Reader) BankAPI {
	return &BankApiStruct{
		Method:        POST,
		URL:           "https://api.finicity.com/aggregation/v2/customers",
		Body:          data,
		ContentType:   "application/json",
		Authorization: "Bearer 0e" + GetEnv("FINICITY_API_KEY"),
	}
}

func (api *BankApiStruct) Call() (int, []byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(string(api.Method), api.URL, api.Body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Add("Content-Type", api.ContentType)
	req.Header.Add("Authorization", api.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, body, nil
}

func (api *BankApiStruct) TransferFunds() (int, []byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(string(api.Method), api.URL, api.Body)
	if err != nil {
		return 500, nil, err
	}
	req.Header.Add("Content-Type", api.ContentType)
	req.Header.Add("Authorization", api.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		return 500, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 500, nil, err
	}
	return resp.StatusCode, body, nil
}
