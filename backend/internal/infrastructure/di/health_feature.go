package di

import (
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	healthapp "github.com/vinaycharlie01/shroute/backend/internal/application/health"
	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
)

// newHealthHandler builds the health HTTP handler from a set of Pingers.
// Health is "special" because it needs every Pinger from every feature, so
// it's wired directly in New rather than as a standalone Feature.
func newHealthHandler(deps ...ports.Pinger) *handlers.Health {
	return handlers.NewHealth(healthapp.NewService(deps...))
}
