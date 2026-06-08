package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionStamping builds logstatsd with -ldflags -X stamping a
// version + commit into buildinfo, runs `-version`, and asserts the
// stamped values appear. Proves the link-time stamping mechanism
// end-to-end without Docker. Always runs (only needs the go toolchain).
// (Containerization itself is L27; here the daemon just keeps its
// version-stamping as it gains observability.)
func TestVersionStamping(t *testing.T) {
	const pkg = "github.com/ristkari-dev/go-training/lessons/28-observability/solutions/cmd/logstatsd"
	const buildinfoPkg = "github.com/ristkari-dev/go-training/lessons/28-observability/solutions/warmup/buildinfo"

	bin := filepath.Join(t.TempDir(), "logstatsd")
	ldflags := "-X " + buildinfoPkg + ".Version=test-v9 -X " + buildinfoPkg + ".Commit=abc123"
	// Build by IMPORT PATH (not ./...): the test's CWD is this package's
	// dir, so a relative path would double-nest.
	build := exec.Command("go", "build", "-ldflags", ldflags, "-o", bin, pkg)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	out, err := exec.Command(bin, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("run -version: %v\n%s", err, out)
	}
	got := string(out)
	for _, want := range []string{"test-v9", "abc123"} {
		if !strings.Contains(got, want) {
			t.Errorf("-version output %q missing %q", strings.TrimSpace(got), want)
		}
	}
}
