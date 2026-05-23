// Package aggregator is the lesson 17 reference implementation.
// Walk now takes a per-file timeout and returns a WalkResult struct.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ristkari-dev/go-training/lessons/17-select-timers/solutions/internal/logparse"
)

type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// WalkResult is what Walk returns. Counts is the aggregated level
// counts; TimedOut lists files that exceeded the per-file timeout.
type WalkResult struct {
	Counts   map[string]int
	TimedOut []string
}

// Walk lists files in dir and processes each in a goroutine. Returns
// the aggregated counts plus any files that timed out (timeout=0
// disables the timeout).
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
				// Early return on real error. The buffered resultsCh (sized
				// len(files)) absorbs any in-flight workers' sends, so no
				// goroutine leaks even though we stop reading.
				return result, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
			}
			for level, count := range r.counts {
				result.Counts[level] += count
			}
		case <-timeoutCh:
			// Pick one remaining pending path arbitrarily.
			for p := range pending {
				result.TimedOut = append(result.TimedOut, p)
				delete(pending, p)
				break
			}
		}
	}

	return result, nil
}

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
