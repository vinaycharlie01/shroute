// Package postgres provides the Postgres outbound adapter: a pgx connection
// pool plus the ports.Pinger implementation used by the health use case.
// Repository implementations for specific aggregates should live alongside
// this file as the domain grows.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Adapter wraps a pgx connection pool.
type Adapter struct {
	pool *pgxpool.Pool
}

// New creates a Postgres adapter and verifies connectivity with a ping.
func New(ctx context.Context, dsn string) (*Adapter, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return &Adapter{pool: pool}, nil
}

// Pool exposes the underlying pgx pool for repository implementations.
func (a *Adapter) Pool() *pgxpool.Pool {
	return a.pool
}

// Name implements ports.Pinger.
func (a *Adapter) Name() string {
	return "postgres"
}

// Ping implements ports.Pinger.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.pool.Ping(ctx)
}

// Close implements ports.Closer.
func (a *Adapter) Close() error {
	a.pool.Close()
	return nil
}
