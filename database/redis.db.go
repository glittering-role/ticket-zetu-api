package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"ticket-zetu-api/config"
)

var (
	redisClient *redis.Client
)

// InitRedis initializes the Redis client
func InitRedis() {
	appConfig := config.LoadConfig()

	redisClient = redis.NewClient(&redis.Options{
		Addr:         appConfig.RedisAddr,
		Password:     appConfig.RedisPassword,
		DB:           appConfig.RedisDB,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", appConfig.RedisAddr, err)
	}
	log.Printf("Successfully connected to Redis at %s", appConfig.RedisAddr)
}

// GetRedisClient returns the Redis client
func GetRedisClient() *redis.Client {
	return redisClient
}

func SetRedisClient(client *redis.Client) {
	redisClient = client
}

func CloseRedis(client *redis.Client) {
	if err := client.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}
}
