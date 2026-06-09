package main

import "testing"

// TestRunSkeleton is a placeholder so the exercises package compiles and
// tests pass before forwarder.deliver is implemented. It must NOT call run
// (run → Send → deliver, which panics until you implement it). Once deliver
// is done, mirror the solutions main_test.go (forward to an httptest
// aggregator and assert no error).
func TestRunSkeleton(t *testing.T) {
	_ = run
}
