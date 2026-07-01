package settings_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/application/settings"
	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

// --- fakes ---

type fakeSettingsRepo struct {
	data map[string]domainsettings.Document
}

func newFakeSettingsRepo() *fakeSettingsRepo {
	return &fakeSettingsRepo{data: make(map[string]domainsettings.Document)}
}

func (f *fakeSettingsRepo) Get(_ context.Context, key string) (domainsettings.Document, error) {
	doc, ok := f.data[key]
	if !ok {
		return domainsettings.Document{}, domainsettings.ErrNotFound
	}

	return doc, nil
}

func (f *fakeSettingsRepo) Set(_ context.Context, doc domainsettings.Document) (domainsettings.Document, error) {
	f.data[doc.Key] = doc

	return doc, nil
}

func (f *fakeSettingsRepo) Delete(_ context.Context, key string) error {
	delete(f.data, key)

	return nil
}

func (f *fakeSettingsRepo) List(_ context.Context) ([]domainsettings.Document, error) {
	docs := make([]domainsettings.Document, 0, len(f.data))
	for _, doc := range f.data {
		docs = append(docs, doc)
	}

	return docs, nil
}

type fakeFlagsRepo struct {
	data map[string]domainsettings.FeatureFlag
}

func newFakeFlagsRepo() *fakeFlagsRepo {
	return &fakeFlagsRepo{data: make(map[string]domainsettings.FeatureFlag)}
}

func (f *fakeFlagsRepo) Get(_ context.Context, name string) (domainsettings.FeatureFlag, error) {
	flag, ok := f.data[name]
	if !ok {
		return domainsettings.FeatureFlag{}, domainsettings.ErrNotFound
	}

	return flag, nil
}

func (f *fakeFlagsRepo) List(_ context.Context) ([]domainsettings.FeatureFlag, error) {
	flags := make([]domainsettings.FeatureFlag, 0, len(f.data))
	for _, flag := range f.data {
		flags = append(flags, flag)
	}

	return flags, nil
}

func (f *fakeFlagsRepo) Set(_ context.Context, flag domainsettings.FeatureFlag) (domainsettings.FeatureFlag, error) {
	f.data[flag.Name] = flag

	return flag, nil
}

func (f *fakeFlagsRepo) SeedDefaults(_ context.Context, defaults []domainsettings.FeatureFlag) error {
	for _, flag := range defaults {
		if _, ok := f.data[flag.Name]; !ok {
			f.data[flag.Name] = flag
		}
	}

	return nil
}

// --- tests ---

func TestService_GetSetting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seed    map[string]domainsettings.Document
		key     string
		wantErr error
	}{
		{
			name: "returns stored document",
			seed: map[string]domainsettings.Document{
				"log_level": {Key: "log_level", Value: json.RawMessage(`"debug"`)},
			},
			key: "log_level",
		},
		{
			name:    "unknown key",
			key:     "not_a_key",
			wantErr: domainsettings.ErrUnknownKey,
		},
		{
			name:    "not found in repo",
			key:     "log_level",
			wantErr: domainsettings.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newFakeSettingsRepo()
			for k, v := range tt.seed {
				repo.data[k] = v
			}
			svc := settings.NewService(repo, newFakeFlagsRepo())

			_, err := svc.GetSetting(context.Background(), tt.key)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetSetting() error = %v, want %v", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Errorf("GetSetting() unexpected error: %v", err)
			}
		})
	}
}

func TestService_SetSetting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		value   json.RawMessage
		wantErr error
	}{
		{name: "stores valid setting", key: "log_level", value: json.RawMessage(`"info"`)},
		{name: "unknown key", key: "bad_key", value: json.RawMessage(`"x"`), wantErr: domainsettings.ErrUnknownKey},
		{name: "empty value", key: "log_level", value: nil, wantErr: domainsettings.ErrEmptyValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := settings.NewService(newFakeSettingsRepo(), newFakeFlagsRepo())

			_, err := svc.SetSetting(context.Background(), tt.key, tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SetSetting() error = %v, want %v", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Errorf("SetSetting() unexpected error: %v", err)
			}
		})
	}
}

func TestService_DeleteSetting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{name: "deletes known key", key: "log_level"},
		{name: "unknown key rejected", key: "unknown", wantErr: domainsettings.ErrUnknownKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := settings.NewService(newFakeSettingsRepo(), newFakeFlagsRepo())

			err := svc.DeleteSetting(context.Background(), tt.key)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DeleteSetting() error = %v, want %v", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Errorf("DeleteSetting() unexpected error: %v", err)
			}
		})
	}
}

func TestService_SetAndGetFlag(t *testing.T) {
	t.Parallel()

	flag := domainsettings.FeatureFlag{Name: "CACHE_ENABLED", Enabled: true, DefaultValue: "true"}
	svc := settings.NewService(newFakeSettingsRepo(), newFakeFlagsRepo())

	stored, err := svc.SetFlag(context.Background(), flag)
	if err != nil {
		t.Fatalf("SetFlag() error = %v", err)
	}

	if stored.Name != flag.Name || stored.Enabled != flag.Enabled {
		t.Errorf("SetFlag() = %+v, want %+v", stored, flag)
	}

	got, err := svc.GetFlag(context.Background(), flag.Name)
	if err != nil {
		t.Fatalf("GetFlag() error = %v", err)
	}

	if got.Enabled != flag.Enabled {
		t.Errorf("GetFlag().Enabled = %v, want %v", got.Enabled, flag.Enabled)
	}
}

func TestService_PIIFlagsDefaultFalse(t *testing.T) {
	t.Parallel()

	piiNames := []string{"PII_REDACTION_ENABLED", "PII_RESPONSE_SANITIZATION"}

	for _, name := range piiNames {
		found := false

		for _, f := range domainsettings.DefaultFeatureFlags {
			if f.Name != name {
				continue
			}

			found = true

			if f.Enabled {
				t.Errorf("PII flag %q has Enabled=true, must be false (Hard Rule #20)", name)
			}

			if f.DefaultValue != "false" {
				t.Errorf("PII flag %q DefaultValue = %q, must be \"false\"", name, f.DefaultValue)
			}
		}

		if !found {
			t.Errorf("PII flag %q not found in DefaultFeatureFlags", name)
		}
	}
}
