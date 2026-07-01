package settings_test

import (
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

func TestIsKnownKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "known key", key: "log_level", want: true},
		{name: "unknown key", key: "nonexistent", want: false},
		{name: "empty key", key: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := settings.IsKnownKey(tt.key); got != tt.want {
				t.Errorf("IsKnownKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestCategoryFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		want settings.Category
	}{
		{name: "general key", key: "log_level", want: settings.CategoryGeneral},
		{name: "ui key", key: "theme", want: settings.CategoryUI},
		{name: "security key", key: "require_api_key", want: settings.CategorySecurity},
		{name: "routing key", key: "default_model", want: settings.CategoryRouting},
		{name: "unknown returns empty", key: "unknown", want: settings.Category("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := settings.CategoryFor(tt.key); got != tt.want {
				t.Errorf("CategoryFor(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestDefaultFeatureFlags_PIIFlagsAreDisabledByDefault(t *testing.T) {
	t.Parallel()

	piiNames := []string{"PII_REDACTION_ENABLED", "PII_RESPONSE_SANITIZATION"}

	for _, name := range piiNames {
		found := false

		for _, f := range settings.DefaultFeatureFlags {
			if f.Name != name {
				continue
			}

			found = true

			if f.Enabled {
				t.Errorf("PII flag %q has Enabled=true, must be false (Hard Rule #20)", name)
			}

			if f.DefaultValue != "false" {
				t.Errorf("PII flag %q has DefaultValue=%q, must be \"false\"", name, f.DefaultValue)
			}
		}

		if !found {
			t.Errorf("PII flag %q missing from DefaultFeatureFlags", name)
		}
	}
}
