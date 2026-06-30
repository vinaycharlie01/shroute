package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

const auditCollectionName = "audit_log"

// AuditRepository implements ports.AuditRepository against a MongoDB
// collection.
type AuditRepository struct {
	coll *mongo.Collection
}

// NewAuditRepository builds an AuditRepository over db's audit_log
// collection.
func NewAuditRepository(db *mongo.Database) *AuditRepository {
	return &AuditRepository{coll: db.Collection(auditCollectionName)}
}

type auditDocument struct {
	ID        bson.ObjectID  `bson:"_id,omitempty"`
	Actor     string         `bson:"actor"`
	Action    string         `bson:"action"`
	Target    string         `bson:"target"`
	Metadata  map[string]any `bson:"metadata,omitempty"`
	CreatedAt time.Time      `bson:"created_at"`
}

// Append implements ports.AuditRepository.
func (r *AuditRepository) Append(ctx context.Context, e audit.Entry) (audit.Entry, error) {
	doc := auditDocument{
		ID:        bson.NewObjectID(),
		Actor:     e.Actor,
		Action:    e.Action,
		Target:    e.Target,
		Metadata:  e.Metadata,
		CreatedAt: time.Now().UTC(),
	}

	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		return audit.Entry{}, fmt.Errorf("mongodb: audit: append: %w", err)
	}

	return toEntry(doc), nil
}

// List implements ports.AuditRepository: returns the most recent entries,
// newest first.
func (r *AuditRepository) List(ctx context.Context, limit int) ([]audit.Entry, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := r.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("mongodb: audit: list: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	var docs []auditDocument
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("mongodb: audit: list: %w", err)
	}

	entries := make([]audit.Entry, len(docs))
	for i, d := range docs {
		entries[i] = toEntry(d)
	}

	return entries, nil
}

func toEntry(d auditDocument) audit.Entry {
	return audit.Entry{
		ID:        d.ID.Hex(),
		Actor:     d.Actor,
		Action:    d.Action,
		Target:    d.Target,
		Metadata:  d.Metadata,
		CreatedAt: d.CreatedAt,
	}
}
