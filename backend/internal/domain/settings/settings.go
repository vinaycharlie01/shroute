// Package settings contains the domain model for application-level
// configuration: key-value settings documents and feature-flag toggles.
package settings

import (
	"encoding/json"
	"time"
)

// Category classifies settings by their functional area.
type Category string

const (
	CategoryGeneral  Category = "general"
	CategorySecurity Category = "security"
	CategoryRouting  Category = "routing"
	CategoryUI       Category = "ui"
)

// knownKeys maps each accepted setting key to its canonical category.
var knownKeys = map[string]Category{
	"log_level":          CategoryGeneral,
	"max_tokens_default": CategoryGeneral,
	"theme":              CategoryUI,
	"require_api_key":    CategorySecurity,
	"allowed_origins":    CategorySecurity,
	"rate_limit_enabled": CategorySecurity,
	"default_model":      CategoryRouting,
	"fallback_enabled":   CategoryRouting,
}

// Document is a settings entry: an opaque JSON value stored under a
// well-known key.
type Document struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	Category  Category        `json:"category"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// FeatureFlag is a named boolean toggle with a declared default.
type FeatureFlag struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DefaultValue string `json:"default_value"`
}

// IsKnownKey reports whether key is in the registry of accepted setting keys.
func IsKnownKey(key string) bool {
	_, ok := knownKeys[key]

	return ok
}

// CategoryFor returns the category for a known key. Callers should check
// IsKnownKey first; an unknown key returns the zero value.
func CategoryFor(key string) Category {
	return knownKeys[key]
}
