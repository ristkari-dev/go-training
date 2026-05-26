# Plan U — Lesson 18 (sync & memory model) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 18 of Phase 3 — students learn the `sync` package (`Mutex`, `RWMutex`, `WaitGroup`, `Once`), `sync/atomic`, the race detector formally, and the Go memory model intuition (happens-before via channels and mutexes). The aggregator gains a second implementation: `WalkLocked` (mutex + WaitGroup) alongside L17's `Walk` (channel-based). CLI gets `-mode=channel|mutex` to pick at runtime. **`make test-race` Makefile target added** and featured in Phase 3+ daily habits from this lesson onward.

**Architecture:** Same per-lesson pattern as Plans D-T. Seven tasks (one extra for Makefile target + race-detector integration). Skeleton tests in warmup + main subpackages (Phase 2/3 default). Heavy-explanatory slide deck with **five** concept blocks. Third lesson under the Phase 3 design.

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `encoding/json`, `errors`, `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `regexp`, `sort`, `strconv`, `strings`, `sync`, `sync/atomic`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan U: lesson 18 is complete; `make test` AND `make test-race` green; `cmd/aggregator -mode=mutex` works end-to-end; race detector clean on all of L18's code.

### Design decisions (2 user-approved + 7 plan-recommended)

**User-approved via brainstorming:**

1. **Five slide concepts.** (1) Mutex + memory model intuition; (2) RWMutex; (3) WaitGroup; (4) Once + sync/atomic (bundled); (5) Share by communicating vs share by locking — meta concept comparing Walk and WalkLocked side-by-side.

2. **Both Walk implementations live in the same package.** `Walk(dir, timeout)` (channel-based, verbatim L17) AND `WalkLocked(dir, timeout)` (new, mutex+WaitGroup) both in `internal/aggregator/aggregator.go`. Same WalkResult; same tests parameterized over both. CLI gets `-mode=channel|mutex` flag.

**Plan-recommended:**

3. **Counter warmup has BOTH Counter (correct, mutex-protected) and CounterUnsafe (no mutex, buggy).** Tests verify Counter only. CounterUnsafe exists in the file for slides/README to walk through — students run it manually under `-race` to see the data race detected. Testing CounterUnsafe would either flake (race may not manifest on all hardware) or fail under `-race` (which would break `make test-race`).

4. **`make test-race` Makefile target added in Task 1.** Sits alongside `make test` (which stays as the fast smoke). Phase 3 design says `make test` keeps meaning "fast smoke" (Phase 1-2 lessons have no concurrency); Phase 3+ adds `make test-race`. Same shape as existing `make test` (excludes `/exercises$` packages); adds `-race` flag.

5. **WalkLocked uses `sync.Mutex` + `sync.WaitGroup`.** Shared state: `sharedCounts map[string]int`, `sharedTimedOut []string`, both guarded by a single `sync.Mutex`. Workers `wg.Add(1)` before launch + `defer wg.Done()`. Main goroutine `wg.Wait()` then reads (no mutex needed after Wait — happens-before guaranteed by WaitGroup semantics).

6. **Per-file timeout still uses select + time.After in WalkLocked.** Timing is orthogonal to coordination style. The "wait for a result OR timeout" pattern needs select; that's true regardless of whether the "result" comes via channel send (Walk) or via shared-state update (WalkLocked). WalkLocked uses a `doneCh chan struct{}` per file that workers close on completion; the reduce loop selects on `<-doneCh` vs `<-time.After(timeout)`.

7. **Aggregator tests parameterized via `t.Run("Walk", ...)` + `t.Run("WalkLocked", ...)`.** Both implementations exercise the same 7 sub-tests (single-file, multiple-files, empty-dir, malformed-line-returns-error, nonexistent-dir-returns-error, subdirs-skipped, all-files-timed-out). Total ~14 sub-tests. The parameterization sets up the lesson's centerpiece comparison.

8. **CLI default mode is `channel`** (L17 default behavior unchanged). `-mode=mutex` is opt-in. Preserves backward compatibility with L17 smoke tests.

9. **`sync/atomic` example uses `int64` and `atomic.AddInt64`/`atomic.LoadInt64`** — not the generic `atomic.Int64` type (Go 1.19+). The generic type is cleaner but the standalone functions are still more common in real code and connect to the lesson's "low-level primitive" framing. Mention the generic type in slides as the modern alternative.

---

## Plans F-T lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24) covers nested subpackages.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `gofmt -w .` and `go vet ./...` mentions in README. **NEW: `make test-race` becomes daily habit from this lesson.**
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).
8. gofmt 1.19+ normalises godoc list indentation — `gofmt -w` if triggered.
9. **Build-index slug fix needed**: current `tools/build-index/main.go:58` has `Slug: "sync"`; lesson directory is `18-sync-memory-model`. Fix in Task 1.

---

## File Structure

After Plan U (~22 files):

```
lessons/18-sync-memory-model/
├── README.md                                       (Task 5)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 4)
├── exercises/
│   ├── warmup/counter/
│   │   ├── counter.go                              (Task 2 — Counter + CounterUnsafe)
│   │   └── counter_test.go                         (Task 2 — SKELETON; tests Counter only)
│   ├── cmd/aggregator/
│   │   ├── main.go                                 (Task 3 — adds -mode flag)
│   │   └── main_test.go                            (Task 3 — SKELETON; integration + mode test)
│   └── internal/
│       ├── logparse/                               (Task 3 — verbatim from L17)
│       │   ├── logparse.go
│       │   └── logparse_test.go
│       └── aggregator/
│           ├── aggregator.go                       (Task 3 — Walk + WalkLocked + processFile)
│           └── aggregator_test.go                  (Task 3 — SKELETON; parameterized t.Run)
└── solutions/  (mirrored)

