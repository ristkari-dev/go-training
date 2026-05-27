# Plan W — Lesson 20 (Concurrency patterns) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 20 of Phase 3 — the "named patterns" lesson. Worker pool, fan-out, fan-in, pipeline, and the canonical errgroup idiom (a lesson-local `internal/errgroupx/` package that mirrors `golang.org/x/sync/errgroup`'s API). The aggregator gains a third implementation `WalkPool(ctx, dir, timeout, workers)` alongside L17's `Walk` and L18's `WalkLocked`. CLI gets `-mode=channel|mutex|pool` + `-workers=N`. The warmup is the generic `FanIn[T any](chans ...<-chan T) <-chan T` (uses L12 generics).

**Architecture:** Same per-lesson pattern as Plans D-V. Seven tasks (one extra for the errgroupx subpackage). Skeleton tests in warmup + errgroupx + main + cmd. Heavy-explanatory slide deck with **five** concept blocks (worker pool, fan-out, fan-in, pipeline, errgroupx).

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `context`, `encoding/json`, `errors`, `flag`, `fmt`, `io`, `os`, `os/exec`, `os/signal`, `path/filepath`, `regexp`, `runtime`, `sort`, `strconv`, `strings`, `sync`, `sync/atomic`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan W: lesson 20 is complete; `make test` + `make test-race` green; all three Walk variants work via `cmd/aggregator -mode=...`.

### Design decisions (2 user-approved + 8 plan-recommended)

**User-approved via brainstorming:**

1. **Three Walk variants coexist** — Walk (channels), WalkLocked (mutex), WalkPool (worker pool). CLI picks via `-mode=channel|mutex|pool` + `-workers=N` (only meaningful for pool). Tests parameterize over all three via `runWalkSuite`.

2. **Five slide concepts** — worker pool, fan-out, fan-in, pipeline, errgroupx. Each named pattern gets its own slot.

**Plan-recommended:**

3. **`errgroupx` mirrors `golang.org/x/sync/errgroup`'s API exactly.** Same method names (`WithContext`, `Go`, `Wait`), same semantics (first-error-wins, ctx cancelled on first error or on Wait). When students later use the real errgroup, the code is identical. Skip advanced methods like `SetLimit` (Go 1.20+); keep the package minimal.

4. **`WalkPool` uses `errgroupx` internally.** Demonstrates the canonical "manage N workers" pattern in the running example. Each worker is a goroutine spawned via `g.Go(...)`. First file error cancels the rest via `errgroupx`'s ctx cancellation.

5. **`WalkPool` worker count: `workers=0` means default.** Default is `runtime.NumCPU()` (sensible for CPU-bound work like log parsing). CLI's `-workers=0` (or missing flag) triggers the default. Documented in slides + README.

6. **Per-file timeout still works in WalkPool.** Each worker, when consuming a path from `pathsCh`, runs the same `select { case <-resultCh: ...; case <-time.After(timeout): ... }` pattern as L17 Walk's reduce loop, just per-worker not per-file-iteration. Per-file behavior preserved.

