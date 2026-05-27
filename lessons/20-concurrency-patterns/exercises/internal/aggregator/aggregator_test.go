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

type walkFunc func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error)

// runWalkSuite is the parameterized test suite. SKELETON: students fill
// in the same 9 sub-tests carried forward from L17/L18/L19 (single-file,
// multiple-files, empty-dir, malformed-line-returns-error,
// nonexistent-dir-returns-error, subdirs-skipped, all-files-timed-out,
// ctx-cancelled-before-walk, ctx-deadline-exceeded). Once filled in,
// TestWalk + TestWalkLocked + TestWalkPool below all run the same suite
// against the three Walk variants.
func runWalkSuite(t *testing.T, fn walkFunc) {
	// TODO: carry forward L19's runWalkSuite sub-tests verbatim. Examples:
	//
	//   t.Run("single-file", func(t *testing.T) {
	//       dir := t.TempDir()
	//       writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n")
	//       result, err := fn(context.Background(), dir, 0)
	//       if err != nil { t.Fatalf("Walk: %v", err) }
	//       want := map[string]int{"INFO": 1}
	//       if !reflect.DeepEqual(result.Counts, want) { t.Errorf(...) }
	//   })
	//
	//   ...plus 8 more sub-tests covering multiple-files, empty-dir,
	//   malformed-line-returns-error, nonexistent-dir-returns-error,
	//   subdirs-skipped, all-files-timed-out (invariant: files accounted
	//   for = file count), ctx-cancelled-before-walk, ctx-deadline-exceeded.
	_ = writeFile
	_ = reflect.DeepEqual
	_ = strings.Contains
	_ = time.Microsecond
	_ = errors.Is
	_ = context.WithCancel
	_ = fn
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}

// TestWalkPool parameterizes runWalkSuite over WalkPool via an adapter
// that closes over a fixed workers=4.
func TestWalkPool(t *testing.T) {
	walkPoolAdapter := func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
		return WalkPool(ctx, dir, timeout, 4)
	}
	runWalkSuite(t, walkPoolAdapter)
}

// TestWalkPoolWorkers verifies the `workers` parameter is respected.
// SKELETON: students should test workers=1 (serial), workers=2, workers=8,
// and workers=0 (default → runtime.NumCPU()) all produce the same WalkResult.
func TestWalkPoolWorkers(t *testing.T) {
	// TODO:
	//   dir := t.TempDir()
	//   writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	//   writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")
	//   writeFile(t, dir, "c.log", "2026-05-21T14:32:00 ERROR c\n")
	//   for _, workers := range []int{1, 2, 8, 0} {
	//       result, err := WalkPool(context.Background(), dir, 0, workers)
	//       if err != nil { t.Fatalf(...) }
	//       want := map[string]int{"INFO": 1, "WARN": 1, "ERROR": 1}
	//       if !reflect.DeepEqual(result.Counts, want) { t.Errorf(...) }
	//   }
}
