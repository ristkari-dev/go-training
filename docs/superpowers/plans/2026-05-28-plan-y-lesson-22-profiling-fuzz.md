# Plan Y — Lesson 22 (Profiling, benchmarking & fuzzing) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 22 — the Phase 3 finale. Students profile the L20 log aggregator, find the regex hotspot, replace it with a hand-written parser (a verified 5.4× speedup, 2 allocs → 0), then fuzz the new parser to catch the edge-case bug the optimization introduced — tying benchmarking, profiling, and fuzzing into one honest narrative.

**Architecture:** Same per-lesson pattern as Plans D-X. Six tasks. Carries forward L20's full aggregator stack (`logparse` + `aggregator` + `errgroupx` + `cmd/aggregator`) verbatim with import paths rewritten, then adds the optimization + measurement surface. Skeleton tests in warmup + the new code (Phase 3 invariant). Five-concept heavy-explanatory slide deck.

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `context`, `flag`, `fmt`, `io`, `net/http`, `net/http/pprof`, `os`, `path/filepath`, `regexp`, `runtime`, `runtime/pprof`, `sort`, `strconv`, `strings`, `sync`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan Y: lesson 22 complete; `make test` + `make test-race` green; benchmarks runnable; fuzz target ships with regression seeds; `cmd/aggregator-profile` writes CPU + heap profiles and serves a live pprof endpoint; slug `22-profiling-fuzz` lands in the index.

### Design decisions (3 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **Five slide concepts**: benchmarking · CPU profiling · heap/alloc profiling + escape analysis · `net/http/pprof` · fuzzing.
2. **Full integrated optimization+fuzz arc**: profile → regex `FindStringSubmatch` is the hotspot → replace `parseLine` with a hand-written manual parser → benchmark proves the speedup → fuzz finds the naive parser's panic → fix it.
3. **Slug `22-profiling-fuzz`**: requires updating `tools/build-index/main.go` (slug `profiling` → `profiling-fuzz`; Title → "Profiling & fuzzing"; Blurb → "pprof · bench · fuzz").

**Plan-recommended:**

4. **Carry forward L20's full stack** (`internal/{logparse,aggregator,errgroupx}` + `cmd/aggregator`) with import paths rewritten `20-concurrency-patterns` → `22-profiling-fuzz`. The aggregator/errgroupx/cmd are verbatim; `logparse` is where the optimization happens.

5. **Keep the regex parser** (`parseLineRegex` + `logLineRE`) in the solution's `logparse.go` as the fuzz/benchmark **oracle**. `Parse` switches to `parseLineFast`. `unused` counts same-package test usage, so the regex stays lint-clean. (Contingency: if `unused` flags it, move the regex oracle into a `_test.go` file in `package logparse`.)

6. **Robustness + one-directional oracle fuzzing.** `FuzzParseLineFast` asserts (a) the fast parser never panics, and (b) **whenever the regex reference accepts, the fast parser accepts identically**. The reverse is intentionally NOT required — the manual parser may be marginally more lenient on exotic whitespace, and the regex's `.`/`$`/`\s` semantics are fiddly to mirror exactly. This one-directional oracle survived 2M+ random execs with zero divergence during plan verification. `go test` (no `-fuzz`) only runs the committed `f.Add` seeds, which are verified.

