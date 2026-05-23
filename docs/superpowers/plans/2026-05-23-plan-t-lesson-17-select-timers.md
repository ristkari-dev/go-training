# Plan T — Lesson 17 (Select & timers) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 17 of Phase 3 — students learn `select` for channel multiplexing and timer primitives for time-bounded operations. The aggregator from L16 evolves: gains a per-file timeout via `select { case res := <-ch: ...; case <-time.After(d): skip }`. Walk's return type changes from `(map[string]int, error)` to `(WalkResult, error)` where `WalkResult` carries both `Counts` and `TimedOut []string`. Sets up L19 (`context`) by introducing the close-of-done-channel cancellation pattern.

**Architecture:** Same per-lesson pattern as Plans D-S. Six tasks. Skeleton tests in warmup + main subpackages (Phase 2/3 default). Heavy-explanatory slide deck with **five** concept blocks. Second lesson under the Phase 3 design.

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `encoding/json`, `errors`, `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `regexp`, `sort`, `strconv`, `strings`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan T: lesson 17 is complete; `make test` green; `cmd/aggregator -timeout=<dur>` works end-to-end; race detector clean.

### Design decisions (2 user-approved + 6 plan-recommended)

**User-approved via brainstorming:**

1. **Five slide concepts.** (1) `select` basics; (2) `time.After` + timeouts; (3) `time.NewTicker`; (4) `default` branch (non-blocking select); (5) **Cancellation via close-of-done-channel** as its own slot — foundation for L19's context lesson. Done-channel cancellation gets a dedicated treatment (Motivation/Basics/Worked/Common-mistake/Recap) so students see the pattern formally before L19 generalizes it via `context.Context`.

2. **`Walk` returns `(WalkResult, error)`.** `WalkResult struct { Counts map[string]int; TimedOut []string }`. Library stays I/O-free; CLI extracts TimedOut and prints to stderr. Cleaner engineering than "library writes to stderr directly" and cleaner semantics than "timeouts as errors via sentinel" (success-with-timeouts has nil error).

**Plan-recommended:**

3. **`timeout=0` means "no timeout"** via the nil-channel-never-fires trick:
   ```go
   var timeoutCh <-chan time.Time
   if timeout > 0 { timeoutCh = time.After(timeout) }
   select { case r := <-perFileCh: ...; case <-timeoutCh: ... }
   ```
   A nil channel in a select never receives. Clean implementation; no special-case branch. CLI's `-timeout` flag defaults to 0 (no timeout) — matches Go's "zero value should be useful" idiom.

4. **Warmup `wait.WaitWithTimeout` uses a sentinel `ErrTimeout`.** `var ErrTimeout = errors.New("wait: timeout")`. Callers `errors.Is(err, wait.ErrTimeout)` to distinguish timeout from real failures. L11 callback (sentinel errors are a tool students already know).

5. **`time.NewTicker` worked example is hypothetical, not in the aggregator code.** Tickers don't fit the aggregator's "process N files then exit" shape. Slides cover ticker via a small "heartbeat printer" example; the aggregator code doesn't use one. Mentioned in README as a sneak preview of L20's worker-pool stats reporting.

6. **Timeouts are NOT errors** — they're expected behavior in real log aggregation (some files are huge/network-mounted/locked). Walk returns nil error on success-with-timeouts; populates `WalkResult.TimedOut`. Real file errors (open failure, parse error from non-timeout source) still abort with wrapped error per L16 semantics.

7. **`processFile` signature unchanged from L16.** Worker goroutines still take `(path string, resultsCh chan<- fileResult)`. The timeout `select` happens in the REDUCE loop in main, not inside workers. Workers don't know about timeouts; they just process and send. This separates concerns cleanly: workers do work, main does coordination.

8. **Per-file timeout, not whole-Walk timeout.** Each file gets its own `time.After(timeout)` clock in the reduce loop's select. If timeout is 5s and we have 10 files all taking 4s, total Walk time can be 40s+ but no file is "slow." If one file takes 6s with 5s timeout, it's skipped; others still get their 5s budget. This matches the design's "per-file timeout" framing.

---

## Plans F-S lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24) covers nested subpackages.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `gofmt -w .` and `go vet ./...` mentions in README.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).
8. gofmt 1.19+ normalises godoc list indentation — `gofmt -w` if triggered.
9. **Build-index slug fix needed**: current `tools/build-index/main.go:57` has `Slug: "select"`; the lesson directory is `17-select-timers`. Fix in Task 1 (precedent: Plan Q fixed L14 slug; Plan S fixed L16 slug).

---

## File Structure

After Plan T (~20 files, same as L16):

```
lessons/17-select-timers/
├── README.md                                       (Task 5)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 4)
├── exercises/
│   ├── warmup/wait/
│   │   ├── wait.go                                 (Task 2)
│   │   └── wait_test.go                            (Task 2 — SKELETON)
│   ├── cmd/aggregator/
│   │   ├── main.go                                 (Task 3 — adds -timeout flag)
│   │   └── main_test.go                            (Task 3 — SKELETON; integration + timeout test)
│   └── internal/
│       ├── logparse/                               (Task 3 — verbatim from L16)
│       │   ├── logparse.go
│       │   └── logparse_test.go
│       └── aggregator/
│           ├── aggregator.go                       (Task 3 — Walk evolves with timeout + WalkResult)
│           └── aggregator_test.go                  (Task 3 — SKELETON)
└── solutions/  (mirrored)
```

---

## Conventions

- **Branch:** `feature/plan-t-lesson-17-select-timers`
- **Commit messages:** Conventional Commits
- **Carry-forward sources:**
  - `lessons/16-goroutines-channels/solutions/internal/logparse/{logparse.go, logparse_test.go}` (verbatim)
  - `lessons/16-goroutines-channels/solutions/internal/aggregator/aggregator.go` (REFACTORED — Walk signature changes; processFile unchanged)
  - `lessons/16-goroutines-channels/solutions/cmd/aggregator/main.go` (REFACTORED — adds -timeout flag, handles WalkResult)

---

## Task 1: Scaffold + restructure + slug fix

Same dance as Plans J/K/L/M/N/O/P/Q/R/S.

- [ ] **Step 1:** `make new-lesson NAME=17-select-timers`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/17-select-timers/exercises/warmup.go
rm lessons/17-select-timers/exercises/warmup_test.go
rm lessons/17-select-timers/exercises/main.go
rm lessons/17-select-timers/exercises/main_test.go
rm lessons/17-select-timers/solutions/warmup.go
rm lessons/17-select-timers/solutions/warmup_test.go
rm lessons/17-select-timers/solutions/main.go
rm lessons/17-select-timers/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree.

- [ ] **Step 4:** Fix L17 slug in `tools/build-index/main.go`:

Find: `{Number: "17", Slug: "select", Title: "Select & timers", Blurb: "select · time.After", Phase: 3},`

Change to: `{Number: "17", Slug: "select-timers", Title: "Select & timers", Blurb: "select · time.After", Phase: 3},`

Verify by running:
```bash
go test ./tools/build-index/...
make slides-build
grep -q "17-select-timers" dist/index.html && echo "✓ found"
ls dist/lessons/17-select-timers/ && echo "✓ dir created"
rm -rf dist
```

- [ ] **Step 5:** Commit:

```bash
git add lessons/17-select-timers/ tools/build-index/main.go
git commit -m "feat(lessons): scaffold lesson 17-select-timers with empty subpackage layout

