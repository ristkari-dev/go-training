// Package testutil provides small helpers for tests in lesson 15.
//
// PROVIDED — students don't implement this file; they use AssertGolden
// from cmd/expenses/main_test.go.
package testutil

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

// Update enables overwriting golden files when go test is invoked with
// `-update`. The flag is registered at package init.
//
// To regenerate a golden file after legitimate output changes:
//
//	go test -update ./path/to/test/
var Update = flag.Bool("update", false, "regenerate golden files")

// AssertGolden compares got against the contents of the file at path.
// On mismatch, fails the test with a diff. If -update is set, writes
// got to path instead of comparing (regenerates the golden file).
//
// Usage:
//
//	output := runMyCommand(...)
//	testutil.AssertGolden(t, output, "testdata/expected.golden")
func AssertGolden(t *testing.T, got []byte, path string) {
	t.Helper()
	if *Update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("update %s: %v", path, err)
		}
		t.Logf("updated %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output mismatch with %s:\n--- got (%d bytes) ---\n%s\n--- want (%d bytes) ---\n%s",
			path, len(got), got, len(want), want)
	}
}