7. **`cmd/aggregator-profile` is provided working code** (students study it, like L10's CLI). Generates a synthetic dataset (or `-dir`), runs `WalkPool`, writes `-cpuprofile`/`-memprofile` via `runtime/pprof`, optionally serves `-httppprof=addr` (blank-imports `net/http/pprof`).

8. **`f.Add` seeds, not a `testdata/fuzz` corpus dir.** The formerly-crashing inputs (`"0"`, `""`, `"INFO"`) become permanent in-code seeds — simpler to ship, same regression value.

### Verified facts (prototyped before writing this plan)

- Manual parser matches the regex oracle on the full table + 2M+ random fuzz execs (one-directional oracle). Zero divergence.
- Benchmark on Apple M1 Max: `parseLineRegex` = 698.7 ns/op, 128 B/op, **2 allocs/op**; `parseLineFast` = 130.1 ns/op, **0 B/op, 0 allocs/op**. ~5.4× faster, allocation-free.
- The naive parser (`strings.IndexByte(s, ' ')` then `s[:i]` without the `-1` guard) panics on no-space input; `go test -fuzz` finds the crasher `"0"` in well under a second.
- `.golangci.yml` enables errcheck/govet/ineffassign/staticcheck/unused. **gosec is NOT enabled** → no G108 concern for the pprof endpoint. errcheck is excluded for all `lessons/.*/(exercises|solutions)/` → ignored `http.ListenAndServe` error is fine. staticcheck+unused excluded for `lessons/.*/exercises/`.

---

## Plans F-X lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages (`lessons/.*/exercises/`).
3. Common-mistake content in README + slides (one per concept).
4. Slides + README written inline by the controller.
5. `make test-race` daily-habits continues from L18.
6. Skeleton tests pass vacuously (empty case slice; `panic("TODO")` body; benchmark/fuzz stubs that no-op).
7. Timing-sensitive tests use generous deadlines / invariant-only assertions (CI-flake lessons from L17-L20). The fuzz target under `go test` runs only fixed seeds — deterministic.

---

## File structure

```
lessons/22-profiling-fuzz/
├── README.md                                          (Task 6)
├── slides/
│   ├── index.html (scaffolded), slides.md, assets/.gitkeep   (Task 5)
├── exercises/
│   ├── warmup/bench/
│   │   ├── bench.go                                   (Task 2 — GenLines)
│   │   └── bench_test.go                              (Task 2 — SKELETON + BenchmarkParse)
│   ├── cmd/
│   │   ├── aggregator/{main.go, main_test.go}         (Task 1 — carried verbatim)
│   │   └── aggregator-profile/
│   │       ├── main.go                                (Task 4 — SKELETON)
│   │       └── main_test.go                           (Task 4 — SKELETON)
│   └── internal/
│       ├── logparse/
│       │   ├── logparse.go                            (Task 1 carry → Task 3 adds parseLineFast SKELETON)
│       │   ├── logparse_test.go                       (Task 1 — carried verbatim)
│       │   ├── logparse_bench_test.go                 (Task 3 — SKELETON)
│       │   └── fuzz_test.go                           (Task 3 — SKELETON)
│       ├── aggregator/{aggregator.go, aggregator_test.go}   (Task 1 — carried verbatim)
│       └── errgroupx/{errgroupx.go, errgroupx_test.go}      (Task 1 — carried verbatim)
└── solutions/   (mirrored, with full implementations)
```

**File count:** ~32 (biggest lesson yet; most is verbatim carry-forward).

---

## Task 1: Scaffold + carry forward the L20 aggregator stack

Establish a green, carried-forward baseline before any new work.

- [ ] **Step 1:** Scaffold and remove the flat stubs:

```bash
make new-lesson NAME=22-profiling-fuzz
rm lessons/22-profiling-fuzz/exercises/main.go \
   lessons/22-profiling-fuzz/exercises/main_test.go \
   lessons/22-profiling-fuzz/exercises/warmup.go \
   lessons/22-profiling-fuzz/exercises/warmup_test.go \
   lessons/22-profiling-fuzz/solutions/main.go \
   lessons/22-profiling-fuzz/solutions/main_test.go \
   lessons/22-profiling-fuzz/solutions/warmup.go \
   lessons/22-profiling-fuzz/solutions/warmup_test.go
```

- [ ] **Step 2:** Copy the L20 carry-forward trees (both exercises + solutions), then rewrite import paths:

```bash
SRC=lessons/20-concurrency-patterns
DST=lessons/22-profiling-fuzz
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/cmd"
  cp -R "$SRC/$side/internal/logparse"    "$DST/$side/internal/logparse"
  cp -R "$SRC/$side/internal/aggregator"  "$DST/$side/internal/aggregator"
  cp -R "$SRC/$side/internal/errgroupx"   "$DST/$side/internal/errgroupx"
  cp -R "$SRC/$side/cmd/aggregator"       "$DST/$side/cmd/aggregator"
done
# Rewrite module import paths 20-concurrency-patterns → 22-profiling-fuzz.
grep -rl '20-concurrency-patterns' "$DST" | while read -r f; do
  sed -i '' 's#lessons/20-concurrency-patterns#lessons/22-profiling-fuzz#g' "$f"
done
```

> Note: the `sed -i ''` form is the macOS/BSD in-place syntax (the dev machine is darwin). The course module path is `github.com/ristkari-dev/go-training`; only the lesson segment changes.

- [ ] **Step 3:** The L20 `fanin` warmup is NOT carried (L22 has its own `bench` warmup). Confirm no stray `warmup/fanin` was copied (we only copied `internal/` + `cmd/aggregator`).

- [ ] **Step 4:** Verify the carried-forward baseline builds + tests pass:

```bash
gofmt -l lessons/22-profiling-fuzz/
go build ./lessons/22-profiling-fuzz/...
go test ./lessons/22-profiling-fuzz/... 2>&1 | tail -20
go test -race ./lessons/22-profiling-fuzz/solutions/... 2>&1 | tail -5
go vet ./lessons/22-profiling-fuzz/...
```

Expected: gofmt empty; build clean; all carried tests pass (exercises + solutions); -race clean; vet clean. The doc comments still say "lesson 20" — that's cosmetic and acceptable for carried-forward files; the package behavior and import paths are correct.

- [ ] **Step 5:** Commit:

```bash
git add lessons/22-profiling-fuzz/exercises lessons/22-profiling-fuzz/solutions
git commit -m "chore(lesson-22): scaffold + carry forward L20 aggregator stack

cp logparse/aggregator/errgroupx/cmd-aggregator from L20, import
paths rewritten 20-concurrency-patterns → 22-profiling-fuzz. Green
baseline before the profiling/optimization work."
```

---

## Task 2: Warm-up — `bench` (GenLines + BenchmarkParse)

A profiling lesson's warm-up is to write a benchmark. `GenLines` is the deterministic synthetic-data generator (unit-testable); `BenchmarkParse` measures the carried-forward `logparse.Parse` baseline.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/22-profiling-fuzz/{exercises,solutions}/warmup/bench`

- [ ] **Step 2:** Create `lessons/22-profiling-fuzz/exercises/warmup/bench/bench.go`:

```go
// Package bench is the lesson 22 warm-up: a deterministic synthetic
// log-line generator used to feed benchmarks.
//
// Benchmarks need a fixed, repeatable input so ns/op is comparable
// across runs. GenLines builds n well-formed log lines that
// logparse.Parse accepts — no randomness, no I/O.
package bench

import (
	"fmt"
	"strings"
)

// levels cycles through the three valid log levels.
var levels = []string{"INFO", "WARN", "ERROR"}

// GenLines returns n synthetic log lines joined by '\n' (one trailing
// newline-free block suitable for strings.NewReader). Each line is
// well-formed: "2006-01-02T15:04:SS LEVEL message ...".
//
// Hint:
//   var b strings.Builder
//   for i := 0; i < n; i++ {
//       fmt.Fprintf(&b, "2026-01-02T15:04:%02d %s message number %d\n",
//           i%60, levels[i%3], i)
//   }
//   return b.String()
func GenLines(n int) string {
	_ = fmt.Fprintf
	_ = strings.Builder{}
	_ = levels
	panic("TODO: build n well-formed log lines into a strings.Builder; return the string")
}
```

- [ ] **Step 3:** Create `lessons/22-profiling-fuzz/exercises/warmup/bench/bench_test.go` (SKELETON):

```go
package bench

import (
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/exercises/internal/logparse"
)

// TestGenLines is a SKELETON. Verify GenLines(n) produces n parseable
// lines.
func TestGenLines(t *testing.T) {
	cases := []int{
		// TODO: e.g. 0, 1, 10, 100
	}
	for _, n := range cases {
		got := strings.Count(GenLines(n), "\n")
		if got != n {
			t.Errorf("GenLines(%d) has %d newlines, want %d", n, got, n)
		}
	}
}

// BenchmarkParse measures the baseline logparse.Parse throughput over
// a fixed synthetic input. Run with: go test -bench=. -benchmem
func BenchmarkParse(b *testing.B) {
	// TODO:
	//   data := GenLines(1000)
	//   b.ReportAllocs()
	//   b.ResetTimer()
	//   for i := 0; i < b.N; i++ {
	//       _, _ = logparse.Parse(strings.NewReader(data))
	//   }
	_ = logparse.Parse
}
```

- [ ] **Step 4:** Create `lessons/22-profiling-fuzz/solutions/warmup/bench/bench.go`:

```go
// Package bench is the lesson 22 warm-up reference implementation.
package bench

import (
	"fmt"
	"strings"
)

var levels = []string{"INFO", "WARN", "ERROR"}

// GenLines returns n deterministic, well-formed log lines.
func GenLines(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "2026-01-02T15:04:%02d %s message number %d\n", i%60, levels[i%3], i)
	}
	return b.String()
}
```

- [ ] **Step 5:** Create `lessons/22-profiling-fuzz/solutions/warmup/bench/bench_test.go`:

```go
package bench

import (
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/internal/logparse"
)

func TestGenLines(t *testing.T) {
	for _, n := range []int{0, 1, 10, 100} {
		got := strings.Count(GenLines(n), "\n")
		if got != n {
			t.Errorf("GenLines(%d) has %d newlines, want %d", n, got, n)
		}
	}
}

func TestGenLinesParseable(t *testing.T) {
	entries, err := logparse.Parse(strings.NewReader(GenLines(50)))
	if err != nil {
		t.Fatalf("Parse(GenLines(50)) error: %v", err)
	}
	if len(entries) != 50 {
		t.Errorf("parsed %d entries, want 50", len(entries))
	}
}

// BenchmarkParse measures baseline logparse.Parse throughput.
// Run with: go test -bench=. -benchmem ./.../warmup/bench/
func BenchmarkParse(b *testing.B) {
	data := GenLines(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logparse.Parse(strings.NewReader(data))
	}
}
```

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/22-profiling-fuzz/
go test ./lessons/22-profiling-fuzz/exercises/warmup/bench/... 2>&1 | tail -5
go test -v ./lessons/22-profiling-fuzz/solutions/warmup/bench/... 2>&1 | tail -15
go test -bench=. -benchmem ./lessons/22-profiling-fuzz/solutions/warmup/bench/... 2>&1 | tail -8
make test
go vet ./lessons/22-profiling-fuzz/...

git add lessons/22-profiling-fuzz/exercises/warmup lessons/22-profiling-fuzz/solutions/warmup
git commit -m "feat(lesson-22): warmup — bench (GenLines synthetic data + BenchmarkParse baseline)"
```

Expected: exercises pass vacuously; solutions tests PASS; `BenchmarkParse` prints ns/op + B/op + allocs/op.

---

## Task 3: logparse optimization — `parseLineFast` + benchmark + fuzz

The heart of the lesson. Add the hand-written parser, switch `Parse` to use it, add the comparison benchmark, and add the fuzz target with regression seeds.

**Files:** modify `internal/logparse/logparse.go` (both trees); create `logparse_bench_test.go` + `fuzz_test.go` (both trees).

### 3a — exercises `logparse.go` (add `parseLineFast` SKELETON; `Parse` keeps using the regex parser so carried tests stay green)

- [ ] **Step 1:** Edit `lessons/22-profiling-fuzz/exercises/internal/logparse/logparse.go`. The carried-forward file currently has `Parse`, `parseLine` (regex), `CountByLevel`. Make these edits:
  1. Rename `parseLine` → `parseLineRegex` (the reference/oracle). Update `Parse` to call `parseLineRegex` for now.
  2. Add the `parseLineFast` skeleton + helpers below.

The resulting exercises `logparse.go`:

```go
// Package logparse is the lesson 22 reference implementation.
//
// L22 optimization arc: parseLineRegex (carried from L14/L20) is the
// reference. You implement parseLineFast — a hand-written parser that
// avoids the per-line regexp allocation. Profile first, then optimize.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

var logLineRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR)\s+(.+)$`)

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r and returns parsed entries.
func Parse(r io.Reader) ([]LogEntry, error) {
	s := bufio.NewScanner(r)
	out := []LogEntry{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		// TODO (L22): once parseLineFast is implemented and benchmarked,
		// switch this call to parseLineFast.
		e, err := parseLineRegex(text)
		if err != nil {
			return out, fmt.Errorf("logparse: line %d: %w", line, err)
		}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return out, fmt.Errorf("logparse: %w", err)
	}
	return out, nil
}

// parseLineRegex is the carried-forward reference parser. Kept as the
// fuzz/benchmark oracle even after Parse switches to parseLineFast.
func parseLineRegex(s string) (LogEntry, error) {
	m := logLineRE.FindStringSubmatch(s)
	if m == nil {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	t, err := time.Parse(timeLayout, m[1])
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", m[1], err)
	}
	return LogEntry{Time: t, Level: m[2], Message: m[3]}, nil
}

// parseLineFast is the hand-written parser you implement. No regexp,
// no per-call slice allocation.
//
// The format is fixed: "2006-01-02T15:04:05 LEVEL message".
//   - timestamp is exactly 19 chars
//   - one or more whitespace (\s = [\t\n\f\r ])
//   - level is INFO|WARN|ERROR
//   - one or more whitespace
//   - non-empty message containing no '\n' (regex '.' excludes newline)
//
// CRITICAL: guard every index. strings.IndexByte returns -1 when the
// byte is absent; slicing s[:-1] PANICS. Fuzzing will find this if you
// forget (try seed "0"). Return an error, never panic.
//
// Hint:
//   const tsLen = 19
//   if len(s) < tsLen { return LogEntry{}, errMalformed(s) }
//   tsStr, rest := s[:tsLen], s[tsLen:]
//   rest, ok := cutLeadingSpace(rest); if !ok { return ... }
//   find next \s in rest → sp; if sp < 0 { return ... }
//   level := rest[:sp]; validate INFO|WARN|ERROR
//   msg, ok := cutLeadingSpace(rest[sp:]); if !ok || msg == "" { return ... }
//   if strings.IndexByte(msg, '\n') >= 0 { return ... }   // '.' excludes '\n'
//   t, err := time.Parse(timeLayout, tsStr); ...
func parseLineFast(s string) (LogEntry, error) {
	_ = cutLeadingSpace
	_ = isLogSpace
	panic("TODO: hand-written parser; guard every index; match parseLineRegex when it accepts")
}

// isLogSpace matches RE2's \s class: [\t\n\f\r ].
func isLogSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\f', '\r':
		return true
	}
	return false
}

// cutLeadingSpace strips a run of >=1 \s chars from the front. ok is
// false when there was no leading \s char at all (mirrors regex \s+).
func cutLeadingSpace(s string) (rest string, ok bool) {
	i := 0
	for i < len(s) && isLogSpace(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// CountByLevel tallies entries per level.
func CountByLevel(entries []LogEntry) map[string]int {
	out := map[string]int{}
	for _, e := range entries {
		out[e.Level]++
	}
	return out
}
```

