// Package dedup is a bounded, concurrency-safe set of seen idempotency
// keys — the heart of server-side deduplication. Reference impl.
package dedup

import "sync"

type Store struct {
	mu    sync.Mutex
	cap   int
	seen  map[string]struct{}
	order []string
}

func New(capacity int) *Store {
	if capacity < 1 {
		capacity = 1
	}
	return &Store{cap: capacity, seen: make(map[string]struct{}, capacity)}
}

func (s *Store) Seen(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return true
	}
	s.seen[key] = struct{}{}
	s.order = append(s.order, key)
	if len(s.order) > s.cap {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.seen, oldest)
	}
	return false
}