Makefile                                            (Task 1 — adds test-race target)
tools/build-index/main.go                           (Task 1 — slug fix)
```

---

## Conventions

- **Branch:** `feature/plan-u-lesson-18-sync-memory-model`
- **Commit messages:** Conventional Commits
- **Carry-forward sources:**
  - `lessons/17-select-timers/solutions/internal/logparse/{logparse.go, logparse_test.go}` (verbatim)
  - `lessons/17-select-timers/solutions/internal/aggregator/aggregator.go` (Walk + processFile + WalkResult carry forward; WalkLocked is NEW)
  - `lessons/17-select-timers/solutions/cmd/aggregator/main.go` (carries forward + adds -mode flag)

---

## Task 1: Scaffold + restructure + slug fix + Makefile target

Same dance as Plans J-T. Plus two extras: build-index slug fix (L18 was `sync` placeholder) AND new `test-race` Makefile target.

- [ ] **Step 1:** `make new-lesson NAME=18-sync-memory-model`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/18-sync-memory-model/exercises/warmup.go
rm lessons/18-sync-memory-model/exercises/warmup_test.go
rm lessons/18-sync-memory-model/exercises/main.go
rm lessons/18-sync-memory-model/exercises/main_test.go
rm lessons/18-sync-memory-model/solutions/warmup.go
rm lessons/18-sync-memory-model/solutions/warmup_test.go
rm lessons/18-sync-memory-model/solutions/main.go
rm lessons/18-sync-memory-model/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree.

- [ ] **Step 4:** Fix L18 slug in `tools/build-index/main.go`:

Find: `{Number: "18", Slug: "sync", Title: "sync & memory model", Blurb: "Mutex · race detector", Phase: 3},`

Change to: `{Number: "18", Slug: "sync-memory-model", Title: "sync & memory model", Blurb: "Mutex · race detector", Phase: 3},`

Verify:
```bash
go test ./tools/build-index/...
make slides-build
grep -q "18-sync-memory-model" dist/index.html && echo "✓ in index"
ls dist/lessons/18-sync-memory-model/ && echo "✓ dir created"
rm -rf dist
```

- [ ] **Step 5:** Add `test-race` target to `Makefile`. Insert ABOVE the existing `test-exercises` target (so it sits next to `test`):

```makefile
.PHONY: test-race
test-race: ## Run tests with the race detector (excludes exercises; 5-20× slower)
	@pkgs=$$(go list ./... | grep -v '/exercises$$'); \
	if [ -z "$$pkgs" ]; then echo "no testable packages"; exit 0; fi; \
	go test -race $$pkgs