> Note: in the exercises tree, `parseLineFast` panics (TODO) and is referenced only by the skeleton `fuzz_test.go`/`logparse_bench_test.go` (which no-op). `Parse` uses `parseLineRegex`, so all carried-forward aggregator tests stay green. The `lessons/.*/exercises/` lint exclusion silences `unused`/staticcheck on the stubs.

### 3b — exercises `logparse_bench_test.go` (SKELETON)

- [ ] **Step 2:** Create `lessons/22-profiling-fuzz/exercises/internal/logparse/logparse_bench_test.go`:

```go
package logparse

import "testing"

// sink prevents the compiler from eliminating the parse result as dead
// code (a classic benchmarking mistake).
var sink LogEntry

// BenchmarkParseLineRegex / BenchmarkParseLineFast compare the two
// parsers. Run: go test -bench=ParseLine -benchmem
//
// TODO: build a fixed line, loop b.N times calling each parser, assign
// to sink, ReportAllocs. See the solution for the shape.
func BenchmarkParseLineRegex(b *testing.B) {
	_ = sink
	b.Skip("TODO: implement BenchmarkParseLineRegex")
}

func BenchmarkParseLineFast(b *testing.B) {
	b.Skip("TODO: implement BenchmarkParseLineFast")
}
```

### 3c — exercises `fuzz_test.go` (SKELETON)

