// Package handlers contains HTTP inbound adapters: they translate
// http.Request/ResponseWriter into application use-case calls and back.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	domainhealth "github.com/vinaycharlie01/shroute/backend/internal/domain/health"
)

// healthChecker is the subset of the health application service the handler
// depends on, declared locally so this package stays decoupled from the
// concrete application/health import (and is trivially mockable in tests).
type healthChecker interface {
	Check(ctx context.Context) domainhealth.Status
}

// Health handles GET /healthz and GET /readyz.
type Health struct {
	checker healthChecker
}

// NewHealth builds a Health handler over the given checker.
func NewHealth(checker healthChecker) *Health {
	return &Health{checker: checker}
}

// RegisterRoutes implements httpadapter.RouteRegistrar.
func (h *Health) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Live)
	mux.HandleFunc("GET /readyz", h.Ready)
}

type healthResponse struct {
	Status       string             `json:"status"`
	Dependencies []dependencyStatus `json:"dependencies,omitempty"`
}

type dependencyStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Live handles GET /healthz: a cheap liveness probe that never checks
// downstream dependencies, only that the process is up and serving.
func (h *Health) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: string(domainhealth.StateUp)})
}

// Ready handles GET /readyz: a readiness probe that checks every registered
// dependency and returns 200 unless the aggregate state is down.
func (h *Health) Ready(w http.ResponseWriter, r *http.Request) {
	status := h.checker.Check(r.Context())

	resp := healthResponse{Status: string(status.State)}
	for _, d := range status.Dependencies {
		resp.Dependencies = append(resp.Dependencies, dependencyStatus{
			Name:   d.Name,
			Status: string(d.State),
			Error:  d.Error,
		})
	}

	code := http.StatusOK
	if status.State == domainhealth.StateDown {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, resp)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