Also fix build-index master list: L17 slug was 'select' (legacy
placeholder); update to 'select-timers' to match the actual lesson
directory."
```

---

## Task 2: Author the warm-up — `wait` subpackage

`WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error)` — block on ch or return (zero T, ErrTimeout) after d. Demonstrates `select { case v := <-ch: ...; case <-time.After(d): ... }` in its tightest form.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/17-select-timers/{exercises,solutions}/warmup/wait`

- [ ] **Step 2:** Create `lessons/17-select-timers/exercises/warmup/wait/wait.go`:

```go
// Package wait is the lesson 17 warm-up: blocking on a channel with a
// timeout via select + time.After.
//
// WaitWithTimeout returns the first value sent on ch, or ErrTimeout
// if d elapses first. The implementation is the textbook timeout
// pattern in three lines of select.
//
// Generic via [T any] (lesson 12) so the same function works for any
// channel element type.
package wait

import (
	"errors"
	"time"
)

// ErrTimeout is returned by WaitWithTimeout when the deadline d
// elapses before a value arrives on ch.
//
// Callers can errors.Is(err, ErrTimeout) to distinguish timeout
// from other errors (which WaitWithTimeout doesn't itself produce —
// the type signature allows it for future extensibility).
var ErrTimeout = errors.New("wait: timeout")

// WaitWithTimeout returns the first value from ch, or (zero T,
// ErrTimeout) if d elapses first.
//
// Examples:
//
//	ch := make(chan int, 1); ch <- 42
//	v, err := WaitWithTimeout(ch, time.Second)   // v=42, err=nil
//
//	emptyCh := make(chan string)
//	v, err := WaitWithTimeout(emptyCh, time.Millisecond)
//	// v="", err=ErrTimeout
//
// Hint:
//   1. var zero T
//   2. select {
//        case v := <-ch:
//            return v, nil
//        case <-time.After(d):
//            return zero, ErrTimeout
//      }
//
// The `var zero T` idiom (L12 callback) gives you T's zero value
// without knowing the concrete type. Don't write `return T{}, ...`
// (T might not be a struct) or `return *new(T), ...` (ugly).
func WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error) {
	_ = time.After
	panic("TODO: var zero T; select with case v := <-ch and case <-time.After(d)")
}
```

