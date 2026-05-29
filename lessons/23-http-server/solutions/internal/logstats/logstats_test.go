package logstats

import (
	"sync"
	"testing"
)

func TestStore(t *testing.T) {
	s := NewStore()
	s.Merge(map[string]int{"INFO": 2, "WARN": 1})
	s.Merge(map[string]int{"INFO": 3, "ERROR": 1})

	counts, total := s.Snapshot()
	if counts["INFO"] != 5 || counts["WARN"] != 1 || counts["ERROR"] != 1 {
		t.Errorf("counts = %v", counts)
	}
	if total != 7 {
		t.Errorf("total = %d, want 7", total)
	}
}

func TestSnapshotIsCopy(t *testing.T) {
	s := NewStore()
	s.Merge(map[string]int{"INFO": 1})
	snap, _ := s.Snapshot()
	snap["INFO"] = 999 // mutating the copy must not affect the store
	again, _ := s.Snapshot()
	if again["INFO"] != 1 {
		t.Errorf("Snapshot returned the internal map, not a copy: %v", again)
	}
}

func TestConcurrentMerge(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Merge(map[string]int{"INFO": 1})
		}()
	}
	wg.Wait()
	counts, _ := s.Snapshot()
	if counts["INFO"] != 100 {
		t.Errorf("INFO = %d, want 100", counts["INFO"])
	}
}
