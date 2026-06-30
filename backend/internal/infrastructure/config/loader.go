package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Load reads {configDir}/config.base.yaml, overlays
// {configDir}/config.{env}.yaml, applies environment-variable overrides
// (after loading dotenvPath into the process environment, when present),
// and validates the result against the struct tags on Config.
//
// env is normally one of: local, development, staging, production. When
// empty, it falls back to the APP_ENV environment variable, then "local".
func Load(configDir, dotenvPath, env string) (*Config, error) {
	if dotenvPath != "" {
		if err := godotenv.Load(dotenvPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: load dotenv: %w", err)
		}
	}

	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = "local"
	}

	cfg := &Config{}
	if err := mergeYAML(cfg, filepath.Join(configDir, "config.base.yaml")); err != nil {
		return nil, err
	}
	if err := mergeYAML(cfg, filepath.Join(configDir, "config."+env+".yaml")); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return cfg, nil
}

// mergeYAML unmarshals path onto cfg, leaving fields already set untouched
// when the file is absent or omits them. A missing overlay file (e.g. no
// config.production.yaml) is not an error.
func mergeYAML(cfg *Config, path string) error {
	data, err := os.ReadFile(path) //nolint:gosec // path is built from the trusted configDir startup parameter, not external input
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("config: read %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("config: parse %s: %w", path, err)
	}

	return nil
}

// applyEnvOverrides applies a small, explicit set of environment-variable
// overrides on top of the merged YAML config. Only operational and
// security-sensitive fields (host/port/log level/credentials) are
// overridable this way; the rest of the configuration surface lives in
// YAML so it stays diffable and reviewable.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.Env = v
	}
	if v := os.Getenv("APP_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("APP_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("APP_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("APP_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if v := os.Getenv("APP_POSTGRES_DSN"); v != "" {
		cfg.Postgres.DSN = v
		cfg.Postgres.Enabled = true
	}
	if v := os.Getenv("APP_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
		cfg.Redis.Enabled = true
	}
	if v := os.Getenv("APP_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
}
