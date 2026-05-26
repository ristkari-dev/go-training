# Plan V — Lesson 19 (context) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 19 of Phase 3 — generalize L17's close-of-done-channel cancellation into the standard library `context.Context` interface. Students learn the `Context` type (4 methods), derivation via `WithCancel`/`WithTimeout`/`WithDeadline`, the `Done()`+`Err()` check pattern, and the first-parameter propagation idiom. The aggregator's `Walk` and `WalkLocked` BOTH gain `ctx context.Context` as the new first parameter; workers check `ctx.Done()` between operations. The CLI's `main` uses `signal.NotifyContext(ctx, os.Interrupt)` so Ctrl-C cancels the walk gracefully. Sets up L20 (worker pool + errgroup).

**Architecture:** Same per-lesson pattern as Plans D-U. Six tasks. Skeleton tests in warmup + main (Phase 3 invariant). Heavy-explanatory slide deck with **four** concept blocks. Fourth lesson under the Phase 3 design.

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `context`, `encoding/json`, `errors`, `fmt`, `io`, `os`, `os/exec`, `os/signal`, `path/filepath`, `regexp`, `sort`, `strconv`, `strings`, `sync`, `sync/atomic`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan V: lesson 19 complete; `make test` AND `make test-race` green; `cmd/aggregator` Ctrl-C handler works; race-clean.

### Design decisions (1 user-approved + 8 plan-recommended)

**User-approved via brainstorming:**

1. **Four slide concepts.** (1) `context.Context` interface + first-parameter propagation; (2) Derivation (`WithCancel`/`WithTimeout`/`WithDeadline` + `defer cancel()`); (3) `Done()` + `Err()` check pattern; (4) Common mistakes + `context.Value` bundled. Matches L16/L17 rhythm.

**Plan-recommended:**

2. **`ctx` is the NEW first parameter for both Walk and WalkLocked.** Signatures change from L18:
   - L18: `Walk(dir, timeout)` / `WalkLocked(dir, timeout)`
   - L19: `Walk(ctx, dir, timeout)` / `WalkLocked(ctx, dir, timeout)`
   
   Each lesson evolves the aggregator; this is the canonical "context is always first" idiom.

3. **On cancellation, Walk returns `(partialResult, ctx.Err())`.** `ctx.Err()` is either `context.Canceled` or `context.DeadlineExceeded`. Callers use `errors.Is(err, context.Canceled)` to distinguish. No new fields on WalkResult.

4. **Per-file timeout (L17) and ctx cancellation (L19) are orthogonal.** Both coexist:
   - Per-file timeout fires → file added to `TimedOut`, continue
   - ctx cancelled → return immediately with partial + `ctx.Err()`
   
   Walk's reduce loop's select now has THREE cases: result, time.After, ctx.Done.

5. **Warmup `sleepctx.SleepWithCtx` uses `select { case <-time.After(d): return nil; case <-ctx.Done(): return ctx.Err() }`.** The textbook pattern in three lines. Returns `nil` on clean completion, `ctx.Err()` on cancellation.

6. **CLI `main()` uses `signal.NotifyContext(context.Background(), os.Interrupt)`.** Returns a ctx that cancels on SIGINT. `defer stop()` releases the signal handler. `run(ctx, args, stdout, stderr) error` is the testable entry point.

7. **CLI's cancellation handling**: on `errors.Is(err, context.Canceled)`, print partial counts to stdout AND "cancelled by user" to stderr, then exit 0 (user-requested cancellation isn't a failure). On other errors, exit 1 per L18 behavior.

8. **WalkLocked under ctx**: outer goroutine's select gains `ctx.Done()` case. On ctx cancellation, outer goroutine exits without doing the merge. The inner goroutine (file processor) doesn't know about ctx — it just sends on the doneCh; if nobody reads, the buffered-size-1 channel absorbs it cleanly. After `wg.Wait()`, main checks `ctx.Err()` first; returns it if non-nil; else returns the firstErr (atomic.Pointer pattern from L18).

9. **No new build-index slug fix needed.** Current slug `context` matches directory `19-context`. (Pleasant surprise — most prior Phase 3 lessons needed a fix.)

---

