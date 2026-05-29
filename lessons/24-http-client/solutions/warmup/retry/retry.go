// Package retry runs a function with bounded retries and exponential
// backoff. It is the lesson 24 warm-up reference implementation.
package retry

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// permanent wraps an error to signal that Do must stop immediately
// rather than retry. Mirrors cenkalti/backoff.Permanent.
type permanent struct{ err error }

func (p *permanent) Error() string { return p.err.Error() }
func (p *permanent) Unwrap() error { return p.err }

// Permanent marks err as non-retryable: Do returns it (unwrapped) at
// once instead of retrying.
func Permanent(err error) error { return &permanent{err: err} }

// Do calls fn up to attempts times. On a non-nil error it sleeps with
// exponential backoff (base * 2^n) plus full jitter before the next
// try, returning early if ctx is cancelled. A fn error wrapped with
// Permanent stops the loop immediately. Returns nil on the first
// success, or the last (unwrapped) error.
func Do(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	var err error
	for n := 0; n < attempts; n++ {
		err = fn()
		if err == nil {
			return nil
		}
		var p *permanent
		if errors.As(err, &p) {
			return p.err // terminal — do not retry
		}
		if n == attempts-1 {
			break // no sleep after the final attempt
		}
		// Full jitter: sleep in [0, base*2^n].
		backoff := base << n
		d := time.Duration(rand.Int64N(int64(backoff) + 1))
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}
	return err
}
