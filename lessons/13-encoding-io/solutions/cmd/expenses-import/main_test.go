package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/solutions/expense"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "expenses-import")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	return bin
}

func TestImporterGoldenPath(t *testing.T) {
	bin := buildBinary(t)
	jsonPath := filepath.Join(t.TempDir(), "expenses.json")

	cmd := exec.Command(bin, "-file="+jsonPath)
	cmd.Stdin = bytes.NewReader([]byte("2026-05-21,4.50,coffee\n2026-05-21,12.00,lunch\n"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run importer: %v (stderr: %s)", err, stderr.String())
	}

	// Verify the JSON file exists and contains both expenses.
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read result file: %v", err)
	}
	var got []expense.Expense
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d expenses, want 2; data=%s", len(got), string(data))
	}
	if got[0].Category != "coffee" || got[1].Category != "lunch" {
		t.Errorf("unexpected categories: %+v", got)
	}
}

func TestImporterMalformed(t *testing.T) {
	bin := buildBinary(t)
	jsonPath := filepath.Join(t.TempDir(), "expenses.json")

	cmd := exec.Command(bin, "-file="+jsonPath)
	cmd.Stdin = bytes.NewReader([]byte("2026-05-21,oops,coffee\n"))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
	if !strings.Contains(stderr.String(), "line 1") {
		t.Errorf("stderr should mention line 1, got %q", stderr.String())
	}
}

func TestImporterMissingFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	cmd.Stdin = bytes.NewReader([]byte(""))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing -file, got nil")
	}
	if !strings.Contains(stderr.String(), "-file") {
		t.Errorf("stderr should mention -file, got %q", stderr.String())
	}
}
