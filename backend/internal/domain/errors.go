// Package domain contains framework-agnostic business entities and rules
// shared across the application core. It must not import anything from
// adapters or infrastructure.
package domain

import "errors"

// Sentinel domain errors. Adapters translate these into transport-specific
// representations (HTTP status codes, gRPC codes, etc).
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrUnavailable  = errors.New("dependency unavailable")
)
