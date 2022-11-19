package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
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
		fmt.Printf("Error connecting to Redis: %v\n", err)
		os.Exit(1)
	}

	return rdb
}