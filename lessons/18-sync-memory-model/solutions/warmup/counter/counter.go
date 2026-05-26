// Package counter is the lesson 18 warm-up reference implementation.
package counter

import "sync"

// Counter is safe for concurrent use.
type Counter struct {
	mu sync.Mutex
	n  int
}

// Inc atomically increments the counter by 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

// Value returns the current counter value.
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// CounterUnsafe is the same interface without the mutex. Used only
// for slides/README demonstrations — DELIBERATELY buggy.
type CounterUnsafe struct {
	n int
}

func (c *CounterUnsafe) Inc()       { c.n++ }
func (c *CounterUnsafe) Value() int { return c.n }