```

Verify the target appears in `make help`:
```bash
make help | grep test-race
# Expected: "  test-race              Run tests with the race detector ..."
```

Verify it runs (will take longer than `make test`):
```bash
make test-race 2>&1 | tail -5
# Expected: all packages report "ok"; no race warnings.
```

- [ ] **Step 6:** Commit:

```bash
git add lessons/18-sync-memory-model/ tools/build-index/main.go Makefile
git commit -m "feat(lessons): scaffold lesson 18-sync-memory-model with empty subpackage layout

Also:
- Fix build-index master list: L18 slug was 'sync' (legacy placeholder);
  update to 'sync-memory-model' to match the actual lesson directory.
- Add 'make test-race' Makefile target. Phase 3 design specified this
  for L18 onward; sits alongside 'make test' which stays as fast smoke.
  Excludes /exercises packages (same as make test); adds -race flag."
```

---

## Task 2: Author the warm-up — `counter` subpackage

`Counter` (mutex-protected, correct) + `CounterUnsafe` (no mutex, buggy). Tests verify Counter only. CounterUnsafe is for pedagogy — students run it manually under `-race` to see the failure.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/18-sync-memory-model/{exercises,solutions}/warmup/counter`

- [ ] **Step 2:** Create `lessons/18-sync-memory-model/exercises/warmup/counter/counter.go`:

```go
// Package counter is the lesson 18 warm-up: a thread-safe counter
// (with sync.Mutex) alongside a deliberately-buggy version (without
// the mutex) for pedagogical comparison.
//
// Counter is what students implement. CounterUnsafe exists in this
// file so students can run it manually under `go test -race` and see
// the race detector flag the data race.
//
// The unit tests only test Counter (the correct version). Testing
// CounterUnsafe would either flake (the race doesn't always manifest)
// or fail under -race (breaking `make test-race`). Instead, the slides
// and README walk through running CounterUnsafe manually.
package counter

import "sync"

// Counter is safe for concurrent use. Multiple goroutines may call
// Inc and Value simultaneously.
//
// Hint:
//   - Embed a sync.Mutex (or hold a pointer; embedded is simpler here).
//   - Inc: Lock; n++; Unlock. Or defer Unlock.
//   - Value: Lock; v := n; Unlock; return v.
//
// The zero value of Counter is a usable, unlocked counter at 0. No
// constructor needed (this is the idiomatic Go pattern for sync types).
type Counter struct {
	mu sync.Mutex
	n  int
}

// Inc atomically increments the counter by 1.
func (c *Counter) Inc() {
	_ = c.mu
	panic("TODO: c.mu.Lock(); c.n++; c.mu.Unlock()")
}

// Value returns the current counter value.
func (c *Counter) Value() int {
	panic("TODO: c.mu.Lock(); v := c.n; c.mu.Unlock(); return v")
}

// CounterUnsafe is the SAME interface without the mutex. It has a
// data race on n if Inc is called concurrently. The unit tests do NOT
// test this type — it exists so you can run it manually under
// `go test -race` and see the race detector report a data race.
//
// To see the failure firsthand, create a temporary test like:
//
//	func TestCounterUnsafeRaces(t *testing.T) {
//	    var c CounterUnsafe
//	    var wg sync.WaitGroup
//	    for i := 0; i < 100; i++ {
//	        wg.Add(1)
//	        go func() { defer wg.Done(); for j := 0; j < 100; j++ { c.Inc() } }()
//	    }
//	    wg.Wait()
//	    t.Logf("CounterUnsafe.Value() = %d (expected 10000)", c.Value())
//	}
//
// Run it with: go test -race -run TestCounterUnsafeRaces ./...
// The race detector will print "DATA RACE" + a stack trace pointing
// at c.n in CounterUnsafe.Inc and Value. Without -race the test
// "passes" but the value is usually less than 10000 (the unsynchronized
// increments lose updates).
type CounterUnsafe struct {
	n int
}

func (c *CounterUnsafe) Inc()    { c.n++ }
func (c *CounterUnsafe) Value() int { return c.n }
```

