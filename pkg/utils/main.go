package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime"

	"github.com/google/uuid"
)

// generate random uuid
func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

func StructToJSON(i interface{}) io.Reader {
	json, _ := json.Marshal(i)
	return bytes.NewReader(json)
}

func GetEnv(key string) string {
	if key == "PORT" {
		if os.Getenv(key) == "" {
			return ":8080"
		}
	}
	return os.Getenv(key)
}

// Get name of function
func GetFunctionNameV1(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Colorise text in the terminal
func Colorise(color string, text string) string {
	colors := map[string]string{
		"reset":   "\033[0m",
		"black":   "\033[30m",
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
	}
	return colors[color] + text + colors["reset"]
}

// Get the name of the current function
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

// Signup godoc
// @Summary Create a new account
// @Description Create a new account
// @Tags auth
// @Accept  json
// @Produce  json
// @Param account body hp.Account true "Account"
// @Success 200 {object} hp.Account
// @Failure 400 {object} hp.Error
// @Failure 500 {object} hp.Error
// @Router /auth/signup [post]

// func GetSocial(c *gin.Context) {
// 	collection := config.MI.DB.Collection("socials")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	var socials []models.Social
// 	cur, err := collection.Find(ctx, bson.M{})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while getting socials", "GetSocial"))
// 		return
// 	}
// 	defer cur.Close(ctx)
// 	for cur.Next(ctx) {
// 		var social models.Social
// 		err := cur.Decode(&social)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while decoding socials", "GetSocial"))
// 			return
// 		}
// 		socials = append(socials, social)
// 	}
// 	if err := cur.Err(); err != nil {
// 		c.JSON(http.StatusInternalServerError, helpers.SetError(err, "Error while getting socials", "GetSocial"))
// 		return
// 	}
// 	c.JSON(http.StatusOK, helpers.SetSuccess("Socials found", socials, "GetSocial"))
// }

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
