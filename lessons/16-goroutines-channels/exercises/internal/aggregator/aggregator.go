// Package aggregator walks a directory of log files in parallel and
// returns the combined level counts.
//
// This is the running example for Phase 3. Lesson 16 builds the basics:
// one goroutine per file, results collected via a channel, single main
// reducer. Future lessons add timeouts (L17), shared-state coordination
// (L18), context cancellation (L19), and a bounded worker pool (L20).
//
// The aggregator pattern — multiple producers, single consumer — is
// the most basic concurrency idiom in Go and the foundation for
// everything that follows.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/exercises/internal/logparse"
)

// fileResult holds one file's parsed level counts plus any error
// encountered while reading or parsing.
//
// Workers send a fileResult on the results channel for each file they
// process. The main goroutine collects N results and reduces.
type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// Walk lists files in dir, spawns one goroutine per file, and returns
// the combined level counts from logparse.Parse across all files.
//
// On any file error (open failure, parse error), Walk returns the
// partial counts collected so far plus a wrapped error mentioning the
// offending file. Same shape as L13's csvimport.Parse.
//
// Hint:
//  1. entries, err := os.ReadDir(dir)
//  2. if err != nil → return nil, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
//  3. files := []string{} ; for each entry, skip dirs, append filepath.Join(dir, name) to files
//  4. resultsCh := make(chan fileResult, len(files))  // buffered so workers don't block
//  5. for each file path: go processFile(path, resultsCh)
//  6. combined := map[string]int{}
//  7. for i := 0; i < len(files); i++ {
//     r := <-resultsCh
//     if r.err != nil → return combined, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
//     for level, count := range r.counts { combined[level] += count }
//     }
//  8. return combined, nil
//
// processFile is a helper: open the file, call logparse.Parse, send
// either {path, counts, nil} or {path, nil, err} on resultsCh.
func Walk(dir string) (map[string]int, error) {
	_ = os.ReadDir
	_ = filepath.Join
	_ = logparse.Parse
	_ = fmt.Errorf
	panic("TODO: read dir, spawn one goroutine per file, collect results, reduce")
}
