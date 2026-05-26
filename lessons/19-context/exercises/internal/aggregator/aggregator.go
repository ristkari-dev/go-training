// Package aggregator walks a directory of log files in parallel.
//
// CHANGED in lesson 19: both Walk and WalkLocked take ctx context.Context
// as the new FIRST parameter. Workers check ctx.Done() between operations.
// On ctx cancellation: returns (partial WalkResult, ctx.Err()).
//
// Per-file timeout (L17) still applies independently — ctx cancellation
// stops EVERYTHING; per-file timeout skips one slow file.
//
// processFile/processFileForLocked unchanged from L18 (they don't know
// about ctx — the cancellation happens in the reduce/coordination layer).
package aggregator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ristkari-dev/go-training/lessons/19-context/exercises/internal/logparse"
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

// Walk is the channel-based aggregator. ctx cancellation returns
// (partial, ctx.Err()).
//
// Hint:
//  1. Same setup as L18 (read dir, list files, spawn one goroutine per file).
//  2. The reduce loop's select gains a third case: ctx.Done().
//  3. On case <-ctx.Done(): return result, ctx.Err()
//
// The pending-map iteration on timeout stays the same as L17.
func Walk(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
	_ = ctx.Done
	_ = os.ReadDir
	_ = filepath.Join
	_ = logparse.Parse
	_ = fmt.Errorf
	_ = time.After
	panic("TODO: same as L18 Walk but select also checks ctx.Done(); on cancel return result + ctx.Err()")
}

// WalkLocked is the mutex-based aggregator. Same ctx propagation.
//
// Hint:
//  1. Same setup as L18 (read dir, sync.Mutex, sync.WaitGroup, atomic.Pointer[error]).
//  2. Each outer goroutine's select gains case <-ctx.Done(): return.
//     (When ctx cancels, outer worker exits without doing the merge.)
//  3. After wg.Wait(), check ctx.Err() FIRST before firstErr:
//     if err := ctx.Err(); err != nil { return result, err }
//     if perr := firstErr.Load(); perr != nil { return result, *perr }
//     return result, nil
func WalkLocked(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
	_ = ctx.Done
	_ = sync.Mutex{}
	_ = sync.WaitGroup{}
	_ = atomic.Pointer[error]{}
	panic("TODO: same as L18 WalkLocked but select also checks ctx.Done(); after Wait check ctx.Err() before firstErr")
}

// processFile (channel-based, unchanged from L18) — opens file, parses,
// sends fileResult.
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

// processFileForLocked (unchanged from L18) — returns fileResult directly.
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