- [ ] **Step 3:** Create `lessons/18-sync-memory-model/exercises/warmup/counter/counter_test.go` (SKELETON):

```go
package counter

import (
	"sync"
	"testing"
)

// TestCounter is a SKELETON. The full test spawns 100 goroutines, each
// doing 100 increments (10,000 total), then asserts Counter.Value()
// equals exactly 10000. This passes under `go test -race` because
// Counter uses a mutex.
//
// Hint:
//   1. var c Counter
//   2. var wg sync.WaitGroup
//   3. for i := 0; i < 100; i++ {
//        wg.Add(1)
//        go func() { defer wg.Done(); for j := 0; j < 100; j++ { c.Inc() } }()
//      }
//   4. wg.Wait()
//   5. if got := c.Value(); got != 10000 → t.Errorf("got %d, want 10000", got)
func TestCounter(t *testing.T) {
	// TODO: write the test per the hint above.
	_ = sync.WaitGroup{}
}
```

- [ ] **Step 4:** Create `lessons/18-sync-memory-model/solutions/warmup/counter/counter.go`:

```go
// Package counter is the lesson 18 warm-up reference implementation.
package counter

import "sync"

// Counter is safe for concurrent use.
type Counter struct {
	mu sync.Mutex
	n  int
}

// Inc atomically increments the counter by 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

// Value returns the current counter value.
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// CounterUnsafe is the same interface without the mutex. Used only
// for slides/README demonstrations — DELIBERATELY buggy.
type CounterUnsafe struct {
	n int
}

func (c *CounterUnsafe) Inc()    { c.n++ }
func (c *CounterUnsafe) Value() int { return c.n }
```

- [ ] **Step 5:** Create `lessons/18-sync-memory-model/solutions/warmup/counter/counter_test.go`:

```go
package counter

import (
	"sync"
	"testing"
)

// TestCounter spawns 100 goroutines, each doing 100 increments
// (10,000 total). Asserts the final value is exactly 10000. Passes
// cleanly under `go test -race` because Counter is mutex-protected.
func TestCounter(t *testing.T) {
	var c Counter
	var wg sync.WaitGroup

	const goroutines = 100
	const incPerG = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incPerG; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	want := goroutines * incPerG
	if got := c.Value(); got != want {
		t.Errorf("Counter.Value() = %d, want %d", got, want)
	}
}
```

