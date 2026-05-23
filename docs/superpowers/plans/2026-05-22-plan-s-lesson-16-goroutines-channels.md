# Plan S — Lesson 16 (Goroutines & channels) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 16 of Phase 3 — students' first exposure to Go's concurrency primitives. Five concepts: the `go` keyword and goroutine model, unbuffered channels (synchronous handoff), buffered channels (capacity), `range`/`close`/goroutine lifecycle, and the **aggregator pattern** — multiple producers, single consumer — named explicitly so students recognize it across L17-L20 as the running example evolves. The warmup implements a tiny `Echo` function (range over input channel, send through new output channel, close on EOF). The main exercise builds `internal/aggregator/Walk(dir)` — walks a directory, spawns one goroutine per file, sends per-file level counts through a results channel; plus a `cmd/aggregator/` binary with end-to-end integration tests.

**Architecture:** Same per-lesson pattern as Plans D-R. Six tasks. Skeleton tests in warmup + main subpackages (Phase 2/3 default). Heavy-explanatory slide deck with **five** concept blocks. First lesson under the Phase 3 design (`docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md`).

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `encoding/json`, `errors`, `flag`, `fmt`, `io`, `os`, `os/exec`, `path/filepath`, `regexp`, `sort`, `strconv`, `strings`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan S: lesson 16 is complete; `make test` green; `cmd/aggregator` works end-to-end against a tempdir of fixture log files; `internal/aggregator` tests pass under `-race`.

### Design decisions (2 user-approved + 6 plan-recommended)

**User-approved via brainstorming:**

1. **Five slide concepts with the "aggregator pattern" as the named closing concept.** Concepts 1-4 cover the primitives; concept 5 names the producer-consumer pattern students will see again as the aggregator evolves through L17-L20. The named pattern carries forward.

2. **Full `os/exec` integration test for `cmd/aggregator`** — match L13's `cmd/expenses-import` precedent. Build binary into `t.TempDir`, exec it with fixture log files, assert stdout content. Two sub-tests: golden path (3 fixture files with known counts) + malformed file (exit 1 with stderr message).

**Plan-recommended:**

3. **`aggregator.Walk(dir string) (map[string]int, error)` signature.** Returns partial map + wrapped error on file failures. Same shape as L13's `csvimport.Parse(io.Reader)` for visual continuity ("io-consuming function returns typed slice/map + error").

4. **`internal/logparse` carries forward verbatim from L14.** Only import paths change — same source, same tests, same fuzz preview from L15 (carried via L15's `internal/logparse` rather than the L14 original — they're identical because L15 carried L14 verbatim).

5. **Race detector mentioned in slides but NOT in daily habits.** Per Phase 3 design, `make test-race` becomes a daily-habits READMEs feature from L18 onward. L16's slides introduce `go test -race` as a tool that exists; the dedicated Makefile target lands in L18's plan.

6. **Goroutine model in concept 1 emphasizes "no return values; communicate via channels."** The single most-confusing thing for students coming from sync languages — `go f()` returns immediately, no Future/Promise. Lead with this so the rest of the concepts make sense.

7. **No buffered-channel "capacity choice" rule beyond rules of thumb.** Concept 3 covers unbuffered vs buffered with the simple guide: "unbuffered when you need to know the receiver got it; buffered when N producers can run ahead of one consumer." The full "how to pick a capacity" discussion is L20 (worker pools) territory.

8. **Echo warmup uses a goroutine internally.** `Echo(in <-chan string) <-chan string` MUST return the output channel immediately and finish via a spawned goroutine; otherwise Echo would have to drain `in` synchronously, defeating the "channels enable concurrency" lesson. The skeleton hint comment explicitly mentions this pattern.

---

## Plans F-R lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24, `lessons/.*/exercises/`) covers nested subpackages — no config changes needed.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `gofmt -w .` and `go vet ./...` mentions in README.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).
8. gofmt 1.19+ normalises godoc list indentation — `gofmt -w` if triggered.
9. The build-index master list slug for L16 needs verification (precedent: L14 fix `stdlib` → `time-strings-regex`; L15 was correct as `structure`).

---

## File Structure

After Plan S (~20 files, similar to L11/L13):

