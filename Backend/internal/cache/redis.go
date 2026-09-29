// Package cache wraps the Redis client used for caching and shared state
// across the api, worker and scheduler processes.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache is a thin JSON-oriented wrapper around a Redis client.
type Cache struct {
	client *redis.Client
	logger *zap.Logger
}

// New creates a Cache and verifies connectivity with a ping.
func New(ctx context.Context, addr, password string, db int, logger *zap.Logger) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("cache: ping redis: %w", err)
	}
	return &Cache{client: client, logger: logger}, nil
}

// Client exposes the underlying client for components that need raw access
// (for example the Redis Streams pipeline).
func (c *Cache) Client() *redis.Client { return c.client }

// GetJSON loads a JSON value. The boolean reports whether the key existed.
func (c *Cache) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	raw, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		c.logger.Warn("cache: dropping malformed value", zap.String("key", key), zap.Error(err))
		_ = c.client.Del(ctx, key)
		return false, nil
	}
	return true, nil
}

// SetJSON stores a JSON value with a TTL.
func (c *Cache) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	raw, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, raw, ttl).Err()
}

// Delete removes keys.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

// Ping checks connectivity (used by readiness probes).
func (c *Cache) Ping(ctx context.Context) error { return c.client.Ping(ctx).Err() }

// Close releases the connection pool.
func (c *Cache) Close() error { return c.client.Close() }
