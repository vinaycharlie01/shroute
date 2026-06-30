// Package mongodb provides the MongoDB outbound adapter: a mongo.Client plus
// the ports.Pinger implementation used by the health use case. Repository
// implementations for specific collections should live alongside this file
// as the domain grows.
package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DatabaseName is the single logical database every collection-specific
// repository in this package reads from and writes to.
const DatabaseName = "omnirouter"

// Adapter wraps a mongo.Client.
type Adapter struct {
	client *mongo.Client
}

// New creates a MongoDB adapter and verifies connectivity with a ping.
func New(ctx context.Context, uri string) (*Adapter, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongodb: connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)

		return nil, fmt.Errorf("mongodb: ping: %w", err)
	}

	return &Adapter{client: client}, nil
}

// Client exposes the underlying mongo client for repository implementations.
func (a *Adapter) Client() *mongo.Client {
	return a.client
}

// Database returns a handle to DatabaseName on the shared client.
// Collection-specific repositories build their *mongo.Collection from this.
func (a *Adapter) Database() *mongo.Database {
	return a.client.Database(DatabaseName)
}

// Name implements ports.Pinger.
func (a *Adapter) Name() string {
	return "mongodb"
}

// Ping implements ports.Pinger.
func (a *Adapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx, nil)
}

// Close implements ports.Closer.
func (a *Adapter) Close() error {
	return a.client.Disconnect(context.Background())
}
