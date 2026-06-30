// Package cache holds the application service that orchestrates cache
// management operations. It depends only on the CacheStore port, not on
// any concrete outbound adapter.
package cache

import (
	"context"
	"fmt"

	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	domaincache "github.com/vinaycharlie01/shroute/backend/internal/domain/cache"
)

// Service provides cache management use cases: stats, listing, and flush.
type Service struct {
	store ports.CacheStore
}

// NewService builds a cache service backed by the given store.
func NewService(store ports.CacheStore) *Service {
	return &Service{store: store}
}

// Stats returns aggregate cache statistics.
func (s *Service) Stats(ctx context.Context) (domaincache.Stats, error) {
	st, err := s.store.Stats(ctx)
	if err != nil {
		return domaincache.Stats{}, fmt.Errorf("cache: Stats: %w", err)
	}

	return st, nil
}

// List returns up to limit entries matching the given key prefix. If prefix
// is empty the caller must explicitly set all true to avoid accidentally
// listing the entire cache; otherwise List returns ErrNoPrefix.
func (s *Service) List(ctx context.Context, prefix string, limit int) ([]domaincache.Entry, error) {
	if prefix == "" {
		return nil, ErrNoPrefix
	}

	entries, err := s.store.List(ctx, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("cache: List: %w", err)
	}

	return entries, nil
}

// Flush removes every key matching the given prefix. An empty prefix is
// rejected unless all is explicitly true, to prevent accidental full flushes.
func (s *Service) Flush(ctx context.Context, prefix string) error {
	if prefix == "" {
		return ErrNoPrefix
	}

	if err := s.store.Flush(ctx, prefix); err != nil {
		return fmt.Errorf("cache: Flush: %w", err)
	}

	return nil
}
