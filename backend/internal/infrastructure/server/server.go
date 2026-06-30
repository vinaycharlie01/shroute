// Package server wraps net/http.Server with graceful shutdown driven by a
// context cancellation (typically tied to OS signals in cmd/server/main.go).
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server wraps a standard http.Server.
type Server struct {
	httpServer      *http.Server
	log             *slog.Logger
	shutdownTimeout time.Duration
}

// Config carries the parameters needed to construct a Server.
type Config struct {
	Host            string
	Port            int
	Handler         http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Logger          *slog.Logger
}

// New builds a Server from cfg.
func New(cfg Config) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port)),
			Handler:      cfg.Handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
		log:             cfg.Logger,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run starts the server and blocks until ctx is canceled, at which point it
// attempts a graceful shutdown bounded by shutdownTimeout. It returns nil on
// a clean shutdown, or the first error encountered otherwise.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http_server_starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server: listen and serve: %w", err)

			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("http_server_shutting_down")
		// context.Background() is intentional here: ctx is already done
		// (that's why we're in this branch), so deriving the shutdown
		// deadline from it would cancel immediately and skip the grace
		// period entirely.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		//nolint:contextcheck // shutdownCtx is intentionally derived from context.Background(), see comment above
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server: shutdown: %w", err)
		}

		s.log.Info("http_server_stopped")

		return nil
	}
}
