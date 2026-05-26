// Package sleepctx is the lesson 19 warm-up: cancellable sleep via
// select + ctx.Done.
//
// SleepWithCtx sleeps for d, returning nil on clean completion or
// ctx.Err() if ctx is cancelled first. The textbook pattern is three
// lines of select:
//
//	select {
//	case <-time.After(d):  return nil
//	case <-ctx.Done():     return ctx.Err()
//	}
//
// This composes the L17 "select with timeout" pattern with the L19
// "ctx as a cancellation signal" idiom. Any function that does
// blocking work should accept ctx for the same reason — to be
// cancellable.
package sleepctx

import (
	"context"
	"time"
)

// SleepWithCtx blocks for d. Returns nil on clean completion, or
// ctx.Err() if ctx is cancelled first.
//
// ctx.Err() returns one of:
//   - context.Canceled — caller invoked cancel()
//   - context.DeadlineExceeded — ctx had a deadline that passed
//   - nil — only possible if ctx wasn't done yet (won't happen here
//     because we only consult Err inside the Done case)
//
// Callers distinguish via errors.Is(err, context.Canceled) etc.
//
// Hint:
//
//	select {
//	case <-time.After(d):
//	    return nil
//	case <-ctx.Done():
//	    return ctx.Err()
//	}
func SleepWithCtx(ctx context.Context, d time.Duration) error {
	_ = time.After
	panic("TODO: select with case <-time.After(d) and case <-ctx.Done()")
}
