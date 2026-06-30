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

// Config is the root, strongly-typed application configuration. The `env`
// struct tags are read by Load via github.com/caarlos0/env: any field
// carrying one is automatically overridable by that environment variable
// with no code change to the loader. Add an `env` tag to make a new field
// overridable; omit it to keep a field YAML-only.
type Config struct {
	Env    string       `yaml:"env"    env:"APP_ENV"    validate:"required,oneof=local development staging production"`
	Server ServerConfig `yaml:"server" validate:"required"`
	Log    LogConfig    `yaml:"log"    validate:"required"`
	Mongo  MongoConfig  `yaml:"mongo"`
	Redis  RedisConfig  `yaml:"redis"`
}

// ServerConfig controls the inbound HTTP server.
type ServerConfig struct {
	Host            string   `yaml:"host" env:"APP_SERVER_HOST" validate:"required"`
	Port            int      `yaml:"port" env:"APP_SERVER_PORT" validate:"required,min=1,max=65535"`
	ReadTimeout     Duration `yaml:"readTimeout" validate:"required"`
	WriteTimeout    Duration `yaml:"writeTimeout" validate:"required"`
	ShutdownTimeout Duration `yaml:"shutdownTimeout" validate:"required"`
	AllowedOrigins  []string `yaml:"allowedOrigins" env:"APP_SERVER_ALLOWED_ORIGINS" envSeparator:","`
}

// LogConfig controls the structured logger.
type LogConfig struct {
	Level  string `yaml:"level"  env:"APP_LOG_LEVEL"  validate:"required,oneof=debug info warn error"`
	Format string `yaml:"format" env:"APP_LOG_FORMAT" validate:"required,oneof=json console"`
}

// MongoConfig controls the MongoDB outbound adapter. Enabled defaults to
// false so the foundation runs with zero external dependencies until a
// feature actually needs persistence. Setting APP_MONGO_URI implies Enabled
// (see loader.go); Enabled itself has no env tag so it can't be flipped
// independently of providing a URI.
type MongoConfig struct {
	Enabled bool   `yaml:"enabled"`
	URI     string `yaml:"uri" env:"APP_MONGO_URI" validate:"required_if=Enabled true"`
}

// RedisConfig controls the Redis outbound adapter. Enabled defaults to false
// for the same reason as MongoConfig; setting APP_REDIS_ADDR implies Enabled.
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr"     env:"APP_REDIS_ADDR"     validate:"required_if=Enabled true"`
	Password string `yaml:"password" env:"APP_REDIS_PASSWORD"`
	DB       int    `yaml:"db"       env:"APP_REDIS_DB"`
}
