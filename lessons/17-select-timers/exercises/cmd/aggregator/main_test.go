package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAggregatorGoldenPath is a SKELETON.
func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: build binary; create fixture dir; run; assert stdout.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorTimeout is a SKELETON. Run with -timeout=1ns; assert
// stderr contains "timeout:" for each fixture file; stdout is empty
// (or just trailing newline).
func TestAggregatorTimeout(t *testing.T) {
	// TODO: build binary; fixture dir with 2 files; run -timeout=1ns;
	// assert stderr mentions each fixture path.
}