7. **`processFile` and `processFileForLocked` carry forward unchanged from L19.** WalkPool uses a third helper `processFileForPool` that returns `fileResult` directly (same as L19's processFileForLocked) — workers in errgroupx call it.

8. **Tests parameterize WalkPool via an adapter** that closes over a fixed `workers=4`:
   ```go
   walkPoolAdapter := func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
       return WalkPool(ctx, dir, timeout, 4)
   }
   runWalkSuite(t, walkPoolAdapter)
   ```
   This way `runWalkSuite` (with the existing `walkFunc` type) tests all three Walk variants identically. Plus 1 dedicated `TestWalkPoolWorkers` that verifies the `workers` parameter is respected (e.g., calling with workers=1 still works correctly, just serially).

9. **CLI `-workers=N` ignored (with stderr warning) unless `-mode=pool`.** Validates that users notice the flag is mode-specific. Empty string / `0` / missing → default.

10. **Slug fix needed**: current `tools/build-index/main.go:60` has `Slug: "patterns"`; lesson directory is `20-concurrency-patterns`. Fix in Task 1.

---

## Plans F-V lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `make test-race` daily-habits feature continues from L18.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files).
8. gofmt 1.19+ godoc list normalization tolerance.
9. **CI flake lessons from PRs #34/#35/#36**: aggregator timeout test asserts invariant (`files accounted for = file count`), not "exactly N timed out." Walk variants must skip merging files already claimed by timeout (`if _, stillPending := pending[r.path]; !stillPending { continue }`). CLI integration test counts stdout newlines + stderr "timeout:" prefixes for invariant check.

---

## File Structure

After Plan W (~22 files; slightly bigger than L19 due to errgroupx subpackage):

```
lessons/20-concurrency-patterns/
├── README.md                                       (Task 6)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 5)
├── exercises/
│   ├── warmup/fanin/
│   │   ├── fanin.go                                (Task 2)
│   │   └── fanin_test.go                           (Task 2 — SKELETON)
│   ├── cmd/aggregator/
│   │   ├── main.go                                 (Task 4)
│   │   └── main_test.go                            (Task 4 — SKELETON)
│   └── internal/
│       ├── logparse/                               (Task 4 — verbatim from L19)
│       ├── errgroupx/
│       │   ├── errgroupx.go                        (Task 3)
│       │   └── errgroupx_test.go                   (Task 3 — SKELETON)
│       └── aggregator/
│           ├── aggregator.go                       (Task 4 — adds WalkPool)
│           └── aggregator_test.go                  (Task 4 — SKELETON; 3-way parameterized)
└── solutions/  (mirrored)
```

---

## Conventions

- **Branch:** `feature/plan-w-lesson-20-concurrency-patterns`
- **Commit messages:** Conventional Commits
- **Carry-forward sources:**
  - `lessons/19-context/solutions/internal/logparse/{logparse.go, logparse_test.go}` (verbatim)
  - `lessons/19-context/solutions/internal/aggregator/aggregator.go` (carries forward Walk + WalkLocked + processFile + processFileForLocked; WalkPool is NEW)
  - `lessons/19-context/solutions/cmd/aggregator/main.go` (gains -mode=pool + -workers)

---

## Task 1: Scaffold + restructure + slug fix

- [ ] **Step 1:** `make new-lesson NAME=20-concurrency-patterns`
- [ ] **Step 2:** Delete 8 flat scaffolder files (same as Plans J-V).
- [ ] **Step 3:** Verify 4-file scaffolded tree.
- [ ] **Step 4:** Fix L20 slug in `tools/build-index/main.go`:

Find: `{Number: "20", Slug: "patterns", Title: "Concurrency patterns", Blurb: "worker pool · errgroup", Phase: 3},`

Change to: `{Number: "20", Slug: "concurrency-patterns", Title: "Concurrency patterns", Blurb: "worker pool · errgroup", Phase: 3},`

Verify slides build + dist/lessons/20-concurrency-patterns/ created.

- [ ] **Step 5:** Commit:

```bash
git add lessons/20-concurrency-patterns/ tools/build-index/main.go
git commit -m "feat(lessons): scaffold lesson 20-concurrency-patterns with empty subpackage layout

Also fix build-index master list: L20 slug was 'patterns' (legacy
placeholder); update to 'concurrency-patterns' to match the actual
lesson directory."
```

---

## Task 2: Author the warm-up — `fanin` subpackage

`FanIn[T any](chans ...<-chan T) <-chan T` — multiplex N input channels into a single output. Uses sync.WaitGroup to know when all inputs close. A closer goroutine closes the output once all inputs are done.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/20-concurrency-patterns/{exercises,solutions}/warmup/fanin`

- [ ] **Step 2:** Create `lessons/20-concurrency-patterns/exercises/warmup/fanin/fanin.go`:

```go
// Package fanin is the lesson 20 warm-up: multiplex N input channels
// into a single output channel.
//
// FanIn is the canonical "many producers, one consumer" multiplexer.
// Each input gets a forwarding goroutine that reads its channel and
// sends to the output. A closer goroutine watches a WaitGroup and
// closes the output once all forwarders are done.
//
// Generic via [T any] (lesson 12) so the same function works for any
// channel element type.
package fanin

import "sync"

// FanIn returns a new channel that receives every value sent on any
// of chans. The output is closed when ALL inputs are closed.
//
// Examples:
//
//	a := make(chan int); b := make(chan int)
//	out := FanIn(a, b)
//	// values sent on a or b appear on out
//	// close(a) AND close(b) → out closes too
//
// Hint:
//   1. out := make(chan T)
//   2. var wg sync.WaitGroup
//   3. for each input channel: wg.Add(1); spawn goroutine that:
//        for v := range input { out <- v }
//        wg.Done()
//   4. go func() { wg.Wait(); close(out) }()
//   5. return out
//
// The closer goroutine pattern (step 4) is the key insight: we can't
// close inline because we don't know when all forwarders finish. The
// goroutine waits for them all then closes once.
func FanIn[T any](chans ...<-chan T) <-chan T {
	_ = sync.WaitGroup{}
	panic("TODO: forward goroutines + closer goroutine; see hint")
}
```

- [ ] **Step 3:** Create `lessons/20-concurrency-patterns/exercises/warmup/fanin/fanin_test.go` (SKELETON):

```go
package fanin

import (
	"sort"
	"testing"
)

// TestFanInTwo is a SKELETON. Two inputs; send N values on each;
// assert all 2N values appear on the output (in any order).
func TestFanInTwo(t *testing.T) {
	// TODO:
	//   a := make(chan int, 3); b := make(chan int, 3)
	//   a <- 1; a <- 2; a <- 3; close(a)
	//   b <- 4; b <- 5; b <- 6; close(b)
	//   out := FanIn[int](a, b)
	//   var got []int
	//   for v := range out { got = append(got, v) }
	//   sort.Ints(got)
	//   want := []int{1, 2, 3, 4, 5, 6}
	//   compare
	_ = sort.Ints
}

// TestFanInEmpty is a SKELETON. Zero inputs; out should close immediately.
func TestFanInEmpty(t *testing.T) {
	// TODO:
	//   out := FanIn[int]()
	//   for range out { t.Errorf("expected no values") }
}

// TestFanInClosesWithAllInputs is a SKELETON. After closing all inputs,
// out should close. Use a timeout via time.After to detect hangs.
func TestFanInClosesWithAllInputs(t *testing.T) {
	// TODO:
	//   a := make(chan int); b := make(chan int)
	//   out := FanIn[int](a, b)
	//   close(a); close(b)
	//   select {
	//   case _, ok := <-out:
	//       if ok { t.Errorf("expected closed channel") }
	//   case <-time.After(time.Second):
	//       t.Fatal("output didn't close")
	//   }
}
```

- [ ] **Step 4:** Create `lessons/20-concurrency-patterns/solutions/warmup/fanin/fanin.go`:

```go
// Package fanin is the lesson 20 warm-up reference implementation.
package fanin

import "sync"

// FanIn multiplexes N input channels into one. Output closes when
// all inputs close.
func FanIn[T any](chans ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, ch := range chans {
		wg.Add(1)
		go func(input <-chan T) {
			defer wg.Done()
			for v := range input {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
```

- [ ] **Step 5:** Create `lessons/20-concurrency-patterns/solutions/warmup/fanin/fanin_test.go`:

```go
package fanin

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestFanInTwo(t *testing.T) {
	a := make(chan int, 3)
	b := make(chan int, 3)
	a <- 1
	a <- 2
	a <- 3
	close(a)
	b <- 4
	b <- 5
	b <- 6
	close(b)

	out := FanIn[int](a, b)
	var got []int
	for v := range out {
		got = append(got, v)
	}
	sort.Ints(got)

	want := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanIn = %v, want %v", got, want)
	}
}

func TestFanInThree(t *testing.T) {
	a := make(chan string, 1)
	b := make(chan string, 1)
	c := make(chan string, 1)
	a <- "alpha"
	b <- "beta"
	c <- "gamma"
	close(a)
	close(b)
	close(c)

	out := FanIn[string](a, b, c)
	var got []string
	for v := range out {
		got = append(got, v)
	}
	sort.Strings(got)

	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanIn = %v, want %v", got, want)
	}
}

func TestFanInEmpty(t *testing.T) {
	out := FanIn[int]()
	for range out {
		t.Errorf("expected no values from empty FanIn")
	}
}

func TestFanInClosesWithAllInputs(t *testing.T) {
	a := make(chan int)
	b := make(chan int)
	out := FanIn[int](a, b)

	close(a)
	close(b)

	select {
	case _, ok := <-out:
		if ok {
			t.Errorf("expected closed channel after both inputs closed")
		}
	case <-time.After(time.Second):
		t.Fatal("output didn't close within 1s — closer goroutine didn't fire")
	}
}
```

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/20-concurrency-patterns/
go test ./lessons/20-concurrency-patterns/exercises/warmup/fanin/... -v 2>&1 | tail -10
go test ./lessons/20-concurrency-patterns/solutions/warmup/fanin/... -v 2>&1 | tail -20
go test -race ./lessons/20-concurrency-patterns/solutions/warmup/fanin/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/20-concurrency-patterns/exercises/warmup/ lessons/20-concurrency-patterns/solutions/warmup/
git commit -m "feat(lesson-20): warmup — fanin (FanIn[T any] via forwarders + closer)"
```

Expected: gofmt empty; exercises pass vacuously; solutions 4 tests PASS; -race clean.

---

## Task 3: Author the `internal/errgroupx/` package

Lesson-local errgroup. Mirrors `golang.org/x/sync/errgroup`'s API: `WithContext`, `Go`, `Wait`. ~50 lines.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/20-concurrency-patterns/{exercises,solutions}/internal/errgroupx`

- [ ] **Step 2:** Create `lessons/20-concurrency-patterns/exercises/internal/errgroupx/errgroupx.go`:

```go
// Package errgroupx is the lesson 20 "manage N workers" library —
// a lesson-local reimplementation of golang.org/x/sync/errgroup.
//
// The pattern: spawn N goroutines that may return errors. When the
// first one fails, cancel the rest via a shared context, then return
// that first error. If they all succeed, return nil.
//
// API mirrors x/sync/errgroup exactly:
//
//	g, ctx := errgroupx.WithContext(parent)
//	for _, item := range items {
//	    item := item
//	    g.Go(func() error {
//	        return process(ctx, item)
//	    })
//	}
//	if err := g.Wait(); err != nil { return err }
//
// In production code, use the real golang.org/x/sync/errgroup. This
// package teaches the pattern from the inside.
package errgroupx

import (
	"context"
	"sync"
)

// Group is a collection of goroutines working on subtasks that are
// part of the same overall task.
//
// A zero Group is valid, but a Group created by WithContext also
// holds an associated context that is cancelled on the first error
// from any subtask or on Wait returning.
type Group struct {
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// WithContext returns a new Group and a derived Context. The Context
// is cancelled when the first non-nil error is returned by any Go-
// spawned function, or when Wait returns, whichever occurs first.
//
// Hint:
//   1. ctx, cancel := context.WithCancelCause(parent)  // Go 1.20+
//   2. return &Group{cancel: cancel}, ctx
//
// WithCancelCause is like WithCancel but lets you record WHY the
// context was cancelled (useful for debugging; matches real errgroup).
func WithContext(ctx context.Context) (*Group, context.Context) {
	_ = context.WithCancelCause
	panic("TODO: WithCancelCause + return Group with cancel set")
}

// Go calls f in a new goroutine. The first call to return a non-nil
// error cancels the group's context (if any).
//
// Hint:
//   1. g.wg.Add(1)
//   2. go func() {
//        defer g.wg.Done()
//        if err := f(); err != nil {
//            g.errOnce.Do(func() {
//                g.err = err
//                if g.cancel != nil {
//                    g.cancel(err)
//                }
//            })
//        }
//      }()
//
// errOnce ensures we only record the FIRST error. Subsequent failures
// are silently dropped (the first one represents the cause).
func (g *Group) Go(f func() error) {
	_ = sync.Once{}
	panic("TODO: wg.Add + spawn goroutine + errOnce.Do capture")
}

// Wait blocks until all functions launched by Go have returned, then
// returns the first non-nil error (if any) from them.
//
// Hint:
//   1. g.wg.Wait()
//   2. if g.cancel != nil { g.cancel(g.err) }
//   3. return g.err
//
// The cancel call ensures the ctx is cancelled even if no error
// occurred — releases timers and any goroutines watching ctx.Done.
func (g *Group) Wait() error {
	panic("TODO: wg.Wait + cancel + return g.err")
}
```

- [ ] **Step 3:** Create `lessons/20-concurrency-patterns/exercises/internal/errgroupx/errgroupx_test.go` (SKELETON):

```go
package errgroupx

import (
	"context"
	"errors"
	"testing"
)

// TestSuccess is a SKELETON. Spawn three goroutines that all return
// nil; assert Wait returns nil.
func TestSuccess(t *testing.T) {
	// TODO:
	//   g, _ := WithContext(context.Background())
	//   for i := 0; i < 3; i++ { g.Go(func() error { return nil }) }
	//   if err := g.Wait(); err != nil → t.Errorf("...")
	_ = context.Background
}

// TestFirstErrorWins is a SKELETON. Spawn three goroutines; one returns
// a sentinel error; the others succeed. Assert Wait returns that error.
func TestFirstErrorWins(t *testing.T) {
	// TODO:
	//   g, _ := WithContext(context.Background())
	//   sentinel := errors.New("first fail")
	//   g.Go(func() error { return nil })
	//   g.Go(func() error { return sentinel })
	//   g.Go(func() error { return nil })
	//   if err := g.Wait(); !errors.Is(err, sentinel) → t.Errorf("...")
	_ = errors.Is
}

// TestCtxCancelledOnError is a SKELETON. Verify that when one goroutine
// returns an error, the group's ctx is cancelled (other goroutines can
// see ctx.Done() fire).
func TestCtxCancelledOnError(t *testing.T) {
	// TODO:
	//   g, ctx := WithContext(context.Background())
	//   sentinel := errors.New("fail")
	//   g.Go(func() error { return sentinel })
	//   <-ctx.Done()  // should fire quickly
	//   if !errors.Is(ctx.Err(), context.Canceled) → t.Errorf("...")
}
```

- [ ] **Step 4:** Create `lessons/20-concurrency-patterns/solutions/internal/errgroupx/errgroupx.go`:

```go
// Package errgroupx is the lesson 20 reference implementation —
// a tiny errgroup mirroring golang.org/x/sync/errgroup.
package errgroupx

import (
	"context"
	"sync"
)

// Group manages a collection of goroutines working on subtasks.
type Group struct {
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// WithContext returns a Group and a derived Context. The Context is
// cancelled on the first error or when Wait returns.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go spawns f as a goroutine. First error cancels the group's ctx.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
	}()
}

// Wait blocks until all Go-spawned functions return, then returns the
// first non-nil error (if any).
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}
```

- [ ] **Step 5:** Create `lessons/20-concurrency-patterns/solutions/internal/errgroupx/errgroupx_test.go`:

```go
package errgroupx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSuccess(t *testing.T) {
	g, _ := WithContext(context.Background())
	for range 3 {
		g.Go(func() error { return nil })
	}
	if err := g.Wait(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestFirstErrorWins(t *testing.T) {
	g, _ := WithContext(context.Background())
	sentinel := errors.New("first fail")

	g.Go(func() error { return nil })
	g.Go(func() error { return sentinel })
	g.Go(func() error { return nil })

	if err := g.Wait(); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCtxCancelledOnError(t *testing.T) {
	g, ctx := WithContext(context.Background())
	sentinel := errors.New("fail")

	g.Go(func() error { return sentinel })

	select {
	case <-ctx.Done():
		// Expected — first error cancels ctx.
	case <-time.After(time.Second):
		t.Fatal("ctx never cancelled after first error")
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", ctx.Err())
	}

	_ = g.Wait()
}

func TestCtxCancelledOnWait(t *testing.T) {
	// Even with no errors, ctx should be cancelled when Wait returns
	// (releases ctx-watching goroutines that aren't part of the group).
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return nil })
	_ = g.Wait()

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("expected ctx cancelled after Wait, got %v", ctx.Err())
	}
}
```

> Note on `for range 3`: Go 1.22+ "range over integer" syntax. Cleaner than `for i := 0; i < 3; i++` when you don't need the index. The package is Go 1.23.

> Note on `_ = g.Wait()` in TestCtxCancelledOnError: we already verified ctx cancellation; we call Wait afterwards to drain the goroutine but don't care about its error.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/20-concurrency-patterns/
go test ./lessons/20-concurrency-patterns/exercises/internal/errgroupx/... -v 2>&1 | tail -10
go test ./lessons/20-concurrency-patterns/solutions/internal/errgroupx/... -v 2>&1 | tail -20
go test -race ./lessons/20-concurrency-patterns/solutions/internal/errgroupx/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/20-concurrency-patterns/exercises/internal/errgroupx/ lessons/20-concurrency-patterns/solutions/internal/errgroupx/
git commit -m "feat(lesson-20): internal/errgroupx — lesson-local errgroup (WithContext + Go + Wait)"
```

Expected: gofmt empty; exercises pass vacuously; solutions 4 tests PASS; -race clean.

---

## Task 4: Main — logparse carry-forward + aggregator gains WalkPool + cmd `-mode=pool` + `-workers`

Largest task. Aggregator's third Walk variant; CLI's third mode.

**Files (10 total):**

### Step 1: Create directories

```bash
mkdir -p lessons/20-concurrency-patterns/{exercises,solutions}/internal/logparse
mkdir -p lessons/20-concurrency-patterns/{exercises,solutions}/internal/aggregator
mkdir -p lessons/20-concurrency-patterns/{exercises,solutions}/cmd/aggregator
```

### Step 2-3: `internal/logparse/` (verbatim from L19)

- [ ] **Step 2:** Forward-port `lessons/19-context/solutions/internal/logparse/logparse.go` to both exercises and solutions trees.

- [ ] **Step 3:** Same for `logparse_test.go`.

### Step 4-7: `internal/aggregator/` (carries Walk+WalkLocked + adds WalkPool)

- [ ] **Step 4:** Create `lessons/20-concurrency-patterns/exercises/internal/aggregator/aggregator.go`. Carry Walk + WalkLocked + processFile + processFileForLocked verbatim from L19 (with import-path rewrites). Add the NEW WalkPool stub:

```go
// (carry-forward header + Walk + WalkLocked + processFile + processFileForLocked from L19)

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
//   1. Upfront ctx check (same as L19).
//   2. workers <= 0 → workers = runtime.NumCPU()
//   3. List files (same as L19).
//   4. pathsCh := make(chan string, len(files)); fill with paths; close.
//   5. resultsCh := make(chan fileResult, len(files))
//   6. g, ctx := errgroupx.WithContext(ctx)
//   7. for i := 0; i < workers; i++ {
//        g.Go(func() error {
//            for path := range pathsCh {
//                r := workerProcess(ctx, path, timeout)
//                if r.err != nil && !isTimeout(r.err) {
//                    return r.err  // first real error cancels group
//                }
//                resultsCh <- r
//                if ctx.Err() != nil { return ctx.Err() }
//            }
//            return nil
//        })
//      }
//   8. go func() { g.Wait(); close(resultsCh) }()
//   9. result := WalkResult{Counts: map[string]int{}}
//      for r := range resultsCh {
//          if r.timedOut {
//              result.TimedOut = append(result.TimedOut, r.path)
//              continue
//          }
//          for level, count := range r.counts {
//              result.Counts[level] += count
//          }
//      }
//  10. return result, g.Wait()  // pick up any error; ctx.Err() included
func WalkPool(ctx context.Context, dir string, timeout time.Duration, workers int) (WalkResult, error) {
	_ = errgroupx.WithContext
	_ = runtime.NumCPU
	panic("TODO: bounded worker pool; pathsCh + N workers + errgroupx; same semantics as Walk/WalkLocked")
}
```


> The hint above gestures at the worker-pool design. The full implementation needs a way for workers to signal "this file timed out" through `resultsCh` so the reducer puts it in TimedOut. Approach: use a **package-private sentinel** `errFileTimedOut` carried via `fileResult.err`. Real errors don't pass through resultsCh — they propagate up through `g.Go`'s error return (which cancels the group). The reducer only sees success-with-counts or timeout-sentinel.

- [ ] **Step 5:** Create `lessons/20-concurrency-patterns/exercises/internal/aggregator/aggregator_test.go` (SKELETON):

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
	// Same 9 sub-tests as L19 (single-file, multiple-files, empty-dir,
	// malformed-line-returns-error, nonexistent-dir-returns-error,
	// subdirs-skipped, all-files-timed-out, ctx-cancelled-before-walk,
	// ctx-deadline-exceeded). Bodies TODO-stubbed.
	_ = writeFile
	_ = reflect.DeepEqual
	_ = strings.Contains
	_ = time.Microsecond
	_ = errors.Is
	_ = context.WithCancel
	_ = fn
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}

func TestWalkPool(t *testing.T) {
	walkPoolAdapter := func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
		return WalkPool(ctx, dir, timeout, 4)
	}
	runWalkSuite(t, walkPoolAdapter)
}

// TestWalkPoolWorkers verifies the `workers` parameter is respected.
// SKELETON.
func TestWalkPoolWorkers(t *testing.T) {
	// TODO: workers=1 → serial; workers=8 → parallel; both produce
	// same WalkResult. Mainly verifies the parameter doesn't crash.
}
```

- [ ] **Step 6:** Create `lessons/20-concurrency-patterns/solutions/internal/aggregator/aggregator.go`. Carry forward Walk + WalkLocked + processFile + processFileForLocked from L19 (path-rewritten); add WalkPool:

```go
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

	"github.com/ristkari-dev/go-training/lessons/20-concurrency-patterns/solutions/internal/errgroupx"
	"github.com/ristkari-dev/go-training/lessons/20-concurrency-patterns/solutions/internal/logparse"
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

// (Walk and WalkLocked and processFile and processFileForLocked are
// carried forward from L19 verbatim — see the L19 plan for source.
// Import paths rewritten to 20-concurrency-patterns/solutions/...)

// WalkPool walks dir using a BOUNDED worker pool of size `workers`.
// Spawns exactly `workers` goroutines (instead of one per file).
// Workers consume paths from a shared channel and send results
// through a results channel; main reduces.
//
// Uses errgroupx for first-error-wins + ctx cancellation. workers=0
// means runtime.NumCPU() (sensible for CPU-bound log parsing).
//
// Same WalkResult / same ctx-cancellation semantics as Walk/WalkLocked.
// Per-file timeout via per-worker select.
func WalkPool(ctx context.Context, dir string, timeout time.Duration, workers int) (WalkResult, error) {
	if err := ctx.Err(); err != nil {
		return WalkResult{}, err
	}

	if workers <= 0 {
		workers = runtime.NumCPU()
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

	// pathsCh: producer-closed channel of file paths workers consume.
	pathsCh := make(chan string, len(files))
	for _, p := range files {
		pathsCh <- p
	}
	close(pathsCh)

	// resultsCh: workers send fileResult here; reducer collects.
	resultsCh := make(chan fileResult, len(files))

	g, ctx := errgroupx.WithContext(ctx)

	for range workers {
		g.Go(func() error {
			for path := range pathsCh {
				// Per-file timeout via inner goroutine + select.
				innerCh := make(chan fileResult, 1)
				go func() {
					innerCh <- processFileForPool(path)
				}()

				var timeoutCh <-chan time.Time
				if timeout > 0 {
					timeoutCh = time.After(timeout)
				}

				select {
				case r := <-innerCh:
					if r.err != nil {
						return fmt.Errorf("aggregator: %s: %w", r.path, r.err)
					}
					resultsCh <- r
				case <-timeoutCh:
					resultsCh <- fileResult{path: path, err: errFileTimedOut}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	// Closer goroutine: wait for workers, then close resultsCh so the
	// reducer's range exits.
	go func() {
		_ = g.Wait()
		close(resultsCh)
	}()

	result := WalkResult{Counts: map[string]int{}}
	for r := range resultsCh {
		if errors.Is(r.err, errFileTimedOut) {
			result.TimedOut = append(result.TimedOut, r.path)
			continue
		}
		for level, count := range r.counts {
			result.Counts[level] += count
		}
	}

	// Pick up the first error (or ctx.Err()) from the group.
	if err := g.Wait(); err != nil {
		return result, err
	}
	// Defensive: if ctx was cancelled but no error was captured.
	if err := ctx.Err(); err != nil {
		return result, err
	}
	return result, nil
}

// processFileForPool — same body as processFile/processFileForLocked,
// returns fileResult directly.
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

// (Plus: full Walk + WalkLocked + processFile + processFileForLocked
// from L19 solutions, with import paths rewritten.)

// unused import guards (remove after carrying forward L19's full source):
var _ = sync.Mutex{}
var _ = atomic.Pointer[error]{}
```

> Note on the closer goroutine: `go func() { _ = g.Wait(); close(resultsCh) }()` is the classic "wait for all workers then close" pattern from L16 slides. The first Wait swallows the error (a goroutine can't return it); the SECOND Wait in main captures it.

> Note on calling `g.Wait()` twice: it's safe — Wait returns the same captured error each time. The first call (in the closer) just needs to signal "all done"; the second call (in main) extracts the error.

- [ ] **Step 7:** Create the solutions test file. Carry the L19 runWalkSuite + 9 sub-tests verbatim, add the WalkPool adapter test, and add TestWalkPoolWorkers. Full content too long to embed here — copy `lessons/19-context/solutions/internal/aggregator/aggregator_test.go` verbatim, add the imports for errgroupx if needed, then append:

```go
func TestWalkPool(t *testing.T) {
	walkPoolAdapter := func(ctx context.Context, dir string, timeout time.Duration) (WalkResult, error) {
		return WalkPool(ctx, dir, timeout, 4)
	}
	runWalkSuite(t, walkPoolAdapter)
}

func TestWalkPoolWorkers(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")
	writeFile(t, dir, "c.log", "2026-05-21T14:32:00 ERROR c\n")

	cases := []struct {
		name    string
		workers int
	}{
		{"workers=1 (serial)", 1},
		{"workers=2", 2},
		{"workers=8", 8},
		{"workers=0 (default)", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := WalkPool(context.Background(), dir, 0, tc.workers)
			if err != nil {
				t.Fatalf("WalkPool: %v", err)
			}
			want := map[string]int{"INFO": 1, "WARN": 1, "ERROR": 1}
			if !reflect.DeepEqual(result.Counts, want) {
				t.Errorf("workers=%d: Counts = %v, want %v", tc.workers, result.Counts, want)
			}
		})
	}
}
```

### Step 8-11: `cmd/aggregator/` (gains -mode=pool + -workers)

- [ ] **Step 8:** Create `lessons/20-concurrency-patterns/exercises/cmd/aggregator/main.go`. Carry forward L19's main.go; extend `-mode` to accept "pool"; add `-workers` flag; pass `workers` to WalkPool. Pseudocode:

```go
// Add to parseFlags:
//   case strings.HasPrefix(a, "-workers="):
//       w, err := strconv.Atoi(a[len("-workers="):])
//       if err != nil || w < 0 { return ..., fmt.Errorf("invalid -workers: %s", a) }
//       workers = w

// Add to mode switch in run:
//   case "pool":
//       result, walkErr = aggregator.WalkPool(ctx, dir, timeout, workers)
//   default: ... add "pool" to the error message

// Optional: print warning if -workers set but mode != pool.
```

Skeleton has the stub. Solutions has the full impl.

- [ ] **Step 9:** Create `lessons/20-concurrency-patterns/exercises/cmd/aggregator/main_test.go` (SKELETON). Carry L19's tests. Plus a TODO-stub for TestAggregatorPool.

- [ ] **Step 10:** Create `lessons/20-concurrency-patterns/solutions/cmd/aggregator/main.go` (full impl with -workers flag).

- [ ] **Step 11:** Create `lessons/20-concurrency-patterns/solutions/cmd/aggregator/main_test.go`. Carry L19's tests verbatim. Add:

```go
func TestAggregatorPool(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 WARN slow\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:32:00 ERROR oops\n")

	cmd := exec.Command(bin, "-dir="+logDir, "-mode=pool", "-workers=2")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("aggregator: %v (stderr: %s)", err, stderr.String())
	}

	want := "ERROR: 1\nINFO: 1\nWARN: 1\n"
	if stdout.String() != want {
		t.Errorf("stdout mismatch:\ngot:\n%s\nwant:\n%s", stdout.String(), want)
	}
}
```

### Step 12: Verify and commit Task 4

- [ ] **Step 12:** Standard sweep:

```bash
gofmt -l lessons/20-concurrency-patterns/
go test ./lessons/20-concurrency-patterns/exercises/... 2>&1 | tail -5
go test ./lessons/20-concurrency-patterns/solutions/... 2>&1 | tail -10
go test -race ./lessons/20-concurrency-patterns/solutions/... 2>&1 | tail -10
make test
make test-race
golangci-lint run ./...
go vet ./...

# Smoke
TMP=$(mktemp -d /tmp/lesson20-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO ok" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow" > "$TMP/logs/b.log"
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir="$TMP/logs"
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir="$TMP/logs" -mode=mutex
go run ./lessons/20-concurrency-patterns/solutions/cmd/aggregator -dir="$TMP/logs" -mode=pool -workers=2
rm -rf "$TMP"

git add lessons/20-concurrency-patterns/exercises/internal/ lessons/20-concurrency-patterns/solutions/internal/ lessons/20-concurrency-patterns/exercises/cmd/ lessons/20-concurrency-patterns/solutions/cmd/
git commit -m "feat(lesson-20): main — aggregator gains WalkPool (errgroupx-based); cmd adds -mode=pool + -workers"
```

Expected: all green; three modes produce identical output (`INFO: 1\nWARN: 1\n`).

---

## Task 5: Slide deck — 5 concepts

Heavy-explanatory pattern per concept. ~35 slides.

**File:** `lessons/20-concurrency-patterns/slides/slides.md`

Concept order:

1. **Worker pool** — bounded parallelism. `pathsCh` + N consumer goroutines. When to use. Common-mistake: too many workers.
2. **Fan-out** — distribute work to N consumers. Worked: producer + N workers. Common-mistake: unbounded fan-out.
3. **Fan-in** — multiplex N producers into one consumer. Worked: FanIn warmup. Common-mistake: forgetting to close merged channel.
4. **Pipeline** — chained stages. Worked: read CSV → parse → filter → write (referencing L13). Common-mistake: not propagating ctx through pipeline stages.
5. **`errgroupx`** — canonical "manage N workers." Worked: WalkPool. Common-mistake: forgetting to call Wait (goroutines leak).

- [ ] **Step 1-7:** Author the deck.
- [ ] **Step 8:** Verify slides build + commit:

```bash
make slides-build
grep -q "20-concurrency-patterns" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/20-concurrency-patterns/slides/
git commit -m "feat(lesson-20): slides — concurrency patterns (5 concepts)"
```

---

## Task 6: README

Mirrors slide concepts (5 sections); features `make test-race`.

**File:** `lessons/20-concurrency-patterns/README.md`

- [ ] **Step 1:** Author the README (~270 lines):
  - "What you'll learn" — 5 bullets.
  - "What's different from L19" — third Walk variant + errgroupx package.
  - "The package layout" — annotated tree.
  - Per-concept sections (5).
  - "Exercise: warmup — fanin".
  - "Exercise: internal/errgroupx — implement the canonical errgroup".
  - "Exercise: main — WalkPool + CLI flags".
  - "Daily habits" — gofmt + vet + test + test-race.
  - "How to run" — three modes smoke + workers comparison.
  - "Going further" — Read (golang.org/x/sync/errgroup source; Go blog Pipelines and cancellation); Try (benchmark all three modes; add pipeline stages).

- [ ] **Step 2:** Commit:

```bash
git add lessons/20-concurrency-patterns/README.md
git commit -m "docs(lesson-20): README — concurrency patterns self-study"
```

---

## Task 7: End-to-end verification + final review + PR

```bash
make test
make test-race
go test -v ./lessons/20-concurrency-patterns/solutions/...
go test -race ./lessons/20-concurrency-patterns/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/20-concurrency-patterns/
make slides-build && grep -q "20-concurrency-patterns" dist/index.html && rm -rf dist
```

Dispatch feature-dev:code-reviewer over `git diff main..HEAD`. Apply any fixes. Push branch + `gh pr create`.

---

## Critical file paths

To be created:

- `lessons/20-concurrency-patterns/` (directory)
- `lessons/20-concurrency-patterns/README.md`
- `lessons/20-concurrency-patterns/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/20-concurrency-patterns/exercises/warmup/fanin/{fanin.go, fanin_test.go}`
- `lessons/20-concurrency-patterns/exercises/internal/errgroupx/{errgroupx.go, errgroupx_test.go}`
- `lessons/20-concurrency-patterns/exercises/internal/logparse/{logparse.go, logparse_test.go}`
- `lessons/20-concurrency-patterns/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}`
- `lessons/20-concurrency-patterns/exercises/cmd/aggregator/{main.go, main_test.go}`
- `lessons/20-concurrency-patterns/solutions/...` (mirrored)

To be modified:

- `tools/build-index/main.go` — update L20 slug from `patterns` to `concurrency-patterns` (Task 1)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md` (Phase 3 design)
- `lessons/19-context/solutions/...` (carry-forward source)
- `Makefile` (no changes; `test-race` target added in L18)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