> Note: this test deliberately exercises `c.Inc()` from many goroutines simultaneously. Without the mutex (i.e., using CounterUnsafe), it would race AND produce a wrong value almost always. With the mutex, it's deterministic and -race-clean.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/18-sync-memory-model/
go test ./lessons/18-sync-memory-model/exercises/warmup/counter/... -v 2>&1 | tail -10
go test ./lessons/18-sync-memory-model/solutions/warmup/counter/... -v 2>&1 | tail -10
go test -race ./lessons/18-sync-memory-model/solutions/warmup/counter/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/18-sync-memory-model/exercises/warmup/ lessons/18-sync-memory-model/solutions/warmup/
git commit -m "feat(lesson-18): warmup — counter (sync.Mutex correct + CounterUnsafe pedagogical)"
```

Expected: gofmt empty; exercises pass vacuously; solutions TestCounter PASS (10000 from 100×100 goroutines); -race clean; make test green; make test-race green.

---

## Task 3: Main — logparse carry-forward + aggregator with WalkLocked + cmd with -mode flag

The biggest task. Aggregator gains a second implementation; CLI gains a mode selector.

**Files (12 total — 6 per side):**

### Step 1: Create directories

```bash
mkdir -p lessons/18-sync-memory-model/{exercises,solutions}/internal/logparse
mkdir -p lessons/18-sync-memory-model/{exercises,solutions}/internal/aggregator
mkdir -p lessons/18-sync-memory-model/{exercises,solutions}/cmd/aggregator
```

### Step 2-3: `internal/logparse/` (verbatim from L17)

- [ ] **Step 2:** Forward-port `lessons/17-select-timers/solutions/internal/logparse/logparse.go` to both `exercises/internal/logparse/logparse.go` and `solutions/internal/logparse/logparse.go`. Identical content.

- [ ] **Step 3:** Same for `logparse_test.go`.

### Step 4-7: `internal/aggregator/` (carries Walk forward + adds WalkLocked)

- [ ] **Step 4:** Create `lessons/18-sync-memory-model/exercises/internal/aggregator/aggregator.go`:

```go
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
//   1. List files (same as Walk).
//   2. result := WalkResult{Counts: map[string]int{}}
//   3. var mu sync.Mutex
//   4. var wg sync.WaitGroup
//   5. For each file: wg.Add(1); spawn goroutine that:
//        - opens file, parses with logparse.Parse
//        - on success: mu.Lock(); merge counts into result.Counts; mu.Unlock()
//        - sends a done signal on a per-file channel
//   6. Main loop: for each file, select on doneCh vs time.After(timeout).
//      On done: continue to next file. On timeout: mu.Lock(); append
//      to result.TimedOut; mu.Unlock().
//   7. wg.Wait() at the end ensures all workers finish before we
//      return (no goroutine leaks).
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
```

- [ ] **Step 5:** Create `lessons/18-sync-memory-model/exercises/internal/aggregator/aggregator_test.go` (SKELETON):

```go
package aggregator

