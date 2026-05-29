package breaker

import (
	"errors"
	"testing"
	"time"
)

// fakeClock is a manually-advanced time source for deterministic tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time      { return c.t }
func (c *fakeClock) add(d time.Duration) { c.t = c.t.Add(d) }

// TestBreaker is a SKELETON. Cover: opens after maxFailures; fails fast
// while open (fn not called); half-open→closed on trial success;
// half-open→open on trial failure. Use WithClock(clk.now) + clk.add to
// drive the cooldown deterministically.
func TestBreaker(t *testing.T) {
	// TODO:
	//   clk := &fakeClock{t: time.Unix(0, 0)}
	//   b := New(3, time.Minute, WithClock(clk.now))
	//   ... fail 3 times, assert Open, assert fast-fail, advance clock, trial ...
	_ = errors.New
	_ = New
}
