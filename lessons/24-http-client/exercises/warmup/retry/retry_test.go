package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestDo is a SKELETON. Cover: eventual success after N failures,
// exhaustion returns the last error, Permanent stops at one call, and
// ctx cancellation aborts the backoff.
func TestDo(t *testing.T) {
	// TODO:
	//   calls := 0
	//   err := Do(context.Background(), 5, time.Microsecond, func() error {
	//       calls++; if calls < 3 { return errors.New("x") }; return nil
	//   })
	//   assert err == nil && calls == 3
	_ = errors.New
	_ = context.Background
	_ = time.Microsecond
}
