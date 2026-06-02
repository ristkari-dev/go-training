package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionStamping builds logstatsd with -ldflags -X stamping a
// version + commit into buildinfo, runs `-version`, and asserts the
// stamped values appear. Proves the link-time stamping mechanism
// end-to-end without Docker. Always runs (only needs the go toolchain).
func TestVersionStamping(t *testing.T) {
	const pkg = "github.com/ristkari-dev/go-training/lessons/27-container/solutions/cmd/logstatsd"
	const buildinfoPkg = "github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo"

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

// TestDockerBuild builds + runs the image. Opt-in: skips unless
// DOCKER_TEST=1 and docker is on PATH, so normal CI stays fast.
func TestDockerBuild(t *testing.T) {
	if os.Getenv("DOCKER_TEST") != "1" {
		t.Skip("set DOCKER_TEST=1 to run the docker build test")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not on PATH")
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "..")) // → repo root
	if err != nil {
		t.Fatal(err)
	}
	tag := "logstatsd:l27test"
	build := exec.Command("docker", "build",
		"-f", filepath.Join("lessons", "27-container", "Dockerfile"),
		"--build-arg", "VERSION=docker-v9",
		"-t", tag, ".")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("docker build: %v\n%s", err, out)
	}
	defer exec.Command("docker", "rmi", "-f", tag).Run() //nolint:errcheck

	out, err := exec.Command("docker", "run", "--rm", tag, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("docker run -version: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "docker-v9") {
		t.Errorf("docker -version output %q missing docker-v9", strings.TrimSpace(string(out)))
	}
}
