// Package di wires together the application's domain, application, and
// adapter layers into a runnable Container. It is the only place in the
// codebase allowed to import every layer at once.
package di

import (
	"context"
	"fmt"
	"log/slog"

	httpadapter "github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/postgres"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/redis"
	healthapp "github.com/vinaycharlie01/shroute/backend/internal/application/health"
	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/logger"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/server"
)

// Container holds every wired component needed to run the service.
type Container struct {
	Config *config.Config
	Logger *slog.Logger
	Server *server.Server

	closers []ports.Closer
}

// New builds a fully wired Container from cfg: outbound adapters (only the
// ones enabled in cfg), application services, the inbound HTTP router, and
// the HTTP server.
func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	log := logger.New(logger.Config{Level: cfg.Log.Level, Format: cfg.Log.Format})

	c := &Container{Config: cfg, Logger: log}

	var deps []ports.Pinger

	if cfg.Postgres.Enabled {
		pg, err := postgres.New(ctx, cfg.Postgres.DSN)
		if err != nil {
			return nil, fmt.Errorf("di: postgres: %w", err)
		}
		deps = append(deps, pg)
		c.closers = append(c.closers, pg)
	}

	if cfg.Redis.Enabled {
		rd, err := redis.New(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			return nil, fmt.Errorf("di: redis: %w", err)
		}
		deps = append(deps, rd)
		c.closers = append(c.closers, rd)
	}

	healthService := healthapp.NewService(deps...)
	healthHandler := handlers.NewHealth(healthService)

	router := httpadapter.NewRouter(httpadapter.RouterConfig{
		Logger:         log,
		Health:         healthHandler,
		AllowedOrigins: cfg.Server.AllowedOrigins,
	})

	c.Server = server.New(server.Config{
		Host:            cfg.Server.Host,
		Port:            cfg.Server.Port,
		Handler:         router,
		ReadTimeout:     cfg.Server.ReadTimeout.Duration(),
		WriteTimeout:    cfg.Server.WriteTimeout.Duration(),
		ShutdownTimeout: cfg.Server.ShutdownTimeout.Duration(),
		Logger:          log,
	})

	return c, nil
}

// Close releases every resource held by outbound adapters (connection
// pools, clients), in reverse wiring order.
func (c *Container) Close() error {
	var firstErr error
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i].Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
