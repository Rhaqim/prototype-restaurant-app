package cache

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v9"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Cache variables
var (
	RedisAddr     = os.Getenv("REDIS_ADDR")
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	RedisDB, _    = strconv.Atoi(os.Getenv("REDIS_DB"))
)

// InitCache initializes the redis cache
func InitCache() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		DB:       RedisDB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error connecting to redis: %v", err)
	}

	return client
}

type Cache struct {
	client *redis.Client
	Key    string
	Value  interface{}
}

func NewCache(key string, value interface{}) *Cache {
	return &Cache{
		client: InitCache(),
		Key:    key,
		Value:  value,
	}
}

// Set sets a key-value pair in the cache
func (c *Cache) Set() error {
	defer c.client.Close()
	err := c.client.Set(context.Background(), c.Key, c.Value, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get gets a value from the cache
func (c *Cache) Get() (interface{}, error) {
	defer c.client.Close()
	val, err := c.client.Get(context.Background(), c.Key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// Get List gets a list from the cache
func (c *Cache) GetList() ([]string, error) {
	// defer c.client.Close()
	val, err := c.client.LRange(context.Background(), c.Key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}

// Set List sets a list in the cache
func (c *Cache) SetList() error {
	// defer c.client.Close()
	err := c.client.LPush(context.Background(), c.Key, c.Value).Err()
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a key-value pair from the cache
func (c *Cache) Delete() error {
	// defer c.client.Close()
	err := c.client.Del(context.Background(), c.Key).Err()
	if err != nil {
		return err
	}

	return nil
}

const ErrCacheNotFound = redis.Nil

func (c *Cache) Exists() (bool, error) {
	// defer c.client.Close()
	exists, err := c.client.Exists(context.Background(), c.Key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (c *Cache) HMSet() error {
	// defer c.client.Close()
	err := c.client.HSet(context.Background(), c.Key, c.Value).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) HGetAll() (map[string]string, error) {
	// defer c.client.Close()
	val, err := c.client.HGetAll(context.Background(), c.Key).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *Cache) SAdd() error {
	// defer c.client.Close()
	err := c.client.SAdd(context.Background(), c.Key, c.Value).Err()
	if err != nil {
		return err
	}

	return nil
}