import (
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

// walkFunc is the common type of Walk and WalkLocked, used to
// parameterize tests so both implementations run the same suite.
type walkFunc func(dir string, timeout time.Duration) (WalkResult, error)

// runWalkSuite executes the same set of sub-tests against a walkFunc.
// Skeleton: cases are TODO-stubbed; both Walk and WalkLocked panic
// inside, so we don't actually call them.
func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write a fixture file; call fn(dir, 0); assert Counts.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO.
		_ = strings.Contains
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		// TODO.
	})

	t.Run("subdirs-skipped", func(t *testing.T) {
		// TODO.
	})

	t.Run("all-files-timed-out", func(t *testing.T) {
		// TODO. Use 1*time.Microsecond per L17 lesson (race-detector headroom).
		_ = time.Microsecond
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

- [ ] **Step 6:** Create `lessons/18-sync-memory-model/solutions/internal/aggregator/aggregator.go`:

```go
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
```

> Note on WalkLocked's "wait for all workers" vs Walk's "early return on error": this is a deliberate behavior difference and documents the tradeoff. Walk can return immediately because its buffered channel absorbs straggler sends; WalkLocked needs to wg.Wait() before reading result safely. The test suite's `malformed-line-returns-error` test asserts only that an error is returned and mentions the bad file — doesn't compare exact result content.

> Note on `atomic.Pointer[error]`: Go 1.19+ feature. Generic atomic pointer. Used here so firstErr storage is safe without needing the same mutex that guards result. Demonstrates sync/atomic alongside the mutex; foreshadows the L20 errgroup pattern.

> Note on the per-worker goroutine inside the outer worker goroutine: this is the "spawn-and-timeout" pattern from L17. The outer goroutine waits on either the inner's completion or the timeout; the inner does the actual file work. Goroutine count grows by 2× files; in real code with many files you'd switch to a worker pool (L20).

- [ ] **Step 7:** Create `lessons/18-sync-memory-model/solutions/internal/aggregator/aggregator_test.go`:

```go
package aggregator

import (
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

type walkFunc func(dir string, timeout time.Duration) (WalkResult, error)

func runWalkSuite(t *testing.T, fn walkFunc) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log",
			"2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")

		result, err := fn(dir, 0)
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

		result, err := fn(dir, 0)
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
		result, err := fn(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts, got %v", result.Counts)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// Note: only bad.log errors, so any goroutine ordering eventually
		// hits it. Walk returns immediately on the first error; WalkLocked
		// waits for all workers via wg.Wait then returns firstErr.
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")

		_, err := fn(dir, 0)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := fn("/nonexistent/path/that/does/not/exist", 0)
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

		result, err := fn(dir, 0)
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

		// 1µs (not 1ns) gives ~1000× headroom over scheduler/race
		// overhead — see L17 plan for rationale.
		result, err := fn(dir, 1*time.Microsecond)
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
}

func TestWalk(t *testing.T) {
	runWalkSuite(t, Walk)
}

func TestWalkLocked(t *testing.T) {
	runWalkSuite(t, WalkLocked)
}
```

### Step 8-11: `cmd/aggregator/` (adds -mode flag)

- [ ] **Step 8:** Create `lessons/18-sync-memory-model/exercises/cmd/aggregator/main.go`:

```go
// Package main is the lesson 18 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>                          # mode=channel (default)
//	aggregator -dir=<path> -mode=channel            # explicit channel
//	aggregator -dir=<path> -mode=mutex              # mutex-based
//	aggregator -dir=<path> -mode=mutex -timeout=5s  # combine
//
// Walks the directory, parses each log file in parallel, prints the
// combined level counts to stdout. Files that exceed -timeout are
// printed to stderr. -mode selects which Walk variant to use:
// channel-based (L17) or mutex-based (L18).
//
// Exit codes: 0 on success; 1 on any error.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point.
//
// Hint:
//   1. Parse -dir, -timeout, -mode flags. -mode defaults to "channel".
//   2. Pick walk function: channel → aggregator.Walk; mutex → aggregator.WalkLocked.
//      Invalid -mode value → return error.
//   3. result, err := walkFn(dir, timeout)
//   4. Print TimedOut to stderr, sorted Counts to stdout (per L17).
func run(args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = aggregator.WalkLocked
	_ = sort.Strings
	_ = time.ParseDuration
	panic("TODO: parse -dir + -timeout + -mode; pick Walk or WalkLocked; print results")
}
```

- [ ] **Step 9:** Create `lessons/18-sync-memory-model/exercises/cmd/aggregator/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: see L17 pattern.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorMode is a SKELETON. Run with -mode=mutex and assert
// the output matches the default (-mode=channel) for the same input.
func TestAggregatorMode(t *testing.T) {
	// TODO: build binary; create fixture dir; run twice (default and
	// -mode=mutex); assert stdout matches.
}
```

- [ ] **Step 10:** Create `lessons/18-sync-memory-model/solutions/cmd/aggregator/main.go`:

```go
// Package main is the lesson 18 aggregator CLI reference implementation.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/18-sync-memory-model/solutions/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type walkFunc func(dir string, timeout time.Duration) (aggregator.WalkResult, error)

func run(args []string, stdout io.Writer, stderr io.Writer) error {
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

	result, err := fn(dir, timeout)
	if err != nil {
		return err
	}

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
	return nil
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

- [ ] **Step 11:** Create `lessons/18-sync-memory-model/solutions/cmd/aggregator/main_test.go`:

```go
package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

// TestAggregatorMode runs the binary twice — once with default mode
// (channel) and once with -mode=mutex — over the same fixture. Asserts
// stdout matches between the two runs. Demonstrates that both Walk
// variants produce identical output for the same input.
func TestAggregatorMode(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 WARN b\n")
	writeFile(t, logDir, "b.log", "2026-05-21T14:32:00 ERROR c\n")

	runOnce := func(mode string) string {
		t.Helper()
		args := []string{"-dir=" + logDir}
		if mode != "" {
			args = append(args, "-mode="+mode)
		}
		cmd := exec.Command(bin, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("aggregator (mode=%q): %v (stderr: %s)", mode, err, stderr.String())
		}
		return stdout.String()
	}

	defaultOut := runOnce("")
	channelOut := runOnce("channel")
	mutexOut := runOnce("mutex")

	if defaultOut != channelOut {
		t.Errorf("default mode should match -mode=channel:\ndefault:\n%s\nchannel:\n%s",
			defaultOut, channelOut)
	}
	if channelOut != mutexOut {
		t.Errorf("channel and mutex modes should produce identical output:\nchannel:\n%s\nmutex:\n%s",
			channelOut, mutexOut)
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

// TestAggregatorInvalidMode asserts the CLI rejects unknown -mode values.
func TestAggregatorInvalidMode(t *testing.T) {
	bin := buildBinary(t)
	logDir := t.TempDir()
	writeFile(t, logDir, "a.log", "2026-05-21T14:30:00 INFO ok\n")

	cmd := exec.Command(bin, "-dir="+logDir, "-mode=bogus")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for invalid -mode, got nil")
	}
	if !strings.Contains(stderr.String(), "-mode") {
		t.Errorf("stderr should mention -mode, got %q", stderr.String())
	}
}
```

### Step 12: Verify and commit

- [ ] **Step 12:** Verify and commit Task 3:

```bash
gofmt -l lessons/18-sync-memory-model/
go test ./lessons/18-sync-memory-model/exercises/... 2>&1 | tail -10
go test ./lessons/18-sync-memory-model/solutions/... 2>&1 | tail -20
go test -race ./lessons/18-sync-memory-model/solutions/... 2>&1 | tail -10
make test
make test-race
golangci-lint run ./...
go vet ./...

# Smoke
TMP=$(mktemp -d /tmp/lesson18-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO ok" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow" > "$TMP/logs/b.log"
# Default mode=channel
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir="$TMP/logs"
# Explicit mutex
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir="$TMP/logs" -mode=mutex
# Timeout in mutex mode
go run ./lessons/18-sync-memory-model/solutions/cmd/aggregator -dir="$TMP/logs" -mode=mutex -timeout=1ns
rm -rf "$TMP"

git add lessons/18-sync-memory-model/exercises/internal/ lessons/18-sync-memory-model/solutions/internal/ lessons/18-sync-memory-model/exercises/cmd/ lessons/18-sync-memory-model/solutions/cmd/
git commit -m "feat(lesson-18): main — aggregator gains WalkLocked (mutex+WG); cmd adds -mode flag"
```

Expected: gofmt empty; exercises pass vacuously; solutions all PASS; -race clean on both Walk and WalkLocked; smoke shows mutex output matches channel output.

---

## Task 4: Slide deck — 5 concepts

Heavy-explanatory pattern (Motivation/Basics/Worked-example/Common-mistake/Recap) per concept.

**File:** `lessons/18-sync-memory-model/slides/slides.md`

~32 slides. Concept order matches design:

1. **`sync.Mutex` + memory model intuition** — `mu.Lock()` / `defer mu.Unlock()`; zero value usable; happens-before via mutex and channel. Common-mistake: copying a Mutex.
2. **`sync.RWMutex`** — RLock/RUnlock for readers; Lock/Unlock for writers. Worked: read-heavy cache. Common-mistake: using RWMutex when reads aren't truly hot.
3. **`sync.WaitGroup`** — Add/Done/Wait. Worked: WalkLocked. Common-mistake: Add inside the goroutine (race with Wait).
4. **`sync.Once` + `sync/atomic` (bundled)** — Once for lazy init; atomic for lock-free single-word ops. Worked: config singleton + request counter. Common-mistake: mismatched atomic/non-atomic access.
5. **Share by communicating vs share by locking** — Walk vs WalkLocked, side by side. Common-mistake: assuming one style is "better" — both are correct; pick the simpler one for the problem.

- [ ] **Step 1-7:** Author the deck.

- [ ] **Step 8:** Verify and commit:

```bash
make slides-build
grep -q "18-sync-memory-model" dist/index.html && echo "✓ in index"
ls dist/lessons/18-sync-memory-model/
rm -rf dist

git add lessons/18-sync-memory-model/slides/
git commit -m "feat(lesson-18): slides — sync & memory model (5 concepts)"
```

---

## Task 5: README

Mirrors slide concepts (5 sections); features `make test-race` in daily-habits block (FIRST lesson where it does).

**File:** `lessons/18-sync-memory-model/README.md`

- [ ] **Step 1:** Author the README (~270 lines):
  - "What you'll learn" — 5 bullets.
  - "What's different from L17" — second aggregator implementation; race detector becomes daily habit via `make test-race`.
  - "The package layout" — annotated tree.
  - Per-concept sections (5).
  - "Exercise: warm-up — counter" with the "run CounterUnsafe under -race manually" demo instructions.
  - "Exercise: main — aggregator with both modes" — refactor + WalkLocked + CLI flag.
  - "Daily habits" — gofmt -w, go vet, go test, **`make test-race`** featured prominently.
  - "How to run" — including -mode=mutex smoke + the CounterUnsafe race demo.
  - "Going further" — Read (Go memory model spec); Try (benchmark Walk vs WalkLocked).

- [ ] **Step 2:** Commit:

```bash
git add lessons/18-sync-memory-model/README.md
git commit -m "docs(lesson-18): README — sync & memory model self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1-7:** Standard sweep (make test, make test-race, solutions verbose, exercises vacuous-pass, static analysis, slides build, binary smoke covering all -mode/-timeout combinations).

---

## Task 7: Final code review + push PR

Same pattern as L16/L17 final reviews — dispatch a feature-dev:code-reviewer subagent over `git diff main..HEAD`. Apply any fixes the reviewer surfaces. Then `gh pr create` per the established template.

---

## Verification

```bash
make test                                                                      # green
make test-race                                                                 # green (NEW)
go test -v ./lessons/18-sync-memory-model/solutions/... 2>&1 | grep -E "PASS|FAIL"  # ~20 PASS lines (14 aggregator + Counter + ~5 cmd)
go vet ./... && golangci-lint run ./...                                        # clean
gofmt -l lessons/18-sync-memory-model/                                         # empty
make slides-build && grep -q "18-sync-memory-model" dist/index.html && rm -rf dist  # green
```

## Critical file paths

To be created:

- `lessons/18-sync-memory-model/` (directory)
- `lessons/18-sync-memory-model/README.md`
- `lessons/18-sync-memory-model/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/18-sync-memory-model/exercises/warmup/counter/{counter.go, counter_test.go}`
- `lessons/18-sync-memory-model/exercises/cmd/aggregator/{main.go, main_test.go}`
- `lessons/18-sync-memory-model/exercises/internal/logparse/{logparse.go, logparse_test.go}`
- `lessons/18-sync-memory-model/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}`
- `lessons/18-sync-memory-model/solutions/...` (mirrored)

To be modified:

- `tools/build-index/main.go` — update L18 slug from `sync` to `sync-memory-model` (Task 1).
- `Makefile` — add `test-race` target (Task 1).

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md` (Phase 3 design — locks the `make test-race` decision)
- `lessons/17-select-timers/solutions/...` (carry-forward source for logparse + Walk + cmd/aggregator)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