- [ ] **Step 3:** Create `lessons/17-select-timers/exercises/warmup/wait/wait_test.go` (SKELETON):

```go
package wait

import (
	"errors"
	"testing"
	"time"
)

// TestWaitImmediate is a SKELETON. Use a buffered channel with one
// value already in it; WaitWithTimeout should return that value
// without ever hitting the timeout.
func TestWaitImmediate(t *testing.T) {
	// TODO:
	//   ch := make(chan int, 1); ch <- 42
	//   v, err := WaitWithTimeout(ch, time.Second)
	//   if err != nil → t.Fatalf("unexpected error: %v", err)
	//   if v != 42 → t.Errorf("got %d, want 42", v)
	_ = time.Second
}

// TestWaitTimeout is a SKELETON. Use a channel that nobody sends on;
// WaitWithTimeout should hit the timeout and return ErrTimeout.
func TestWaitTimeout(t *testing.T) {
	// TODO:
	//   ch := make(chan int)
	//   _, err := WaitWithTimeout(ch, time.Millisecond)
	//   if !errors.Is(err, ErrTimeout) → t.Errorf("expected ErrTimeout, got %v", err)
	_ = errors.Is
}

// TestWaitGenericString is a SKELETON. Verify the function works
// with T=string (not just int).
func TestWaitGenericString(t *testing.T) {
	// TODO:
	//   ch := make(chan string, 1); ch <- "hi"
	//   v, err := WaitWithTimeout(ch, time.Second)
	//   assert v == "hi", err == nil
}
```

- [ ] **Step 4:** Create `lessons/17-select-timers/solutions/warmup/wait/wait.go`:

```go
// Package wait is the lesson 17 warm-up reference implementation.
package wait

import (
	"errors"
	"time"
)

// ErrTimeout is returned by WaitWithTimeout when d elapses first.
var ErrTimeout = errors.New("wait: timeout")

// WaitWithTimeout returns the first value from ch, or (zero T,
// ErrTimeout) after d.
func WaitWithTimeout[T any](ch <-chan T, d time.Duration) (T, error) {
	var zero T
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(d):
		return zero, ErrTimeout
	}
}
```

- [ ] **Step 5:** Create `lessons/17-select-timers/solutions/warmup/wait/wait_test.go`:

