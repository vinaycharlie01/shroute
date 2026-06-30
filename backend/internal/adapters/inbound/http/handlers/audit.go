package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domainaudit "github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

// auditRecorder is the subset of the audit application service the handler
// depends on, declared locally so this package stays decoupled from the
// concrete application/audit import (and is trivially mockable in tests).
type auditRecorder interface {
	Record(ctx context.Context, e domainaudit.Entry) (domainaudit.Entry, error)
	List(ctx context.Context, limit int) ([]domainaudit.Entry, error)
}

// Audit handles POST /api/audit and GET /api/audit.
type Audit struct {
	svc auditRecorder
}

// NewAudit builds an Audit handler over the given service.
func NewAudit(svc auditRecorder) *Audit {
	return &Audit{svc: svc}
}

type auditEntryResponse struct {
	ID        string         `json:"id,omitempty"`
	Actor     string         `json:"actor"`
	Action    string         `json:"action"`
	Target    string         `json:"target"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
}

type auditListResponse struct {
	Entries []auditEntryResponse `json:"entries"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Record handles POST /api/audit: decodes a JSON audit entry, validates and
// persists it, and returns the stored entry.
func (h *Audit) Record(w http.ResponseWriter, r *http.Request) {
	var e domainaudit.Entry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

		return
	}

	stored, err := h.svc.Record(r.Context(), e)
	if err != nil {
		if isValidationError(err) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusCreated, toEntryResponse(stored))
}

// List handles GET /api/audit: returns the most recent audit entries,
// newest first. An optional ?limit= query param caps the result count.
func (h *Audit) List(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid limit"})

		return
	}

	entries, err := h.svc.List(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	resp := auditListResponse{Entries: make([]auditEntryResponse, len(entries))}
	for i, e := range entries {
		resp.Entries[i] = toEntryResponse(e)
	}
	writeJSON(w, http.StatusOK, resp)
}

func isValidationError(err error) bool {
	return errors.Is(err, domainaudit.ErrMissingActor) ||
		errors.Is(err, domainaudit.ErrMissingAction) ||
		errors.Is(err, domainaudit.ErrMissingTarget)
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	return strconv.Atoi(raw)
}

func toEntryResponse(e domainaudit.Entry) auditEntryResponse {
	resp := auditEntryResponse{
		ID:       e.ID,
		Actor:    e.Actor,
		Action:   e.Action,
		Target:   e.Target,
		Metadata: e.Metadata,
	}
	if !e.CreatedAt.IsZero() {
		resp.CreatedAt = e.CreatedAt.Format(http.TimeFormat)
	}

	return resp
}
