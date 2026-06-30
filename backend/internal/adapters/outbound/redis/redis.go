// Package redis provides the Redis outbound adapter: a go-redis client plus
// the ports.Pinger implementation used by the health use case. Cache
// implementations for specific use cases should live alongside this file.
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Adapter wraps a go-redis client.
type Adapter struct {
	client *redis.Client
}

// New creates a Redis adapter and verifies connectivity with a ping.
func New(ctx context.Context, addr, password string, db int) (*Adapter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return &Adapter{client: client}, nil
}

// Client exposes the underlying go-redis client for cache implementations.
func (a *Adapter) Client() *redis.Client {
	return a.client
}

// Name implements ports.Pinger.
func (a *Adapter) Name() string {
	return "redis"
}

// Ping implements ports.Pinger.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}

// Close implements ports.Closer.
func (a *Adapter) Close() error {
	return a.client.Close()
}