## Plans F-U lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages.
3. Common-mistake content in README (4 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `make test-race` continues to feature in daily-habits README.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files).
8. gofmt 1.19+ normalises godoc list indentation — `gofmt -w` if triggered.
9. Empirical timeout values: 1µs for unit-level aggregator tests; 1ns for subprocess-level CLI tests. (Verified through L17 + L18 — `1ms` is too long; file work completes before timer fires.)

---

## File Structure

After Plan V (~20 files, same as L17/L18):

```
lessons/19-context/
├── README.md                                       (Task 5)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 4)
├── exercises/
│   ├── warmup/sleepctx/
│   │   ├── sleepctx.go                             (Task 2)
│   │   └── sleepctx_test.go                        (Task 2 — SKELETON)
│   ├── cmd/aggregator/
│   │   ├── main.go                                 (Task 3 — NotifyContext + ctx propagation)
│   │   └── main_test.go                            (Task 3 — SKELETON; integration + cancel test)
│   └── internal/
│       ├── logparse/                               (Task 3 — verbatim from L18)
│       │   ├── logparse.go
│       │   └── logparse_test.go
│       └── aggregator/
│           ├── aggregator.go                       (Task 3 — Walk + WalkLocked both get ctx)
│           └── aggregator_test.go                  (Task 3 — SKELETON; +cancel tests)
└── solutions/  (mirrored)
```

---

## Conventions

- **Branch:** `feature/plan-v-lesson-19-context`
- **Commit messages:** Conventional Commits
- **Carry-forward sources:**
  - `lessons/18-sync-memory-model/solutions/internal/logparse/{logparse.go, logparse_test.go}` (verbatim)
  - `lessons/18-sync-memory-model/solutions/internal/aggregator/aggregator.go` (Walk + WalkLocked both get ctx prepended; processFile/processFileForLocked unchanged)
  - `lessons/18-sync-memory-model/solutions/cmd/aggregator/main.go` (gains signal.NotifyContext + ctx propagation + cancel handling)

---

## Task 1: Scaffold + restructure

Same dance as Plans J-U. No slug fix needed (current `context` slug matches).

- [ ] **Step 1:** `make new-lesson NAME=19-context`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/19-context/exercises/warmup.go
rm lessons/19-context/exercises/warmup_test.go
rm lessons/19-context/exercises/main.go
rm lessons/19-context/exercises/main_test.go
rm lessons/19-context/solutions/warmup.go
rm lessons/19-context/solutions/warmup_test.go
rm lessons/19-context/solutions/main.go
rm lessons/19-context/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree + slides build.

```bash
find lessons/19-context -type f | sort
make slides-build
grep -q "19-context" dist/index.html && echo "✓ in index"
ls dist/lessons/19-context/ && echo "✓ dir created"
rm -rf dist
```

- [ ] **Step 4:** Commit:

```bash
git add lessons/19-context/
git commit -m "feat(lessons): scaffold lesson 19-context with empty subpackage layout"
```

---

## Task 2: Author the warm-up — `sleepctx` subpackage

`SleepWithCtx(ctx, d) error` — sleep d unless ctx is cancelled first. Textbook three-line select pattern.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/19-context/{exercises,solutions}/warmup/sleepctx`

- [ ] **Step 2:** Create `lessons/19-context/exercises/warmup/sleepctx/sleepctx.go`:

```go
// Package sleepctx is the lesson 19 warm-up: cancellable sleep via
// select + ctx.Done.
//
// SleepWithCtx sleeps for d, returning nil on clean completion or
// ctx.Err() if ctx is cancelled first. The textbook pattern is three
// lines of select:
//
//	select {
//	case <-time.After(d):  return nil
//	case <-ctx.Done():     return ctx.Err()
//	}
//
// This composes the L17 "select with timeout" pattern with the L19
// "ctx as a cancellation signal" idiom. Any function that does
// blocking work should accept ctx for the same reason — to be
// cancellable.
package sleepctx

import (
	"context"
	"time"
)

