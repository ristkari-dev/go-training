package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestImporterGoldenPath is a SKELETON. Build the binary into a temp
// dir, exec it with -file=<tmp.json> and CSV piped via stdin, then
// verify the resulting JSON file exists and contains the expected
// entries.
//
// NOTE: This test is skipped unless `go build` is available — students
// running the test in a stripped-down env may not have a Go toolchain.
// The skip happens automatically if exec.LookPath("go") fails.
func TestImporterGoldenPath(t *testing.T) {
	// TODO:
	//   if _, err := exec.LookPath("go"); err != nil { t.Skip("no go toolchain") }
	//   build the binary with `go build -o <tmp>/expenses-import ./...`
	//   exec it: cmd := exec.Command(bin, "-file="+jsonPath); cmd.Stdin = bytes.NewReader([]byte("..."));
	//   verify the json file content matches expected
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}
