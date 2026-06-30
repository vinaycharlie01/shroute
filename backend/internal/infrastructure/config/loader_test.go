package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/infrastructure/config"
)

const baseYAML = `
env: local
server:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 10s
  writeTimeout: 10s
  shutdownTimeout: 15s
log:
  level: info
  format: console
`

const devOverlayYAML = `
env: development
log:
  level: debug
`

func writeConfigDir(t *testing.T, base, overlay, overlayName string) string {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "config.base.yaml"), []byte(base), 0o600); err != nil {
		t.Fatalf("write base config: %v", err)
	}
	if overlay != "" {
		if err := os.WriteFile(filepath.Join(dir, overlayName), []byte(overlay), 0o600); err != nil {
			t.Fatalf("write overlay config: %v", err)
		}
	}

	return dir
}

func TestLoad_BaseOnly(t *testing.T) {
	dir := writeConfigDir(t, baseYAML, "", "")

	cfg, err := config.Load(dir, "", "local")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "local" {
		t.Errorf("Env = %q, want %q", cfg.Env, "local")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout.Duration().String() != "10s" {
		t.Errorf("Server.ReadTimeout = %v, want 10s", cfg.Server.ReadTimeout.Duration())
	}
}

func TestLoad_OverlayMergesOnTopOfBase(t *testing.T) {
	dir := writeConfigDir(t, baseYAML, devOverlayYAML, "config.development.yaml")

	cfg, err := config.Load(dir, "", "development")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want %q (from overlay)", cfg.Log.Level, "debug")
	}
	// Untouched-by-overlay field must keep its base value.
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080 (from base)", cfg.Server.Port)
	}
}

func TestLoad_EnvVarOverrides(t *testing.T) {
	dir := writeConfigDir(t, baseYAML, "", "")

	t.Setenv("APP_SERVER_PORT", "9090")
	t.Setenv("APP_LOG_LEVEL", "warn")

	cfg, err := config.Load(dir, "", "local")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090 (env override)", cfg.Server.Port)
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q (env override)", cfg.Log.Level, "warn")
	}
}

func TestLoad_ValidationFailsOnMissingRequiredField(t *testing.T) {
	const invalid = `
env: local
server:
  host: "0.0.0.0"
  port: 8080
log:
  level: info
  format: console
`
	dir := writeConfigDir(t, invalid, "", "")

	if _, err := config.Load(dir, "", "local"); err == nil {
		t.Fatal("Load() error = nil, want validation error for missing timeouts")
	}
}

func TestLoad_DefaultsToLocalEnvWhenUnset(t *testing.T) {
	dir := writeConfigDir(t, baseYAML, "", "")

	cfg, err := config.Load(dir, "", "")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Env != "local" {
		t.Errorf("Env = %q, want %q", cfg.Env, "local")
	}
}
