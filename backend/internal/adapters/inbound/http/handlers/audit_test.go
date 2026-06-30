package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	domainaudit "github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

type stubAuditRecorder struct {
	recordErr error
	listErr   error
	stored    domainaudit.Entry
	entries   []domainaudit.Entry
}

func (s stubAuditRecorder) Record(_ context.Context, _ domainaudit.Entry) (domainaudit.Entry, error) {
	if s.recordErr != nil {
		return domainaudit.Entry{}, s.recordErr
	}

	return s.stored, nil
}

func (s stubAuditRecorder) List(_ context.Context, _ int) ([]domainaudit.Entry, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}

	return s.entries, nil
}

func TestAudit_Record(t *testing.T) {
	t.Parallel()

	t.Run("valid entry returns 201", func(t *testing.T) {
		t.Parallel()
		stored := domainaudit.Entry{ID: "1", Actor: "user-1", Action: "login", Target: "session", CreatedAt: time.Now()}
		h := handlers.NewAudit(stubAuditRecorder{stored: stored})

		body, _ := json.Marshal(domainaudit.Entry{Actor: "user-1", Action: "login", Target: "session"})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/audit", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Record(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/audit", bytes.NewReader([]byte("not json")))
		rec := httptest.NewRecorder()

		h.Record(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error returns 400 with message", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{recordErr: domainaudit.ErrMissingActor})

		body, _ := json.Marshal(domainaudit.Entry{Action: "login", Target: "session"})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/audit", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Record(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("repo error returns 500 with generic message", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{recordErr: context.DeadlineExceeded})

		body, _ := json.Marshal(domainaudit.Entry{Actor: "user-1", Action: "login", Target: "session"})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/audit", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Record(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		var body2 map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body2); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body2["error"] != "internal error" {
			t.Errorf("error field = %v, want %q (must not leak details)", body2["error"], "internal error")
		}
	})
}

func TestAudit_List(t *testing.T) {
	t.Parallel()

	t.Run("returns entries", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{
			entries: []domainaudit.Entry{{ID: "1", Actor: "user-1", Action: "login", Target: "session"}},
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/audit", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var body map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		entries, ok := body["entries"].([]any)
		if !ok || len(entries) != 1 {
			t.Fatalf("entries = %v, want 1 entry", body["entries"])
		}
	})

	t.Run("invalid limit returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/audit?limit=abc", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("repo error returns 500", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewAudit(stubAuditRecorder{listErr: context.DeadlineExceeded})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/audit", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}
