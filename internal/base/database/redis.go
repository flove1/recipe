package database

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(connectionString string) (*redis.Client, error) {
	opts, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, fmt.Errorf("redis options parse err: %w", err)
	}

	client := redis.NewClient(opts)
	return client, nil
}
