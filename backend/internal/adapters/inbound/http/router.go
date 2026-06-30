// Package httpadapter wires the inbound HTTP adapter: middleware chain,
// routes, and handlers. It depends only on the application layer's use
// cases, never on outbound adapters directly.
package httpadapter

import (
	"net/http"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/middleware"
)

// RouteRegistrar is the interface handlers implement to self-register their
// HTTP routes on the mux. Any handler that implements this interface can be
// plugged into the router without changing RouterConfig.
type RouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

// RouterConfig carries the dependencies needed to build the router.
type RouterConfig struct {
	Routes         []RouteRegistrar
	AllowedOrigins []string
}

// NewRouter builds the application's http.Handler: middleware chain (request
// ID, logging, recovery, CORS) wrapping the route table.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	for _, r := range cfg.Routes {
		r.RegisterRoutes(mux)
	}

	var handler http.Handler = mux
	handler = middleware.CORS(cfg.AllowedOrigins)(handler)
	handler = middleware.Recover(handler)
	handler = middleware.Logging()(handler)
	handler = middleware.RequestID(handler)

	return handler
}
