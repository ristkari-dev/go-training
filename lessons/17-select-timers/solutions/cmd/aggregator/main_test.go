package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestAggregatorGoldenPath(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log",
		"2026-05-21T14:30:00 INFO a\n"+
			"2026-05-21T14:31:00 INFO b\n"+
			"2026-05-21T14:32:00 WARN c\n")
	writeFile(t, logDir, "b.log",
		"2026-05-21T14:33:00 ERROR d\n")

	cmd := exec.Command(bin, "-dir="+logDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	want := "ERROR: 1\nINFO: 2\nWARN: 1\n"
	if stdout.String() != want {
		t.Errorf("stdout mismatch:\ngot:\n%s\nwant:\n%s", stdout.String(), want)
	}
	if stderr.String() != "" {
		t.Errorf("expected empty stderr, got %q", stderr.String())
	}
}

func TestAggregatorMalformed(t *testing.T) {
	// Same non-determinism caveat as L16. Safe because only bad.log errors.
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
	writeFile(t, logDir, "bad.log", "garbage line\n")

	cmd := exec.Command(bin, "-dir="+logDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
	if !strings.Contains(stderr.String(), "bad.log") {
		t.Errorf("stderr should mention bad.log, got %q", stderr.String())
	}
}

func TestAggregatorMissingFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing -dir, got nil")
	}
	if !strings.Contains(stderr.String(), "-dir") {
		t.Errorf("stderr should mention -dir, got %q", stderr.String())
	}
}

func TestAggregatorTimeout(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:31:00 WARN b\n")

	cmd := exec.Command(bin, "-dir="+logDir, "-timeout=1ns")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	// Integration test: verify the binary handles -timeout without
	// crashing AND every file is accounted for (either as a count line
	// in stdout or a "timeout:" line in stderr). Don't assert HOW many
	// timed out — environment-dependent (slow CI may produce 0 timeouts;
	// fast machines may produce all). Unit tests cover timeout semantics.
	//
	// Each fixture has one log entry, so the line count across both
	// streams should equal the file count (2).
	stdoutLines := strings.Count(stdout.String(), "\n")
	stderrTimeouts := strings.Count(stderr.String(), "timeout:")
	if total := stdoutLines + stderrTimeouts; total != 2 {
		t.Errorf("expected 2 files accounted for, got %d (stdout=%q stderr=%q)",
			total, stdout.String(), stderr.String())
	}
}
