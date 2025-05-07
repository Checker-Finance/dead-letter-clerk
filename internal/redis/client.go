package redis

import (
	"context"
	"dead-letter-clerk/internal/config"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	Client *goredis.Client
}

// NewClient initializes and returns a Redis client based on config.
func NewClient(cfg config.RedisConfig) *Client {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Client{Client: rdb}
}

// GetString retrieves a string value from Redis.
func (r *Client) GetString(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", nil // key does not exist
	}
	return val, err
}

// SetString stores a string value in Redis.
func (r *Client) SetString(ctx context.Context, key string, value string) error {
	return r.Client.Set(ctx, key, value, 0).Err()
}

// Close closes the Redis connection.
func (r *Client) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return fmt.Errorf("redis client not initialized")
}