- [ ] **Step 3:** Create `lessons/22-profiling-fuzz/exercises/internal/logparse/fuzz_test.go`:

```go
package logparse

import "testing"

// FuzzParseLineFast is a SKELETON. The shipped target asserts:
//   (1) parseLineFast never panics, and
//   (2) whenever parseLineRegex ACCEPTS, parseLineFast agrees identically.
//
// Run with: go test -fuzz=FuzzParseLineFast
//
// TODO: seed with well-formed lines AND the no-space crashers ("0", "",
// "INFO"); in f.Fuzz compare parseLineFast vs parseLineRegex one-way.
func FuzzParseLineFast(f *testing.F) {
	f.Add("2026-01-02T15:04:05 INFO ok")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = parseLineFast(s)
		_ = parseLineRegex
	})
}
```

### 3d — solutions `logparse.go` (implement `parseLineFast`; `Parse` uses it)

- [ ] **Step 4:** Edit `lessons/22-profiling-fuzz/solutions/internal/logparse/logparse.go` to the full reference. Same structure as 3a but `parseLineFast` is implemented and `Parse` calls it:

```go
// Package logparse is the lesson 22 reference implementation.
//
// parseLineRegex (carried from L14/L20) is kept as the fuzz/benchmark
// oracle; Parse now uses the allocation-free parseLineFast.
package logparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

var logLineRE = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+(INFO|WARN|ERROR)\s+(.+)$`)

const timeLayout = "2006-01-02T15:04:05"

// Parse reads log lines from r and returns parsed entries.
func Parse(r io.Reader) ([]LogEntry, error) {
	s := bufio.NewScanner(r)
	out := []LogEntry{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		e, err := parseLineFast(text)
		if err != nil {
			return out, fmt.Errorf("logparse: line %d: %w", line, err)
		}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return out, fmt.Errorf("logparse: %w", err)
	}
	return out, nil
}

// parseLineRegex is the carried-forward reference parser, retained as
// the fuzz/benchmark oracle.
func parseLineRegex(s string) (LogEntry, error) {
	m := logLineRE.FindStringSubmatch(s)
	if m == nil {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	t, err := time.Parse(timeLayout, m[1])
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", m[1], err)
	}
	return LogEntry{Time: t, Level: m[2], Message: m[3]}, nil
}

