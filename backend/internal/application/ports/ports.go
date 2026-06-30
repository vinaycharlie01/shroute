// Package ports declares the outbound interfaces the application core
// depends on. Concrete implementations live in internal/adapters/outbound;
// the application layer never imports them directly, only these interfaces.
package ports

import (
	"context"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
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
type AuditRepository interface {
	Append(ctx context.Context, e audit.Entry) (audit.Entry, error)
	List(ctx context.Context, limit int) ([]audit.Entry, error)
}
