package utils

import (
	"io"
	"net/http"
	"reflect"
	"runtime"

	"github.com/google/uuid"
)

// generate random uuid
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

// Call external API
func CallAPI(url string, method string, body io.Reader) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, nil
}

// Get name of function
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
