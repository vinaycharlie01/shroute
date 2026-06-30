// Package cache holds the pure domain types for the cache management
// feature: stats, entries, and configuration. It imports only the standard
// library.
package cache

import "time"

// Stats represents cache-wide aggregate statistics.
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	SizeBytes  int64 `json:"size_bytes"`
	EntryCount int64 `json:"entry_count"`
}

// Entry represents a single cached key/value pair's metadata.
type Entry struct {
	Key       string        `json:"key"`
	SizeBytes int64         `json:"size_bytes"`
	TTL       time.Duration `json:"ttl"`
}