```
lessons/16-goroutines-channels/
├── README.md                                       (Task 5)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 4)
├── exercises/
│   ├── warmup/echo/
│   │   ├── echo.go                                 (Task 2)
│   │   └── echo_test.go                            (Task 2 — SKELETON)
│   ├── cmd/aggregator/
│   │   ├── main.go                                 (Task 3 — CLI)
│   │   └── main_test.go                            (Task 3 — SKELETON; integration test)
│   └── internal/
│       ├── logparse/                               (Task 3 — verbatim L15)
│       │   ├── logparse.go
│       │   └── logparse_test.go
│       └── aggregator/
│           ├── aggregator.go                       (Task 3 — Walk function)
│           └── aggregator_test.go                  (Task 3 — SKELETON)
└── solutions/  (mirrored)
```

---

## Conventions

- **Branch:** `feature/plan-s-lesson-16-goroutines-channels`
- **Commit messages:** Conventional Commits
- **Carry-forward sources:**
  - `lessons/15-structure/solutions/internal/logparse/{logparse.go, logparse_test.go}` (forward-port verbatim; path-rewrite only)

---

## Task 1: Scaffold + restructure

Same dance as Plans J/K/L/M/N/O/P/Q/R.

- [ ] **Step 1:** `make new-lesson NAME=16-goroutines-channels`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/16-goroutines-channels/exercises/warmup.go
rm lessons/16-goroutines-channels/exercises/warmup_test.go
rm lessons/16-goroutines-channels/exercises/main.go
rm lessons/16-goroutines-channels/exercises/main_test.go
rm lessons/16-goroutines-channels/solutions/warmup.go
rm lessons/16-goroutines-channels/solutions/warmup_test.go
rm lessons/16-goroutines-channels/solutions/main.go
rm lessons/16-goroutines-channels/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree.

- [ ] **Step 4:** Check `tools/build-index/main.go` master list. L16's row should match the slug `16-goroutines-channels`. If it's a placeholder like `concurrency` or similar, update it. Re-run `make slides-build` and verify `dist/lessons/16-goroutines-channels/` is created (precedent: Plan Q fixed L14's slug).

- [ ] **Step 5:** Commit:

```bash
git add lessons/16-goroutines-channels/
git commit -m "feat(lessons): scaffold lesson 16-goroutines-channels with empty subpackage layout"
```

If the build-index slug was updated, commit that separately first or bundle:
```bash
git add tools/build-index/main.go
git commit -m "chore(tools): update build-index slug for lesson 16"
```

---

## Task 2: Author the warm-up — `echo` subpackage

`Echo(in <-chan string) <-chan string` — read all from in, send each through a new output channel, close output on input EOF. Spawned via goroutine so the function returns immediately.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/16-goroutines-channels/{exercises,solutions}/warmup/echo`

- [ ] **Step 2:** Create `lessons/16-goroutines-channels/exercises/warmup/echo/echo.go`:

```go
// Package echo is the lesson 16 warm-up: a tiny demonstration of
// goroutines + channels working together.
//
// Echo takes an input channel of strings and returns a new output
// channel of strings. The output gets every value the input received,
// in order. When the input is closed (no more values), the output is
// closed too.
//
// The pedagogical point: Echo returns the output channel IMMEDIATELY.
// The reading-and-writing happens inside a spawned goroutine. This is
// the canonical "function that does work concurrently" shape in Go —
// not a Future, not a Promise, just a channel you can read from while
// the work happens in the background.
package echo

// Echo returns a new channel that receives everything sent on in,
// in order. When in is closed, the returned channel is closed too.
//
// Hint:
//   1. out := make(chan string)
//   2. go func() {
//        for v := range in {
//            out <- v
//        }
//        close(out)
//      }()
//   3. return out
//
// Three things to internalise:
//   - out is created BEFORE the goroutine starts.
//   - The goroutine closes out when in closes.
//   - Echo returns immediately; the work happens later.
func Echo(in <-chan string) <-chan string {
	panic("TODO: make output channel; spawn goroutine that range-copies and closes; return output")
}
```

- [ ] **Step 3:** Create `lessons/16-goroutines-channels/exercises/warmup/echo/echo_test.go` (SKELETON):

```go
package echo

import (
	"testing"
)

// TestEcho is a SKELETON. Cover at least:
//   - Sending 3 values produces 3 outputs in order
//   - Closing input → closing output (the test should range over the
//     output and finish without deadlocking)
//   - Empty input (closed immediately) → empty output (closed immediately)
func TestEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		// TODO: at least 3 cases.
		// {"three", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		// {"empty", []string{}, []string{}},
		// {"one", []string{"only"}, []string{"only"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   in := make(chan string)
			//   out := Echo(in)
			//   go func() {
			//       for _, v := range tc.send { in <- v }
			//       close(in)
			//   }()
			//   var got []string
			//   for v := range out { got = append(got, v) }
			//   compare got to tc.want
			_ = tc
		})
	}
}
```

- [ ] **Step 4:** Create `lessons/16-goroutines-channels/solutions/warmup/echo/echo.go`:

```go
// Package echo is the lesson 16 warm-up reference implementation.
package echo

