package redis

import (
	"os"
	"time"

	"context"

	"github.com/go-redis/redis/v8"
)

// Client is a shared Redis client.
var Client *redis.Client

// Init initializes the Redis client using env vars.
func Init() error {
	addr := os.Getenv("REDIS_ADDR")
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Client.Ping(ctx).Err()
}