// SleepWithCtx blocks for d. Returns nil on clean completion, or
// ctx.Err() if ctx is cancelled first.
//
// ctx.Err() returns one of:
//   - context.Canceled — caller invoked cancel()
//   - context.DeadlineExceeded — ctx had a deadline that passed
//   - nil — only possible if ctx wasn't done yet (won't happen here
//     because we only consult Err inside the Done case)
//
// Callers distinguish via errors.Is(err, context.Canceled) etc.
//
// Hint:
//   select {
//   case <-time.After(d):
//       return nil
//   case <-ctx.Done():
//       return ctx.Err()
//   }
func SleepWithCtx(ctx context.Context, d time.Duration) error {
	_ = time.After
	panic("TODO: select with case <-time.After(d) and case <-ctx.Done()")
}
```

- [ ] **Step 3:** Create `lessons/19-context/exercises/warmup/sleepctx/sleepctx_test.go` (SKELETON):

```go
package sleepctx

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestSleepCompletes is a SKELETON. Use a ctx that won't be cancelled
// (context.Background()) and a short duration; assert nil error.
func TestSleepCompletes(t *testing.T) {
	// TODO:
	//   err := SleepWithCtx(context.Background(), 10*time.Millisecond)
	//   if err != nil → t.Errorf("expected nil, got %v", err)
	_ = context.Background
	_ = time.Millisecond
}

// TestSleepCancelled is a SKELETON. Use a ctx that's cancelled before
// SleepWithCtx is called; assert returns context.Canceled via
// errors.Is.
func TestSleepCancelled(t *testing.T) {
	// TODO:
	//   ctx, cancel := context.WithCancel(context.Background())
	//   cancel()  // cancel BEFORE calling SleepWithCtx
	//   err := SleepWithCtx(ctx, time.Hour)  // would sleep forever without cancel
	//   if !errors.Is(err, context.Canceled) → t.Errorf("expected context.Canceled, got %v", err)
	_ = errors.Is
}

// TestSleepDeadline is a SKELETON. Use context.WithTimeout(parent, 10ms)
// and SleepWithCtx for 1*time.Hour; assert returns
// context.DeadlineExceeded.
func TestSleepDeadline(t *testing.T) {
	// TODO:
	//   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	//   defer cancel()
	//   err := SleepWithCtx(ctx, time.Hour)
	//   if !errors.Is(err, context.DeadlineExceeded) → t.Errorf("expected DeadlineExceeded, got %v", err)
}
```

- [ ] **Step 4:** Create `lessons/19-context/solutions/warmup/sleepctx/sleepctx.go`:

```go
// Package sleepctx is the lesson 19 warm-up reference implementation.
package sleepctx

import (
	"context"
	"time"
)

// SleepWithCtx blocks for d or until ctx is cancelled.
func SleepWithCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

- [ ] **Step 5:** Create `lessons/19-context/solutions/warmup/sleepctx/sleepctx_test.go`:

```go
package sleepctx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSleepCompletes(t *testing.T) {
	err := SleepWithCtx(context.Background(), 10*time.Millisecond)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestSleepCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel BEFORE the sleep starts — Done is already closed
	err := SleepWithCtx(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSleepCancelledMidway(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := SleepWithCtx(ctx, time.Hour) // would block forever; cancel rescues
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSleepDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := SleepWithCtx(ctx, time.Hour)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
```

> Note on TestSleepCompletes's 10ms duration: short enough for tests to run quickly, long enough to avoid scheduler flake. Same rationale as L17.

