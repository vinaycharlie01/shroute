package middleware

import (
	"context"
	"net/http"

	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/logger"
)

// Recover converts a panic in any downstream handler into a 500 response
// instead of crashing the server, and logs the panic value.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverPanic(w, r.Context())
		next.ServeHTTP(w, r)
	})
}

// recoverPanic must be called directly via defer for recover() to take effect.
func recoverPanic(w http.ResponseWriter, ctx context.Context) {
	if rec := recover(); rec != nil {
		logger.FromContext(ctx).Error("panic_recovered", "panic", rec)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