```go
package wait

import (
	"errors"
	"testing"
	"time"
)

func TestWaitImmediate(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	v, err := WaitWithTimeout(ch, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Errorf("got %d, want 42", v)
	}
}

func TestWaitTimeout(t *testing.T) {
	ch := make(chan int)

	v, err := WaitWithTimeout(ch, 10*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
	if v != 0 {
		t.Errorf("expected zero value 0, got %d", v)
	}
}

func TestWaitGenericString(t *testing.T) {
	ch := make(chan string, 1)
	ch <- "hi"

	v, err := WaitWithTimeout(ch, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hi" {
		t.Errorf("got %q, want %q", v, "hi")
	}
}

func TestWaitZeroDuration(t *testing.T) {
	// d=0 fires immediately. The select picks one of the two ready
	// cases at random; either the channel receive (if ch has a value)
	// or the time.After(0) which fires immediately. With an empty
	// channel, only time.After is ready → ErrTimeout.
	ch := make(chan int)
	_, err := WaitWithTimeout(ch, 0)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout with d=0 on empty channel, got %v", err)
	}
}
```

> Note on TestWaitTimeout's 10ms duration: short enough that the test runs quickly, long enough to avoid scheduler flakiness. Tests using sub-microsecond timeouts can occasionally race with goroutine scheduling and produce spurious failures.

> Note on TestWaitZeroDuration: with `time.After(0)`, the timer fires immediately. The select sees both cases simultaneously if the channel ALSO has a value; the runtime picks one at random. With an empty channel (this test), only the timeout case is ready, so it always fires.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/17-select-timers/
go test ./lessons/17-select-timers/exercises/warmup/wait/... -v 2>&1 | tail -10
go test ./lessons/17-select-timers/solutions/warmup/wait/... -v 2>&1 | tail -20
go test -race ./lessons/17-select-timers/solutions/warmup/wait/... 2>&1 | tail -5
make test
golangci-lint run ./...
go vet ./...

git add lessons/17-select-timers/exercises/warmup/ lessons/17-select-timers/solutions/warmup/
git commit -m "feat(lesson-17): warmup — wait (WaitWithTimeout with select + time.After)"
```

Expected: gofmt empty; exercises pass vacuously; solutions 4 tests PASS; -race clean; make test green.

---

## Task 3: Main — logparse carry-forward + aggregator evolution + cmd/aggregator with -timeout

The biggest task. Aggregator gets `WalkResult` return + per-file timeout. CLI adds `-timeout` flag and stderr printing of timed-out paths.

**Files (12 total — 6 per side):**

### Step 1: Create directories

```bash
mkdir -p lessons/17-select-timers/{exercises,solutions}/internal/logparse
mkdir -p lessons/17-select-timers/{exercises,solutions}/internal/aggregator
mkdir -p lessons/17-select-timers/{exercises,solutions}/cmd/aggregator
```

### Step 2-3: `internal/logparse/` (verbatim from L16)

- [ ] **Step 2:** Forward-port `lessons/16-goroutines-channels/solutions/internal/logparse/logparse.go` to both `exercises/internal/logparse/logparse.go` and `solutions/internal/logparse/logparse.go`. Identical content; no path rewrites needed (the package has no internal imports).

- [ ] **Step 3:** Same for `logparse_test.go`. Both trees get the same full implementation + tests.

### Step 4-7: `internal/aggregator/` (EVOLVES from L16)

- [ ] **Step 4:** Create `lessons/17-select-timers/exercises/internal/aggregator/aggregator.go`:

```go
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
//   1. entries, err := os.ReadDir(dir)
//   2. if err != nil → return WalkResult{}, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
//   3. Build files []string from entries (skip subdirs).
//   4. resultsCh := make(chan fileResult, len(files))
//   5. for each path: go processFile(path, resultsCh)
//   6. Track which paths have been "claimed" so timeouts know which
//      file timed out. Easiest: ASSIGN each select iteration to a file
//      in launch order — but that's not how concurrency works.
//      Better: use a map[string]bool of paths-still-pending; each
//      received fileResult removes r.path from the map; a timeout
//      removes ONE remaining path arbitrarily.
//
//      Even simpler: for each file, do an inner select that waits
//      for THAT file's specific result OR the timeout. This means
//      one timer per file — exactly the "per-file timeout" semantics.
//      But we don't know which fileResult will arrive next on the
//      shared channel; results come in completion order.
//
//      SIMPLEST WORKING APPROACH: keep a set of pending paths. The
//      reduce loop's select has two cases: (a) read a fileResult,
//      remove r.path from pending; (b) timeout, pick any remaining
//      pending path and add it to TimedOut. Iterate until pending
//      is empty.
//
// Sketch:
//   pending := map[string]struct{}{}
//   for _, p := range files { pending[p] = struct{}{} }
//   for len(pending) > 0 {
//       var timeoutCh <-chan time.Time
//       if timeout > 0 { timeoutCh = time.After(timeout) }
//       select {
//       case r := <-resultsCh:
//           delete(pending, r.path)
//           if r.err != nil → return result, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
//           for level, count := range r.counts { result.Counts[level] += count }
//       case <-timeoutCh:
//           // Pick one remaining pending path arbitrarily, mark timed out.
//           for p := range pending {
//               result.TimedOut = append(result.TimedOut, p)
//               delete(pending, p)
//               break  // just one per timeout
//           }
//       }
//   }
//   return result, nil
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
```

- [ ] **Step 5:** Create `lessons/17-select-timers/exercises/internal/aggregator/aggregator_test.go` (SKELETON):

```go
package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// writeFile is a test helper.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// TestWalkNoTimeout is a SKELETON. With timeout=0 (no timeout), Walk
// behaves like L16: all files processed; WalkResult.Counts has the
// aggregate; WalkResult.TimedOut is empty.
func TestWalkNoTimeout(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		// TODO: write one fixture file; call Walk(dir, 0); assert
		// result.Counts and len(result.TimedOut) == 0.
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO: similar pattern, assert merged counts.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO: empty dir; assert Counts is empty map (not nil),
		// TimedOut is empty.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO: write a malformed file; assert err != nil and mentions
		// the file path.
	})
}

