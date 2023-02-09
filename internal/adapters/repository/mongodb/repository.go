package repository

import (
	"context"

	// "github.com/dutchapp/backend/internal/config"
	// "github.com/dutchapp/backend/internal/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo db adapter
type Mongo struct {
	client *mongo.Client
}

// NewMongo
func NewMongo() *Mongo {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		logger.Shutdown("Problem connecting to the MongoDB database:", err)
	}

	return &Mongo{
		client: client,
	}
}

// Open collection
func (m *Mongo) Open(collection string) *mongo.Collection {

	col := m.client.Database(config.DatabaseName).Collection(collection)

	return col
}

// Disconnect
func (m *Mongo) Disconnect() error {

	err := m.client.Disconnect(context.Background())

	return err
}
