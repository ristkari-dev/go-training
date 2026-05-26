package counter

import (
	"sync"
	"testing"
)

// TestCounter is a SKELETON. The full test spawns 100 goroutines, each
// doing 100 increments (10,000 total), then asserts Counter.Value()
// equals exactly 10000. This passes under `go test -race` because
// Counter uses a mutex.
//
// Hint:
//  1. var c Counter
//  2. var wg sync.WaitGroup
//  3. for i := 0; i < 100; i++ {
//     wg.Add(1)
//     go func() { defer wg.Done(); for j := 0; j < 100; j++ { c.Inc() } }()
//     }
//  4. wg.Wait()
//  5. if got := c.Value(); got != 10000 → t.Errorf("got %d, want 10000", got)
func TestCounter(t *testing.T) {
	// TODO: write the test per the hint above.
	_ = sync.WaitGroup{}
}