// Echo returns a new channel that receives everything sent on in.
func Echo(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for v := range in {
			out <- v
		}
		close(out)
	}()
	return out
}
```

- [ ] **Step 5:** Create `lessons/16-goroutines-channels/solutions/warmup/echo/echo_test.go`:

```go
package echo

import (
	"reflect"
	"testing"
)

func TestEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		{"three", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"one", []string{"only"}, []string{"only"}},
		{"empty", []string{}, []string{}},
		{"unicode", []string{"héllo", "wörld"}, []string{"héllo", "wörld"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := make(chan string)
			out := Echo(in)

			// Producer: send all values, then close.
			go func() {
				for _, v := range tc.send {
					in <- v
				}
				close(in)
			}()

			// Consumer: collect everything Echo emits.
			var got []string
			for v := range out {
				got = append(got, v)
			}

			// Normalise empty slice vs nil for comparison.
			if got == nil {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Echo(%v): got %v, want %v", tc.send, got, tc.want)
			}
		})
	}
}

// TestEchoClosesOutput asserts that if the producer never closes the
// input, Echo's goroutine is "leaked" (still running waiting for input)
// — and conversely, if we DO close the input, Echo closes its output.
// This documents the contract: Echo's lifetime is tied to its input.
func TestEchoClosesOutput(t *testing.T) {
	in := make(chan string)
	out := Echo(in)

	close(in)

	// out should close soon after in does. range over out completes
	// without blocking.
	for range out {
		t.Errorf("expected no values from out after closing in")
	}
}
```

> Note on the empty-slice/nil dance in the test: `range out` against an immediately-closed channel produces zero iterations, so `got` stays as a zero-value `[]string` (which is `nil`). `reflect.DeepEqual(nil, []string{})` is false. The `if got == nil { got = []string{} }` normalisation keeps the comparison clean.

> Note on `TestEchoClosesOutput`: this is a separate test specifically documenting the "Echo's lifetime is tied to its input" contract. Important for the next lesson (L17) where students learn about timeouts and partial cancellation; the contract for Echo is "input closes → output closes."

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/16-goroutines-channels/
go test ./lessons/16-goroutines-channels/exercises/warmup/echo/... -v 2>&1 | tail -10
go test ./lessons/16-goroutines-channels/solutions/warmup/echo/... -v 2>&1 | tail -20
go test -race ./lessons/16-goroutines-channels/solutions/warmup/echo/... 2>&1 | tail -5
make test
golangci-lint run ./...
go vet ./...

git add lessons/16-goroutines-channels/exercises/warmup/ lessons/16-goroutines-channels/solutions/warmup/
git commit -m "feat(lesson-16): warmup — echo (goroutine + channel round-trip)"
```

Expected: gofmt empty; exercises pass vacuously; solutions TestEcho (4 sub-tests) + TestEchoClosesOutput PASS; -race clean; make test green; lint 0 issues; vet clean.

---

## Task 3: Main — `internal/logparse` carry-forward + `internal/aggregator` + `cmd/aggregator` (with integration test)

The biggest task. Forward-port logparse from L15, build the aggregator package with Walk, build the cmd binary with full integration test.

**Files (12 total — 6 per side):**

### Step 1: Create directories

```bash
mkdir -p lessons/16-goroutines-channels/{exercises,solutions}/internal/logparse
mkdir -p lessons/16-goroutines-channels/{exercises,solutions}/internal/aggregator
mkdir -p lessons/16-goroutines-channels/{exercises,solutions}/cmd/aggregator
```

### Step 2-3: `internal/logparse/` (verbatim carry-forward from L15)

- [ ] **Step 2:** Read `lessons/15-structure/solutions/internal/logparse/logparse.go` and copy to both `exercises/internal/logparse/logparse.go` and `solutions/internal/logparse/logparse.go`. No content changes needed (the package imports nothing internal).

