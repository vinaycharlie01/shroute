package di

import (
	"context"

	httpadapter "github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	auditapp "github.com/vinaycharlie01/shroute/backend/internal/application/audit"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
)

type auditFeature struct{}

func (auditFeature) Wire(ctx context.Context, cfg *config.Config, registry *SharedRegistry) (FeatureOutput, error) {
	if !cfg.Mongo.Enabled || registry.Mongo == nil {
		return FeatureOutput{}, nil
	}

	auditRepo := mongodb.NewAuditRepository(registry.Mongo.Database())
	auditService := auditapp.NewService(auditRepo)
	auditHandler := handlers.NewAudit(auditService)

	return FeatureOutput{
		Routes: []httpadapter.RouteRegistrar{auditHandler},
	}, nil
}
