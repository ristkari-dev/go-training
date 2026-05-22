package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAggregatorGoldenPath is a SKELETON. The full test exists in the
// solutions tree. Here it stays a stub so the package compiles.
//
// The solution builds the binary into t.TempDir, runs it against a
// fixture directory of log files, and asserts stdout has the expected
// level counts in alphabetical order.
func TestAggregatorGoldenPath(t *testing.T) {
	// TODO:
	//   if _, err := exec.LookPath("go"); err != nil { t.Skip("no go toolchain") }
	//   Build binary; create fixture dir; run; assert stdout.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}