- [ ] **Step 3:** Same for `logparse_test.go`. Both trees get the same full implementation + tests.

### Step 4-5: `internal/aggregator/` (NEW)

- [ ] **Step 4:** Create `lessons/16-goroutines-channels/exercises/internal/aggregator/aggregator.go`:

```go
// Package aggregator walks a directory of log files in parallel and
// returns the combined level counts.
//
// This is the running example for Phase 3. Lesson 16 builds the basics:
// one goroutine per file, results collected via a channel, single main
// reducer. Future lessons add timeouts (L17), shared-state coordination
// (L18), context cancellation (L19), and a bounded worker pool (L20).
//
// The aggregator pattern — multiple producers, single consumer — is
// the most basic concurrency idiom in Go and the foundation for
// everything that follows.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/exercises/internal/logparse"
)

// fileResult holds one file's parsed level counts plus any error
// encountered while reading or parsing.
//
// Workers send a fileResult on the results channel for each file they
// process. The main goroutine collects N results and reduces.
type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// Walk lists files in dir, spawns one goroutine per file, and returns
// the combined level counts from logparse.Parse across all files.
//
// On any file error (open failure, parse error), Walk returns the
// partial counts collected so far plus a wrapped error mentioning the
// offending file. Same shape as L13's csvimport.Parse.
//
// Hint:
//   1. entries, err := os.ReadDir(dir)
//   2. if err != nil → return nil, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
//   3. files := []string{} ; for each entry, skip dirs, append filepath.Join(dir, name) to files
//   4. resultsCh := make(chan fileResult, len(files))  // buffered so workers don't block
//   5. for each file path: go processFile(path, resultsCh)
//   6. combined := map[string]int{}
//   7. for i := 0; i < len(files); i++ {
//        r := <-resultsCh
//        if r.err != nil → return combined, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
//        for level, count := range r.counts { combined[level] += count }
//      }
//   8. return combined, nil
//
// processFile is a helper: open the file, call logparse.Parse, send
// either {path, counts, nil} or {path, nil, err} on resultsCh.
func Walk(dir string) (map[string]int, error) {
	_ = os.ReadDir
	_ = filepath.Join
	_ = logparse.Parse
	_ = fmt.Errorf
	panic("TODO: read dir, spawn one goroutine per file, collect results, reduce")
}
```

- [ ] **Step 5:** Create `lessons/16-goroutines-channels/exercises/internal/aggregator/aggregator_test.go` (SKELETON):

```go
package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeFile is a test helper. Used inside subtests to create fixture
// log files in a t.TempDir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

// TestWalk is a SKELETON. Cover at least:
//   - Single file with known counts → counts match
//   - Multiple files → merged counts
//   - Empty dir → empty map (or nil — either acceptable)
//   - Malformed log line in one file → wrapped error
func TestWalk(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		// TODO:
		//   dir := t.TempDir()
		//   writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")
		//   got, err := Walk(dir)
		//   if err != nil → t.Fatalf("...")
		//   want := map[string]int{"INFO": 1, "WARN": 1}
		//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
		_ = writeFile
		_ = reflect.DeepEqual
	})

	t.Run("multiple-files", func(t *testing.T) {
		// TODO: write 3 files; assert merged counts.
	})

	t.Run("empty-dir", func(t *testing.T) {
		// TODO: t.TempDir() then Walk; expect empty map + nil error.
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		// TODO: write a file with a garbage line; assert err is non-nil
		// and mentions the file path.
	})
}
```

- [ ] **Step 6:** Create `lessons/16-goroutines-channels/solutions/internal/aggregator/aggregator.go`:

