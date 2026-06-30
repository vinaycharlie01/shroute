package di

import (
	"context"

	httpadapter "github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
)

// SharedRegistry holds adapters that may be shared across features.
// A feature reads the adapter it needs; nil means the adapter was not
// configured / started.
type SharedRegistry struct {
	Mongo *mongodb.Adapter
	Redis *redis.Adapter
}

// FeatureOutput is the set of components a feature contributes back to the
// Container. Routes are registered on the HTTP mux, Closers are released on
// shutdown, and Pingers feed the health-check feature.
type FeatureOutput struct {
	Routes  []httpadapter.RouteRegistrar
	Closers []ports.Closer
	Pingers []ports.Pinger
}

// Feature is the interface every pluggable feature module implements. Each
// feature owns its own wire-up (outbound adapter → application service →
// inbound handler → routes) so that New [container.go] never needs to change
// when a feature is added or removed.
type Feature interface {
	Wire(ctx context.Context, cfg *config.Config, registry *SharedRegistry) (FeatureOutput, error)
}
