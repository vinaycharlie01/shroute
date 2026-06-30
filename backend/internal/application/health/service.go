// Package health implements the health-check use case: it asks every
// registered outbound dependency to report liveness and aggregates the
// result into a domain.Status.
package health

import (
	"context"
	"sync"

	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	domainhealth "github.com/vinaycharlie01/shroute/backend/internal/domain/health"
)

// Service is the health-check use case. It is infrastructure-agnostic: it
// only knows about the ports.Pinger interface, not concrete adapters.
type Service struct {
	deps []ports.Pinger
}

// NewService builds a health Service over the given set of pingable
// outbound dependencies. An empty set is valid (no dependencies to check).
func NewService(deps ...ports.Pinger) *Service {
	return &Service{deps: deps}
}

// Check pings every registered dependency concurrently and returns the
// aggregated health status.
func (s *Service) Check(ctx context.Context) domainhealth.Status {
	results := make([]domainhealth.DependencyStatus, len(s.deps))

	var wg sync.WaitGroup
	for i, dep := range s.deps {
		wg.Add(1)
		go func(i int, dep ports.Pinger) {
			defer wg.Done()
			results[i] = pingOne(ctx, dep)
		}(i, dep)
	}
	wg.Wait()

	return domainhealth.Status{
		State:        domainhealth.Overall(results),
		Dependencies: results,
	}
}

func pingOne(ctx context.Context, dep ports.Pinger) domainhealth.DependencyStatus {
	if err := dep.Ping(ctx); err != nil {
		return domainhealth.DependencyStatus{
			Name:  dep.Name(),
			State: domainhealth.StateDown,
			Error: err.Error(),
		}
	}

	return domainhealth.DependencyStatus{Name: dep.Name(), State: domainhealth.StateUp}
}
