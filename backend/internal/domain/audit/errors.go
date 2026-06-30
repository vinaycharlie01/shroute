package audit

import "errors"

// Sentinel validation errors for Entry.Validate.
var (
	ErrMissingActor  = errors.New("audit: actor is required")
	ErrMissingAction = errors.New("audit: action is required")
	ErrMissingTarget = errors.New("audit: target is required")
)
