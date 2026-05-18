package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runBinary builds the expenses binary into a temp location and returns a
// helper that runs it with the given args, returning stdout, stderr, and
// any error from the process (non-zero exit codes show up as *exec.ExitError).
//
// This is the canonical "test a Go binary by building it once and running
// it many times" pattern. The build cost is paid once; each call to the
// returned function is fast (just exec).
type runFn func(args ...string) (stdout, stderr string, err error)

func buildAndRun(t *testing.T) runFn {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "expenses")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return func(args ...string) (string, string, error) {
		cmd := exec.Command(binPath, args...)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
}

// TestAddSingle is a SKELETON. Test that `add 2026-05-18 4.50 coffee`:
//   - exits with code 0
//   - prints "added: 2026-05-18  €4.50    coffee" to stdout
//   - creates the JSON file at the -file path
//   - the JSON file contains exactly one expense
func TestAddSingle(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO:
	//   stdout, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee")
	//   if err != nil → t.Fatalf with stderr
	//   assert stdout contains "added: 2026-05-18  €4.50    coffee"
	//   assert os.ReadFile(tmpFile) succeeds and the JSON parses as 1 entry
	_ = run
	_ = tmpFile
	_ = os.ReadFile
}

// TestAddThenList is a SKELETON. Test that after adding three expenses,
// `list` prints them in insertion order, one per line.
func TestAddThenList(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO: add three expenses, then `list`, assert all three lines present
	// in stdout in insertion order. Use strings.Contains or split by "\n".
	_ = run
	_ = tmpFile
}

// TestAddThenSummary is a SKELETON. Test that after adding three expenses
// (coffee 4.50, lunch 12, groceries 23.50), `summary` prints the expected
// totals + the "biggest category: groceries (€23.50)" line.
func TestAddThenSummary(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO: add 3, then `summary`, assert key lines present:
	//   "3 expenses, total €40.00"
	//   "biggest category: groceries (€23.50)"
	// Don't assert the bar-chart whitespace — it's locale/font-fragile.
	_ = run
	_ = tmpFile
}
