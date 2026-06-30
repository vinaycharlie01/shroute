//go:build integration

package containers

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

// MongoDB starts a MongoDB testcontainer, registers its termination on tb's
// cleanup, and returns its connection URI.
func MongoDB(ctx context.Context, tb testing.TB) string {
	tb.Helper()

	container, err := tcmongodb.Run(ctx, "mongo:7")
	if err != nil {
		tb.Fatalf("start mongodb container: %v", err)
	}
	tb.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			tb.Logf("terminate mongodb container: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		tb.Fatalf("mongodb connection string: %v", err)
	}

	return uri
}
