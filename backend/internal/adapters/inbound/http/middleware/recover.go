package middleware

import (
	"net/http"

	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/logger"
)

// Recover converts a panic in any downstream handler into a 500 response
// instead of crashing the server, and logs the panic value.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.FromContext(r.Context()).Error("panic_recovered", "panic", rec)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
