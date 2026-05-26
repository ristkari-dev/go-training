package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: see L17 pattern.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorMode is a SKELETON. Run with -mode=mutex and assert
// the output matches the default (-mode=channel) for the same input.
func TestAggregatorMode(t *testing.T) {
	// TODO: build binary; create fixture dir; run twice (default and
	// -mode=mutex); assert stdout matches.
}