// parseLineFast is the allocation-free hand-written parser. It matches
// parseLineRegex whenever the regex accepts; it never panics.
func parseLineFast(s string) (LogEntry, error) {
	const tsLen = 19 // 2006-01-02T15:04:05
	if len(s) < tsLen {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	tsStr := s[:tsLen]
	rest := s[tsLen:]

	// \s+ between timestamp and level.
	rest, ok := cutLeadingSpace(rest)
	if !ok {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	// level runs up to the next \s char.
	sp := -1
	for i := 0; i < len(rest); i++ {
		if isLogSpace(rest[i]) {
			sp = i
			break
		}
	}
	if sp < 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	level := rest[:sp]
	if level != "INFO" && level != "WARN" && level != "ERROR" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	// \s+ between level and message; message non-empty, no '\n'.
	msg, ok := cutLeadingSpace(rest[sp:])
	if !ok || msg == "" {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}
	if strings.IndexByte(msg, '\n') >= 0 {
		return LogEntry{}, fmt.Errorf("malformed log line %q", s)
	}

	t, err := time.Parse(timeLayout, tsStr)
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", tsStr, err)
	}
	return LogEntry{Time: t, Level: level, Message: msg}, nil
}

// isLogSpace matches RE2's \s class: [\t\n\f\r ].
func isLogSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\f', '\r':
		return true
	}
	return false
}

// cutLeadingSpace strips a run of >=1 \s chars from the front. ok is
// false when there was no leading \s char at all (mirrors regex \s+).
func cutLeadingSpace(s string) (rest string, ok bool) {
	i := 0
	for i < len(s) && isLogSpace(s[i]) {
		i++
	}
	return s[i:], i > 0
}

// CountByLevel tallies entries per level.
func CountByLevel(entries []LogEntry) map[string]int {
	out := map[string]int{}
	for _, e := range entries {
		out[e.Level]++
	}
	return out
}
```

### 3e — solutions `logparse_bench_test.go`

- [ ] **Step 5:** Create `lessons/22-profiling-fuzz/solutions/internal/logparse/logparse_bench_test.go`:

```go
package logparse

import "testing"

// sink prevents dead-code elimination of the parse result.
var sink LogEntry

const benchLine = "2026-01-02T15:04:05 INFO message number 42 here"

func BenchmarkParseLineRegex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, _ := parseLineRegex(benchLine)
		sink = e
	}
}

func BenchmarkParseLineFast(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e, _ := parseLineFast(benchLine)
		sink = e
	}
}
```

### 3f — solutions `fuzz_test.go` + table test

- [ ] **Step 6:** Create `lessons/22-profiling-fuzz/solutions/internal/logparse/fuzz_test.go`:

```go
package logparse

import "testing"

// TestParseLineFastMatchesRegex checks the fast parser agrees with the
// regex reference across well-formed and malformed inputs.
func TestParseLineFastMatchesRegex(t *testing.T) {
	cases := []string{
		"2026-01-02T15:04:05 INFO server started",
		"2026-01-02T15:04:05 WARN slow query",
		"2026-01-02T15:04:05 ERROR boom",
		"2026-01-02T15:04:05   INFO   extra spaces",
		"2026-01-02T15:04:05\tINFO\ttabs",
		"2026-01-02T15:04:05 INFO trailing  ",
		"2026-13-02T15:04:05 INFO bad month",
		"not a log line",
		"",
		"INFO",
		"2026-01-02T15:04:05 DEBUG unknown level",
		"2026-01-02T15:04:05 INFO ",
	}
	for _, s := range cases {
		want, errRe := parseLineRegex(s)
		got, errFast := parseLineFast(s)
		if errRe == nil {
			if errFast != nil {
				t.Errorf("regex accepted %q but fast rejected: %v", s, errFast)
				continue
			}
			if got != want {
				t.Errorf("mismatch for %q: fast=%+v regex=%+v", s, got, want)
			}
		}
	}
}

// FuzzParseLineFast asserts (1) parseLineFast never panics, and
// (2) whenever parseLineRegex accepts, parseLineFast agrees identically.
//
// The seeds include the no-space inputs ("0", "", "INFO") that crash a
// naive parser — permanent regression cases. Run the random fuzzer with:
//   go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=30s
func FuzzParseLineFast(f *testing.F) {
	seeds := []string{
		"2026-01-02T15:04:05 INFO ok",
		"2026-01-02T15:04:05 WARN w",
		"2026-01-02T15:04:05 ERROR e",
		"2026-01-02T15:04:05   INFO   spaces",
		"",
		"0",
		"INFO",
		"x",
		"2026-01-02T15:04:05",
		"2026-01-02T15:04:05 ",
		"2026-01-02T15:04:05 INFO ",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, errFast := parseLineFast(s) // must never panic
		want, errRe := parseLineRegex(s)
		if errRe == nil {
			if errFast != nil {
				t.Errorf("regex accepted %q but fast rejected: %v", s, errFast)
			} else if got != want {
				t.Errorf("mismatch for %q: fast=%+v regex=%+v", s, got, want)
			}
		}
	})
}
```

- [ ] **Step 7:** Verify and commit:

```bash
gofmt -l lessons/22-profiling-fuzz/
go test ./lessons/22-profiling-fuzz/exercises/internal/logparse/... 2>&1 | tail -5
go test -v ./lessons/22-profiling-fuzz/solutions/internal/logparse/... 2>&1 | tail -20
go test -bench=ParseLine -benchmem ./lessons/22-profiling-fuzz/solutions/internal/logparse/... 2>&1 | tail -8
# Random fuzz smoke (10s) — expect PASS, no crasher:
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=10s ./lessons/22-profiling-fuzz/solutions/internal/logparse/... 2>&1 | tail -8
# Clean up any fuzz cache the smoke run created (we ship f.Add seeds only):
git status --porcelain lessons/22-profiling-fuzz/solutions/internal/logparse/testdata 2>/dev/null
make test
make test-race
go vet ./lessons/22-profiling-fuzz/...
golangci-lint run ./lessons/22-profiling-fuzz/...

