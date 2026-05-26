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

func TestWalkNoTimeout(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")

		result, err := Walk(dir, 0)
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

		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}

		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts, got %v", result.Counts)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// Note: same non-deterministic ordering caveat as L16 — only
		// bad.log errors, so any goroutine ordering eventually hits it.
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")

		_, err := Walk(dir, 0)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := Walk("/nonexistent/path/that/does/not/exist", 0)
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

		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("subdirs should be skipped: Counts = %v, want %v", result.Counts, want)
		}
	})
}

func TestWalkAllFilesTimedOut(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")

	// 1µs (not 1ns) gives ~1000× headroom over scheduler/race overhead.
	// 1ns mostly works, but under -race instrumentation (5-20× overhead)
	// a worker can occasionally finish before the timer's first iteration.
	// A microsecond is still effectively instant for this test's purpose.
	result, err := Walk(dir, 1*time.Microsecond)
	if err != nil {
		t.Fatalf("timeouts should not error: %v", err)
	}
	if len(result.Counts) != 0 {
		t.Errorf("expected empty Counts when all files time out, got %v", result.Counts)
	}
	if len(result.TimedOut) != 2 {
		t.Errorf("expected 2 timed-out files, got %d: %v", len(result.TimedOut), result.TimedOut)
	}
}
