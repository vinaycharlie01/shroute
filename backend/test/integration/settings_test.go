//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/vinaycharlie01/shroute/backend/internal/adapters/outbound/mongodb"
	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
	"github.com/vinaycharlie01/shroute/backend/test/containers"
)

func TestSettingsRepository_SetGetDeleteAgainstRealMongoDB(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	uri := containers.MongoDB(ctx, t)

	adapter, err := mongodb.New(ctx, uri)
	if err != nil {
		t.Fatalf("mongodb.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Logf("close adapter: %v", err)
		}
	})

	repo := mongodb.NewSettingsRepository(adapter.Database())

	// Set
	doc := domainsettings.Document{
		Key:      "log_level",
		Value:    json.RawMessage(`"debug"`),
		Category: domainsettings.CategoryGeneral,
	}

	stored, err := repo.Set(ctx, doc)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if stored.UpdatedAt.IsZero() {
		t.Error("Set() did not populate UpdatedAt")
	}

	// Get
	got, err := repo.Get(ctx, "log_level")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Key != "log_level" {
		t.Errorf("Get() Key = %q, want %q", got.Key, "log_level")
	}

	// List
	docs, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("List() returned %d docs, want 1", len(docs))
	}

	// Delete
	if err := repo.Delete(ctx, "log_level"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Get after delete should return ErrNotFound.
	if _, err := repo.Get(ctx, "log_level"); !errors.Is(err, domainsettings.ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestFeatureFlagRepository_SeedDefaultsAndSet(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	uri := containers.MongoDB(ctx, t)

	adapter, err := mongodb.New(ctx, uri)
	if err != nil {
		t.Fatalf("mongodb.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := adapter.Close(); err != nil {
			t.Logf("close adapter: %v", err)
		}
	})

	repo := mongodb.NewFeatureFlagRepository(adapter.Database())

	// Seed defaults.
	if err := repo.SeedDefaults(ctx, domainsettings.DefaultFeatureFlags); err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	// Seeded PII flag must be disabled.
	piiFlag, err := repo.Get(ctx, "PII_REDACTION_ENABLED")
	if err != nil {
		t.Fatalf("Get(PII_REDACTION_ENABLED) error = %v", err)
	}

	if piiFlag.Enabled {
		t.Error("seeded PII_REDACTION_ENABLED has Enabled=true, must be false (Hard Rule #20)")
	}

	// Operator overrides a flag.
	if _, err := repo.Set(ctx, domainsettings.FeatureFlag{Name: "CACHE_ENABLED", Enabled: false}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Re-seeding must not overwrite the operator's value.
	if err := repo.SeedDefaults(ctx, domainsettings.DefaultFeatureFlags); err != nil {
		t.Fatalf("SeedDefaults() second call error = %v", err)
	}

	got, err := repo.Get(ctx, "CACHE_ENABLED")
	if err != nil {
		t.Fatalf("Get(CACHE_ENABLED) after second seed error = %v", err)
	}

	if got.Enabled {
		t.Error("second SeedDefaults() overwrote operator Set() value, want Enabled=false")
	}
}
