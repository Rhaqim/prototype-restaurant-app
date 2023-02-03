package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func Now(format string) time.Time {
	if format == "date" {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	return time.Now().UTC()
}

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
	if os.Getenv(key) == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}

	return os.Getenv(key)
}

// Get name of function
func GetFunctionNameV1(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Get the name of the current function
func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
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

func LoadJsonFile(path string) []byte {
	jsonFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	return jsonData
}

// Generic function to fetch data from mongoDB

func FetchDataFromMongoDB(ctx context.Context, collection *mongo.Collection, filter interface{}, result *interface{}) error {
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return err
	}
	return nil
}

// Generating random string with time and key as seed
func RandomString(length int, key string) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano() * int64(len(key))))

	var letters = []byte(letterBytes)

	b := make([]byte, length)

	for i := range b {
		b[i] = letters[seededRand.Intn(len(letters))]
	}

	return string(b)
}

func ToJSON(i interface{}) []byte {
	json, err := json.Marshal(i)
	if err != nil {
		log.Println("Error Marshalling object", err)
	}
	log.Println(string(json))
	return json
}

func ToJsonString(i interface{}) string {
	json, err := json.Marshal(i)
	if err != nil {
		log.Println("Error Marshalling object", err)
	}
	return string(json)
}

func FromJSON(data []byte, i interface{}) {
	err := json.Unmarshal(data, &i)
	if err != nil {
		log.Println("Error Unmarshalling object", err)
	}
}

// Generate 10 digit reference number with uuid
func GenerateReferenceNumber() string {
	ref := RandomString(10, GenerateUUID())

	// conver ref to uppercase
	return strings.ToUpper(ref)
}

// Slugify string
func Slugify(s string) string {
	// regex to remove all non-word characters
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return ""
	}
	// convert to lowercase
	s = strings.ToLower(s)
	// replace all non-word characters with a dash
	s = reg.ReplaceAllString(s, "-")
	// remove all dashes at the end of the string
	s = strings.TrimSuffix(s, "-")
	return s
}

/* Base API caller */
func BaseAPICaller(url string, method string, body io.Reader) (int, []byte) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println("Error creating request", err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error making request", err)
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response body", err)
	}
	return res.StatusCode, bodyBytes
}

func GetReservationTimes(openTime string, closeTime string, currentTime string) []string {

	var times []string

	// get the open time
	openTimeInt, _ := strconv.Atoi(strings.Replace(openTime, ":", "", -1))

	// get the close time
	closeTimeInt, _ := strconv.Atoi(strings.Replace(closeTime, ":", "", -1))

	// get the current time
	currentTimeInt, _ := strconv.Atoi(strings.Replace(currentTime, ":", "", -1))

	// get the reservation times
	for i := openTimeInt; i <= closeTimeInt; i += 100 {
		if i >= currentTimeInt {
			times = append(times, strconv.Itoa(i))
		}
	}

	return times
}
