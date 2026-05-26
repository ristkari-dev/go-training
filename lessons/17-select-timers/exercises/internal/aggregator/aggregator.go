// Package aggregator walks a directory of log files in parallel and
// returns the combined level counts.
//
// CHANGED from lesson 16:
//   - Walk's signature is now Walk(dir string, timeout time.Duration)
//     (WalkResult, error). Returns a struct instead of just a map so
//     callers can also see which files timed out.
//   - Per-file timeout: each file's result-collection step is wrapped
//     in a select { case res := <-ch: ...; case <-time.After(timeout):
//     skip }. timeout=0 means "no timeout."
//   - Timeouts are NOT errors. They populate WalkResult.TimedOut and
//     execution continues. Real file errors (open, parse) still abort.
//
// UNCHANGED from lesson 16:
//   - One goroutine per file via processFile.
//   - Buffered channel sized len(files) so workers never block.
//   - processFile sends exactly one fileResult per file (success or error).
//   - Real-error semantics: first non-timeout error aborts the reduce
//     loop and returns wrapped.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ristkari-dev/go-training/lessons/17-select-timers/exercises/internal/logparse"
)

// fileResult is one file's parsed level counts plus any error.
// Workers send a fileResult on the results channel for each file.
type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// WalkResult is what Walk returns. Counts is the aggregated level
// counts (was the L16 return value). TimedOut lists files that
// exceeded the per-file timeout — they contribute nothing to Counts.
type WalkResult struct {
	Counts   map[string]int
	TimedOut []string
}

// Walk lists files in dir, spawns one goroutine per file, and returns
// the combined level counts plus any files that exceeded the timeout.
//
// timeout=0 means "no timeout" (the select's time.After branch becomes
// a nil channel, which never fires). Pass time.Duration values like
// 5*time.Second otherwise.
//
// Real file errors (open failure, parse error from logparse) still
// abort with a wrapped error mentioning the offending file. Timeouts
// don't error — they populate TimedOut.
//
// Hint:
//
//  1. entries, err := os.ReadDir(dir)
//
//  2. if err != nil → return WalkResult{}, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
//
//  3. Build files []string from entries (skip subdirs).
//
//  4. resultsCh := make(chan fileResult, len(files))
//
//  5. for each path: go processFile(path, resultsCh)
//
//  6. Track which paths have been "claimed" so timeouts know which
//     file timed out. Easiest: ASSIGN each select iteration to a file
//     in launch order — but that's not how concurrency works.
//     Better: use a map[string]bool of paths-still-pending; each
//     received fileResult removes r.path from the map; a timeout
//     removes ONE remaining path arbitrarily.
//
//     Even simpler: for each file, do an inner select that waits
//     for THAT file's specific result OR the timeout. This means
//     one timer per file — exactly the "per-file timeout" semantics.
//     But we don't know which fileResult will arrive next on the
//     shared channel; results come in completion order.
//
//     SIMPLEST WORKING APPROACH: keep a set of pending paths. The
//     reduce loop's select has two cases: (a) read a fileResult,
//     remove r.path from pending; (b) timeout, pick any remaining
//     pending path and add it to TimedOut. Iterate until pending
//     is empty.
//
// Sketch:
//
//	pending := map[string]struct{}{}
//	for _, p := range files { pending[p] = struct{}{} }
//	for len(pending) > 0 {
//	    var timeoutCh <-chan time.Time
//	    if timeout > 0 { timeoutCh = time.After(timeout) }
//	    select {
//	    case r := <-resultsCh:
//	        delete(pending, r.path)
//	        if r.err != nil → return result, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
//	        for level, count := range r.counts { result.Counts[level] += count }
//	    case <-timeoutCh:
//	        // Pick one remaining pending path arbitrarily, mark timed out.
//	        for p := range pending {
//	            result.TimedOut = append(result.TimedOut, p)
//	            delete(pending, p)
//	            break  // just one per timeout
//	        }
//	    }
//	}
//	return result, nil
//
// Note: the per-file timeout fires once per pending file. If timeout=5s
// and 3 files are slow, the reduce loop spends up to 5s per slow file
// (worst case 15s total). This matches the design: each file gets its
// own time budget.
func Walk(dir string, timeout time.Duration) (WalkResult, error) {
	_ = os.ReadDir
	_ = filepath.Join
	_ = logparse.Parse
	_ = fmt.Errorf
	_ = time.After
	panic("TODO: read dir, spawn workers, reduce with select + per-file timeout")
}

// processFile (unchanged from L16) opens the file, parses with
// logparse.Parse, sends exactly one fileResult on resultsCh.
func processFile(path string, resultsCh chan<- fileResult) {
	f, err := os.Open(path)
	if err != nil {
		resultsCh <- fileResult{path: path, err: err}
		return
	}
	defer f.Close()

	entries, err := logparse.Parse(f)
	if err != nil {
		resultsCh <- fileResult{path: path, err: err}
		return
	}

	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Level]++
	}
	resultsCh <- fileResult{path: path, counts: counts}
}