> Note on TestSleepCancelledMidway: spawns a cancelling goroutine after 10ms. Tests the realistic case where cancellation arrives mid-sleep. The other two cases (pre-cancelled, deadline) cover the corner cases.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/19-context/
go test ./lessons/19-context/exercises/warmup/sleepctx/... -v 2>&1 | tail -10
go test ./lessons/19-context/solutions/warmup/sleepctx/... -v 2>&1 | tail -20
go test -race ./lessons/19-context/solutions/warmup/sleepctx/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/19-context/exercises/warmup/ lessons/19-context/solutions/warmup/
git commit -m "feat(lesson-19): warmup — sleepctx (SleepWithCtx via select + ctx.Done)"
```

Expected: gofmt empty; exercises pass vacuously; solutions 4 tests PASS; -race clean.

---

## Task 3: Main — logparse carry-forward + aggregator with ctx + cmd with SIGINT handler

Big task. Both Walk and WalkLocked gain ctx as first parameter; CLI's main uses signal.NotifyContext.

**Files (12 total — 6 per side):**

### Step 1: Create directories

```bash
mkdir -p lessons/19-context/{exercises,solutions}/internal/logparse
mkdir -p lessons/19-context/{exercises,solutions}/internal/aggregator
mkdir -p lessons/19-context/{exercises,solutions}/cmd/aggregator
```

### Step 2-3: `internal/logparse/` (verbatim from L18)

- [ ] **Step 2:** Forward-port `lessons/18-sync-memory-model/solutions/internal/logparse/logparse.go` to both `exercises/internal/logparse/` and `solutions/internal/logparse/`. Identical content.

- [ ] **Step 3:** Same for `logparse_test.go`.

### Step 4-7: `internal/aggregator/` (Walk + WalkLocked both gain ctx)

- [ ] **Step 4:** Create `lessons/19-context/exercises/internal/aggregator/aggregator.go`:

```go
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
//   1. Same setup as L18 (read dir, list files, spawn one goroutine per file).
//   2. The reduce loop's select gains a third case: ctx.Done().
//   3. On case <-ctx.Done(): return result, ctx.Err()
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
//   1. Same setup as L18 (read dir, sync.Mutex, sync.WaitGroup, atomic.Pointer[error]).
//   2. Each outer goroutine's select gains case <-ctx.Done(): return.
//      (When ctx cancels, outer worker exits without doing the merge.)
//   3. After wg.Wait(), check ctx.Err() FIRST before firstErr:
//        if err := ctx.Err(); err != nil { return result, err }
//        if perr := firstErr.Load(); perr != nil { return result, *perr }
//        return result, nil
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
```

- [ ] **Step 5:** Create `lessons/19-context/exercises/internal/aggregator/aggregator_test.go` (SKELETON):

```go
package aggregator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// walkFunc — common shape of Walk and WalkLocked. NOTE: ctx is now first.
type walkFunc func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error)

// runWalkSuite runs the same sub-tests against both Walk and WalkLocked.
// Skeleton: cases TODO-stubbed.
func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write fixture; fn(context.Background(), dir, 0); assert Counts.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO
		_ = strings.Contains
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		// TODO
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		// TODO
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		// TODO. 1µs per L17/L18 empirical rationale.
		_ = time.Microsecond
	})

	t.Run("ctx-cancelled-before-walk", func(t *testing.T) {
		// TODO:
		//   ctx, cancel := context.WithCancel(context.Background())
		//   cancel()
		//   _, err := fn(ctx, dir, 0)
		//   if !errors.Is(err, context.Canceled) → t.Errorf("...")
		_ = errors.Is
		_ = context.WithCancel
	})

	t.Run("ctx-deadline-exceeded", func(t *testing.T) {
		// TODO:
		//   ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		//   defer cancel()
		//   _, err := fn(ctx, dir, 0)
		//   if !errors.Is(err, context.DeadlineExceeded) → t.Errorf("...")
	})

	_ = fn
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
```

- [ ] **Step 6:** Create `lessons/19-context/solutions/internal/aggregator/aggregator.go`:

```go
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
```

> Note on ctx priority in WalkLocked: after wg.Wait, we check `ctx.Err()` FIRST and return it if non-nil. This means "user cancelled" trumps "some file failed" in error reporting. Rationale: if the user cancelled, telling them about a file failure is less useful than confirming the cancellation. The partial result is still returned.

> Note on processFile NOT taking ctx: the file work is fast (file I/O completes in ~100µs); adding ctx.Done() checks inside processFile would be over-engineering. The cancellation happens at the coordination level (reduce loop / outer worker), where it has meaningful impact.

- [ ] **Step 7:** Create `lessons/19-context/solutions/internal/aggregator/aggregator_test.go`:

```go
package aggregator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

type walkFunc func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error)

