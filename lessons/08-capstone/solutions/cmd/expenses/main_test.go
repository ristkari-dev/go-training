package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestAddSingle(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	stdout, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee")
	if err != nil {
		t.Fatalf("add failed: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "added: 2026-05-18  €4.50    coffee") {
		t.Errorf("stdout missing expected add confirmation, got: %q", stdout)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("JSON contains %d entries, want 1", len(got))
	}
}

func TestAddThenList(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee"); err != nil {
		t.Fatalf("add 1: %v\n%s", err, stderr)
	}
	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "12.00", "lunch"); err != nil {
		t.Fatalf("add 2: %v\n%s", err, stderr)
	}
	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "23.50", "groceries"); err != nil {
		t.Fatalf("add 3: %v\n%s", err, stderr)
	}

	stdout, stderr, err := run("-file="+tmpFile, "list")
	if err != nil {
		t.Fatalf("list failed: %v\nstderr: %s", err, stderr)
	}
	wantLines := []string{
		"2026-05-18  €4.50    coffee",
		"2026-05-18  €12.00   lunch",
		"2026-05-18  €23.50   groceries",
	}
	for _, w := range wantLines {
		if !strings.Contains(stdout, w) {
			t.Errorf("list output missing line %q\nfull stdout:\n%s", w, stdout)
		}
	}
}

func TestAddThenSummary(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	for _, args := range [][]string{
		{"add", "2026-05-18", "4.50", "coffee"},
		{"add", "2026-05-18", "12.00", "lunch"},
		{"add", "2026-05-18", "23.50", "groceries"},
	} {
		if _, stderr, err := run(append([]string{"-file=" + tmpFile}, args...)...); err != nil {
			t.Fatalf("setup add failed: %v\n%s", err, stderr)
		}
	}

	stdout, stderr, err := run("-file="+tmpFile, "summary")
	if err != nil {
		t.Fatalf("summary failed: %v\nstderr: %s", err, stderr)
	}
	want := []string{
		"3 expenses, total €40.00",
		"by category:",
		"biggest category: groceries (€23.50)",
	}
	for _, w := range want {
		if !strings.Contains(stdout, w) {
			t.Errorf("summary output missing %q\nfull stdout:\n%s", w, stdout)
		}
	}
}
