package di

import (
	"context"
	"fmt"

	httpadapter "github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	settingsapp "github.com/vinaycharlie01/shroute/backend/internal/application/settings"
	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
)

type settingsFeature struct{}

func (settingsFeature) Wire(ctx context.Context, cfg *config.Config, registry *SharedRegistry) (FeatureOutput, error) {
	if !cfg.Mongo.Enabled || registry.Mongo == nil {
		return FeatureOutput{}, nil
	}

	db := registry.Mongo.Database()
	settingsRepo := mongodb.NewSettingsRepository(db)
	flagsRepo := mongodb.NewFeatureFlagRepository(db)

	if err := flagsRepo.SeedDefaults(ctx, domainsettings.DefaultFeatureFlags); err != nil {
		return FeatureOutput{}, fmt.Errorf("di: settings: seed defaults: %w", err)
	}

	svc := settingsapp.NewService(settingsRepo, flagsRepo)
	handler := handlers.NewSettings(svc)

	return FeatureOutput{
		Routes: []httpadapter.RouteRegistrar{handler},
	}, nil
}
