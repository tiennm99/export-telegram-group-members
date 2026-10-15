// Package store owns every Redis key the tool reads or writes. Other packages
// never import go-redis directly.
package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// keyPrefix namespaces every key. It is fixed to keep local setup to one
// REDIS_URL value.
const keyPrefix = "telegram-export"

// Store wraps the shared Redis that holds config, session, and crawl history,
// so the tool runs on any device from just a .env file.
type Store struct {
	rdb *redis.Client
}

// OpenFromEnv loads .env, then connects to REDIS_URL.
func OpenFromEnv(ctx context.Context) (*Store, error) {
	_ = godotenv.Load()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, errors.New("REDIS_URL not set in .env (e.g. rediss://default:<password>@<host>:<port>)")
	}
	return Open(ctx, redisURL)
}

// Open connects to Redis and checks the connection.
func Open(ctx context.Context, redisURL string) (*Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}
	return &Store{rdb: rdb}, nil
}

// Close releases the Redis connection.
func (s *Store) Close() error {
	return s.rdb.Close()
}

// key builds a namespaced Redis key; the single owner of the prefix.
func key(parts ...string) string {
	return strings.Join(append([]string{keyPrefix}, parts...), ":")
}
