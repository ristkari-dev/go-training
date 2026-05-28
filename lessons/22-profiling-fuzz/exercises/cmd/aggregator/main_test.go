package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SKELETONS. Carries forward L19's CLI integration tests + adds a
// TODO-stub for TestAggregatorPool. Fill in once internal/aggregator's
// WalkPool is implemented.

func buildBinary(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "aggregator")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	return bin
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestAggregatorGoldenPath is a SKELETON. Build binary, write log files,
// run with -dir, assert stdout matches expected count lines.
func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: see L19's solutions test as the reference.
	_ = buildBinary
	_ = writeFile
	_ = bytes.Buffer{}
	_ = exec.Command
	_ = fmt.Sprintf
	_ = strings.Contains
	_ = time.Second
}

// TestAggregatorMalformed is a SKELETON. Verify non-zero exit + stderr
// mentions the bad file.
func TestAggregatorMalformed(t *testing.T) {
	// TODO
}

// TestAggregatorMissingFlag is a SKELETON. Verify non-zero exit + stderr
// mentions -dir.
func TestAggregatorMissingFlag(t *testing.T) {
	// TODO
}

// TestAggregatorTimeout is a SKELETON. Verify -timeout=1ns triggers
// timeouts; invariant: total files accounted for = N (sum of stdout count
// lines + stderr "timeout:" lines).
func TestAggregatorTimeout(t *testing.T) {
	// TODO
}

// TestAggregatorCancelled is a SKELETON. Verify SIGINT triggers graceful
// cancellation (exit 0; no "error:" in stderr).
func TestAggregatorCancelled(t *testing.T) {
	// TODO
}

// TestAggregatorPool is a SKELETON for the new -mode=pool path. Verify
// -mode=pool -workers=2 produces the same golden-path stdout.
func TestAggregatorPool(t *testing.T) {
	// TODO:
	//   bin := buildBinary(t)
	//   logDir := t.TempDir()
	//   writeFile(t, logDir, "a.log", "...INFO a\n...WARN slow\n")
	//   writeFile(t, logDir, "b.log", "...ERROR oops\n")
	//   cmd := exec.Command(bin, "-dir="+logDir, "-mode=pool", "-workers=2")
	//   ...
	//   want := "ERROR: 1\nINFO: 1\nWARN: 1\n"
	//   compare stdout
}
