package main

import (
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/testutil"
)

// TestSummaryGolden is a SKELETON. The full test exists in solutions.
// In the exercise tree, the test is a stub so the package compiles.
//
// The solution test seeds a MemoryStore with known expenses, runs
// cmdSummary capturing stdout into a bytes.Buffer, and compares the
// output to testdata/summary.golden via testutil.AssertGolden. With
// `go test -update`, it regenerates the golden file.
func TestSummaryGolden(t *testing.T) {
	// TODO: seed store; run cmdSummary; capture into bytes.Buffer;
	// testutil.AssertGolden(t, buf.Bytes(), filepath.Join("testdata", "summary.golden"))
	_ = testutil.AssertGolden
	_ = filepath.Join
}
