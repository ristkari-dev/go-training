// Package logstats is the lesson 26 reference accumulator.
package logstats

import "sync"

// Store accumulates per-level log counts under a mutex. Safe for
// concurrent use by multiple request handlers.
type Store struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewStore() *Store {
	return &Store{counts: map[string]int{}}
}

// Merge adds the per-level deltas into the store.
func (s *Store) Merge(delta map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for level, n := range delta {
		s.counts[level] += n
	}
}

// Snapshot returns a copy of the counts plus the grand total.
func (s *Store) Snapshot() (map[string]int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.counts))
	total := 0
	for level, n := range s.counts {
		out[level] = n
		total += n
	}
	return out, total
}