// TestWalkAllFilesTimedOut is a SKELETON. With a tiny timeout
// (1 nanosecond), every file's select fires the timeout case before
// the receive case. Assert: Counts is empty; all files appear in
// TimedOut; err is nil (timeouts aren't errors).
func TestWalkAllFilesTimedOut(t *testing.T) {
	// TODO:
	//   dir := t.TempDir()
	//   writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n")
	//   writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN slow\n")
	//   result, err := Walk(dir, 1*time.Nanosecond)
	//   if err != nil → t.Fatalf("timeouts shouldn't error: %v", err)
	//   if len(result.Counts) != 0 → t.Errorf("expected empty counts, got %v", result.Counts)
	//   if len(result.TimedOut) != 2 → t.Errorf("expected 2 timed-out files, got %v", result.TimedOut)
	_ = time.Nanosecond
}
```

- [ ] **Step 6:** Create `lessons/17-select-timers/solutions/internal/aggregator/aggregator.go`:

```go
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
```

> Note on the iteration-over-map-and-break pattern in the timeout case: Go's map iteration order is randomized; picking ONE arbitrary pending key per timeout means tests assert on set membership (`contains "a.log"` in TimedOut), not on order. This is correct for the use case — "some files timed out" is the answer; WHICH file got the first timeout slot is irrelevant.

> Note on the per-file-timeout semantics: each outer-loop iteration restarts the `time.After(timeout)` clock. So timeout=5s and 3 slow files = worst case 15s total. If you wanted a "whole Walk has 5s total" budget, you'd hoist the time.After call outside the loop. The per-file semantics matches the design's framing ("each file gets its own time budget").

- [ ] **Step 7:** Create `lessons/17-select-timers/solutions/internal/aggregator/aggregator_test.go`:

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

func TestWalkNoTimeout(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")

		result, err := Walk(dir, 0)
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

		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}

		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("Counts = %v, want %v", result.Counts, want)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		if len(result.Counts) != 0 {
			t.Errorf("expected empty Counts, got %v", result.Counts)
		}
		if len(result.TimedOut) != 0 {
			t.Errorf("expected no timeouts, got %v", result.TimedOut)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// Note: same non-deterministic ordering caveat as L16 — only
		// bad.log errors, so any goroutine ordering eventually hits it.
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")

		_, err := Walk(dir, 0)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := Walk("/nonexistent/path/that/does/not/exist", 0)
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

		result, err := Walk(dir, 0)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(result.Counts, want) {
			t.Errorf("subdirs should be skipped: Counts = %v, want %v", result.Counts, want)
		}
	})
}

func TestWalkAllFilesTimedOut(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n")
	writeFile(t, dir, "b.log", "2026-05-21T14:31:00 WARN b\n")

	result, err := Walk(dir, 1*time.Nanosecond)
	if err != nil {
		t.Fatalf("timeouts should not error: %v", err)
	}
	if len(result.Counts) != 0 {
		t.Errorf("expected empty Counts when all files time out, got %v", result.Counts)
	}
	if len(result.TimedOut) != 2 {
		t.Errorf("expected 2 timed-out files, got %d: %v", len(result.TimedOut), result.TimedOut)
	}
}
```

