package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: same pattern as L18.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorCancelled is a SKELETON. Builds the binary, runs it
// against a fixture; sends SIGINT while running; asserts exit 0 +
// "cancelled by user" in stderr.
func TestAggregatorCancelled(t *testing.T) {
	// TODO: build binary; exec it; cmd.Process.Signal(os.Interrupt);
	// assert cmd.Wait error is nil (exit 0); stderr contains "cancelled"
}
