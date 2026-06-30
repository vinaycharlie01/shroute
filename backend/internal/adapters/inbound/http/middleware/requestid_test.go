package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/middleware"
)

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	t.Parallel()

	var captured string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = middleware.RequestIDFromContext(r.Context())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.RequestID(next).ServeHTTP(rec, req)

	if captured == "" {
		t.Fatal("expected a generated request ID in context")
	}
	if rec.Header().Get(middleware.RequestIDHeader) != captured {
		t.Errorf("response header %q = %q, want %q", middleware.RequestIDHeader, rec.Header().Get(middleware.RequestIDHeader), captured)
	}
}

func TestRequestID_ReusesInboundHeader(t *testing.T) {
	t.Parallel()

	const inboundID = "test-request-id"

	var captured string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = middleware.RequestIDFromContext(r.Context())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, inboundID)

	middleware.RequestID(next).ServeHTTP(rec, req)

	if captured != inboundID {
		t.Errorf("captured = %q, want %q", captured, inboundID)
	}
}