### Step 8-11: `cmd/aggregator/` (carries forward L16 + adds -timeout flag)

- [ ] **Step 8:** Create `lessons/17-select-timers/exercises/cmd/aggregator/main.go`:

```go
// Package main is the lesson 17 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>                    # no timeout
//	aggregator -dir=<path> -timeout=5s        # 5-second per-file timeout
//
// Walks the directory, parses each log file in parallel, prints the
// combined level counts to stdout. Files that exceed -timeout are
// printed to stderr.
//
// Exit codes:
//   - 0 on success (including success with timeouts)
//   - 1 on any error (missing -dir, bad dir, malformed file, etc.)
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/17-select-timers/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. Parses args, calls aggregator.Walk,
// prints sorted level counts to stdout and any timed-out paths to stderr.
//
// Hint:
//   1. Parse -dir=<path> (required) and -timeout=<dur> (optional, default 0)
//      via time.ParseDuration.
//   2. result, err := aggregator.Walk(dir, timeout)
//   3. if err != nil → return err
//   4. For each result.TimedOut entry: fmt.Fprintf(stderr, "timeout: %s\n", path)
//   5. Sort result.Counts keys alphabetically; print "LEVEL: count" per line to stdout.
func run(args []string, stdout io.Writer, stderr io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = sort.Strings
	_ = time.ParseDuration
	panic("TODO: parse -dir + -timeout; call Walk; print stdout + stderr")
}
```

- [ ] **Step 9:** Create `lessons/17-select-timers/exercises/cmd/aggregator/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAggregatorGoldenPath is a SKELETON.
func TestAggregatorGoldenPath(t *testing.T) {
	// TODO: build binary; create fixture dir; run; assert stdout.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}

// TestAggregatorTimeout is a SKELETON. Run with -timeout=1ns; assert
// stderr contains "timeout:" for each fixture file; stdout is empty
// (or just trailing newline).
func TestAggregatorTimeout(t *testing.T) {
	// TODO: build binary; fixture dir with 2 files; run -timeout=1ns;
	// assert stderr mentions each fixture path.
}
```

- [ ] **Step 10:** Create `lessons/17-select-timers/solutions/cmd/aggregator/main.go`:

```go
// Package main is the lesson 17 aggregator CLI reference implementation.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ristkari-dev/go-training/lessons/17-select-timers/solutions/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	dir, timeout, err := parseFlags(args)
	if err != nil {
		return err
	}

	result, err := aggregator.Walk(dir, timeout)
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

func parseFlags(args []string) (string, time.Duration, error) {
	var (
		dir     string
		timeout time.Duration
	)
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-dir="):
			dir = a[len("-dir="):]
		case strings.HasPrefix(a, "-timeout="):
			t, err := time.ParseDuration(a[len("-timeout="):])
			if err != nil {
				return "", 0, fmt.Errorf("invalid -timeout: %w", err)
			}
			timeout = t
		default:
			return "", 0, fmt.Errorf("unknown flag %q", a)
		}
	}
	if dir == "" {
		return "", 0, fmt.Errorf("-dir=<path> is required")
	}
	return dir, timeout, nil
}
```

- [ ] **Step 11:** Create `lessons/17-select-timers/solutions/cmd/aggregator/main_test.go`:

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
		"2026-05-21T14:30:00 INFO a\n"+
			"2026-05-21T14:31:00 INFO b\n"+
			"2026-05-21T14:32:00 WARN c\n")
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
	if stderr.String() != "" {
		t.Errorf("expected empty stderr, got %q", stderr.String())
	}
}

