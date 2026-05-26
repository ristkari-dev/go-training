package sleepctx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSleepCompletes(t *testing.T) {
	err := SleepWithCtx(context.Background(), 10*time.Millisecond)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestSleepCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel BEFORE the sleep starts — Done is already closed
	err := SleepWithCtx(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSleepCancelledMidway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := SleepWithCtx(ctx, time.Hour) // would block forever; cancel rescues
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSleepDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := SleepWithCtx(ctx, time.Hour)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
