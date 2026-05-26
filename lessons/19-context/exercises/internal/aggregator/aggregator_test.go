package aggregator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// walkFunc — common shape of Walk and WalkLocked. NOTE: ctx is now first.
type walkFunc func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error)

// runWalkSuite runs the same sub-tests against both Walk and WalkLocked.
// Skeleton: cases TODO-stubbed.
func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write fixture; fn(context.Background(), dir, 0); assert Counts.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO
		_ = strings.Contains
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		// TODO
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		// TODO
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		// TODO. 1µs per L17/L18 empirical rationale.
		_ = time.Microsecond
	})

	t.Run("ctx-cancelled-before-walk", func(t *testing.T) {
		// TODO:
		//   ctx, cancel := context.WithCancel(context.Background())
		//   cancel()
		//   _, err := fn(ctx, dir, 0)
		//   if !errors.Is(err, context.Canceled) → t.Errorf("...")
		_ = errors.Is
		_ = context.WithCancel
	})

	t.Run("ctx-deadline-exceeded", func(t *testing.T) {
		// TODO:
		//   ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		//   defer cancel()
		//   _, err := fn(ctx, dir, 0)
		//   if !errors.Is(err, context.DeadlineExceeded) → t.Errorf("...")
	})

	_ = fn
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
