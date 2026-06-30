package audit_test

import (
	"context"
	"errors"
	"testing"

	auditapp "github.com/vinaycharlie01/shroute/backend/internal/application/audit"
	domainaudit "github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

type fakeRepo struct {
	appendErr error
	listErr   error
	stored    []domainaudit.Entry
}

func (f *fakeRepo) Append(_ context.Context, e domainaudit.Entry) (domainaudit.Entry, error) {
	if f.appendErr != nil {
		return domainaudit.Entry{}, f.appendErr
	}
	e.ID = "generated-id"
	f.stored = append(f.stored, e)

	return e, nil
}

func (f *fakeRepo) List(_ context.Context, limit int) ([]domainaudit.Entry, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if limit > len(f.stored) {
		limit = len(f.stored)
	}

	return f.stored[:limit], nil
}

func TestService_Record(t *testing.T) {
	t.Parallel()

	t.Run("valid entry is persisted", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepo{}
		svc := auditapp.NewService(repo)

		stored, err := svc.Record(context.Background(), domainaudit.Entry{
			Actor: "user-1", Action: "login", Target: "session",
		})
		if err != nil {
			t.Fatalf("Record() error = %v", err)
		}
		if stored.ID != "generated-id" {
			t.Errorf("ID = %q, want %q", stored.ID, "generated-id")
		}
	})

	t.Run("invalid entry is rejected before reaching the repo", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepo{}
		svc := auditapp.NewService(repo)

		_, err := svc.Record(context.Background(), domainaudit.Entry{Action: "login", Target: "session"})
		if !errors.Is(err, domainaudit.ErrMissingActor) {
			t.Fatalf("Record() error = %v, want %v", err, domainaudit.ErrMissingActor)
		}
		if len(repo.stored) != 0 {
			t.Errorf("repo.stored = %v, want empty", repo.stored)
		}
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		t.Parallel()
		failure := errors.New("connection refused")
		repo := &fakeRepo{appendErr: failure}
		svc := auditapp.NewService(repo)

		_, err := svc.Record(context.Background(), domainaudit.Entry{
			Actor: "user-1", Action: "login", Target: "session",
		})
		if !errors.Is(err, failure) {
			t.Fatalf("Record() error = %v, want wrapped %v", err, failure)
		}
	})
}

func TestService_List(t *testing.T) {
	t.Parallel()

	t.Run("non-positive limit falls back to default", func(t *testing.T) {
		t.Parallel()
		repo := &fakeRepo{stored: []domainaudit.Entry{{ID: "1"}, {ID: "2"}}}
		svc := auditapp.NewService(repo)

		entries, err := svc.List(context.Background(), 0)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("List() returned %d entries, want 2", len(entries))
		}
	})

	t.Run("repo error is wrapped", func(t *testing.T) {
		t.Parallel()
		failure := errors.New("connection refused")
		repo := &fakeRepo{listErr: failure}
		svc := auditapp.NewService(repo)

		_, err := svc.List(context.Background(), 10)
		if !errors.Is(err, failure) {
			t.Fatalf("List() error = %v, want wrapped %v", err, failure)
		}
	})
}