```go
// Package aggregator is the lesson 16 reference implementation.
package aggregator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/solutions/internal/logparse"
)

type fileResult struct {
	path   string
	counts map[string]int
	err    error
}

// Walk walks dir, spawns one goroutine per file, and returns the
// combined level counts. On any error, returns partial counts +
// wrapped error.
func Walk(dir string) (map[string]int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("aggregator: read dir %s: %w", dir, err)
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

	combined := map[string]int{}
	for i := 0; i < len(files); i++ {
		r := <-resultsCh
		if r.err != nil {
			return combined, fmt.Errorf("aggregator: %s: %w", r.path, r.err)
		}
		for level, count := range r.counts {
			combined[level] += count
		}
	}
	return combined, nil
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

> Note on the buffered channel: `make(chan fileResult, len(files))` is sized so workers never block on send (there are exactly N receivers ahead of them). If we used `make(chan fileResult)` (unbuffered), the first N-1 workers would block until the main goroutine picked up their result — still correct, just slightly less parallel. Buffered is the idiomatic choice when N is known in advance.

> Note on early return: if file 3 of 5 errors, Walk returns immediately. The remaining 2 workers continue running in the background but their results are discarded (the channel buffer holds them). This is technically a goroutine "leak" of finite duration — not a real leak, but the kind of edge case L17/L19 address with `select` + cancellation. For L16 (basics) we accept this tradeoff.

- [ ] **Step 7:** Create `lessons/16-goroutines-channels/solutions/internal/aggregator/aggregator_test.go`:

```go
package aggregator

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func TestWalk(t *testing.T) {
	t.Run("single-file", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO ok\n2026-05-21T14:31:00 WARN slow\n")
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1, "WARN": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Walk = %v, want %v", got, want)
		}
	})

	t.Run("multiple-files", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.log", "2026-05-21T14:30:00 INFO a\n2026-05-21T14:31:00 INFO b\n")
		writeFile(t, dir, "b.log", "2026-05-21T14:32:00 WARN c\n")
		writeFile(t, dir, "c.log", "2026-05-21T14:33:00 ERROR d\n2026-05-21T14:34:00 ERROR e\n2026-05-21T14:35:00 ERROR f\n")
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 2, "WARN": 1, "ERROR": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Walk = %v, want %v", got, want)
		}
	})

	t.Run("empty-dir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk on empty dir: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty map, got %v", got)
		}
	})

	t.Run("malformed-line-returns-error", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "good.log", "2026-05-21T14:30:00 INFO ok\n")
		writeFile(t, dir, "bad.log", "garbage not a log line\n")
		_, err := Walk(dir)
		if err == nil {
			t.Fatal("expected error from malformed file, got nil")
		}
		if !strings.Contains(err.Error(), "bad.log") {
			t.Errorf("error should mention bad.log, got %v", err)
		}
	})

	t.Run("nonexistent-dir-returns-error", func(t *testing.T) {
		_, err := Walk("/nonexistent/path/that/does/not/exist")
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
		got, err := Walk(dir)
		if err != nil {
			t.Fatalf("Walk: %v", err)
		}
		want := map[string]int{"INFO": 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("subdirs should be skipped: got %v, want %v", got, want)
		}
	})
}
```

### Step 8-9: `cmd/aggregator/` (NEW)

- [ ] **Step 8:** Create `lessons/16-goroutines-channels/exercises/cmd/aggregator/main.go`:

```go
// Package main is the lesson 16 aggregator CLI.
//
// Usage:
//
//	aggregator -dir=<path>
//
// Walks the directory, parses each log file in parallel, and prints
// the combined level counts to stdout.
//
// Exit codes:
//   - 0 on success
//   - 1 on any error (missing -dir, bad dir, malformed file, etc.).
//     Error message goes to stderr.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/exercises/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. Parses args, calls aggregator.Walk,
// prints results to stdout in alphabetical order.
//
// Hint:
//   1. Parse -dir=<path> from args (required; error if missing).
//   2. counts, err := aggregator.Walk(dir)
//   3. if err != nil → return err
//   4. Sort keys alphabetically.
//   5. For each key: fmt.Fprintf(stdout, "%s: %d\n", key, counts[key])
func run(args []string, stdout io.Writer) error {
	_ = strings.HasPrefix
	_ = aggregator.Walk
	_ = sort.Strings
	panic("TODO: parse -dir, call aggregator.Walk, print sorted results")
}
```

- [ ] **Step 9:** Create `lessons/16-goroutines-channels/exercises/cmd/aggregator/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAggregatorGoldenPath is a SKELETON. The full test exists in the
// solutions tree. Here it stays a stub so the package compiles.
//
// The solution builds the binary into t.TempDir, runs it against a
// fixture directory of log files, and asserts stdout has the expected
// level counts in alphabetical order.
func TestAggregatorGoldenPath(t *testing.T) {
	// TODO:
	//   if _, err := exec.LookPath("go"); err != nil { t.Skip("no go toolchain") }
	//   Build binary; create fixture dir; run; assert stdout.
	_ = bytes.NewReader
	_ = exec.LookPath
	_ = filepath.Join
}
```

- [ ] **Step 10:** Create `lessons/16-goroutines-channels/solutions/cmd/aggregator/main.go`:

```go
// Package main is the lesson 16 aggregator CLI reference implementation.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/16-goroutines-channels/solutions/internal/aggregator"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	dir, err := parseDir(args)
	if err != nil {
		return err
	}

	counts, err := aggregator.Walk(dir)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(stdout, "%s: %d\n", k, counts[k])
	}
	return nil
}

