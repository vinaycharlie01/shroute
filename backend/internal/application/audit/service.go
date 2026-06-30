// Package audit implements the audit-trail use case: validate and persist
// audit entries, and list recent history.
package audit

import (
	"context"
	"fmt"

	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	domainaudit "github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

// defaultListLimit caps unbounded List calls so a missing/zero limit can't
// accidentally pull the entire collection.
const defaultListLimit = 100

// Service is the audit-trail use case. It depends only on
// ports.AuditRepository, never on a concrete storage adapter.
type Service struct {
	repo ports.AuditRepository
}

// NewService builds an audit Service over the given repository.
func NewService(repo ports.AuditRepository) *Service {
	return &Service{repo: repo}
}

// Record validates and persists a new audit entry, returning the stored
// entry (with its assigned ID and timestamp).
func (s *Service) Record(ctx context.Context, e domainaudit.Entry) (domainaudit.Entry, error) {
	if err := e.Validate(); err != nil {
		return domainaudit.Entry{}, fmt.Errorf("audit: record: %w", err)
	}

	stored, err := s.repo.Append(ctx, e)
	if err != nil {
		return domainaudit.Entry{}, fmt.Errorf("audit: record: %w", err)
	}

	return stored, nil
}

// List returns the most recent audit entries, newest first. A non-positive
// limit falls back to defaultListLimit.
func (s *Service) List(ctx context.Context, limit int) ([]domainaudit.Entry, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}

	entries, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("audit: list: %w", err)
	}

	return entries, nil
}
