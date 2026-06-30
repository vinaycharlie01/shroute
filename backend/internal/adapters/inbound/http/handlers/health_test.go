package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	domainhealth "github.com/vinaycharlie01/shroute/backend/internal/domain/health"
)

type stubChecker struct {
	status domainhealth.Status
}

func (s stubChecker) Check(_ context.Context) domainhealth.Status { return s.status }

func TestHealth_Live(t *testing.T) {
	t.Parallel()

	h := handlers.NewHealth(stubChecker{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil)

	h.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "up" {
		t.Errorf("status field = %v, want %q", body["status"], "up")
	}
}

func TestHealth_Ready(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     domainhealth.Status
		wantCode   int
		wantStatus string
	}{
		{
			name:       "up",
			status:     domainhealth.Status{State: domainhealth.StateUp},
			wantCode:   http.StatusOK,
			wantStatus: "up",
		},
		{
			name: "degraded still returns 200",
			status: domainhealth.Status{
				State: domainhealth.StateDegraded,
				Dependencies: []domainhealth.DependencyStatus{
					{Name: "redis", State: domainhealth.StateDown, Error: "boom"},
				},
			},
			wantCode:   http.StatusOK,
			wantStatus: "degraded",
		},
		{
			name:       "down returns 503",
			status:     domainhealth.Status{State: domainhealth.StateDown},
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: "down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handlers.NewHealth(stubChecker{status: tt.status})
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil)

			h.Ready(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantCode)
			}

			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["status"] != tt.wantStatus {
				t.Errorf("status field = %v, want %q", body["status"], tt.wantStatus)
			}
		})
	}
}
