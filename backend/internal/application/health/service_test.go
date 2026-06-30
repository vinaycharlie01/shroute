package health_test

import (
	"context"
	"errors"
	"testing"

	healthapp "github.com/vinaycharlie01/shroute/backend/internal/application/health"
	domainhealth "github.com/vinaycharlie01/shroute/backend/internal/domain/health"
)

type fakePinger struct {
	name string
	err  error
}

func (f fakePinger) Name() string { return f.name }

func (f fakePinger) Ping(_ context.Context) error { return f.err }

func TestService_Check(t *testing.T) {
	t.Parallel()

	t.Run("no dependencies reports up", func(t *testing.T) {
		t.Parallel()
		svc := healthapp.NewService()
		status := svc.Check(context.Background())

		if status.State != domainhealth.StateUp {
			t.Fatalf("State = %v, want %v", status.State, domainhealth.StateUp)
		}
		if len(status.Dependencies) != 0 {
			t.Fatalf("Dependencies = %v, want empty", status.Dependencies)
		}
	})

	t.Run("all dependencies healthy", func(t *testing.T) {
		t.Parallel()
		svc := healthapp.NewService(fakePinger{name: "mongodb"}, fakePinger{name: "redis"})
		status := svc.Check(context.Background())

		if status.State != domainhealth.StateUp {
			t.Fatalf("State = %v, want %v", status.State, domainhealth.StateUp)
		}
		if len(status.Dependencies) != 2 {
			t.Fatalf("Dependencies = %v, want 2 entries", status.Dependencies)
		}
	})

	t.Run("one dependency failing degrades status", func(t *testing.T) {
		t.Parallel()
		failure := errors.New("connection refused")
		svc := healthapp.NewService(
			fakePinger{name: "mongodb"},
			fakePinger{name: "redis", err: failure},
		)
		status := svc.Check(context.Background())

		if status.State != domainhealth.StateDegraded {
			t.Fatalf("State = %v, want %v", status.State, domainhealth.StateDegraded)
		}

		var redisStatus *domainhealth.DependencyStatus
		for i := range status.Dependencies {
			if status.Dependencies[i].Name == "redis" {
				redisStatus = &status.Dependencies[i]
			}
		}
		if redisStatus == nil {
			t.Fatal("redis dependency status missing")
		}
		if redisStatus.State != domainhealth.StateDown {
			t.Errorf("redis State = %v, want %v", redisStatus.State, domainhealth.StateDown)
		}
		if redisStatus.Error != failure.Error() {
			t.Errorf("redis Error = %q, want %q", redisStatus.Error, failure.Error())
		}
	})
}
