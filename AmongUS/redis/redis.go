package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

// NewClient creates a new Redis client and returns it.
func NewClient() *redis.Client {
	// Initialize Redis client options
	options := &redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // Redis password
		DB:       0,                // Redis database
	}

	// Create and return the Redis client
	return redis.NewClient(options)
}
