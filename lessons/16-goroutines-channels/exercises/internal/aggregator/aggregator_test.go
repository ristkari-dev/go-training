package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeFile is a test helper. Used inside subtests to create fixture
// log files in a t.TempDir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// TestWalk is a SKELETON. Cover at least:
//   - Single file with known counts → counts match
//   - Multiple files → merged counts
//   - Empty dir → empty map (or nil — either acceptable)
//   - Malformed log line in one file → wrapped error
func TestWalk(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		// TODO:
		//   dir := t.TempDir()
		//   writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")
		//   got, err := Walk(dir)
		//   if err != nil → t.Fatalf("...")
		//   want := map[string]int{"INFO": 1, "WARN": 1}
		//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO: write 3 files; assert merged counts.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO: t.TempDir() then Walk; expect empty map + nil error.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO: write a file with a garbage line; assert err is non-nil
		// and mentions the file path.
	})
}
