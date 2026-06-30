// Package config defines the strongly-typed application configuration and
// loads it from YAML files with environment-variable overrides (via
// godotenv) and struct-tag validation.
package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration so YAML values can be written as human
// strings ("5s", "1m30s") instead of raw nanosecond integers.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}

	*d = Duration(parsed)
	return nil
}

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Config is the root, strongly-typed application configuration.
type Config struct {
	Env      string         `yaml:"env" validate:"required,oneof=local development staging production"`
	Server   ServerConfig   `yaml:"server" validate:"required"`
	Log      LogConfig      `yaml:"log" validate:"required"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

// ServerConfig controls the inbound HTTP server.
type ServerConfig struct {
	Host            string   `yaml:"host" validate:"required"`
	Port            int      `yaml:"port" validate:"required,min=1,max=65535"`
	ReadTimeout     Duration `yaml:"readTimeout" validate:"required"`
	WriteTimeout    Duration `yaml:"writeTimeout" validate:"required"`
	ShutdownTimeout Duration `yaml:"shutdownTimeout" validate:"required"`
	AllowedOrigins  []string `yaml:"allowedOrigins"`
}

// LogConfig controls the structured logger.
type LogConfig struct {
	Level  string `yaml:"level" validate:"required,oneof=debug info warn error"`
	Format string `yaml:"format" validate:"required,oneof=json console"`
}

// PostgresConfig controls the Postgres outbound adapter. Enabled defaults to
// false so the foundation runs with zero external dependencies until a
// feature actually needs persistence.
type PostgresConfig struct {
	Enabled bool   `yaml:"enabled"`
	DSN     string `yaml:"dsn" validate:"required_if=Enabled true"`
}

// RedisConfig controls the Redis outbound adapter. Enabled defaults to false
// for the same reason as PostgresConfig.
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr" validate:"required_if=Enabled true"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
