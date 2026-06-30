// Package ports declares the outbound interfaces the application core
// depends on. Concrete implementations live in internal/adapters/outbound;
// the application layer never imports them directly, only these interfaces.
package ports

import "context"

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