func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log",
			"2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")

		result, err := fn(context.Background(), dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1, "WARN": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("multiple-files", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:32:00 WARN c\n")
		writeFile(t, dir, "c.log", "2026-05-21T14:33:00 ERROR d\n2026-05-21T14:34:00 ERROR e\n2026-05-21T14:35:00 ERROR f\n")

		result, err := fn(context.Background(), dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		result, err := fn(context.Background(), dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts, got %v", result.Counts)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")

		_, err := fn(context.Background(), dir, 0)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := fn(context.Background(), "/nonexistent/path/that/does/not/exist", 0)
		if err == nil {
			t.Fatal("expected error from nonexistent dir, got nil")
		}
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n")
		if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "subdir"), "nested.log", "2026-05-21T14:30:00 ERROR ignored\n")

		result, err := fn(context.Background(), dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("subdirs should be skipped: Counts = %v, want %v", result.Counts, want)
		}
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")

		// 1µs per L17/L18 empirical rationale.
		result, err := fn(context.Background(), dir, 1*time.Microsecond)
		if err != nil {
			t.Fatalf("timeouts should not error: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts when all files time out, got %v", result.Counts)
		}
		if len(result.TimedOut) != 2 {
			t.Errorf("expected 2 timed-out files, got %d: %v", len(result.TimedOut), result.TimedOut)
		}
	})

	t.Run("ctx-cancelled-before-walk", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel BEFORE Walk

		_, err := fn(ctx, dir, 0)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("ctx-deadline-exceeded", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")

		// Tight deadline — workers can't finish before it expires.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		defer cancel()

		_, err := fn(ctx, dir, 0)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context.DeadlineExceeded, got %v", err)
		}
	})
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
```

### Step 8-11: `cmd/aggregator/` (signal.NotifyContext + ctx propagation)

- [ ] **Step 8:** Create `lessons/19-context/exercises/cmd/aggregator/main.go`:

```go
// Package main is the lesson 19 aggregator CLI.
//
// CHANGED from L18:
//   - main() wraps the call in signal.NotifyContext(ctx, os.Interrupt)
//     so Ctrl-C cancels the walk gracefully.
//   - run() takes ctx as its new first parameter.
//   - On context.Canceled: prints partial counts to stdout + "cancelled
//     by user" to stderr, exits 0 (user-requested cancellation isn't
//     a failure).
//
// Usage:
//
//	aggregator -dir=<path> [-mode=channel|mutex] [-timeout=<dur>]
//	Ctrl-C cancels mid-walk gracefully.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/19-context/exercises/internal/aggregator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		// User cancellation isn't an error from the user's perspective —
		// print to stderr but exit 0 with whatever partial output we have.
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "cancelled by user")
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point.
//
// Hint:
//   1. Parse -dir, -timeout, -mode (same as L18).
//   2. Pick walkFn: aggregator.Walk or aggregator.WalkLocked.
//   3. result, err := walkFn(ctx, dir, timeout)
//   4. PRINT result.TimedOut to stderr, sorted Counts to stdout — even
//      if err is context.Canceled, because partial output is valuable.
//   5. Return err to main so it can decide the exit code.
func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = aggregator.WalkLocked
	_ = sort.Strings
	_ = time.ParseDuration
	_ = errors.Is
	panic("TODO: parse flags; call walkFn(ctx, ...); print partial result; return err")
}
```

- [ ] **Step 9:** Create `lessons/19-context/exercises/cmd/aggregator/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: same pattern as L18.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorCancelled is a SKELETON. Builds the binary, runs it
// against a fixture; sends SIGINT while running; asserts exit 0 +
// "cancelled by user" in stderr.
func TestAggregatorCancelled(t *testing.T) {
	// TODO: build binary; exec it; cmd.Process.Signal(os.Interrupt);
	// assert cmd.Wait error is nil (exit 0); stderr contains "cancelled"
}
```

- [ ] **Step 10:** Create `lessons/19-context/solutions/cmd/aggregator/main.go`:

```go
// Package main is the lesson 19 aggregator CLI reference implementation.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/19-context/solutions/internal/aggregator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "cancelled by user")
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type walkFunc func(ctx context.Context, dir string, timeout time.Duration) (aggregator.WalkResult, error)

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	dir, timeout, mode, err := parseFlags(args)
	if err != nil {
		return err
	}

	var fn walkFunc
	switch mode {
	case "channel":
		fn = aggregator.Walk
	case "mutex":
		fn = aggregator.WalkLocked
	default:
		return fmt.Errorf("invalid -mode %q (want channel|mutex)", mode)
	}

	result, walkErr := fn(ctx, dir, timeout)

	// Print partial output even on error/cancellation — caller may want
	// what we DID collect.
	for _, path := range result.TimedOut {
		fmt.Fprintf(stderr, "timeout: %s\n", path)
	}

	keys := make([]string, 0, len(result.Counts))
	for k := range result.Counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(stdout, "%s: %d\n", k, result.Counts[k])
	}

	return walkErr
}

