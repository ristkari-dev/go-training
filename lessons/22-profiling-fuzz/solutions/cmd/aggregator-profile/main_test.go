package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesProfiles(t *testing.T) {
	tmp := t.TempDir()
	cpu := filepath.Join(tmp, "cpu.prof")
	mem := filepath.Join(tmp, "heap.prof")

	var out, errBuf bytes.Buffer
	code := run([]string{
		"-n=20", "-lines=10",
		"-cpuprofile=" + cpu,
		"-memprofile=" + mem,
	}, &out, &errBuf)

	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errBuf.String())
	}
	for _, p := range []string{cpu, mem} {
		fi, err := os.Stat(p)
		if err != nil {
			t.Errorf("profile %s not written: %v", p, err)
			continue
		}
		if fi.Size() == 0 {
			t.Errorf("profile %s is empty", p)
		}
	}
	if out.Len() == 0 {
		t.Errorf("expected a summary line on stdout")
	}
}

func TestRunBadFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"-nonsense"}, &out, &errBuf); code != 2 {
		t.Errorf("bad flag exit code = %d, want 2", code)
	}
}
