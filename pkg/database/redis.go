package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

// RdbCon holds database connection string
var RdbCon *redis.Client

// ConnectRedis connects to the Redis database
// returns a pointer to the Redis client and an error if the connection fails
func ConnectRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
