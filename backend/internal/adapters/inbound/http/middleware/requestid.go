package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

// RequestIDKey is the context key under which the per-request ID is stored.
const RequestIDKey contextKey = "request_id"

// RequestIDHeader is the response/request header carrying the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestID assigns a unique ID to every request (reusing an inbound
// X-Request-ID header when present) and stores it on the request context so
// downstream handlers and the logging middleware can correlate log lines.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts the request ID stored by RequestID, if any.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)

	return id
}