func parseDir(args []string) (string, error) {
	for _, a := range args {
		if strings.HasPrefix(a, "-dir=") {
			return a[len("-dir="):], nil
		}
	}
	return "", fmt.Errorf("-dir=<path> is required")
}
```

- [ ] **Step 11:** Create `lessons/16-goroutines-channels/solutions/cmd/aggregator/main_test.go`:

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
```

### Step 12: Verify and commit

- [ ] **Step 12:** Verify and commit Task 3:

```bash
gofmt -l lessons/16-goroutines-channels/
go test ./lessons/16-goroutines-channels/exercises/... -v 2>&1 | tail -20
go test ./lessons/16-goroutines-channels/solutions/... -v 2>&1 | tail -40
go test -race ./lessons/16-goroutines-channels/solutions/... 2>&1 | tail -5
make test
golangci-lint run ./...
go vet ./...

# Smoke test the binary
TMP=$(mktemp -d /tmp/lesson16-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO test" > "$TMP/logs/x.log"
go run ./lessons/16-goroutines-channels/solutions/cmd/aggregator -dir="$TMP/logs"
rm -rf "$TMP"

git add lessons/16-goroutines-channels/exercises/internal/ lessons/16-goroutines-channels/solutions/internal/ lessons/16-goroutines-channels/exercises/cmd/ lessons/16-goroutines-channels/solutions/cmd/
git commit -m "feat(lesson-16): main — aggregator (Walk with one goroutine per file) + cmd/aggregator"
```

Expected: gofmt empty; exercises pass vacuously; solutions all PASS (6 sub-tests for Walk + 3 for cmd integration); -race clean; smoke shows "INFO: 1" output.

---

## Task 4: Slide deck — 5 concepts

Heavy-explanatory pattern (Motivation/Basics/Worked-example/Common-mistake/Recap) per concept. Plus cover + roadmap + Practice + Closing thought + What-we-learned + Up-next.

**File:** `lessons/16-goroutines-channels/slides/slides.md`

~32 slides. Concept order matches design:

