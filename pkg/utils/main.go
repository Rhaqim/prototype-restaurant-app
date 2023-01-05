package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"reflect"
	"runtime"
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
