// Package httpadapter wires the inbound HTTP adapter: middleware chain,
// routes, and handlers. It depends only on the application layer's use
// cases, never on outbound adapters directly.
package httpadapter

import (
	"log/slog"
	"net/http"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/middleware"
)

// RouterConfig carries the dependencies needed to build the router.
type RouterConfig struct {
	Logger         *slog.Logger
	Health         *handlers.Health
	AllowedOrigins []string
}

// NewRouter builds the application's http.Handler: middleware chain (request
// ID, logging, recovery, CORS) wrapping the route table.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", cfg.Health.Live)
	mux.HandleFunc("GET /readyz", cfg.Health.Ready)

	var handler http.Handler = mux
	handler = middleware.CORS(cfg.AllowedOrigins)(handler)
	handler = middleware.Recover(handler)
	handler = middleware.Logging(cfg.Logger)(handler)
	handler = middleware.RequestID(handler)

	return handler
}
