// Package sleepctx is the lesson 19 warm-up reference implementation.
package sleepctx

import (
	"context"
	"time"
)

// SleepWithCtx blocks for d or until ctx is cancelled.
func SleepWithCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