git add lessons/22-profiling-fuzz/exercises/internal/logparse lessons/22-profiling-fuzz/solutions/internal/logparse
git commit -m "feat(lesson-22): logparse — parseLineFast (alloc-free) + benchmark + oracle fuzz"
```

Expected: exercises vacuous-pass; solutions table + fuzz-seeds PASS; benchmark shows fast ≪ regex with 0 allocs; 10s random fuzz finds nothing. If the random-fuzz smoke writes a `testdata/fuzz/` corpus entry, do NOT commit it (we ship `f.Add` seeds only) — `git checkout`/`rm` it before staging.

> Contingency (unused): if `golangci-lint` flags `parseLineRegex`/`logLineRE` as unused in the solution despite test usage, move them into `fuzz_test.go` (package logparse) as the test-only oracle and delete from `logparse.go`.

---

## Task 4: `cmd/aggregator-profile`

Provided working code (students study + run it). Generates a synthetic dataset (or `-dir`), runs `WalkPool` under `runtime/pprof` CPU + heap collection, optionally serves a live `net/http/pprof` endpoint.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/22-profiling-fuzz/{exercises,solutions}/cmd/aggregator-profile`

- [ ] **Step 2:** Create `lessons/22-profiling-fuzz/exercises/cmd/aggregator-profile/main.go` (SKELETON):

```go
// Package main is the lesson 22 aggregator profiler.
//
// Runs WalkPool under load and writes CPU + heap profiles for analysis
// with `go tool pprof`. Optionally serves a live pprof endpoint.
//
// Usage:
//   aggregator-profile -n=2000 -cpuprofile=cpu.prof -memprofile=heap.prof
//   aggregator-profile -dir=/path/to/logs -duration=5s
//   aggregator-profile -httppprof=:6060 -duration=30s   # live: localhost:6060/debug/pprof/
//
// Then: go tool pprof -http=:8080 cpu.prof
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the testable entry point. See the solution for the full
// implementation (flag parsing, synthetic dataset, pprof start/stop,
// WalkPool loop, heap dump).
func run(args []string, stdout, stderr io.Writer) int {
	_ = args
	_ = stdout
	_ = stderr
	panic("TODO: parse flags; build/resolve dataset; StartCPUProfile; loop WalkPool until -duration; WriteHeapProfile")
}
```

- [ ] **Step 3:** Create `lessons/22-profiling-fuzz/exercises/cmd/aggregator-profile/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"testing"
)

// TestRunWritesProfiles is a SKELETON. Run with a tiny synthetic
// dataset + temp profile paths; assert exit 0 and non-empty profiles.
func TestRunWritesProfiles(t *testing.T) {
	// TODO:
	//   dir := t.TempDir()
	//   cpu := filepath.Join(dir, "cpu.prof"); mem := filepath.Join(dir, "heap.prof")
	//   code := run([]string{"-n=20", "-lines=10", "-cpuprofile="+cpu, "-memprofile="+mem}, &out, &errBuf)
	//   assert code == 0; assert both files exist + non-empty
	_ = bytes.NewBuffer
}
```

- [ ] **Step 4:** Create `lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile/main.go`:

```go
// Package main is the lesson 22 aggregator profiler reference impl.
//
// Runs WalkPool under load and writes CPU + heap profiles for analysis
// with `go tool pprof`. Optionally serves a live pprof endpoint via the
// net/http/pprof side-effect import.
//
// Usage:
//   aggregator-profile -n=2000 -cpuprofile=cpu.prof -memprofile=heap.prof
//   aggregator-profile -dir=/path/to/logs -duration=5s
//   aggregator-profile -httppprof=:6060 -duration=30s
//
// Then: go tool pprof -http=:8080 cpu.prof
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	_ "net/http/pprof" // registers /debug/pprof handlers on DefaultServeMux

	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/internal/aggregator"
	"github.com/ristkari-dev/go-training/lessons/22-profiling-fuzz/solutions/warmup/bench"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aggregator-profile", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "directory of log files (empty → generate synthetic)")
	n := fs.Int("n", 2000, "synthetic file count when -dir is empty")
	lines := fs.Int("lines", 50, "lines per synthetic file")
	workers := fs.Int("workers", 0, "worker count (0 = NumCPU)")
	duration := fs.Duration("duration", 0, "loop WalkPool until this elapses (0 = single pass)")
	cpuProfile := fs.String("cpuprofile", "", "write CPU profile to this file")
	memProfile := fs.String("memprofile", "", "write heap profile to this file")
	httpAddr := fs.String("httppprof", "", "serve live pprof on this addr (e.g. :6060)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Optional live pprof endpoint. The blank import registered the
	// handlers; we just need a server on DefaultServeMux.
	if *httpAddr != "" {
		go func() {
			fmt.Fprintf(stderr, "live pprof: http://%s/debug/pprof/\n", *httpAddr)
			if err := http.ListenAndServe(*httpAddr, nil); err != nil {
				fmt.Fprintln(stderr, "pprof http:", err)
			}
		}()
	}

	// Resolve the dataset directory.
	target := *dir
	if target == "" {
		d, err := generateDataset(*n, *lines)
		if err != nil {
			fmt.Fprintln(stderr, "generate dataset:", err)
			return 1
		}
		defer os.RemoveAll(d)
		target = d
	}

	// CPU profile spans the whole workload.
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Fprintln(stderr, "create cpuprofile:", err)
			return 1
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(stderr, "start cpu profile:", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	// Run WalkPool once, or loop until -duration elapses (so CPU
	// profiling gathers enough samples).
	var (
		result aggregator.WalkResult
		err    error
		passes int
		start  = time.Now()
	)
	for {
		result, err = aggregator.WalkPool(context.Background(), target, 0, *workers)
		if err != nil {
			fmt.Fprintln(stderr, "walk:", err)
			return 1
		}
		passes++
		if *duration <= 0 || time.Since(start) >= *duration {
			break
		}
	}

	// Heap profile is a snapshot — take it after the workload, post-GC.
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fmt.Fprintln(stderr, "create memprofile:", err)
			return 1
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintln(stderr, "write heap profile:", err)
			return 1
		}
	}

	total := 0
	for _, c := range result.Counts {
		total += c
	}
	fmt.Fprintf(stdout, "passes=%d files=%d entries=%d timedOut=%d\n",
		passes, *n, total, len(result.TimedOut))
	return 0
}

// generateDataset writes n files of `lines` well-formed log lines each
// into a fresh temp dir and returns its path. Caller removes it.
func generateDataset(n, lines int) (string, error) {
	dir, err := os.MkdirTemp("", "aggprofile-")
	if err != nil {
		return "", err
	}
	block := bench.GenLines(lines)
	for i := 0; i < n; i++ {
		name := filepath.Join(dir, fmt.Sprintf("log-%05d.log", i))
		if err := os.WriteFile(name, []byte(block), 0o644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}
```