func parseFlags(args []string) (string, time.Duration, string, error) {
	var (
		dir     string
		timeout time.Duration
		mode    = "channel"
	)
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-dir="):
			dir = a[len("-dir="):]
		case strings.HasPrefix(a, "-timeout="):
			t, err := time.ParseDuration(a[len("-timeout="):])
			if err != nil {
				return "", 0, "", fmt.Errorf("invalid -timeout: %w", err)
			}
			timeout = t
		case strings.HasPrefix(a, "-mode="):
			mode = a[len("-mode="):]
		default:
			return "", 0, "", fmt.Errorf("unknown flag %q", a)
		}
	}
	if dir == "" {
		return "", 0, "", fmt.Errorf("-dir=<path> is required")
	}
	return dir, timeout, mode, nil
}
```

- [ ] **Step 11:** Create `lessons/19-context/solutions/cmd/aggregator/main_test.go`:

```go
package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not available")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "aggregator")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	return bin
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAggregatorGoldenPath(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log",
		"2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n2026-05-21T14:32:00 WARN c\n")
	writeFile(t, logDir, "b.log",
		"2026-05-21T14:33:00 ERROR d\n")

	cmd := exec.Command(bin, "-dir="+logDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	want := "ERROR: 1\nINFO: 2\nWARN: 1\n"
	if stdout.String() != want {
		t.Errorf("stdout mismatch:\ngot:\n%s\nwant:\n%s", stdout.String(), want)
	}
}

func TestAggregatorMalformed(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
	writeFile(t, logDir, "bad.log", "garbage line\n")

	cmd := exec.Command(bin, "-dir="+logDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
	if !strings.Contains(stderr.String(), "bad.log") {
		t.Errorf("stderr should mention bad.log, got %q", stderr.String())
	}
}

func TestAggregatorMissingFlag(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing -dir, got nil")
	}
	if !strings.Contains(stderr.String(), "-dir") {
		t.Errorf("stderr should mention -dir, got %q", stderr.String())
	}
}

