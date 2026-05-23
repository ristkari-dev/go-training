package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// writeFile is a test helper.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// TestWalkNoTimeout is a SKELETON. With timeout=0 (no timeout), Walk
// behaves like L16: all files processed; WalkResult.Counts has the
// aggregate; WalkResult.TimedOut is empty.
func TestWalkNoTimeout(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write one fixture file; call Walk(dir, 0); assert
		// result.Counts and len(result.TimedOut) == 0.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO: similar pattern, assert merged counts.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO: empty dir; assert Counts is empty map (not nil),
		// TimedOut is empty.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO: write a malformed file; assert err != nil and mentions
		// the file path.
	})
}

// TestWalkAllFilesTimedOut is a SKELETON. With a tiny timeout
// (1 nanosecond), every file's select fires the timeout case before
// the receive case. Assert: Counts is empty; all files appear in
// TimedOut; err is nil (timeouts aren't errors).
func TestWalkAllFilesTimedOut(t *testing.T) {
	// TODO:
	//   dir := t.TempDir()
	//   writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n")
	//   writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN slow\n")
	//   result, err := Walk(dir, 1*time.Nanosecond)
	//   if err != nil → t.Fatalf("timeouts shouldn't error: %v", err)
	//   if len(result.Counts) != 0 → t.Errorf("expected empty counts, got %v", result.Counts)
	//   if len(result.TimedOut) != 2 → t.Errorf("expected 2 timed-out files, got %v", result.TimedOut)
	_ = time.Nanosecond
}
