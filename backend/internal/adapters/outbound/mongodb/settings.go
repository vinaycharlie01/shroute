package mongodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	domainsettings "github.com/vinaycharlie01/shroute/backend/internal/domain/settings"
)

const (
	settingsCollectionName = "settings"
	flagsCollectionName    = "feature_flags"
)

// SettingsRepository implements ports.SettingsRepository against a MongoDB
// collection keyed by the setting's string key as _id.
type SettingsRepository struct {
	coll *mongo.Collection
}

// NewSettingsRepository builds a SettingsRepository over db's settings
// collection.
func NewSettingsRepository(db *mongo.Database) *SettingsRepository {
	return &SettingsRepository{coll: db.Collection(settingsCollectionName)}
}

type settingsDocument struct {
	ID        string    `bson:"_id"`
	Value     any       `bson:"value"`
	Category  string    `bson:"category"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// Get implements ports.SettingsRepository.
func (r *SettingsRepository) Get(ctx context.Context, key string) (domainsettings.Document, error) {
	var doc settingsDocument

	err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: key}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domainsettings.Document{}, domainsettings.ErrNotFound
		}

		return domainsettings.Document{}, fmt.Errorf("mongodb: settings: get: %w", err)
	}

	return toSettingsDoc(doc)
}

// Set implements ports.SettingsRepository. The value is unmarshalled from
// JSON into a native Go type so MongoDB stores it as a proper sub-document.
func (r *SettingsRepository) Set(ctx context.Context, doc domainsettings.Document) (domainsettings.Document, error) {
	var v any
	if err := json.Unmarshal(doc.Value, &v); err != nil {
		return domainsettings.Document{}, fmt.Errorf("mongodb: settings: set: unmarshal value: %w", err)
	}

	now := time.Now().UTC()
	filter := bson.D{{Key: "_id", Value: doc.Key}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "value", Value: v},
		{Key: "category", Value: string(doc.Category)},
		{Key: "updated_at", Value: now},
	}}}

	if _, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return domainsettings.Document{}, fmt.Errorf("mongodb: settings: set: %w", err)
	}

	doc.UpdatedAt = now

	return doc, nil
}

// Delete implements ports.SettingsRepository.
func (r *SettingsRepository) Delete(ctx context.Context, key string) error {
	if _, err := r.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: key}}); err != nil {
		return fmt.Errorf("mongodb: settings: delete: %w", err)
	}

	return nil
}

// List implements ports.SettingsRepository.
func (r *SettingsRepository) List(ctx context.Context) ([]domainsettings.Document, error) {
	cur, err := r.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("mongodb: settings: list: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	var docs []settingsDocument
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("mongodb: settings: list: %w", err)
	}

	result := make([]domainsettings.Document, 0, len(docs))

	for _, d := range docs {
		converted, convErr := toSettingsDoc(d)
		if convErr != nil {
			return nil, convErr
		}

		result = append(result, converted)
	}

	return result, nil
}

func toSettingsDoc(d settingsDocument) (domainsettings.Document, error) {
	b, err := json.Marshal(d.Value)
	if err != nil {
		return domainsettings.Document{}, fmt.Errorf("mongodb: settings: marshal value: %w", err)
	}

	return domainsettings.Document{
		Key:       d.ID,
		Value:     json.RawMessage(b),
		Category:  domainsettings.Category(d.Category),
		UpdatedAt: d.UpdatedAt,
	}, nil
}

// FeatureFlagRepository implements ports.FeatureFlagRepository against a
// MongoDB collection keyed by the flag name as _id.
type FeatureFlagRepository struct {
	coll *mongo.Collection
}

// NewFeatureFlagRepository builds a FeatureFlagRepository over db's
// feature_flags collection.
func NewFeatureFlagRepository(db *mongo.Database) *FeatureFlagRepository {
	return &FeatureFlagRepository{coll: db.Collection(flagsCollectionName)}
}

type flagDocument struct {
	ID           string `bson:"_id"`
	Enabled      bool   `bson:"enabled"`
	DefaultValue string `bson:"default_value"`
}

// Get implements ports.FeatureFlagRepository.
func (r *FeatureFlagRepository) Get(ctx context.Context, name string) (domainsettings.FeatureFlag, error) {
	var doc flagDocument

	err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: name}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domainsettings.FeatureFlag{}, domainsettings.ErrNotFound
		}

		return domainsettings.FeatureFlag{}, fmt.Errorf("mongodb: flags: get: %w", err)
	}

	return toFlag(doc), nil
}

// List implements ports.FeatureFlagRepository.
func (r *FeatureFlagRepository) List(ctx context.Context) ([]domainsettings.FeatureFlag, error) {
	cur, err := r.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("mongodb: flags: list: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	var docs []flagDocument
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("mongodb: flags: list: %w", err)
	}

	flags := make([]domainsettings.FeatureFlag, len(docs))
	for i, d := range docs {
		flags[i] = toFlag(d)
	}

	return flags, nil
}

// Set implements ports.FeatureFlagRepository. Only the enabled field is
// updated; default_value is a seeding concern managed by SeedDefaults.
func (r *FeatureFlagRepository) Set(ctx context.Context, flag domainsettings.FeatureFlag) (domainsettings.FeatureFlag, error) {
	filter := bson.D{{Key: "_id", Value: flag.Name}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "enabled", Value: flag.Enabled},
	}}}

	if _, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return domainsettings.FeatureFlag{}, fmt.Errorf("mongodb: flags: set: %w", err)
	}

	stored, err := r.Get(ctx, flag.Name)
	if err != nil {
		return domainsettings.FeatureFlag{}, fmt.Errorf("mongodb: flags: set: %w", err)
	}

	return stored, nil
}

// SeedDefaults implements ports.FeatureFlagRepository. Each flag is inserted
// with $setOnInsert so existing operator-modified values are never overwritten.
func (r *FeatureFlagRepository) SeedDefaults(ctx context.Context, defaults []domainsettings.FeatureFlag) error {
	for _, flag := range defaults {
		filter := bson.D{{Key: "_id", Value: flag.Name}}
		update := bson.D{{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: flag.Name},
			{Key: "enabled", Value: flag.Enabled},
			{Key: "default_value", Value: flag.DefaultValue},
		}}}

		if _, err := r.coll.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
			return fmt.Errorf("mongodb: flags: seed defaults: %w", err)
		}
	}

	return nil
}

func toFlag(d flagDocument) domainsettings.FeatureFlag {
	return domainsettings.FeatureFlag{
		Name:         d.ID,
		Enabled:      d.Enabled,
		DefaultValue: d.DefaultValue,
	}
}
