package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func TestWalk(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1, "WARN": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Walk = %v, want %v", got, want)
		}
	})

	t.Run("multiple-files", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:32:00 WARN c\n")
		writeFile(t, dir, "c.log", "2026-05-21T14:33:00 ERROR d\n2026-05-21T14:34:00 ERROR e\n2026-05-21T14:35:00 ERROR f\n")
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Walk = %v, want %v", got, want)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk on empty dir: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")
		_, err := Walk(dir)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := Walk("/nonexistent/path/that/does/not/exist")
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
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("subdirs should be skipped: got %v, want %v", got, want)
		}
	})
}
