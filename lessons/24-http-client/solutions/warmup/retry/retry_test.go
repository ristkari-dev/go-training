package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoEventualSuccess(t *testing.T) {
	calls := 0
	err := Do(context.Background(), 5, time.Microsecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Do = %v, want nil", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestDoExhausted(t *testing.T) {
	calls := 0
	want := errors.New("always")
	err := Do(context.Background(), 3, time.Microsecond, func() error {
		calls++
		return want
	})
	if !errors.Is(err, want) {
		t.Errorf("Do = %v, want %v", err, want)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestDoPermanentStopsEarly(t *testing.T) {
	calls := 0
	sentinel := errors.New("bad request")
	err := Do(context.Background(), 5, time.Microsecond, func() error {
		calls++
		return Permanent(sentinel)
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Do = %v, want %v (unwrapped)", err, sentinel)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (permanent stops immediately)", calls)
	}
}

func TestDoContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := Do(ctx, 10, 50*time.Millisecond, func() error {
		calls++
		cancel() // cancel during the first attempt; backoff should abort
		return errors.New("transient")
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (cancel aborts backoff)", calls)
	}
}
