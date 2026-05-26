package wait

import (
	"errors"
	"testing"
	"time"
)

// TestWaitImmediate is a SKELETON. Use a buffered channel with one
// value already in it; WaitWithTimeout should return that value
// without ever hitting the timeout.
func TestWaitImmediate(t *testing.T) {
	// TODO:
	//   ch := make(chan int, 1); ch <- 42
	//   v, err := WaitWithTimeout(ch, time.Second)
	//   if err != nil → t.Fatalf("unexpected error: %v", err)
	//   if v != 42 → t.Errorf("got %d, want 42", v)
	_ = time.Second
}

// TestWaitTimeout is a SKELETON. Use a channel that nobody sends on;
// WaitWithTimeout should hit the timeout and return ErrTimeout.
func TestWaitTimeout(t *testing.T) {
	// TODO:
	//   ch := make(chan int)
	//   _, err := WaitWithTimeout(ch, time.Millisecond)
	//   if !errors.Is(err, ErrTimeout) → t.Errorf("expected ErrTimeout, got %v", err)
	_ = errors.Is
}

// TestWaitGenericString is a SKELETON. Verify the function works
// with T=string (not just int).
func TestWaitGenericString(t *testing.T) {
	// TODO:
	//   ch := make(chan string, 1); ch <- "hi"
	//   v, err := WaitWithTimeout(ch, time.Second)
	//   assert v == "hi", err == nil
}
