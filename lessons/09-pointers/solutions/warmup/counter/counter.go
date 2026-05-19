// Package counter is the lesson 09 warm-up reference implementation.
package counter

// Counter holds an int count.
type Counter struct {
	n int
}

// Inc increments the counter by 1. Pointer receiver — mutates the caller's value.
func (c *Counter) Inc() {
	c.n++
}

// Value returns the current count. Value receiver — read-only.
func (c Counter) Value() int {
	return c.n
}
