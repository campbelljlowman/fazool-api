package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"golang.org/x/exp/slog"
)

func GetRedisClient() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")

	rdb := redis.NewClient(&redis.Options{
        Addr:     redisURL,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("Error connecting to Redis", err)
		os.Exit(1)
	}

	return rdb
}