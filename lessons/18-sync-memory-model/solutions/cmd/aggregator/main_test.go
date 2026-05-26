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
		"2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n2026-05-21T14:32:00 WARN c\n")
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
}

func TestAggregatorMalformed(t *testing.T) {
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

// TestAggregatorMode runs the binary twice — once with default mode
// (channel) and once with -mode=mutex — over the same fixture. Asserts
// stdout matches between the two runs. Demonstrates that both Walk
// variants produce identical output for the same input.
func TestAggregatorMode(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 WARN b\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:32:00 ERROR c\n")

	runOnce := func(mode string) string {
		t.Helper()
		args := []string{"-dir=" + logDir}
		if mode != "" {
			args = append(args, "-mode="+mode)
		}
		cmd := exec.Command(bin, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("aggregator (mode=%q): %v (stderr: %s)", mode, err, stderr.String())
		}
		return stdout.String()
	}

	defaultOut := runOnce("")
	channelOut := runOnce("channel")
	mutexOut := runOnce("mutex")

	if defaultOut != channelOut {
		t.Errorf("default mode should match -mode=channel:\ndefault:\n%s\nchannel:\n%s",
			defaultOut, channelOut)
	}
	if channelOut != mutexOut {
		t.Errorf("channel and mutex modes should produce identical output:\nchannel:\n%s\nmutex:\n%s",
			channelOut, mutexOut)
	}
}

func TestAggregatorTimeout(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:31:00 WARN b\n")

	// 1ms (not 1ns) gives subprocess-level headroom. Even though the
	// timer fires nearly instantly, file I/O + subprocess startup
	// under -race can take hundreds of microseconds. 1ms is still
	// effectively instant from a human perspective but reliably
	// shorter than the file work — every file appears in TimedOut.
	cmd := exec.Command(bin, "-dir="+logDir, "-timeout=1ms")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	if stdout.String() != "" {
		t.Errorf("expected empty stdout when everything times out, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "timeout:") {
		t.Errorf("expected 'timeout:' in stderr, got %q", stderr.String())
	}
}

// TestAggregatorInvalidMode asserts the CLI rejects unknown -mode values.
func TestAggregatorInvalidMode(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO ok\n")

	cmd := exec.Command(bin, "-dir="+logDir, "-mode=bogus")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for invalid -mode, got nil")
	}
	if !strings.Contains(stderr.String(), "-mode") {
		t.Errorf("stderr should mention -mode, got %q", stderr.String())
	}
}
