package redis

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client

func Init() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Client.Ping(ctx).Err()
}