func TestAggregatorMalformed(t *testing.T) {
	// Same non-determinism caveat as L16. Safe because only bad.log errors.
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

	// stdout should be empty (no counts to print) or only newline-ish.
	if stdout.String() != "" {
		t.Errorf("expected empty stdout when everything times out, got %q", stdout.String())
	}

	// stderr should mention both files. Don't assert specific order
	// (Go map iteration randomization).
	if !strings.Contains(stderr.String(), "timeout:") {
		t.Errorf("expected 'timeout:' in stderr, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "a.log") || !strings.Contains(stderr.String(), "b.log") {
		t.Errorf("expected both file paths in stderr, got %q", stderr.String())
	}
}
```

### Step 12: Verify and commit

- [ ] **Step 12:** Verify and commit Task 3:

```bash
gofmt -l lessons/17-select-timers/
go test ./lessons/17-select-timers/exercises/... -v 2>&1 | tail -20
go test ./lessons/17-select-timers/solutions/... -v 2>&1 | tail -40
go test -race ./lessons/17-select-timers/solutions/... 2>&1 | tail -5
make test
golangci-lint run ./...
go vet ./...

# Smoke test the binary
TMP=$(mktemp -d /tmp/lesson17-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO ok" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow" > "$TMP/logs/b.log"
# Normal run (no timeout)
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir="$TMP/logs"
# Timeout run (everything times out)
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir="$TMP/logs" -timeout=1ns
rm -rf "$TMP"

git add lessons/17-select-timers/exercises/internal/ lessons/17-select-timers/solutions/internal/ lessons/17-select-timers/exercises/cmd/ lessons/17-select-timers/solutions/cmd/
git commit -m "feat(lesson-17): main — aggregator gains per-file timeout + WalkResult; cmd adds -timeout flag"
```

Expected: gofmt empty; exercises pass vacuously; solutions all PASS (7+ aggregator sub-tests + 4 cmd integration tests); -race clean; smoke shows normal output without timeout and "timeout:" stderr lines with `-timeout=1ns`.

---

## Task 4: Slide deck — 5 concepts

Heavy-explanatory pattern (Motivation/Basics/Worked-example/Common-mistake/Recap) per concept. Plus cover + roadmap + Practice + Closing thought + What-we-learned + Up-next.

**File:** `lessons/17-select-timers/slides/slides.md`

~32 slides. Concept order matches design:

1. **`select` basics** — multiplex N channels; random pick when multiple cases ready; blocks until one is ready. Common-mistake: assuming priority (there is none).
2. **`time.After` + timeouts** — `time.After(d)` returns a `<-chan time.Time`; combined with `select`, gives timeouts. Worked: the aggregator's per-file timeout. Common-mistake: `time.After` allocates a fresh timer per call — in hot loops, prefer `time.NewTimer` + reset.
3. **`time.NewTicker`** — periodic events. `t := time.NewTicker(d); defer t.Stop(); for { <-t.C; ... }`. Worked: hypothetical heartbeat printer. Common-mistake: forgetting `t.Stop()` leaks goroutines.
4. **`default` branch (non-blocking select)** — `select { case v := <-ch: ...; default: ... }` doesn't block. Worked: try-send pattern. Common-mistake: busy-loop antipattern.
5. **Cancellation via close-of-done-channel** — `done := make(chan struct{})`; receivers `case <-done: return`; canceller `close(done)`. Foundation for L19. Worked: a worker that exits when done closes. Common-mistake: sending on `done` instead of closing.

- [ ] **Step 1-7:** Author the deck. Cover → roadmap → 5 concepts → Practice → Closing thought (L19 preview) → What-we-learned → Up-next (Lesson 18: sync & memory model).

- [ ] **Step 8:** Verify and commit:

```bash
make slides-build
grep -q "17-select-timers" dist/index.html && echo "✓ in index"
ls dist/lessons/17-select-timers/
rm -rf dist

git add lessons/17-select-timers/slides/
git commit -m "feat(lesson-17): slides — Select & timers (5 concepts)"
```

---

## Task 5: README

Mirrors slide concepts (5 sections); one Common-mistake paragraph per concept; gofmt/vet daily-habits.

**File:** `lessons/17-select-timers/README.md`

- [ ] **Step 1:** Author the README (~240 lines):
  - "What you'll learn" — 5 bullets.
  - "What's different from L16" — Aggregator gains timeout; WalkResult struct return.
  - "The package layout" — annotated tree.
  - Per-concept sections (5).
  - "Exercise: warm-up — wait" (~3 lines).
  - "Exercise: main — aggregator with timeout" (~6 lines).
  - "Daily habits" — gofmt -w, go vet, go test, go test -race (still pre-L18 daily habits).
  - "How to run" — including timeout smoke.
  - "Going further" — Read (Go blog on timers, the runtime); Try (whole-Walk timeout vs per-file; ticker-based heartbeat in cmd binary).

- [ ] **Step 2:** Commit:

```bash
git add lessons/17-select-timers/README.md
git commit -m "docs(lesson-17): README — Select & timers self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1:** Full sweep.

```bash
make test
```

Expected: all lessons (01-17) + tools pass.

- [ ] **Step 2:** Solutions verbose.

```bash
go test -v ./lessons/17-select-timers/solutions/...
```

Expected (~18 sub-tests across 4 packages, all PASS).

- [ ] **Step 3:** Race detector clean.

```bash
go test -race ./lessons/17-select-timers/...
```

- [ ] **Step 4:** Exercises pass vacuously.

```bash
go test ./lessons/17-select-timers/exercises/...
```

- [ ] **Step 5:** Static analysis.

```bash
go vet ./...
golangci-lint run ./...
gofmt -l lessons/17-select-timers/
```

- [ ] **Step 6:** Slides build.

```bash
make slides-build
grep -q "17-select-timers" dist/index.html
ls dist/lessons/17-select-timers/
rm -rf dist
```

- [ ] **Step 7:** Binary smoke (golden + timeout cases).

```bash
TMP=$(mktemp -d /tmp/lesson17-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO ok" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow" > "$TMP/logs/b.log"

# Golden path (no timeout)
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir="$TMP/logs"
# Expected: "INFO: 1\nWARN: 1\n"

# Timeout path (1ns, everything times out)
go run ./lessons/17-select-timers/solutions/cmd/aggregator -dir="$TMP/logs" -timeout=1ns 2>&1 1>/dev/null
# Expected on stderr: "timeout: <path>/a.log\ntimeout: <path>/b.log\n" (order varies)

rm -rf "$TMP"
```

- [ ] **Step 8:** Final code review subagent → push branch → create PR.

```bash
git push -u origin feature/plan-t-lesson-17-select-timers
gh pr create --title "feat(lesson-17): Select & timers — per-file timeout + WalkResult evolution" --body "..."
```

---

## Verification

After all 6 tasks:

```bash
make test                                                                       # green
go test -v ./lessons/17-select-timers/solutions/... 2>&1 | grep -E "PASS|FAIL"  # ~18 PASS lines
go test -race ./lessons/17-select-timers/...                                    # no races
go vet ./...                                                                    # clean
golangci-lint run ./...                                                         # 0 issues
gofmt -l lessons/17-select-timers/                                              # empty
make slides-build && grep -q "17-select-timers" dist/index.html && rm -rf dist  # green
```

## Critical file paths

To be created:

- `lessons/17-select-timers/` (directory)
- `lessons/17-select-timers/README.md`
- `lessons/17-select-timers/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/17-select-timers/exercises/warmup/wait/{wait.go, wait_test.go}`
- `lessons/17-select-timers/exercises/cmd/aggregator/{main.go, main_test.go}`
- `lessons/17-select-timers/exercises/internal/logparse/{logparse.go, logparse_test.go}`
- `lessons/17-select-timers/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}`
- `lessons/17-select-timers/solutions/...` (mirrored)

To be modified:

- `tools/build-index/main.go` — update L17 slug from `select` to `select-timers` (Task 1).

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md` (Phase 3 design)
- `lessons/16-goroutines-channels/solutions/internal/logparse/{logparse.go, logparse_test.go}` (verbatim carry-forward source)
- `lessons/16-goroutines-channels/solutions/internal/aggregator/aggregator.go` (refactor source — processFile carries forward unchanged; Walk gets new signature)
- `lessons/16-goroutines-channels/solutions/cmd/aggregator/main.go` (refactor source — gains -timeout flag and stderr printing)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
