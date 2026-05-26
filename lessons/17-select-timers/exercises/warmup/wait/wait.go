// Package wait is the lesson 17 warm-up: blocking on a channel with a
// timeout via select + time.After.
//
// WaitWithTimeout returns the first value sent on ch, or ErrTimeout
// if d elapses first. The implementation is the textbook timeout
// pattern in three lines of select.
//
// Generic via [T any] (lesson 12) so the same function works for any
// channel element type.
package wait

import (
	"errors"
	"time"
)

// ErrTimeout is returned by WaitWithTimeout when the deadline d
// elapses before a value arrives on ch.
//
// Callers can errors.Is(err, ErrTimeout) to distinguish timeout
// from other errors (which WaitWithTimeout doesn't itself produce —
// the type signature allows it for future extensibility).
var ErrTimeout = errors.New("wait: timeout")

// WaitWithTimeout returns the first value from ch, or (zero T,
// ErrTimeout) if d elapses first.
//
// Examples:
//
//	ch := make(chan int, 1); ch <- 42
//	v, err := WaitWithTimeout(ch, time.Second)   // v=42, err=nil
//
//	emptyCh := make(chan string)
//	v, err := WaitWithTimeout(emptyCh, time.Millisecond)
//	// v="", err=ErrTimeout
//
// Hint:
//  1. var zero T
//  2. select {
//     case v := <-ch:
//     return v, nil
//     case <-time.After(d):
//     return zero, ErrTimeout
//     }
//
// The `var zero T` idiom (L12 callback) gives you T's zero value
// without knowing the concrete type. Don't write `return T{}, ...`
// (T might not be a struct) or `return *new(T), ...` (ugly).
func WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error) {
	_ = time.After
	panic("TODO: var zero T; select with case v := <-ch and case <-time.After(d)")
}
