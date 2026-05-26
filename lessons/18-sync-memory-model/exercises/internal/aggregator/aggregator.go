// Package aggregator walks a directory of log files in parallel.
//
// NEW in lesson 18: WalkLocked alongside Walk.
//   - Walk: channel-based coordination (verbatim from L17).
//   - WalkLocked: mutex + WaitGroup. Shared map guarded by sync.Mutex.
//
// Both have identical input/output. The lesson's centerpiece is
// comparing the two styles side-by-side: "share by communicating"
// (Walk) vs "share by locking" (WalkLocked). Same correctness; the
// race detector catches bugs in EITHER style.
//
// CLI -mode=channel|mutex picks at runtime.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/exercises/internal/logparse"
)

// fileResult is one file's parsed level counts plus any error.
// Used by Walk (channel-based).
type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// WalkResult is what Walk and WalkLocked both return. Identical shape.
type WalkResult struct {
	Counts   map[string]int
	TimedOut []string
}

// Walk is L17's channel-based implementation, verbatim. Spawns one
// goroutine per file; collects results via buffered channel; reduce
// loop uses select for per-file timeout.
func Walk(dir string, timeout time.Duration) (WalkResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return WalkResult{}, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
	}

	files := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}

	resultsCh := make(chan fileResult, len(files))

	for _, path := range files {
		go processFile(path, resultsCh)
	}

	result := WalkResult{Counts: map[string]int{}}

	pending := map[string]struct{}{}
	for _, p := range files {
		pending[p] = struct{}{}
	}

	for len(pending) > 0 {
		var timeoutCh <-chan time.Time
		if timeout > 0 {
			timeoutCh = time.After(timeout)
		}
		select {
		case r := <-resultsCh:
			delete(pending, r.path)
			if r.err != nil {
				// Early return on real error. The buffered resultsCh
				// (sized len(files)) absorbs in-flight workers' sends,
				// so no goroutine leaks.
				return result, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
			}
			for level, count := range r.counts {
				result.Counts[level] += count
			}
		case <-timeoutCh:
			for p := range pending {
				result.TimedOut = append(result.TimedOut, p)
				delete(pending, p)
				break
			}
		}
	}

	return result, nil
}

// WalkLocked is the mutex-based variant. Same input/output as Walk.
//
// Coordination:
//   - sharedCounts map[string]int + sharedTimedOut []string, both guarded by sync.Mutex.
//   - sync.WaitGroup tracks workers.
//   - Per-file timeout: each worker has its own doneCh chan struct{};
//     the reduce-style loop selects on the done channel vs time.After.
//
// Pedagogically demonstrates "share by locking" alongside Walk's
// "share by communicating." Same race-detector cleanliness when done
// right; the race detector catches bugs in EITHER style.
//
// Hint:
//  1. List files (same as Walk).
//  2. result := WalkResult{Counts: map[string]int{}}
//  3. var mu sync.Mutex
//  4. var wg sync.WaitGroup
//  5. For each file: wg.Add(1); spawn goroutine that:
//     - opens file, parses with logparse.Parse
//     - on success: mu.Lock(); merge counts into result.Counts; mu.Unlock()
//     - sends a done signal on a per-file channel
//  6. Main loop: for each file, select on doneCh vs time.After(timeout).
//     On done: continue to next file. On timeout: mu.Lock(); append
//     to result.TimedOut; mu.Unlock().
//  7. wg.Wait() at the end ensures all workers finish before we
//     return (no goroutine leaks).
//
// Note on the "errors abort" semantics: WalkLocked maintains parity
// with Walk — first real error aborts and returns with wrapped error.
// Implementation needs an error channel + select arm for that.
func WalkLocked(dir string, timeout time.Duration) (WalkResult, error) {
	_ = sync.Mutex{}
	_ = sync.WaitGroup{}
	panic("TODO: see hint in doc comment; coordinate via Mutex + WaitGroup; per-file timeout via doneCh + time.After")
}

// processFile (unchanged from L17) opens the file, parses with
// logparse.Parse, sends exactly one fileResult on resultsCh.
//
// Used only by Walk (channel-based). WalkLocked has its own per-
// worker logic inline since the coordination shape is different.
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
