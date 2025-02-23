package worker

import (
	"log"
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx = context.Background()
)

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Connected to Redis")
}