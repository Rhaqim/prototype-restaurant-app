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

func InitBankApi(method APIMethods, endpoint string, body io.Reader, authorization string) *BankApiStruct {
	return &BankApiStruct{
		Method:        method,
		URL:           "https://api.finicity.com/" + endpoint,
		Body:          body,
		ContentType:   "application/json",
		Authorization: "Bearer " + authorization,
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

type BankAPI interface {
	Call() (int, []byte, error)
}
