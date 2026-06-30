package di

import (
	"context"

	httpadapter "github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	rediscache "github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
	cacheapp "github.com/vinaycharlie01/shroute/backend/internal/application/cache"
	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
)

type cacheFeature struct{}

func (cacheFeature) Wire(ctx context.Context, cfg *config.Config, registry *SharedRegistry) (FeatureOutput, error) {
	if !cfg.Redis.Enabled || registry.Redis == nil {
		return FeatureOutput{}, nil
	}

	cacheStore := rediscache.NewCacheStore(registry.Redis)
	cacheService := cacheapp.NewService(cacheStore)
	cacheHandler := handlers.NewCache(cacheService)

	return FeatureOutput{
		Routes:  []httpadapter.RouteRegistrar{cacheHandler},
		Pingers: []ports.Pinger{},
		Closers: []ports.Closer{},
	}, nil
}
