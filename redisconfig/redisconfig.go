package redisconfig

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var RedisClient *redis.Client

func ConnectRedis() *redis.Client {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// test connection;
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect the redis : %w", err))
	}
	log.Println("Connected to redis successfully")
	RedisClient = rdb
	return rdb
}

func getEnv(key string, defaultValue string) string {
	value, exist := os.LookupEnv(key)
	if exist {
		return value
	}
	return defaultValue
}