> Note: `-dir` empty for tests is fine; the test passes a small `-n`. `bench.GenLines` is the warm-up helper — `cmd/aggregator-profile` reuses it for the synthetic dataset (one more reason the warm-up exists). errcheck is excluded for lesson code, so the deferred `f.Close()` / `os.RemoveAll` ignores are clean.

- [ ] **Step 5:** Create `lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile/main_test.go`:

```go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesProfiles(t *testing.T) {
	tmp := t.TempDir()
	cpu := filepath.Join(tmp, "cpu.prof")
	mem := filepath.Join(tmp, "heap.prof")

	var out, errBuf bytes.Buffer
	code := run([]string{
		"-n=20", "-lines=10",
		"-cpuprofile=" + cpu,
		"-memprofile=" + mem,
	}, &out, &errBuf)

	if code != 0 {
		t.Fatalf("run exit code = %d, stderr=%q", code, errBuf.String())
	}
	for _, p := range []string{cpu, mem} {
		fi, err := os.Stat(p)
		if err != nil {
			t.Errorf("profile %s not written: %v", p, err)
			continue
		}
		if fi.Size() == 0 {
			t.Errorf("profile %s is empty", p)
		}
	}
	if out.Len() == 0 {
		t.Errorf("expected a summary line on stdout")
	}
}

func TestRunBadFlag(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"-nonsense"}, &out, &errBuf); code != 2 {
		t.Errorf("bad flag exit code = %d, want 2", code)
	}
}
```

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/22-profiling-fuzz/
go test ./lessons/22-profiling-fuzz/exercises/cmd/aggregator-profile/... 2>&1 | tail -5
go test -v ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile/... 2>&1 | tail -15
go test -race ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile/... 2>&1 | tail -5
make test
make test-race
go vet ./lessons/22-profiling-fuzz/...
golangci-lint run ./lessons/22-profiling-fuzz/...

# Manual: write + inspect real profiles
go run ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile -n=2000 -duration=2s -cpuprofile=/tmp/cpu.prof -memprofile=/tmp/heap.prof
go tool pprof -top -nodecount=5 /tmp/cpu.prof | head -15
rm -f /tmp/cpu.prof /tmp/heap.prof

git add lessons/22-profiling-fuzz/exercises/cmd/aggregator-profile lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile
git commit -m "feat(lesson-22): cmd/aggregator-profile (runtime/pprof CPU+heap + net/http/pprof endpoint)"
```

Expected: exercises vacuous-pass; solutions tests PASS (profiles non-empty, exit 0); -race clean.

---

## Task 5: Slide deck — 5 concepts

Heavy-explanatory pattern per concept (motivation → basics → worked example → common mistake → recap). ~30 slides.

**File:** `lessons/22-profiling-fuzz/slides/slides.md`

Concept order:

1. **Benchmarking** — `go test -bench=. -benchmem`, `b.N`, `b.ResetTimer`, `b.ReportAllocs`; reading ns/op · B/op · allocs/op; the `sink` trick. Common-mistake: dead-code elimination (compiler discards the unused result → bogus 0.3ns/op).
2. **CPU profiling** — `runtime/pprof.StartCPUProfile` / `go test -cpuprofile`; `go tool pprof` (`top`, `list`, `-http` flame graph). Worked example: profiling the aggregator reveals `regexp.FindStringSubmatch`. Common-mistake: run too short → no samples.
3. **Heap / allocation profiling + escape analysis** — `-memprofile`, `pprof -alloc_space` vs `-inuse_space`, `go build -gcflags=-m` to see escapes; why `FindStringSubmatch` allocates a `[]string` per call (2 allocs/op → 0). Common-mistake: confusing alloc_space (cumulative) with inuse_space (live).
4. **`net/http/pprof`** — `import _ "net/http/pprof"`, `/debug/pprof/` endpoints, `go tool pprof http://host/debug/pprof/profile?seconds=30`. Common-mistake: exposing the pprof endpoint on a public interface (info leak / DoS surface).
5. **Fuzzing** — `func FuzzX(f *testing.F)`, `f.Add` seeds, `f.Fuzz`, `go test -fuzz`, the `testdata/fuzz` corpus, oracle/differential fuzzing. Worked example: the naive `parseLineFast` panics on `"0"`; fuzz finds it in <1s; fix the index guard; seed it forever. Common-mistake: a non-deterministic fuzz target (flaky → useless corpus).

- [ ] **Step 1-7:** Author the deck following the L19/L21 format (title-slide-grid header, "What we'll cover", "The story so far", 5 concept sections, Practice, Closing thought, What we learned, Up next → Phase 4 / Lesson 23 HTTP servers).
- [ ] **Step 8:** Verify slides build + commit:

```bash
make slides-build
grep -q "22-profiling-fuzz" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/22-profiling-fuzz/slides/
git commit -m "feat(lesson-22): slides — profiling, benchmarking & fuzzing (5 concepts)"
```

> Note: the index check requires the build-index slug update (Task 6 Step 1). If running Task 5 before Task 6, the grep will fail — either reorder so the build-index update lands first, or accept the grep failing until Task 6.

---

## Task 6: build-index update + README + e2e verify + final review + PR

- [ ] **Step 1:** Update `tools/build-index/main.go`. Change the lesson 22 entry:

