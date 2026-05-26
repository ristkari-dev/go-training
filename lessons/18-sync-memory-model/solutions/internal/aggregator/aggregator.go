// Package aggregator is the lesson 18 reference implementation.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/solutions/internal/logparse"
)

type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

type WalkResult struct {
	Counts   map[string]int
	TimedOut []string
}

// Walk is L17's channel-based implementation, verbatim.
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

// WalkLocked is the mutex-based variant. Same input/output as Walk;
// different internals. Coordinates via sync.Mutex on the shared map
// + sync.WaitGroup for worker completion.
//
// First real error is reported via firstErr (sync/atomic). Once set,
// the reduce loop returns it after wg.Wait() completes. We don't
// return immediately on error like Walk does — we let in-flight
// workers complete and only THEN return. This is a small behavior
// difference vs Walk; documented in the test suite.
func WalkLocked(dir string, timeout time.Duration) (WalkResult, error) {
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

	result := WalkResult{Counts: map[string]int{}}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// firstErr holds the first real (non-timeout) error from any
	// worker. Using a pointer + atomic so we don't need the mutex for
	// this single field; demonstrates sync/atomic alongside the
	// mutex-protected map.
	var firstErr atomic.Pointer[error]

	for _, path := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			doneCh := make(chan fileResult, 1)
			go func() {
				doneCh <- processFileForLocked(path)
			}()

			var timeoutCh <-chan time.Time
			if timeout > 0 {
				timeoutCh = time.After(timeout)
			}

			select {
			case r := <-doneCh:
				if r.err != nil {
					// First-error-wins: only the first worker to set
					// firstErr actually sets it.
					perr := r.err
					wrapped := fmt.Errorf("aggregator: %s: %w", r.path, perr)
					firstErr.CompareAndSwap(nil, &wrapped)
					return
				}
				mu.Lock()
				for level, count := range r.counts {
					result.Counts[level] += count
				}
				mu.Unlock()
			case <-timeoutCh:
				mu.Lock()
				result.TimedOut = append(result.TimedOut, path)
				mu.Unlock()
			}
		}(path)
	}

	wg.Wait()

	// After Wait, no other goroutines touch result or firstErr — safe
	// to read without the mutex (happens-before via WaitGroup.Wait).
	if perr := firstErr.Load(); perr != nil {
		return result, *perr
	}
	return result, nil
}

// processFile (channel-based) — unchanged from L17.
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

// processFileForLocked is the WalkLocked-style worker. Same body as
// processFile but returns the fileResult directly (sent via a per-
// worker channel by the caller).
func processFileForLocked(path string) fileResult {
	f, err := os.Open(path)
	if err != nil {
		return fileResult{path: path, err: err}
	}
	defer f.Close()

	entries, err := logparse.Parse(f)
	if err != nil {
		return fileResult{path: path, err: err}
	}

	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Level]++
	}
	return fileResult{path: path, counts: counts}
}
