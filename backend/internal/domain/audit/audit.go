// Package audit contains the domain model for the audit trail: who did
// what, when, recorded for later inspection.
package audit

import "time"

// Entry is a single recorded action.
type Entry struct {
	ID        string         `json:"id,omitempty"`
	Actor     string         `json:"actor"`
	Action    string         `json:"action"`
	Target    string         `json:"target"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitzero"`
}

// Validate reports whether the entry has the minimum fields required to be
// recorded: an actor, an action, and a target. CreatedAt and ID are
// populated by the application/adapter layers, not the caller.
func (e Entry) Validate() error {
	switch {
	case e.Actor == "":
		return ErrMissingActor
	case e.Action == "":
		return ErrMissingAction
	case e.Target == "":
		return ErrMissingTarget
	default:
		return nil
	}
}
