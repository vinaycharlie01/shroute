package settings

import "errors"

// Sentinel errors for settings validation and repository operations.
var (
	ErrNotFound   = errors.New("settings: not found")
	ErrUnknownKey = errors.New("settings: unknown key")
	ErrEmptyValue = errors.New("settings: value must not be empty")
)
