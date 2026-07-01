// Package settings implements the settings and feature-flag use cases.
package settings

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vinaycharlie01/shroute/backend/internal/application/ports"
	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

// Service implements the settings and feature-flag use cases. It depends
// only on ports interfaces, never on concrete storage adapters.
type Service struct {
	settingsRepo ports.SettingsRepository
	flagsRepo    ports.FeatureFlagRepository
}

// NewService builds a Service over the given repositories.
func NewService(settingsRepo ports.SettingsRepository, flagsRepo ports.FeatureFlagRepository) *Service {
	return &Service{settingsRepo: settingsRepo, flagsRepo: flagsRepo}
}

// GetSetting retrieves a single setting by its well-known key.
func (s *Service) GetSetting(ctx context.Context, key string) (domainsettings.Document, error) {
	if !domainsettings.IsKnownKey(key) {
		return domainsettings.Document{}, fmt.Errorf("settings: get: %w", domainsettings.ErrUnknownKey)
	}

	doc, err := s.settingsRepo.Get(ctx, key)
	if err != nil {
		return domainsettings.Document{}, fmt.Errorf("settings: get: %w", err)
	}

	return doc, nil
}

// ListSettings returns all stored settings documents.
func (s *Service) ListSettings(ctx context.Context) ([]domainsettings.Document, error) {
	docs, err := s.settingsRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: list: %w", err)
	}

	return docs, nil
}

// SetSetting stores a JSON value for a known key, returning the persisted
// document.
func (s *Service) SetSetting(
	ctx context.Context,
	key string,
	value json.RawMessage,
) (domainsettings.Document, error) {
	if !domainsettings.IsKnownKey(key) {
		return domainsettings.Document{}, fmt.Errorf("settings: set: %w", domainsettings.ErrUnknownKey)
	}

	if len(value) == 0 {
		return domainsettings.Document{}, fmt.Errorf("settings: set: %w", domainsettings.ErrEmptyValue)
	}

	doc := domainsettings.Document{
		Key:      key,
		Value:    value,
		Category: domainsettings.CategoryFor(key),
	}

	stored, err := s.settingsRepo.Set(ctx, doc)
	if err != nil {
		return domainsettings.Document{}, fmt.Errorf("settings: set: %w", err)
	}

	return stored, nil
}

// DeleteSetting removes a setting by its well-known key.
func (s *Service) DeleteSetting(ctx context.Context, key string) error {
	if !domainsettings.IsKnownKey(key) {
		return fmt.Errorf("settings: delete: %w", domainsettings.ErrUnknownKey)
	}

	if err := s.settingsRepo.Delete(ctx, key); err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

// GetFlag retrieves a single feature flag by name.
func (s *Service) GetFlag(ctx context.Context, name string) (domainsettings.FeatureFlag, error) {
	flag, err := s.flagsRepo.Get(ctx, name)
	if err != nil {
		return domainsettings.FeatureFlag{}, fmt.Errorf("settings: get flag: %w", err)
	}

	return flag, nil
}

// ListFlags returns all stored feature flags.
func (s *Service) ListFlags(ctx context.Context) ([]domainsettings.FeatureFlag, error) {
	flags, err := s.flagsRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: list flags: %w", err)
	}

	return flags, nil
}

// SetFlag stores a feature flag's enabled state, returning the persisted flag.
func (s *Service) SetFlag(ctx context.Context, flag domainsettings.FeatureFlag) (domainsettings.FeatureFlag, error) {
	stored, err := s.flagsRepo.Set(ctx, flag)
	if err != nil {
		return domainsettings.FeatureFlag{}, fmt.Errorf("settings: set flag: %w", err)
	}

	return stored, nil
}