func TestAggregatorTimeout(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:31:00 WARN b\n")

	cmd := exec.Command(bin, "-dir="+logDir, "-timeout=1ns")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	if stdout.String() != "" {
		t.Errorf("expected empty stdout when everything times out, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "timeout:") {
		t.Errorf("expected 'timeout:' in stderr, got %q", stderr.String())
	}
}

// TestAggregatorCancelled starts the binary and signals SIGINT after a
// brief delay. The binary should: receive the signal, ctx cancels,
// Walk returns context.Canceled, main prints "cancelled by user" to
// stderr and exits 0.
//
// To make the test deterministic, we use a directory with many files
// (so the binary doesn't complete before SIGINT arrives).
func TestAggregatorCancelled(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	// Many files so the walk takes long enough for SIGINT to arrive.
	for i := 0; i < 100; i++ {
		writeFile(t, logDir, fmt.Sprintf("f%d.log", i), "2026-05-21T14:30:00 INFO ok\n")
	}

	cmd := exec.Command(bin, "-dir="+logDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Give the binary a moment to start, then signal.
	time.Sleep(5 * time.Millisecond)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("signal: %v", err)
	}

	err := cmd.Wait()
	// On cancellation, exit code is 0 (we treat cancel as success).
	// However: the binary may have FINISHED before SIGINT arrived. In
	// that case, no "cancelled" message; just exit 0 with full output.
	// Tests should tolerate either outcome.
	if err != nil {
		t.Fatalf("wait: %v (stderr: %s)", err, stderr.String())
	}
	// stderr MAY contain "cancelled by user" (if SIGINT won the race)
	// OR be empty (if walk finished first). Either is correct behavior.
	// Just verify no crashy stderr.
	if strings.Contains(stderr.String(), "error:") {
		t.Errorf("unexpected error in stderr: %q", stderr.String())
	}
}
```

Add the missing `fmt` import at top of test file (use it in the fmt.Sprintf for fixture names). Actually let me re-check — yes the import would need to be added.

Update the import block in step 11's main_test.go to include `"fmt"`:

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)
```

### Step 12: Verify and commit

- [ ] **Step 12:** Verify and commit Task 3:

```bash
gofmt -l lessons/19-context/
go test ./lessons/19-context/exercises/... 2>&1 | tail -10
go test ./lessons/19-context/solutions/... -v 2>&1 | tail -30
go test -race ./lessons/19-context/solutions/... 2>&1 | tail -10
make test
make test-race
golangci-lint run ./...
go vet ./...

# Smoke
TMP=$(mktemp -d /tmp/lesson19-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO ok" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow" > "$TMP/logs/b.log"
# Normal
go run ./lessons/19-context/solutions/cmd/aggregator -dir="$TMP/logs"
# Mutex mode
go run ./lessons/19-context/solutions/cmd/aggregator -dir="$TMP/logs" -mode=mutex
rm -rf "$TMP"

git add lessons/19-context/exercises/internal/ lessons/19-context/solutions/internal/ lessons/19-context/exercises/cmd/ lessons/19-context/solutions/cmd/
git commit -m "feat(lesson-19): main — aggregator Walk + WalkLocked gain ctx; cmd handles SIGINT"
```

Expected: gofmt empty; exercises pass vacuously; solutions all PASS (9 sub-tests via runWalkSuite × 2 = 18, plus cmd integration); -race clean; make test-race green; smoke shows both modes working.

---

## Task 4: Slide deck — 4 concepts

Heavy-explanatory pattern. ~30 slides.

**File:** `lessons/19-context/slides/slides.md`

Concept order matches design:

1. **`context.Context` interface + first-parameter propagation**
2. **Derivation (WithCancel/WithTimeout/WithDeadline + defer cancel())**
3. **Done() + Err() check pattern**
4. **Common mistakes + context.Value (bundled)**

- [ ] **Step 1-7:** Author the deck.
- [ ] **Step 8:** Verify and commit:

```bash
make slides-build
grep -q "19-context" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/19-context/slides/
git commit -m "feat(lesson-19): slides — context (4 concepts)"
```

---

## Task 5: README

Mirrors slide concepts. Features make test-race in daily habits.

**File:** `lessons/19-context/README.md`

- [ ] **Step 1:** Author the README (~250 lines):
  - "What you'll learn" — 4 bullets.
  - "What's different from L18" — both Walk variants get ctx; CLI handles SIGINT.
  - "The package layout" — annotated tree.
  - Per-concept sections (4).
  - "Exercise: warm-up — sleepctx".
  - "Exercise: main — aggregator with ctx + SIGINT handler".
  - "Daily habits" — gofmt + vet + test + test-race.
  - "How to run" — including SIGINT smoke (Ctrl-C during walk).
  - "Going further" — Read (Go blog "Go Concurrency Patterns: Context"); Try (errgroup preview — pass ctx through and aggregate first-error-wins; L20 will formalise).

- [ ] **Step 2:** Commit:

```bash
git add lessons/19-context/README.md
git commit -m "docs(lesson-19): README — context self-study"
```

---

## Task 6: End-to-end verification + final review + PR

Standard sweep + final review subagent + push branch + create PR.

```bash
make test
make test-race
go test -v ./lessons/19-context/solutions/...
go test -race ./lessons/19-context/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/19-context/
make slides-build && grep -q "19-context" dist/index.html && rm -rf dist
```

Dispatch feature-dev:code-reviewer over `git diff main..HEAD`. Apply any fixes. Then `gh pr create`.

---

## Critical file paths

To be created:

- `lessons/19-context/` (directory)
- `lessons/19-context/README.md`
- `lessons/19-context/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/19-context/exercises/warmup/sleepctx/{sleepctx.go, sleepctx_test.go}`
- `lessons/19-context/exercises/cmd/aggregator/{main.go, main_test.go}`
- `lessons/19-context/exercises/internal/logparse/{logparse.go, logparse_test.go}`
- `lessons/19-context/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}`
- `lessons/19-context/solutions/...` (mirrored)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md` (Phase 3 design)
- `lessons/18-sync-memory-model/solutions/...` (carry-forward source for logparse + Walk + WalkLocked + cmd/aggregator)
- `tools/build-index/main.go` (no changes — L19 slug `context` matches directory)
- `Makefile` (no changes — `test-race` target added in L18)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
