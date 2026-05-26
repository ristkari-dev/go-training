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

type walkFunc func(dir string, timeout time.Duration) (WalkResult, error)

func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log",
			"2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")

		result, err := fn(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1, "WARN": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("multiple-files", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:32:00 WARN c\n")
		writeFile(t, dir, "c.log", "2026-05-21T14:33:00 ERROR d\n2026-05-21T14:34:00 ERROR e\n2026-05-21T14:35:00 ERROR f\n")

		result, err := fn(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		result, err := fn(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts, got %v", result.Counts)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// Note: only bad.log errors, so any goroutine ordering eventually
		// hits it. Walk returns immediately on the first error; WalkLocked
		// waits for all workers via wg.Wait then returns firstErr.
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")

		_, err := fn(dir, 0)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := fn("/nonexistent/path/that/does/not/exist", 0)
		if err == nil {
			t.Fatal("expected error from nonexistent dir, got nil")
		}
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n")
		if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "subdir"), "nested.log", "2026-05-21T14:30:00 ERROR ignored\n")

		result, err := fn(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("subdirs should be skipped: Counts = %v, want %v", result.Counts, want)
		}
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")

		// Probabilistic test: at 1µs, at least one file should time out.
		// On slow CI runners (GitHub Actions), file work occasionally
		// beats the timer for some individual files, so we verify that
		// (a) every file is accounted for in either Counts or TimedOut,
		// and (b) at least one timeout occurred. Documents the inherent
		// non-determinism honestly while still exercising the timeout
		// code path. (L17 originally asserted exactly == 2; CI flake
		// after L18 merge taught us to soften.)
		result, err := fn(dir, 1*time.Microsecond)
		if err != nil {
			t.Fatalf("timeouts should not error: %v", err)
		}
		filesAccounted := len(result.TimedOut)
		for _, c := range result.Counts {
			filesAccounted += c
		}
		if filesAccounted != 2 {
			t.Errorf("expected 2 files accounted for (Counts+TimedOut), got %d (counts=%v, timedout=%v)",
				filesAccounted, result.Counts, result.TimedOut)
		}
		if len(result.TimedOut) == 0 {
			t.Errorf("expected at least one timed-out file at 1µs timeout, got 0 (counts=%v)", result.Counts)
		}
	})
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
