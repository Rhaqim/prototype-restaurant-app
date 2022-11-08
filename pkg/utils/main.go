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
