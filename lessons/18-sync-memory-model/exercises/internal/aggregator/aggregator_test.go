package aggregator

import (
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

// walkFunc is the common type of Walk and WalkLocked, used to
// parameterize tests so both implementations run the same suite.
type walkFunc func(dir string, timeout time.Duration) (WalkResult, error)

// runWalkSuite executes the same set of sub-tests against a walkFunc.
// Skeleton: cases are TODO-stubbed; both Walk and WalkLocked panic
// inside, so we don't actually call them.
func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write a fixture file; call fn(dir, 0); assert Counts.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO.
		_ = strings.Contains
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		// TODO.
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		// TODO.
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		// TODO. Use 1*time.Microsecond per L17 lesson (race-detector headroom).
		_ = time.Microsecond
	})

	_ = fn
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
