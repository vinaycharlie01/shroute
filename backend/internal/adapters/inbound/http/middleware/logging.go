package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/logger"
)

// statusRecorder captures the status code written by the handler so it can
// be logged after the request completes.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Logging logs one structured line per request: method, path, status,
// duration, and request ID. It attaches a request-scoped logger (carrying
// the request ID) to the request context for downstream handlers to use via
// logger.FromContext.
func Logging(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := RequestIDFromContext(r.Context())

			reqLog := base.With("request_id", requestID)
			ctx := logger.WithContext(r.Context(), reqLog)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			reqLog.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
