package sleepctx

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestSleepCompletes is a SKELETON. Use a ctx that won't be cancelled
// (context.Background()) and a short duration; assert nil error.
func TestSleepCompletes(t *testing.T) {
	// TODO:
	//   err := SleepWithCtx(context.Background(), 10*time.Millisecond)
	//   if err != nil → t.Errorf("expected nil, got %v", err)
	_ = context.Background
	_ = time.Millisecond
}

// TestSleepCancelled is a SKELETON. Use a ctx that's cancelled before
// SleepWithCtx is called; assert returns context.Canceled via
// errors.Is.
func TestSleepCancelled(t *testing.T) {
	// TODO:
	//   ctx, cancel := context.WithCancel(context.Background())
	//   cancel()  // cancel BEFORE calling SleepWithCtx
	//   err := SleepWithCtx(ctx, time.Hour)  // would sleep forever without cancel
	//   if !errors.Is(err, context.Canceled) → t.Errorf("expected context.Canceled, got %v", err)
	_ = errors.Is
}

// TestSleepDeadline is a SKELETON. Use context.WithTimeout(parent, 10ms)
// and SleepWithCtx for 1*time.Hour; assert returns
// context.DeadlineExceeded.
func TestSleepDeadline(t *testing.T) {
	// TODO:
	//   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	//   defer cancel()
	//   err := SleepWithCtx(ctx, time.Hour)
	//   if !errors.Is(err, context.DeadlineExceeded) → t.Errorf("expected DeadlineExceeded, got %v", err)
}
