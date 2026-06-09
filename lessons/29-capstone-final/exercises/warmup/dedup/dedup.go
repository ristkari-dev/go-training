// Package dedup is a bounded, concurrency-safe set of seen idempotency
// keys — the heart of server-side deduplication.
package dedup

import "sync"

// Store remembers up to `cap` recently-seen keys (FIFO eviction).
type Store struct {
	mu    sync.Mutex
	cap   int
	seen  map[string]struct{}
	order []string // insertion order, for bounded eviction
}

func New(capacity int) *Store {
	if capacity < 1 {
		capacity = 1
	}
	return &Store{cap: capacity, seen: make(map[string]struct{}, capacity)}
}

// Seen reports whether key was already recorded. On first sight it
// records the key (evicting the oldest if over capacity) and returns
// false; on a repeat it returns true. Safe for concurrent use.
//
// Hint:
//
//	lock; if _, ok := s.seen[key]; ok { return true }
//	s.seen[key] = struct{}{}; s.order = append(s.order, key)
//	if len(s.order) > s.cap { evict s.order[0] from both order and seen }
//	return false
func (s *Store) Seen(key string) bool {
	_ = sync.Mutex{}
	panic("TODO: bounded, concurrency-safe seen-set; first sight false (record), repeat true")
}