1. **Goroutines** — `go` keyword, no return values, cheap (KB stack), thousands easy. Common-mistake: trying to `result := go f()` (doesn't compile).
2. **Unbuffered channels** — `make(chan T)`, synchronous handoff, the rendezvous model. Common-mistake: send-without-receiver deadlock.
3. **Buffered channels** — `make(chan T, N)`, capacity, when to use which. Common-mistake: thinking "bigger buffer = faster" (not true; it's a backpressure tuning knob).
4. **`range`, `close`, goroutine lifecycle** — `for v := range ch`, `close(ch)`, zero-value-on-closed. Common-mistake: goroutine leaks.
5. **The aggregator pattern** — multiple producers, single consumer. The lesson's running example formalized; named so students recognize it across L17-L20.

- [ ] **Step 1-7:** Author the deck. Cover → roadmap → 5 concepts (5 slides each) → Practice → Closing thought (mention `-race`) → What-we-learned → Up-next (Lesson 17: select & timers).

- [ ] **Step 8:** Verify and commit:

```bash
make slides-build
grep -q "16-goroutines-channels" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/16-goroutines-channels/slides/
git commit -m "feat(lesson-16): slides — Goroutines & channels (5 concepts)"
```

---

## Task 5: README

Mirrors slide concepts (5 sections); one Common-mistake paragraph per concept; gofmt/vet daily-habits block; "Phase 3 begins" header noting the running example is now the log aggregator.

**File:** `lessons/16-goroutines-channels/README.md`

- [ ] **Step 1:** Author the README (~250 lines):
  - "What you'll learn" — 5 bullets.
  - "What's different from L15" — Phase 3 begins. New running example (log aggregator); tracker stays where L15 left it. Race detector introduced as a tool.
  - "The package layout" — annotated tree.
  - Per-concept sections (5).
  - "Exercise: warm-up — echo" (~3 lines).
  - "Exercise: main — aggregator" (~6 lines).
  - "Daily habits" — gofmt -w, go vet, go test. Mention `go test -race ./...` as a new tool worth running occasionally; full `make test-race` integration arrives in L18.
  - "How to run" — both subpackage tests + binary smoke test.
  - "Going further" — Read (Effective Go on goroutines, the Go blog "Share memory by communicating"); Try (run the aggregator on /var/log; benchmark with vs without goroutines; add a -workers=N flag — full worker pool comes in L20 but a sneak preview is fine).

- [ ] **Step 2:** Commit:

```bash
git add lessons/16-goroutines-channels/README.md
git commit -m "docs(lesson-16): README — Goroutines & channels self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1:** Full sweep.

```bash
make test
```

Expected: all lessons (01-16) + tools pass.

- [ ] **Step 2:** Solutions verbose.

```bash
go test -v ./lessons/16-goroutines-channels/solutions/...
```

Expected:
- `echo.TestEcho` — 4 sub-tests PASS
- `echo.TestEchoClosesOutput` PASS
- `logparse.*` — carried L14/L15 tests PASS
- `aggregator.TestWalk` — 6 sub-tests PASS
- `cmd/aggregator.TestAggregatorGoldenPath` PASS
- `cmd/aggregator.TestAggregatorMalformed` PASS
- `cmd/aggregator.TestAggregatorMissingFlag` PASS

Total: ~15+ sub-tests.

- [ ] **Step 3:** Race detector clean.

```bash
go test -race ./lessons/16-goroutines-channels/...
```

Expected: ok (no race warnings).

- [ ] **Step 4:** Exercises pass vacuously.

```bash
go test ./lessons/16-goroutines-channels/exercises/...
```

- [ ] **Step 5:** Static analysis.

```bash
go vet ./...
golangci-lint run ./...
gofmt -l lessons/16-goroutines-channels/
```

Expected: clean.

- [ ] **Step 6:** Slides build.

```bash
make slides-build
grep -q "16-goroutines-channels" dist/index.html
ls dist/lessons/16-goroutines-channels/
rm -rf dist
```

- [ ] **Step 7:** Binary smoke (matches Task 3 step 12 pattern).

```bash
TMP=$(mktemp -d /tmp/lesson16-XXXXXX)
mkdir -p "$TMP/logs"
echo "2026-05-21T14:30:00 INFO server started" > "$TMP/logs/a.log"
echo "2026-05-21T14:31:00 WARN slow query" > "$TMP/logs/b.log"
echo "2026-05-21T14:32:00 ERROR connection refused" > "$TMP/logs/c.log"
go run ./lessons/16-goroutines-channels/solutions/cmd/aggregator -dir="$TMP/logs"
# Expected stdout: "ERROR: 1\nINFO: 1\nWARN: 1\n"
rm -rf "$TMP"
```

- [ ] **Step 8:** Final code review subagent → push branch → create PR.

```bash
git push -u origin feature/plan-s-lesson-16-goroutines-channels
gh pr create --title "feat(lesson-16): Goroutines & channels — basics + log aggregator (Phase 3 begins)" --body "..."
```

---

## Verification

After all 6 tasks:

```bash
make test                                                                          # green
go test -v ./lessons/16-goroutines-channels/solutions/... 2>&1 | grep -E "PASS|FAIL"  # ~15+ PASS lines
go test -race ./lessons/16-goroutines-channels/...                                 # no races
go vet ./...                                                                       # clean
golangci-lint run ./...                                                            # 0 issues
gofmt -l lessons/16-goroutines-channels/                                           # empty
make slides-build && grep -q "16-goroutines-channels" dist/index.html && rm -rf dist  # green
```

## Critical file paths

To be created:

- `lessons/16-goroutines-channels/` (directory)
- `lessons/16-goroutines-channels/README.md`
- `lessons/16-goroutines-channels/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/16-goroutines-channels/exercises/warmup/echo/{echo.go, echo_test.go}`
- `lessons/16-goroutines-channels/exercises/cmd/aggregator/{main.go, main_test.go}`
- `lessons/16-goroutines-channels/exercises/internal/logparse/{logparse.go, logparse_test.go}`
- `lessons/16-goroutines-channels/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}`
- `lessons/16-goroutines-channels/solutions/...` (mirrored)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-22-phase-3-concurrency-design.md` (Phase 3 design)
- `lessons/15-structure/solutions/internal/logparse/{logparse.go, logparse_test.go}` (carry-forward source)
- `tools/build-index/main.go` (verify L16 slug matches `16-goroutines-channels`)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
