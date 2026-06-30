package middleware

import (
	"net/http"
	"slices"
)

// CORS applies a permissive-by-allowlist CORS policy: only origins present
// in allowedOrigins receive Access-Control-Allow-Origin. An allowlist
// containing "*" allows every origin.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := slices.Contains(allowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || slices.Contains(allowedOrigins, origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+RequestIDHeader)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
