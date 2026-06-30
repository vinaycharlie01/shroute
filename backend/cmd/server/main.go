// Command server is the OmniRouter backend entrypoint: it loads
// configuration, wires the dependency-injection container, and runs the
// HTTP server until an OS signal requests a graceful shutdown.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/di"
	"github.com/vinaycharlie01/shroute/backend/internal/version"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configDir := envOr("APP_CONFIG_DIR", "config")
	dotenvPath := envOr("APP_DOTENV_PATH", ".env")

	cfg, err := config.Load(configDir, dotenvPath, os.Getenv("APP_ENV"))
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	container, err := di.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			slog.Error("shutdown_cleanup_failed", "error", closeErr)
		}
	}()

	slog.Info("starting",
		"env", cfg.Env,
		"version", version.Version,
		"commit", version.Commit,
		"build_date", version.BuildDate,
	)

	return container.Server.Run(ctx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