```go
// FROM:
{Number: "22", Slug: "profiling", Title: "Profiling & benchmarks", Blurb: "pprof · go test -bench", Phase: 3},
// TO:
{Number: "22", Slug: "profiling-fuzz", Title: "Profiling & fuzzing", Blurb: "pprof · bench · fuzz", Phase: 3},
```

Verify the build-index tests still pass: `go test ./tools/build-index/...`.

- [ ] **Step 2:** Write `lessons/22-profiling-fuzz/README.md` (~280 lines). Mirror the 5 concepts with per-concept Common-mistake paragraphs. Include:
  - "What's different from L21": L21 was standalone networking; L22 returns to the aggregator to MEASURE and HARDEN it — the Phase 3 capstone-by-another-name.
  - The optimization arc with the actual verified numbers (698→130 ns/op, 2→0 allocs, ~5.4×) framed as "your machine will differ; the shape won't."
  - "How to run": benchmark commands, the profile cmd, `go tool pprof -http`, `go test -fuzz`.
  - `make test-race` in daily habits.
  - "Going further": run the fuzzer for 5 min; reintroduce the bug and watch fuzz catch it; profile `WalkLocked` vs `WalkPool`; add a `-blockprofile`.

- [ ] **Step 3:** Full verification sweep:

```bash
make test
make test-race
go test -v ./lessons/22-profiling-fuzz/solutions/...
go test -race ./lessons/22-profiling-fuzz/...
go test -bench=. -benchmem ./lessons/22-profiling-fuzz/solutions/... 2>&1 | tail -20
go vet ./...
golangci-lint run ./...
gofmt -l lessons/22-profiling-fuzz/
make slides-build && grep -q "22-profiling-fuzz" dist/index.html && echo "✓ index" && rm -rf dist
```

- [ ] **Step 4:** Commit README + build-index:

```bash
git add lessons/22-profiling-fuzz/README.md tools/build-index/main.go
git commit -m "docs(lesson-22): README + build-index slug profiling-fuzz"
```

- [ ] **Step 5:** Dispatch the `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: the manual parser's equivalence to the regex oracle, fuzz determinism under `go test`, no committed `testdata/fuzz` cache, the profile cmd's resource handling (temp dir cleanup, profile file close), and slide/README/code consistency. Apply fixes.

- [ ] **Step 6:** Push + open PR (same pattern as PR #38):

```bash
git push -u origin feature/plan-y-lesson-22-profiling-fuzz
gh pr create --title "Lesson 22 — Profiling, benchmarking & fuzzing" --body "..."
```

---

## Verification (after Task 6)

```bash
# All tests pass (lessons 01-22 + tools)
make test
make test-race

# Lesson 22 exercises pass vacuously
go test ./lessons/22-profiling-fuzz/exercises/...

# Solutions: warmup bench, logparse (table + fuzz seeds), aggregator (carried),
# errgroupx (carried), cmd/aggregator (carried), cmd/aggregator-profile
go test -v ./lessons/22-profiling-fuzz/solutions/...

# Benchmarks run; fast ≪ regex, 0 allocs
go test -bench=. -benchmem ./lessons/22-profiling-fuzz/solutions/...

# Random fuzz finds nothing on the fixed parser (positive signal)
go test -run='^$' -fuzz=FuzzParseLineFast -fuzztime=20s ./lessons/22-profiling-fuzz/solutions/internal/logparse/

# Profile cmd produces real, inspectable profiles
go run ./lessons/22-profiling-fuzz/solutions/cmd/aggregator-profile -n=2000 -duration=2s -cpuprofile=/tmp/cpu.prof -memprofile=/tmp/heap.prof
go tool pprof -top -nodecount=8 /tmp/cpu.prof
rm -f /tmp/cpu.prof /tmp/heap.prof

# Static analysis clean
go vet ./...
golangci-lint run ./...
gofmt -l lessons/22-profiling-fuzz/

# Slides build; lesson 22 lands in the index
make slides-build && grep -q "22-profiling-fuzz" dist/index.html && rm -rf dist
```

## Critical file paths

To be created:
- `lessons/22-profiling-fuzz/` (directory)
- `lessons/22-profiling-fuzz/README.md`
- `lessons/22-profiling-fuzz/slides/{index.html (scaffolded), slides.md, assets/.gitkeep}`
- `lessons/22-profiling-fuzz/exercises/warmup/bench/{bench.go, bench_test.go}`
- `lessons/22-profiling-fuzz/exercises/internal/logparse/{logparse.go, logparse_test.go, logparse_bench_test.go, fuzz_test.go}`
- `lessons/22-profiling-fuzz/exercises/internal/aggregator/{aggregator.go, aggregator_test.go}` (carried)
- `lessons/22-profiling-fuzz/exercises/internal/errgroupx/{errgroupx.go, errgroupx_test.go}` (carried)
- `lessons/22-profiling-fuzz/exercises/cmd/aggregator/{main.go, main_test.go}` (carried)
- `lessons/22-profiling-fuzz/exercises/cmd/aggregator-profile/{main.go, main_test.go}`
- `lessons/22-profiling-fuzz/solutions/...` (mirrored)

To be modified:
- `tools/build-index/main.go` (slug `profiling` → `profiling-fuzz`, Title, Blurb)

To be referenced (not modified):
- `lessons/20-concurrency-patterns/solutions/...` + `exercises/...` (carry-forward source)
- `.golangci.yml` (no changes; gosec not enabled, errcheck excluded for lessons)

## Execution after approval

1. (Branch `feature/plan-y-lesson-22-profiling-fuzz` already created off main.)
2. Commit this plan doc.
3. Execute the 6 tasks via subagent-driven development (controller carries forward + verifies Task 1 directly; dispatches subagents for Tasks 2-4; writes slides/README inline for Tasks 5-6; runs verification directly).
4. Final code review subagent.
5. Push and open PR.
