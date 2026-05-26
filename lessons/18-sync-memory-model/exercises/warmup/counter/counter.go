// Package counter is the lesson 18 warm-up: a thread-safe counter
// (with sync.Mutex) alongside a deliberately-buggy version (without
// the mutex) for pedagogical comparison.
//
// Counter is what students implement. CounterUnsafe exists in this
// file so students can run it manually under `go test -race` and see
// the race detector flag the data race.
//
// The unit tests only test Counter (the correct version). Testing
// CounterUnsafe would either flake (the race doesn't always manifest)
// or fail under -race (breaking `make test-race`). Instead, the slides
// and README walk through running CounterUnsafe manually.
package counter

import "sync"

// Counter is safe for concurrent use. Multiple goroutines may call
// Inc and Value simultaneously.
//
// Hint:
//   - Embed a sync.Mutex (or hold a pointer; embedded is simpler here).
//   - Inc: Lock; n++; Unlock. Or defer Unlock.
//   - Value: Lock; v := n; Unlock; return v.
//
// The zero value of Counter is a usable, unlocked counter at 0. No
// constructor needed (this is the idiomatic Go pattern for sync types).
type Counter struct {
	mu sync.Mutex
	n  int
}

// Inc atomically increments the counter by 1.
func (c *Counter) Inc() {
	panic("TODO: c.mu.Lock(); c.n++; c.mu.Unlock()")
}

// Value returns the current counter value.
func (c *Counter) Value() int {
	panic("TODO: c.mu.Lock(); v := c.n; c.mu.Unlock(); return v")
}

// CounterUnsafe is the SAME interface without the mutex. It has a
// data race on n if Inc is called concurrently. The unit tests do NOT
// test this type — it exists so you can run it manually under
// `go test -race` and see the race detector report a data race.
//
// To see the failure firsthand, create a temporary test like:
//
//	func TestCounterUnsafeRaces(t *testing.T) {
//	    var c CounterUnsafe
//	    var wg sync.WaitGroup
//	    for i := 0; i < 100; i++ {
//	        wg.Add(1)
//	        go func() { defer wg.Done(); for j := 0; j < 100; j++ { c.Inc() } }()
//	    }
//	    wg.Wait()
//	    t.Logf("CounterUnsafe.Value() = %d (expected 10000)", c.Value())
//	}
//
// Run it with: go test -race -run TestCounterUnsafeRaces ./...
// The race detector will print "DATA RACE" + a stack trace pointing
// at c.n in CounterUnsafe.Inc and Value. Without -race the test
// "passes" but the value is usually less than 10000 (the unsynchronized
// increments lose updates).
type CounterUnsafe struct {
	n int
}

func (c *CounterUnsafe) Inc()       { c.n++ }
func (c *CounterUnsafe) Value() int { return c.n }
