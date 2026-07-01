// Package ports declares the outbound interfaces the application core
// depends on. Concrete implementations live in internal/adapters/outbound;
// the application layer never imports them directly, only these interfaces.
//
//go:generate counterfeiter -generate
package ports

import (
	"context"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
	"github.com/vinaycharlie01/shroute/backend/internal/domain/cache"
	"github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

// Pinger is implemented by any outbound dependency adapter that can report
// liveness (database pool, cache client, etc). Health checks depend only on
// this interface, never on a concrete driver.
type Pinger interface {
	Name() string
	Ping(ctx context.Context) error
}

// Closer is implemented by adapters that hold resources requiring an
// explicit shutdown (connection pools, clients).
type Closer interface {
	Close() error
}

// AuditRepository persists and retrieves audit trail entries. The
// application layer depends only on this interface; the concrete
// implementation lives in adapters/outbound/mongodb.
//
//counterfeiter:generate . AuditRepository
type AuditRepository interface {
	Append(ctx context.Context, e audit.Entry) (audit.Entry, error)
	List(ctx context.Context, limit int) ([]audit.Entry, error)
}

// CacheStore provides cache management operations backed by Redis.
// Stats returns aggregate statistics, List returns recent entries by
// prefix, and Flush removes every key matching the prefix.
//
//counterfeiter:generate . CacheStore
type CacheStore interface {
	Stats(ctx context.Context) (cache.Stats, error)
	List(ctx context.Context, prefix string, limit int) ([]cache.Entry, error)
	Flush(ctx context.Context, prefix string) error
}

// SettingsRepository persists and retrieves settings documents keyed by a
// well-known string key.
//
//counterfeiter:generate . SettingsRepository
type SettingsRepository interface {
	Get(ctx context.Context, key string) (settings.Document, error)
	Set(ctx context.Context, doc settings.Document) (settings.Document, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]settings.Document, error)
}

// FeatureFlagRepository persists and retrieves feature flag state.
// SeedDefaults inserts defaults using $setOnInsert semantics so existing
// operator-modified values are never overwritten.
//
//counterfeiter:generate . FeatureFlagRepository
type FeatureFlagRepository interface {
	Get(ctx context.Context, name string) (settings.FeatureFlag, error)
	List(ctx context.Context) ([]settings.FeatureFlag, error)
	Set(ctx context.Context, flag settings.FeatureFlag) (settings.FeatureFlag, error)
	SeedDefaults(ctx context.Context, defaults []settings.FeatureFlag) error
}
