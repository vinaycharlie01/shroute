package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
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

	if err := applyEnvOverrides(cfg); err != nil {
		return nil, err
	}

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

// applyEnvOverrides applies environment-variable overrides on top of the
// merged YAML config. Field-level overrides are driven entirely by the
// `env` struct tags on Config (see config.go) via caarlos0/env: adding
// override support for a new field is a one-line tag addition here, not a
// new branch in this function. Only the Mongo/Redis "providing a URI/addr
// implies the dependency is enabled" cascade is bespoke logic, since that
// relationship can't be expressed as a single field's struct tag.
func applyEnvOverrides(cfg *Config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("config: env overrides: %w", err)
	}

	if os.Getenv("APP_MONGO_URI") != "" {
		cfg.Mongo.Enabled = true
	}
	if os.Getenv("APP_REDIS_ADDR") != "" {
		cfg.Redis.Enabled = true
	}

	return nil
}
