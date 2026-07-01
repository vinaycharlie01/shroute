package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/inbound/http/handlers"
	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

type stubSettingsManager struct {
	listSettingsResult []domainsettings.Document
	listSettingsErr    error
	getSettingResult   domainsettings.Document
	getSettingErr      error
	setSettingResult   domainsettings.Document
	setSettingErr      error
	deleteSettingErr   error
	listFlagsResult    []domainsettings.FeatureFlag
	listFlagsErr       error
	getFlagResult      domainsettings.FeatureFlag
	getFlagErr         error
	setFlagResult      domainsettings.FeatureFlag
	setFlagErr         error
}

func (s *stubSettingsManager) ListSettings(_ context.Context) ([]domainsettings.Document, error) {
	return s.listSettingsResult, s.listSettingsErr
}

func (s *stubSettingsManager) GetSetting(_ context.Context, _ string) (domainsettings.Document, error) {
	return s.getSettingResult, s.getSettingErr
}

func (s *stubSettingsManager) SetSetting(
	_ context.Context,
	_ string,
	_ json.RawMessage,
) (domainsettings.Document, error) {
	return s.setSettingResult, s.setSettingErr
}

func (s *stubSettingsManager) DeleteSetting(_ context.Context, _ string) error {
	return s.deleteSettingErr
}

func (s *stubSettingsManager) ListFlags(_ context.Context) ([]domainsettings.FeatureFlag, error) {
	return s.listFlagsResult, s.listFlagsErr
}

func (s *stubSettingsManager) GetFlag(_ context.Context, _ string) (domainsettings.FeatureFlag, error) {
	return s.getFlagResult, s.getFlagErr
}

func (s *stubSettingsManager) SetFlag(_ context.Context, _ domainsettings.FeatureFlag) (domainsettings.FeatureFlag, error) {
	return s.setFlagResult, s.setFlagErr
}

func TestSettings_List(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with settings", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			listSettingsResult: []domainsettings.Document{
				{Key: "log_level", Value: json.RawMessage(`"debug"`), Category: domainsettings.CategoryGeneral},
			},
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("repo error returns 500", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{listSettingsErr: context.DeadlineExceeded})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestSettings_Get(t *testing.T) {
	t.Parallel()

	t.Run("known key returns 200", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			getSettingResult: domainsettings.Document{Key: "log_level", Value: json.RawMessage(`"info"`)},
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/log_level", nil)
		req.SetPathValue("key", "log_level")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("unknown key returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{getSettingErr: domainsettings.ErrUnknownKey})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/bad", nil)
		req.SetPathValue("key", "bad")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{getSettingErr: domainsettings.ErrNotFound})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/log_level", nil)
		req.SetPathValue("key", "log_level")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestSettings_Set(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 200", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			setSettingResult: domainsettings.Document{Key: "log_level", Value: json.RawMessage(`"warn"`)},
		})

		body, _ := json.Marshal(map[string]any{"value": "warn"})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/settings/log_level", bytes.NewReader(body))
		req.SetPathValue("key", "log_level")
		rec := httptest.NewRecorder()

		h.Set(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/settings/log_level", bytes.NewReader([]byte("bad")))
		req.SetPathValue("key", "log_level")
		rec := httptest.NewRecorder()

		h.Set(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("unknown key returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{setSettingErr: domainsettings.ErrUnknownKey})

		body, _ := json.Marshal(map[string]any{"value": "x"})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/settings/bad", bytes.NewReader(body))
		req.SetPathValue("key", "bad")
		rec := httptest.NewRecorder()

		h.Set(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestSettings_Delete(t *testing.T) {
	t.Parallel()

	t.Run("valid key returns 204", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/settings/log_level", nil)
		req.SetPathValue("key", "log_level")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("unknown key returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{deleteSettingErr: domainsettings.ErrUnknownKey})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/settings/bad", nil)
		req.SetPathValue("key", "bad")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestSettings_ListFlags(t *testing.T) {
	t.Parallel()

	t.Run("returns flags", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			listFlagsResult: []domainsettings.FeatureFlag{
				{Name: "CACHE_ENABLED", Enabled: true, DefaultValue: "true"},
			},
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/flags", nil)
		rec := httptest.NewRecorder()

		h.ListFlags(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestSettings_GetFlag(t *testing.T) {
	t.Parallel()

	t.Run("known flag returns 200", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			getFlagResult: domainsettings.FeatureFlag{Name: "CACHE_ENABLED", Enabled: true},
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/flags/CACHE_ENABLED", nil)
		req.SetPathValue("name", "CACHE_ENABLED")
		rec := httptest.NewRecorder()

		h.GetFlag(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{getFlagErr: domainsettings.ErrNotFound})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/settings/flags/MISSING", nil)
		req.SetPathValue("name", "MISSING")
		rec := httptest.NewRecorder()

		h.GetFlag(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestSettings_SetFlag(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 200", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{
			setFlagResult: domainsettings.FeatureFlag{Name: "CACHE_ENABLED", Enabled: false},
		})

		body, _ := json.Marshal(map[string]any{"enabled": false})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/settings/flags/CACHE_ENABLED", bytes.NewReader(body))
		req.SetPathValue("name", "CACHE_ENABLED")
		rec := httptest.NewRecorder()

		h.SetFlag(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		t.Parallel()
		h := handlers.NewSettings(&stubSettingsManager{})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/settings/flags/X", bytes.NewReader([]byte("bad")))
		req.SetPathValue("name", "X")
		rec := httptest.NewRecorder()

		h.SetFlag(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
