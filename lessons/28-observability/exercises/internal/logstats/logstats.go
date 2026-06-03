// Package logstats is the in-memory level-count accumulator shared
// across all ingest requests of the logstats service. It is safe for
// concurrent use.
package logstats

import "sync"

// Store accumulates per-level log counts under a mutex.
type Store struct {
	mu     sync.Mutex
	counts map[string]int
}

// NewStore returns an empty, ready-to-use Store.
func NewStore() *Store {
	return &Store{counts: map[string]int{}}
}

// Merge adds the per-level deltas into the store.
//
// Hint: lock, then `for level, n := range delta { s.counts[level] += n }`.
func (s *Store) Merge(delta map[string]int) {
	_ = sync.Mutex{}
	panic("TODO: lock, add each delta into s.counts")
}

// Snapshot returns a COPY of the counts plus the grand total. Returning
// a copy (not the internal map) keeps callers from racing on it.
//
// Hint: lock, copy counts into a new map, sum the values.
func (s *Store) Snapshot() (map[string]int, int) {
	panic("TODO: lock, copy counts to a new map, return (copy, total)")
}
