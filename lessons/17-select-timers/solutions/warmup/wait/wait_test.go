package wait

import (
	"errors"
	"testing"
	"time"
)

func TestWaitImmediate(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	v, err := WaitWithTimeout(ch, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestWaitTimeout(t *testing.T) {
	ch := make(chan int)

	v, err := WaitWithTimeout(ch, 10*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
	if v != 0 {
		t.Errorf("expected zero value 0, got %d", v)
	}
}

func TestWaitGenericString(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "hi"

	v, err := WaitWithTimeout(ch, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hi" {
		t.Errorf("got %q, want %q", v, "hi")
	}
}

func TestWaitZeroDuration(t *testing.T) {
	// d=0 fires immediately. The select picks one of the two ready
	// cases at random; either the channel receive (if ch has a value)
	// or the time.After(0) which fires immediately. With an empty
	// channel, only time.After is ready → ErrTimeout.
	ch := make(chan int)
	_, err := WaitWithTimeout(ch, 0)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout with d=0 on empty channel, got %v", err)
	}
}
