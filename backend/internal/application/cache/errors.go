package cache

import "errors"

var (
	// ErrNoPrefix is returned when a List or Flush call omits the key prefix.
	ErrNoPrefix = errors.New("cache: key prefix is required")
)
