// Package aggregator is the lesson 22 exercises tree.
//
// Walk + WalkLocked + their helpers carry forward verbatim from L17/L18/L19
// (students aren't reimplementing the channel- and mutex-based variants
// in L20). The L20 task is to implement WalkPool — the worker-pool
// variant — using internal/errgroupx.
package aggregator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/exercises/internal/errgroupx"
	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/exercises/internal/logparse"
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

// errFileTimedOut is the package-private sentinel used by WalkPool's
// workers to signal "this file timed out" through resultsCh. The
// reducer translates this into a TimedOut entry. Other (real) errors
// don't pass through resultsCh — they trigger g.Go's error return,
// which cancels the group.
var errFileTimedOut = errors.New("aggregator: file timed out")

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
			// A file might arrive AFTER the timeout case already marked
			// it as timed out — skip the merge so we never count a file
			// in BOTH Counts and TimedOut. (Same race fix as L17/L18.)
			if _, stillPending := pending[r.path]; !stillPending {
				continue
			}
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

// WalkPool is the worker-pool variant. Spawns exactly `workers`
// goroutines (instead of one per file). Each worker reads paths from
// a shared channel; results go to a shared resultsCh; main reduces.
// Uses errgroupx for first-error-wins + ctx cancellation.
//
// workers=0 means runtime.NumCPU() as a sensible default for
// CPU-bound work (which log parsing largely is).
//
// Same WalkResult / same ctx-cancellation semantics as Walk +
// WalkLocked. Same per-file timeout (each worker selects on
// time.After per file).
//
// Hint:
//  1. Upfront ctx check (same as L19).
//  2. workers <= 0 → workers = runtime.NumCPU()
//  3. List files (same as L19).
//  4. pathsCh := make(chan string, len(files)); fill with paths; close.
//  5. resultsCh := make(chan fileResult, len(files))
//  6. g, ctx := errgroupx.WithContext(ctx)
//  7. for i := 0; i < workers; i++ {
//     g.Go(func() error {
//     for path := range pathsCh {
//     // per-file timeout via inner goroutine + select
//     // on success → resultsCh <- r
//     // on timeout → resultsCh <- fileResult{path, err: errFileTimedOut}
//     // on real error → return wrapped error (cancels group)
//     // on ctx.Done() → return ctx.Err()
//     }
//     return nil
//     })
//     }
//  8. go func() { g.Wait(); close(resultsCh) }()
//  9. Reduce: range resultsCh; on errFileTimedOut → TimedOut; else merge counts.
//  10. return result, g.Wait()
func WalkPool(ctx context.Context, dir string, timeout time.Duration, workers int) (WalkResult, error) {
	_ = errgroupx.WithContext
	_ = runtime.NumCPU
	_ = errFileTimedOut
	panic("TODO: bounded worker pool; pathsCh + N workers + errgroupx; same semantics as Walk/WalkLocked")
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

// processFileForPool — same body as processFile/processFileForLocked,
// returns fileResult directly. Used by WalkPool's workers.
func processFileForPool(path string) fileResult {
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
