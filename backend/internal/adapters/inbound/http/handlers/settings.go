package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

// settingsManager is the subset of the settings application service the
// handler depends on, declared locally for decoupling and testability.
type settingsManager interface {
	GetSetting(ctx context.Context, key string) (domainsettings.Document, error)
	ListSettings(ctx context.Context) ([]domainsettings.Document, error)
	SetSetting(ctx context.Context, key string, value json.RawMessage) (domainsettings.Document, error)
	DeleteSetting(ctx context.Context, key string) error
	GetFlag(ctx context.Context, name string) (domainsettings.FeatureFlag, error)
	ListFlags(ctx context.Context) ([]domainsettings.FeatureFlag, error)
	SetFlag(ctx context.Context, flag domainsettings.FeatureFlag) (domainsettings.FeatureFlag, error)
}

// Settings handles the settings and feature-flag HTTP routes.
type Settings struct {
	svc settingsManager
}

// NewSettings builds a Settings handler over the given service.
func NewSettings(svc settingsManager) *Settings {
	return &Settings{svc: svc}
}

// RegisterRoutes implements httpadapter.RouteRegistrar. Static paths
// (GET /api/settings/flags) are registered before wildcard paths so the
// Go 1.22 ServeMux resolves them as more specific.
func (h *Settings) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", h.List)
	mux.HandleFunc("GET /api/settings/flags", h.ListFlags)
	mux.HandleFunc("GET /api/settings/flags/{name}", h.GetFlag)
	mux.HandleFunc("PUT /api/settings/flags/{name}", h.SetFlag)
	mux.HandleFunc("GET /api/settings/{key}", h.Get)
	mux.HandleFunc("PUT /api/settings/{key}", h.Set)
	mux.HandleFunc("DELETE /api/settings/{key}", h.Delete)
}

// settingResponse is the JSON envelope returned for a single setting document.
type settingResponse struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value" swaggertype:"object"`
	Category  string          `json:"category"`
	UpdatedAt string          `json:"updated_at,omitempty"`
}

// settingsListResponse is the JSON envelope for a list of settings.
type settingsListResponse struct {
	Settings []settingResponse `json:"settings"`
}

// flagResponse is the JSON envelope for a single feature flag.
type flagResponse struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DefaultValue string `json:"default_value"`
}

// flagsListResponse is the JSON envelope for a list of feature flags.
type flagsListResponse struct {
	Flags []flagResponse `json:"flags"`
}

// setValueRequest is the request body for PUT /api/settings/{key}.
type setValueRequest struct {
	Value json.RawMessage `json:"value" swaggertype:"object"`
}

// setFlagRequest is the request body for PUT /api/settings/flags/{name}.
type setFlagRequest struct {
	Enabled bool `json:"enabled"`
}

// List handles GET /api/settings.
//
// @Summary      List all settings
// @Description  Returns all stored application settings across every category.
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  settingsListResponse
// @Failure      500  {object}  errorResponse
// @Router       /api/settings [get].
func (h *Settings) List(w http.ResponseWriter, r *http.Request) {
	docs, err := h.svc.ListSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	resp := settingsListResponse{Settings: make([]settingResponse, len(docs))}
	for i, d := range docs {
		resp.Settings[i] = toSettingResponse(d)
	}
	writeJSON(w, http.StatusOK, resp)
}

// Get handles GET /api/settings/{key}.
//
// @Summary      Get a setting
// @Description  Returns a single setting document by its well-known key.
// @Tags         Settings
// @Produce      json
// @Param        key  path      string  true  "Setting key (e.g. log_level, theme)"
// @Success      200  {object}  settingResponse
// @Failure      400  {object}  errorResponse  "Unknown setting key"
// @Failure      404  {object}  errorResponse  "Setting not found"
// @Failure      500  {object}  errorResponse
// @Router       /api/settings/{key} [get].
func (h *Settings) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	doc, err := h.svc.GetSetting(r.Context(), key)
	if err != nil {
		if errors.Is(err, domainsettings.ErrUnknownKey) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unknown setting key"})

			return
		}

		if errors.Is(err, domainsettings.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "setting not found"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, toSettingResponse(doc))
}

