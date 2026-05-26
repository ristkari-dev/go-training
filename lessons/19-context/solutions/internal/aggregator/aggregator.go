// Package aggregator is the lesson 19 reference implementation.
package aggregator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ristkari-dev/go-training/lessons/19-context/solutions/internal/logparse"
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

// Walk is the channel-based aggregator with ctx propagation.
func Walk(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
	// Canonical upfront ctx check. Without this, a pre-cancelled ctx
	// races with the workers' first results in the reduce loop's
	// select — both cases are ready, runtime picks randomly. The
	// upfront check guarantees we never spawn workers for an already-
	// cancelled call.
	if err := ctx.Err(); err != nil {
		return WalkResult{}, err
	}

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
		case <-ctx.Done():
			return result, ctx.Err()
		}
	}

	return result, nil
}

// WalkLocked is the mutex-based aggregator with ctx propagation.
func WalkLocked(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
	// Symmetry with Walk: canonical upfront ctx check. Even though
	// WalkLocked's structure means each worker would self-skip the
	// merge on cancel and wg.Wait + ctx.Err priority would handle it,
	// the upfront check saves spawning N goroutines for an already-
	// cancelled call.
	if err := ctx.Err(); err != nil {
		return WalkResult{}, err
	}

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
			case <-ctx.Done():
				// Drop this worker's result; ctx cancellation will be
				// reported by the main goroutine after wg.Wait.
				return
			}
		}(path)
	}

	wg.Wait()

	// Check ctx FIRST — if cancelled, that's the dominant signal.
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if perr := firstErr.Load(); perr != nil {
		return result, *perr
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
