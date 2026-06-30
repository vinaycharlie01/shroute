package audit_test

import (
	"errors"
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/audit"
)

func TestEntry_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entry   audit.Entry
		wantErr error
	}{
		{
			name:    "valid entry",
			entry:   audit.Entry{Actor: "user-1", Action: "login", Target: "session"},
			wantErr: nil,
		},
		{
			name:    "missing actor",
			entry:   audit.Entry{Action: "login", Target: "session"},
			wantErr: audit.ErrMissingActor,
		},
		{
			name:    "missing action",
			entry:   audit.Entry{Actor: "user-1", Target: "session"},
			wantErr: audit.ErrMissingAction,
		},
		{
			name:    "missing target",
			entry:   audit.Entry{Actor: "user-1", Action: "login"},
			wantErr: audit.ErrMissingTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.entry.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
