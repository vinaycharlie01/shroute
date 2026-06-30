//go:build integration

// Package containers provides shared testcontainers lifecycle helpers
// (start, wait, register cleanup) for integration tests, so new test cases
// can spin up a dependency with one call instead of duplicating
// container-management boilerplate.
package containers

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// Redis starts a Redis testcontainer, registers its termination on tb's
// cleanup, and returns its connection address.
func Redis(ctx context.Context, tb testing.TB) string {
	tb.Helper()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		tb.Fatalf("start redis container: %v", err)
	}
	tb.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			tb.Logf("terminate redis container: %v", err)
		}
	})

	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		tb.Fatalf("redis endpoint: %v", err)
	}

	return addr
}
