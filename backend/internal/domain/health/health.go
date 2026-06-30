// Package health contains the domain model for service health reporting.
package health

// State describes the overall health of the service.
type State string

const (
	StateUp       State = "up"
	StateDown     State = "down"
	StateDegraded State = "degraded"
)

// Status is the domain entity describing the current health of the service
// and each of its dependencies.
type Status struct {
	State        State
	Dependencies []DependencyStatus
}

// DependencyStatus describes the health of a single downstream dependency.
type DependencyStatus struct {
	Name  string
	State State
	Error string
}

// Overall derives the aggregate State from the dependency list: any down
// dependency makes the whole status degraded, never a hard failure, since a
// single unavailable dependency should not necessarily take the service out
// of rotation.
func Overall(deps []DependencyStatus) State {
	for _, d := range deps {
		if d.State != StateUp {
			return StateDegraded
		}
	}
	return StateUp
}