// Set handles PUT /api/settings/{key}.
//
// @Summary      Set a setting
// @Description  Creates or updates the JSON value for a well-known setting key.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        key      path      string           true  "Setting key (e.g. log_level, theme)"
// @Param        request  body      setValueRequest  true  "JSON value to store"
// @Success      200      {object}  settingResponse
// @Failure      400      {object}  errorResponse  "Unknown key, empty value, or invalid JSON"
// @Failure      500      {object}  errorResponse
// @Router       /api/settings/{key} [put].
func (h *Settings) Set(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req setValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

		return
	}

	doc, err := h.svc.SetSetting(r.Context(), key, req.Value)
	if err != nil {
		if errors.Is(err, domainsettings.ErrUnknownKey) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unknown setting key"})

			return
		}

		if errors.Is(err, domainsettings.ErrEmptyValue) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "value must not be empty"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, toSettingResponse(doc))
}

// Delete handles DELETE /api/settings/{key}.
//
// @Summary      Delete a setting
// @Description  Removes a setting by its well-known key. Returns 204 on success.
// @Tags         Settings
// @Produce      json
// @Param        key  path  string  true  "Setting key"
// @Success      204
// @Failure      400  {object}  errorResponse  "Unknown setting key"
// @Failure      500  {object}  errorResponse
// @Router       /api/settings/{key} [delete].
func (h *Settings) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := h.svc.DeleteSetting(r.Context(), key); err != nil {
		if errors.Is(err, domainsettings.ErrUnknownKey) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unknown setting key"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListFlags handles GET /api/settings/flags.
//
// @Summary      List feature flags
// @Description  Returns all feature flags and their current enabled state.
// @Tags         Settings
// @Produce      json
// @Success      200  {object}  flagsListResponse
// @Failure      500  {object}  errorResponse
// @Router       /api/settings/flags [get].
func (h *Settings) ListFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.svc.ListFlags(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	resp := flagsListResponse{Flags: make([]flagResponse, len(flags))}
	for i, f := range flags {
		resp.Flags[i] = toFlagResponse(f)
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetFlag handles GET /api/settings/flags/{name}.
//
// @Summary      Get a feature flag
// @Description  Returns a single feature flag by name, including its enabled state and seeded default.
// @Tags         Settings
// @Produce      json
// @Param        name  path      string  true  "Flag name (e.g. CACHE_ENABLED, PII_REDACTION_ENABLED)"
// @Success      200   {object}  flagResponse
// @Failure      404   {object}  errorResponse  "Flag not found"
// @Failure      500   {object}  errorResponse
// @Router       /api/settings/flags/{name} [get].
func (h *Settings) GetFlag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	flag, err := h.svc.GetFlag(r.Context(), name)
	if err != nil {
		if errors.Is(err, domainsettings.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "flag not found"})

			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, toFlagResponse(flag))
}

// SetFlag handles PUT /api/settings/flags/{name}.
//
// @Summary      Set a feature flag
// @Description  Enables or disables a feature flag by name. PII flags are opt-in — do not enable by default.
// @Tags         Settings
// @Accept       json
// @Produce      json
// @Param        name     path      string          true  "Flag name (e.g. CACHE_ENABLED)"
// @Param        request  body      setFlagRequest  true  "Enabled state to apply"
// @Success      200      {object}  flagResponse
// @Failure      400      {object}  errorResponse  "Invalid request body"
// @Failure      500      {object}  errorResponse
// @Router       /api/settings/flags/{name} [put].
func (h *Settings) SetFlag(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var req setFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

		return
	}

	stored, err := h.svc.SetFlag(r.Context(), domainsettings.FeatureFlag{
		Name:    name,
		Enabled: req.Enabled,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, toFlagResponse(stored))
}

func toSettingResponse(d domainsettings.Document) settingResponse {
	resp := settingResponse{
		Key:      d.Key,
		Value:    d.Value,
		Category: string(d.Category),
	}

	if !d.UpdatedAt.IsZero() {
		resp.UpdatedAt = d.UpdatedAt.Format(http.TimeFormat)
	}

	return resp
}

func toFlagResponse(f domainsettings.FeatureFlag) flagResponse {
	return flagResponse{
		Name:         f.Name,
		Enabled:      f.Enabled,
		DefaultValue: f.DefaultValue,
	}
}
