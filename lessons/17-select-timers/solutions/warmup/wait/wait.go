// Package wait is the lesson 17 warm-up reference implementation.
package wait

import (
	"errors"
	"time"
)

// ErrTimeout is returned by WaitWithTimeout when d elapses first.
var ErrTimeout = errors.New("wait: timeout")

// WaitWithTimeout returns the first value from ch, or (zero T,
// ErrTimeout) after d.
func WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error) {
	var zero T
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(d):
		return zero, ErrTimeout
	}
}
