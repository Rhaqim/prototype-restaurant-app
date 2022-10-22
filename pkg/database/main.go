package database

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI is not set")
	}

	log.Println("Connecting to the database...")

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatal("Problem connecting to the MongoDB database:", err)
	}

	return client
}

func OpenCollection(client *mongo.Client, databaseName string, collectionName string) *mongo.Collection {
	return client.Database(databaseName).Collection(collectionName)
}
