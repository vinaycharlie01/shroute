package health_test

import (
	"testing"

	"github.com/vinaycharlie01/shroute/backend/internal/domain/health"
)

func TestOverall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		deps []health.DependencyStatus
		want health.State
	}{
		{name: "no dependencies", deps: nil, want: health.StateUp},
		{
			name: "all up",
			deps: []health.DependencyStatus{{Name: "postgres", State: health.StateUp}},
			want: health.StateUp,
		},
		{
			name: "one down",
			deps: []health.DependencyStatus{
				{Name: "postgres", State: health.StateUp},
				{Name: "redis", State: health.StateDown},
			},
			want: health.StateDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := health.Overall(tt.deps); got != tt.want {
				t.Errorf("Overall() = %v, want %v", got, tt.want)
			}
		})
	}
}
