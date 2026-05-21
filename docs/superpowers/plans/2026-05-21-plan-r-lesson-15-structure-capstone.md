# Plan R — Lesson 15 (Project Structure CAPSTONE) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 15 of Phase 2 — the **CAPSTONE**. Reorganise the tracker that evolved through L08/L09/L10/L11/L13 into a polished v2 with the idiomatic `cmd/` + `internal/` layout. Teach Go's testing toolkit beyond table tests: `t.Helper`, `t.TempDir`, golden file tests with `-update`, benchmarks (`testing.B`), and a fuzz preview (`testing.F`). Adds one significant net-new piece (golden-file integration test for `cmd/expenses summary` + benchmark for `summary.TotalsByCategory` on 10,000 entries + tiny `FuzzParseLine` on `csvimport`) on top of substantial mechanical carry-forwards from prior lessons.

**Architecture:** Same per-lesson pattern as Plans D-Q. Seven tasks (one more than usual due to capstone scope). Skeleton tests in warmup + main subpackages (Phase 2 default). Heavy-explanatory slide deck with **five** concept blocks (cmd+internal → test helpers → golden files → benchmarks → fuzz preview).

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `bytes`, `encoding/json`, `errors`, `flag`, `fmt`, `io`, `io/fs`, `os`, `os/exec`, `path/filepath`, `regexp`, `sort`, `strconv`, `strings`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan R: lesson 15 is complete; `make test` green; both binaries work end-to-end against golden files; benchmark runs and reports ns/op; fuzz seeds run as normal tests.

### Design decisions (2 user-approved + 8 plan-recommended)

**User-approved via brainstorming:**

1. **Tiny standalone warmup**: `warmup/internal/util/format.go` + `warmup/main.go` that imports it. Compiles cleanly (same module, `internal/` access allowed). The slides explain that a sibling module trying to import it would fail. Preserves the Phase 2 "skeleton tests in both warmup and main" invariant.

2. **Five slide concepts with fuzz as its own concept** (not a closing-thought callback). `FuzzParseLine` gets ~20 lines of working code in `internal/csvimport/csvimport_fuzz_test.go` — students can run `go test -fuzz=FuzzParseLine` and see it explore. Phase 3 L22 dives deeper.

**Plan-recommended:**

3. **`cmd/` + `internal/` layout per Phase 2 spec.** `internal/` packages: expense, store, csvimport, summary, testutil. `cmd/` binaries: expenses, expenses-import. Three top-level directories at the exercise root (warmup, cmd, internal) — Go's module system enforces internal/ visibility at the compiler level.

4. **Most carry-forwards are verbatim with only import-path rewrites.** L11's expense, L13's store (with the Encoder/Decoder refactor), L13's csvimport, L13's cmd/expenses-import — all move with their full test suites carrying over.

5. **`cmd/expenses` is REFACTORED to use `internal/summary`.** L11/L13's `cmdSummary` inlined its own summary logic. L15 rewrites cmdSummary to call `internal/summary.TotalsByCategory` + `BiggestCategory` + `Bar`. Net effect: the summary subcommand output is RICHER (includes the biggest-category line + bar chart) and demonstrates internal/ usage at the cmd level.

6. **L08's summary package is carried forward verbatim.** Three functions: `TotalsByCategory(es []expense.Expense) map[string]float64`, `BiggestCategory(totals map[string]float64) (string, float64)`, `Bar(value, max float64, width int) string`. Already tested in L08; tests carry forward.

7. **`internal/testutil/AssertGolden` is PROVIDED** (~25 lines). Students don't implement it. They USE it from `cmd/expenses/main_test.go`. The `-update` flag is a package-level `flag.Bool` in testutil.

8. **Benchmark generates 10,000 random expenses with a fixed seed** for reproducibility. Same seed in exercises and solutions. `BenchmarkTotalsByCategory` calls `summary.TotalsByCategory(es)` inside the `for i := 0; i < b.N; i++ {}` loop with setup before the loop. `b.ResetTimer()` after setup; `b.ReportAllocs()` enabled.

9. **Fuzz function fuzzes `parseLine`** (the package-private helper in csvimport). Seeded with three known-valid lines; the fuzz function asserts `parseLine` never panics. Tagged in the slides as "syntax preview; deep dive in L22." The fuzz test runs as a normal Go test under `go test` (just exercising the seeds); `-fuzz=FuzzParseLine` lets students actually explore. Note: parseLine is unexported, so the fuzz test must live in the same package (no `_test` suffix on the package declaration).

10. **Golden file location: `cmd/expenses/testdata/`**. Go's testing tool automatically excludes `testdata/` from package compilation — the golden files live alongside the test code without polluting the package.

---

## Plans F-Q lessons-learned applied here

1. Subpackages under internal/ keep their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24, `lessons/.*/exercises/`) covers nested subpackages including the new `internal/` tree — no config changes.
3. Common-mistake content in README (5 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `gofmt -w .` and `go vet ./...` mentions in README.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files; build the subpackage tree).
8. gofmt 1.19+ normalises godoc list indentation — `gofmt -w` if triggered.
9. The build-index master list slug for L15 is `structure` (already correct in `tools/build-index/main.go:54`) — directory name must be `15-structure`.

---

## File Structure

After Plan R (~40 files — biggest Phase 2 lesson):

```
lessons/15-structure/
├── README.md                                            (Task 6)
├── slides/{index.html, slides.md, assets/.gitkeep}      (Task 1 + Task 5)
├── exercises/
│   ├── warmup/
│   │   ├── main.go                                      (Task 2)
│   │   └── internal/util/
│   │       ├── format.go                                (Task 2)
│   │       └── format_test.go                           (Task 2 — SKELETON)
│   ├── cmd/
│   │   ├── expenses/
│   │   │   ├── main.go                                  (Task 4 — refactored)
│   │   │   ├── main_test.go                             (Task 4 — golden test SKELETON)
│   │   │   └── testdata/summary.golden                  (Task 4)
│   │   └── expenses-import/
│   │       ├── main.go                                  (Task 4 — verbatim L13)
│   │       └── main_test.go                             (Task 4 — verbatim L13)
│   └── internal/
│       ├── expense/
│       │   ├── expense.go                               (Task 3 — verbatim L11)
│       │   └── expense_test.go                          (Task 3 — verbatim L09)
│       ├── store/
│       │   ├── store.go                                 (Task 3 — verbatim L13)
│       │   └── store_test.go                            (Task 3 — verbatim L13 SKELETON)
│       ├── csvimport/
│       │   ├── csvimport.go                             (Task 3 — verbatim L13)
│       │   ├── csvimport_test.go                        (Task 3 — verbatim L13)
│       │   └── csvimport_fuzz_test.go                   (Task 3 — NEW)
│       ├── summary/
│       │   ├── summary.go                               (Task 3 — verbatim L08)
│       │   ├── summary_test.go                          (Task 3 — verbatim L08)
│       │   └── summary_bench_test.go                    (Task 3 — NEW)
│       └── testutil/
│           └── golden.go                                (Task 3 — PROVIDED)
└── solutions/  (mirrored)
```

Top-level: 4 files (README + 3 slides). Exercise/solution code: ~36 files. Total ~40.

---

## Conventions

- **Branch:** `feature/plan-r-lesson-15-structure-capstone`
- **Commit messages:** Conventional Commits
- **Carry-forward sources** (read-only; reference for path-rewrite work):
  - `lessons/11-errors/solutions/expense/expense.go`
  - `lessons/09-pointers/solutions/expense/expense_test.go` (the test suite for the four methods)
  - `lessons/13-encoding-io/solutions/store/{store.go, store_test.go}`
  - `lessons/13-encoding-io/solutions/csvimport/{csvimport.go, csvimport_test.go}`
  - `lessons/13-encoding-io/solutions/cmd/expenses-import/{main.go, main_test.go}`
  - `lessons/13-encoding-io/solutions/cmd/expenses/main.go` (REFACTORED in L15 to use internal/summary)
  - `lessons/08-capstone/solutions/summary/{summary.go, summary_test.go}`

---

## Task 1: Scaffold + restructure

Same dance as Plans J/K/L/M/N/O/P/Q.

- [ ] **Step 1:** `make new-lesson NAME=15-structure`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/15-structure/exercises/warmup.go
rm lessons/15-structure/exercises/warmup_test.go
rm lessons/15-structure/exercises/main.go
rm lessons/15-structure/exercises/main_test.go
rm lessons/15-structure/solutions/warmup.go
rm lessons/15-structure/solutions/warmup_test.go
rm lessons/15-structure/solutions/main.go
rm lessons/15-structure/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree.

- [ ] **Step 4:** Commit:

```bash
git add lessons/15-structure/
git commit -m "feat(lessons): scaffold lesson 15-structure with empty subpackage layout"
```

---

## Task 2: Author the warm-up — internal/util demo

`warmup/internal/util/format.go` exposes a tiny helper. `warmup/main.go` imports it. Compiles cleanly (same module). The lesson is the COMPILATION SUCCESS — the demo of `internal/` access from siblings/descendants of the parent directory. Slides cover the FAILURE case (sibling module can't import).

**Files (6 total — 3 per side):**

- [ ] **Step 1:** `mkdir -p lessons/15-structure/{exercises,solutions}/warmup/internal/util`

- [ ] **Step 2:** Create `lessons/15-structure/exercises/warmup/internal/util/format.go`:

```go
// Package util is a tiny demo of the internal/ access rule.
//
// This package lives under warmup/internal/util/. Per Go's module
// system, packages under an internal/ directory can ONLY be imported
// from code inside the parent directory or its descendants — i.e.,
// from anything inside warmup/. Code in a sibling directory (say,
// lessons/14-time-strings-regex/...) trying to import this package
// would fail at compile time:
//
//	package warmup/internal/util is not allowed: use of internal package
//
// The compiler enforces it. There's no opt-out.
package util

import "fmt"

// Money formats a euro amount as "€<2-decimal>". E.g., Money(4.5) →
// "€4.50".
//
// Hint: return fmt.Sprintf("€%.2f", amount)
func Money(amount float64) string {
	_ = fmt.Sprintf
	panic("TODO: fmt.Sprintf with %.2f")
}
```

- [ ] **Step 3:** Create `lessons/15-structure/exercises/warmup/internal/util/format_test.go` (SKELETON):

```go
package util

import "testing"

// TestMoney is a SKELETON. Assert that Money returns the expected
// "€<2-decimal>" formatting for at least 3 inputs.
func TestMoney(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		// TODO: at least 3 cases.
		// {"whole-number", 4.0, "€4.00"},
		// {"two-decimals", 4.50, "€4.50"},
		// {"long-decimals-rounded", 4.567, "€4.57"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got := Money(tc.in)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}
```

- [ ] **Step 4:** Create `lessons/15-structure/exercises/warmup/main.go`:

```go
// Package main is the lesson 15 warm-up: importing an internal/ package
// from a sibling directory in the SAME module.
//
// This binary compiles because warmup/main.go and
// warmup/internal/util/format.go share a common ancestor (warmup/).
// The internal/ rule allows imports from descendants of the parent of
// the internal/ directory.
//
// Try moving main.go up one level (to lessons/15-structure/exercises/)
// and re-running `go build`. The build will fail:
//
//	use of internal package github.com/.../warmup/internal/util
//	  not allowed
//
// That's the rule, enforced by the compiler.
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/warmup/internal/util"
)

func main() {
	fmt.Println(util.Money(4.50))
}
```

- [ ] **Step 5:** Create `lessons/15-structure/solutions/warmup/internal/util/format.go`:

```go
// Package util is the lesson 15 warm-up reference implementation.
package util

import "fmt"

// Money formats a euro amount as "€<2-decimal>".
func Money(amount float64) string {
	return fmt.Sprintf("€%.2f", amount)
}
```

- [ ] **Step 6:** Create `lessons/15-structure/solutions/warmup/internal/util/format_test.go`:

```go
package util

import "testing"

func TestMoney(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"whole-number", 4.0, "€4.00"},
		{"two-decimals", 4.50, "€4.50"},
		{"three-decimals-rounded-up", 4.567, "€4.57"},
		{"zero", 0.0, "€0.00"},
		{"large", 12345.67, "€12345.67"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Money(tc.in)
			if got != tc.want {
				t.Errorf("Money(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 7:** Create `lessons/15-structure/solutions/warmup/main.go`:

```go
// Package main is the lesson 15 warm-up reference implementation.
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/warmup/internal/util"
)

func main() {
	fmt.Println(util.Money(4.50))
}
```

- [ ] **Step 8:** Verify and commit:

```bash
gofmt -l lessons/15-structure/
go test ./lessons/15-structure/exercises/warmup/... -v 2>&1 | tail -10
go test ./lessons/15-structure/solutions/warmup/... -v 2>&1 | tail -15
go run ./lessons/15-structure/solutions/warmup  # should print "€4.50"
make test
golangci-lint run ./...
go vet ./...

git add lessons/15-structure/exercises/warmup/ lessons/15-structure/solutions/warmup/
git commit -m "feat(lesson-15): warmup — internal/util demo (compiler-enforced visibility)"
```

Expected: gofmt empty; exercises pass vacuously; solutions TestMoney (5 sub-tests) PASS; warmup binary prints "€4.50"; make test green.

---

## Task 3: Bulk internal/ — expense, store, csvimport (+fuzz), summary (+bench), testutil

**Six subpackages, ~26 files (13 per side).** Mostly mechanical carry-forwards with import-path rewrites. Net-new: `internal/testutil/golden.go`, `internal/summary/summary_bench_test.go`, `internal/csvimport/csvimport_fuzz_test.go`.

The subagent dispatched for this task should READ the carry-forward source files first (paths listed in the "Conventions" section above) and then write each L15 file with the appropriate `solutions/` or `exercises/` import-path rewrite.

### Step 1: Create all directories

```bash
mkdir -p lessons/15-structure/{exercises,solutions}/internal/{expense,store,csvimport,summary,testutil}
```

### Step 2-3: `internal/expense/`

- [ ] **Step 2:** Forward-port `lessons/11-errors/solutions/expense/expense.go` to `lessons/15-structure/exercises/internal/expense/expense.go` (identical content; no path rewrite needed inside the file).

- [ ] **Step 3:** Same for `lessons/15-structure/solutions/internal/expense/expense.go`.

For the test file: L11 didn't have an expense_test.go (the expense tests live in L09). Forward-port `lessons/09-pointers/solutions/expense/expense_test.go` to both exercises and solutions. The L09 test file uses package `expense` (no _test suffix) and exercises Format/IsHigh/ApplyDiscount/Bump.

### Step 4-5: `internal/store/`

- [ ] **Step 4:** Forward-port `lessons/13-encoding-io/solutions/store/store.go` to `lessons/15-structure/exercises/internal/store/store.go`. **Path rewrite:** `lessons/13-encoding-io/solutions/expense` → `lessons/15-structure/exercises/internal/expense`. The exercises version keeps the panicking-stub Load/Save bodies and the helper var-block from L13 exercises (not solutions).

  ACTUALLY: re-check. L13 exercises/store/store.go had Load and Save as panicking stubs. L13 solutions/store/store.go had them implemented. For L15, students are LEARNING about structure, not re-implementing Encoder/Decoder logic. So L15 exercises should have the FULL implementation (with import-path adjustments), not the stub. Otherwise students would have to re-implement store internals twice.

  Final rule for L15 exercises tree: copy `lessons/13-encoding-io/SOLUTIONS/store/store.go` (the full impl) and rewrite import paths to point at `15-structure/exercises/internal/expense`. Same for L15 solutions, pointing at `15-structure/solutions/internal/expense`.

- [ ] **Step 5:** Same forward-port for `solutions/internal/store/store.go`.

For store_test.go: forward-port `lessons/13-encoding-io/solutions/store/store_test.go` to BOTH exercises and solutions (path-adjusted). L15's "exercise" for the store package is NOT to reimplement the tests — it's to move the package. The tests run as-is.

### Step 6-7: `internal/csvimport/` (+ new fuzz file)

- [ ] **Step 6:** Forward-port `lessons/13-encoding-io/solutions/csvimport/csvimport.go` to both trees (import-path-rewritten). Both trees get the full implementation.

- [ ] **Step 7:** Forward-port `lessons/13-encoding-io/solutions/csvimport/csvimport_test.go` to both trees.

- [ ] **Step 8:** Create the NEW fuzz file `lessons/15-structure/exercises/internal/csvimport/csvimport_fuzz_test.go` (SKELETON):

```go
package csvimport

import (
	"testing"
)

// FuzzParseLine is a SKELETON. The full fuzz function exists in the
// solutions tree. Here the seeds + body are placeholders.
//
// Lesson 15 introduces the syntax; Phase 3 lesson 22 covers fuzzing in
// depth. To actually run the fuzzer:
//
//	go test -fuzz=FuzzParseLine ./lessons/15-structure/solutions/internal/csvimport/
//
// (Run from the solutions tree where the implementation is real.)
func FuzzParseLine(f *testing.F) {
	// TODO: f.Add three valid seeds; f.Fuzz a function that calls
	// parseLine and asserts no panic.
	_ = f
}
```

- [ ] **Step 9:** Create `lessons/15-structure/solutions/internal/csvimport/csvimport_fuzz_test.go`:

```go
package csvimport

import (
	"testing"
)

// FuzzParseLine exercises parseLine against arbitrary string input.
// Seeded with three valid CSV lines; the fuzzer mutates from there
// looking for inputs that panic.
//
// Run normally (just the seeds, ~milliseconds):
//
//	go test ./lessons/15-structure/solutions/internal/csvimport/
//
// Run as a fuzzer (until interrupted, explores the input space):
//
//	go test -fuzz=FuzzParseLine ./lessons/15-structure/solutions/internal/csvimport/
//
// parseLine returns an error for malformed input — that's fine. We
// assert it never PANICS. Any panic is a real bug.
func FuzzParseLine(f *testing.F) {
	f.Add("2026-05-21,4.50,coffee")
	f.Add("")
	f.Add("a,b,c")

	f.Fuzz(func(t *testing.T, line string) {
		// parseLine is package-private; this fuzz test lives in the
		// same package so it can call it directly.
		_, _ = parseLine(line) // any panic fails the test
	})
}
```

### Step 10-11: `internal/summary/` (+ new benchmark)

- [ ] **Step 10:** Forward-port `lessons/08-capstone/solutions/summary/summary.go` to both trees (path-rewritten — `lessons/08-capstone/solutions/expense` → `lessons/15-structure/{exercises,solutions}/internal/expense`).

- [ ] **Step 11:** Forward-port `lessons/08-capstone/solutions/summary/summary_test.go` to both trees (same path rewrite).

- [ ] **Step 12:** Create the NEW benchmark file `lessons/15-structure/exercises/internal/summary/summary_bench_test.go` (SKELETON):

```go
package summary

import (
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/expense"
)

// BenchmarkTotalsByCategory is a SKELETON. The full benchmark exists
// in the solutions tree (10,000 random expenses, fixed seed, b.N loop).
// Here we keep the body trivial so the file compiles.
//
// To run benchmarks:
//
//	go test -bench=. ./lessons/15-structure/solutions/internal/summary/
//
// (Run from the solutions tree where the implementation is real.)
func BenchmarkTotalsByCategory(b *testing.B) {
	// TODO: generate 10,000 random expenses (rand with fixed seed);
	// b.ResetTimer(); b.ReportAllocs(); loop calling TotalsByCategory.
	_ = expense.Expense{}
	b.Skip("benchmark not implemented in exercises tree; see solutions")
}
```

- [ ] **Step 13:** Create `lessons/15-structure/solutions/internal/summary/summary_bench_test.go`:

```go
package summary

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

// BenchmarkTotalsByCategory measures TotalsByCategory throughput on a
// realistic 10,000-entry slice with a few dozen distinct categories.
//
// Run with:
//
//	go test -bench=BenchmarkTotalsByCategory -benchmem \
//	  ./lessons/15-structure/solutions/internal/summary/
//
// Expected output looks like (numbers will vary by machine):
//
//	BenchmarkTotalsByCategory-8     5000    234567 ns/op    8192 B/op    1 allocs/op
//
// The benchmark is informational — no assertions. It's there so
// students can see ns/op and B/op for a realistic workload.
func BenchmarkTotalsByCategory(b *testing.B) {
	// Setup: 10,000 random expenses with a fixed seed for reproducibility.
	r := rand.New(rand.NewSource(42))
	categories := []string{"coffee", "lunch", "groceries", "transit", "entertainment", "books", "household"}
	es := make([]expense.Expense, 10000)
	for i := range es {
		es[i] = expense.Expense{
			Date:     fmt.Sprintf("2026-%02d-%02d", 1+r.Intn(12), 1+r.Intn(28)),
			Amount:   r.Float64() * 100,
			Category: categories[r.Intn(len(categories))],
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = TotalsByCategory(es)
	}
}
```

### Step 14: `internal/testutil/golden.go` (provided helper)

- [ ] **Step 14:** Create `lessons/15-structure/exercises/internal/testutil/golden.go`:

```go
// Package testutil provides small helpers for tests in lesson 15.
//
// PROVIDED — students don't implement this file; they use AssertGolden
// from cmd/expenses/main_test.go.
package testutil

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

// Update enables overwriting golden files when go test is invoked with
// `-update`. The flag is registered at package init.
//
// To regenerate a golden file after legitimate output changes:
//
//	go test -update ./path/to/test/
var Update = flag.Bool("update", false, "regenerate golden files")

// AssertGolden compares got against the contents of the file at path.
// On mismatch, fails the test with a diff. If -update is set, writes
// got to path instead of comparing (regenerates the golden file).
//
// Usage:
//
//	output := runMyCommand(...)
//	testutil.AssertGolden(t, output, "testdata/expected.golden")
func AssertGolden(t *testing.T, got []byte, path string) {
	t.Helper()
	if *Update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("update %s: %v", path, err)
		}
		t.Logf("updated %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output mismatch with %s:\n--- got (%d bytes) ---\n%s\n--- want (%d bytes) ---\n%s",
			path, len(got), got, len(want), want)
	}
}
```

- [ ] **Step 15:** Same content for `solutions/internal/testutil/golden.go`.

### Step 16: Verify and commit

- [ ] **Step 16:** Verify and commit Task 3:

```bash
gofmt -l lessons/15-structure/
go test ./lessons/15-structure/exercises/internal/... -v 2>&1 | tail -30
go test ./lessons/15-structure/solutions/internal/... -v 2>&1 | tail -40
go test -bench=. -benchtime=1x ./lessons/15-structure/solutions/internal/summary/  # quick sanity check
make test
golangci-lint run ./...
go vet ./...

git add lessons/15-structure/exercises/internal/ lessons/15-structure/solutions/internal/
git commit -m "feat(lesson-15): internal/* — expense, store, csvimport (+fuzz), summary (+bench), testutil"
```

Expected: gofmt empty; all internal/* tests PASS in both trees; benchmark runs without crashing; make test green.

---

## Task 4: Bulk cmd/ — cmd/expenses (refactored + golden test) + cmd/expenses-import (verbatim L13)

**Two binaries, ~7 files (3-4 per side).** cmd/expenses is REWRITTEN to use `internal/summary`; cmd/expenses-import is verbatim L13.

### Step 1: Create directories

```bash
mkdir -p lessons/15-structure/{exercises,solutions}/cmd/expenses/testdata
mkdir -p lessons/15-structure/{exercises,solutions}/cmd/expenses-import
```

### Step 2-3: cmd/expenses-import (verbatim L13)

- [ ] **Step 2:** Forward-port `lessons/13-encoding-io/solutions/cmd/expenses-import/main.go` to both trees, with imports rewritten:
  - `lessons/13-encoding-io/solutions/csvimport` → `lessons/15-structure/{exercises,solutions}/internal/csvimport`
  - `lessons/13-encoding-io/solutions/store` → `lessons/15-structure/{exercises,solutions}/internal/store`

- [ ] **Step 3:** Forward-port `lessons/13-encoding-io/solutions/cmd/expenses-import/main_test.go` to both trees. Same import rewrites if any; the test only imports `expense`.

### Step 4-7: cmd/expenses (refactored to use internal/summary)

- [ ] **Step 4:** Create `lessons/15-structure/exercises/cmd/expenses/main.go`. This is REWRITTEN from L13 — the `cmdSummary` function delegates to `internal/summary`. Full source:

```go
// Package main is the lesson 15 expense tracker CLI — the v2 polished
// version. The summary subcommand now uses internal/summary for
// totals/biggest/bar; the rest matches L11/L13.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/expense"
	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/store"
	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/summary"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		var pe *store.ParseError
		if errors.As(err, &pe) {
			fmt.Fprintf(os.Stderr, "error: could not parse %s at line %d: %v\n",
				pe.Path, pe.Line, pe.Cause)
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

// run is the testable entry point. Takes args + an io.Writer for
// stdout output (so tests can capture into a bytes.Buffer).
//
// The signature changed from L11/L13 (which used os.Stdout implicitly)
// — passing the Writer explicitly is required for golden file tests.
func run(args []string, stdout interface {
	Write(p []byte) (n int, err error)
}) error {
	storeKind, args, err := parseStoreOption(args)
	if err != nil {
		return err
	}
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	s, err := buildStore(storeKind, path)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(s, args[1:], stdout)
	case "list":
		return cmdList(s, stdout)
	case "summary":
		return cmdSummary(s, stdout)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseStoreOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-store=") {
		v := args[0][len("-store="):]
		if v != "mem" && v != "json" {
			return "", nil, fmt.Errorf("invalid -store value %q (must be mem or json)", v)
		}
		return v, args[1:], nil
	}
	return "json", args, nil
}

func parseFileOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-file=") {
		return args[0][len("-file="):], args[1:], nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".expenses.json"), args, nil
}

func buildStore(kind, path string) (store.Store, error) {
	switch kind {
	case "mem":
		return store.NewMemoryStore(), nil
	case "json":
		return store.NewJSONStore(path), nil
	default:
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}

func cmdAdd(s store.Store, args []string, w interface {
	Write(p []byte) (n int, err error)
}) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := s.Save(es); err != nil {
		return err
	}
	fmt.Fprintln(w, "added:", e.Format())
	return nil
}

func cmdList(s store.Store, w interface {
	Write(p []byte) (n int, err error)
}) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Fprintln(w, e.Format())
	}
	return nil
}

// cmdSummary uses internal/summary for TotalsByCategory + BiggestCategory + Bar.
// Output: totals per category (alphabetical), then biggest line, then bar chart.
func cmdSummary(s store.Store, w interface {
	Write(p []byte) (n int, err error)
}) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}

	var total float64
	for _, e := range es {
		total += e.Amount
	}
	fmt.Fprintf(w, "%d expenses, total €%.2f\n", len(es), total)
	if len(es) == 0 {
		return nil
	}

	totals := summary.TotalsByCategory(es)

	// Sorted by category name for stable output.
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	// Sort via the strings convention.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	fmt.Fprintln(w, "by category:")
	for _, k := range keys {
		fmt.Fprintf(w, "  %-10s €%-7.2f\n", k, totals[k])
	}

	// Biggest line + bar chart.
	name, biggest := summary.BiggestCategory(totals)
	if name != "" {
		fmt.Fprintf(w, "biggest: %s €%.2f\n", name, biggest)
		fmt.Fprintf(w, "  %s\n", summary.Bar(biggest, biggest, 20))
	}

	return nil
}
```

> Note on the inline sort: I used a manual bubble sort instead of `sort.Strings` to avoid pulling in the `sort` package for this single use. Reasonable for a small slice; if the category count grew large, `sort.Strings(keys)` would be the obvious choice. Phase 1 students saw sort.Strings in L08; this is a minor stylistic choice.

> Note on the `interface { Write(p []byte) (n int, err error) }` parameter: this is an inline `io.Writer` interface to avoid importing `io` from main. The cleaner alternative is `import "io"` and `w io.Writer`. Either works. The inline version is uglier but keeps the import list minimal.

Actually let's clean that up — using `io.Writer` directly is cleaner. Update the source above to use `io.Writer` (and add `"io"` to imports).

REVISED main.go body uses `io.Writer` instead of the inline interface:

```go
import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// ... internal imports
)

func run(args []string, stdout io.Writer) error { ... }
func cmdAdd(s store.Store, args []string, w io.Writer) error { ... }
func cmdList(s store.Store, w io.Writer) error { ... }
func cmdSummary(s store.Store, w io.Writer) error { ... }
```

The implementer subagent should use the `io.Writer` version, not the inline interface form. Keep the rest of the body as above.

- [ ] **Step 5:** Create `lessons/15-structure/exercises/cmd/expenses/main_test.go` (SKELETON):

```go
package main

import (
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/testutil"
)

// TestSummaryGolden is a SKELETON. The full test exists in solutions.
// In the exercise tree, the test is a stub so the package compiles.
//
// The solution test seeds a JSON store with known expenses, runs
// `summary`, captures stdout, and compares to testdata/summary.golden
// via testutil.AssertGolden. With `go test -update`, it regenerates
// the golden file.
func TestSummaryGolden(t *testing.T) {
	// TODO: seed store; run "summary"; capture stdout into bytes.Buffer;
	// AssertGolden(t, buf.Bytes(), "testdata/summary.golden")
	_ = testutil.AssertGolden
	_ = filepath.Join
}
```

- [ ] **Step 6:** Create `lessons/15-structure/solutions/cmd/expenses/main.go` — same content as exercises but with `solutions/` import paths.

- [ ] **Step 7:** Create `lessons/15-structure/solutions/cmd/expenses/main_test.go`:

```go
package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/store"
	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/testutil"
)

// TestSummaryGolden runs the summary subcommand against a fixed
// dataset, captures stdout into a bytes.Buffer, and compares to the
// golden file at testdata/summary.golden.
//
// To regenerate the golden file after intentional output changes:
//
//	go test -update ./lessons/15-structure/solutions/cmd/expenses/
//
// The test uses an in-memory MemoryStore to avoid filesystem I/O for
// the test fixture. The CLI itself (run via os.Args) uses JSONStore;
// this test bypasses parseStoreOption / buildStore to seed the store
// directly.
func TestSummaryGolden(t *testing.T) {
	s := store.NewMemoryStore()
	if err := s.Save([]expense.Expense{
		{Date: "2026-05-21", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-21", Amount: 12.00, Category: "lunch"},
		{Date: "2026-05-22", Amount: 80.00, Category: "groceries"},
		{Date: "2026-05-22", Amount: 3.50, Category: "coffee"},
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	var buf bytes.Buffer
	if err := cmdSummary(s, &buf); err != nil {
		t.Fatalf("cmdSummary: %v", err)
	}

	testutil.AssertGolden(t, buf.Bytes(), filepath.Join("testdata", "summary.golden"))
}
```

- [ ] **Step 8:** Create `lessons/15-structure/exercises/cmd/expenses/testdata/summary.golden` (placeholder — students would generate with `-update`):

```
4 expenses, total €100.00
by category:
  coffee     €8.00   
  groceries  €80.00  
  lunch      €12.00  
biggest: groceries €80.00
  ████████████████████
```

Plus the SAME content at `lessons/15-structure/solutions/cmd/expenses/testdata/summary.golden`.

> Note on the exact expected content: the column widths and percentages depend on the cmdSummary print formatting. The solution test will catch any drift via the golden file comparison. If the format changes during implementation, run `go test -update ./...` once and commit the regenerated golden file. The bar `████████████████████` is 20 block characters (width=20 in `summary.Bar`).

### Step 9: Verify and commit

- [ ] **Step 9:** Verify and commit Task 4:

```bash
gofmt -l lessons/15-structure/
go test ./lessons/15-structure/exercises/cmd/... -v 2>&1 | tail -10
go test ./lessons/15-structure/solutions/cmd/... -v 2>&1 | tail -20
make test
golangci-lint run ./...
go vet ./...

# Smoke test the binaries
TMP=$(mktemp -d /tmp/lesson15-XXXXXX)
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/tracker.json" add 2026-05-21 4.50 coffee
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/tracker.json" summary
echo "2026-05-21,12.00,lunch" | go run ./lessons/15-structure/solutions/cmd/expenses-import -file="$TMP/imported.json"
rm -rf "$TMP"

git add lessons/15-structure/exercises/cmd/ lessons/15-structure/solutions/cmd/
git commit -m "feat(lesson-15): cmd/* — expenses (golden file test) + expenses-import (verbatim L13)"
```

Expected: gofmt empty; exercises pass vacuously (skeleton); solutions all PASS (including golden file test); smoke shows binaries working.

---

## Task 5: Slide deck — 5 concepts

Heavy-explanatory pattern (Motivation/Basics/Worked-example/Common-mistake/Recap) per concept.

**File:** `lessons/15-structure/slides/slides.md`

~32 slides. Concept order matches design:

1. **`cmd/` + `internal/`** — project layout. cmd/ recap (L07-L08); internal/ as new material with compiler enforcement.
2. **Test helpers** — `t.Run` subtests, `t.Helper()`, `t.TempDir()`.
3. **Golden file tests** — testdata/.golden + `-update` flag pattern.
4. **Benchmarks** — `testing.B`, `b.N`, `b.ResetTimer`, `b.ReportAllocs`.
5. **Fuzz tests** — `testing.F`, `f.Add`, `f.Fuzz`; `go test -fuzz=`.

- [ ] **Step 1-7:** Author the deck. Cover → roadmap → 5 concepts (5 slides each) → Practice → Closing thought (Phase 2 wrap-up — major recap of all 7 lessons) → What's next (Phase 3 preview).

- [ ] **Step 8:** Verify and commit:

```bash
make slides-build
grep -q "15-structure" dist/index.html && echo "✓ 15-structure in index"
rm -rf dist

git add lessons/15-structure/slides/
git commit -m "feat(lesson-15): slides — Project structure & testing patterns (5 concepts)"
```

---

## Task 6: README

Mirrors slide concepts (5 sections); one Common-mistake paragraph per concept; gofmt/vet daily-habits; "Phase 2 wrap-up" section pointing forward to Phase 3.

**File:** `lessons/15-structure/README.md`

- [ ] **Step 1:** Author the README (~270 lines):
  - "What you'll learn" — 5 bullets.
  - "What's different from L14" — Phase 2 capstone. Brings back tracker, restructures into v2 cmd/+internal/. Introduces internal/ formally, golden file tests, benchmarks, fuzz preview.
  - "The v2 layout" — annotated tree showing cmd/ + internal/.
  - "What you're (mostly) NOT writing" — most code is mechanical carry-forward. The substantive new code is the golden-file integration test, the benchmark, and the fuzz function.
  - Per-concept sections (5).
  - "Exercise: warmup — internal/util demo" (~3 lines).
  - "Exercise: main — v2 reorg" (~10 lines).
  - "Daily habits" + "How to run" + "Going further" (Read: Go blog on testdata convention, the testing package docs; Try: extend the golden file pattern to other commands; benchmark a hot path elsewhere in the tracker).

- [ ] **Step 2:** Commit:

```bash
git add lessons/15-structure/README.md
git commit -m "docs(lesson-15): README — Project structure & testing capstone"
```

---

## Task 7: End-to-end verification

- [ ] **Step 1:** Full sweep.

```bash
make test
```

Expected: all lessons (01-15) + tools pass.

- [ ] **Step 2:** Solution-only verbose for lesson 15.

```bash
go test -v ./lessons/15-structure/solutions/...
```

Expected (~40+ sub-tests, the most of any lesson):
- `util.TestMoney` — 5 sub-tests PASS
- `expense.*` — carry-forward L09 tests PASS
- `store.*` — carry-forward L13 tests PASS
- `csvimport.TestParseGoldenPath` (5), `TestParseMalformed` (3), `FuzzParseLine` (seeds only — runs as 3 cases) PASS
- `summary.*` — carry-forward L08 tests PASS
- `cmd/expenses.TestSummaryGolden` PASS (golden file compared)
- `cmd/expenses-import.TestImporter*` PASS (3 integration tests)

- [ ] **Step 3:** Exercises pass vacuously.

```bash
go test ./lessons/15-structure/exercises/...
```

- [ ] **Step 4:** Benchmark sanity (1 iteration to verify it compiles + runs).

```bash
go test -bench=BenchmarkTotalsByCategory -benchtime=1x ./lessons/15-structure/solutions/internal/summary/
```

Expected output mentions `BenchmarkTotalsByCategory` and `ns/op`.

- [ ] **Step 5:** Fuzz sanity (seeds only — no `-fuzz` flag, runs as normal test).

```bash
go test -run FuzzParseLine ./lessons/15-structure/solutions/internal/csvimport/
```

Expected: ok (the seeds run without panicking).

- [ ] **Step 6:** Static analysis.

```bash
go vet ./...
golangci-lint run ./...
gofmt -l lessons/15-structure/
```

Expected: clean.

- [ ] **Step 7:** Slides build.

```bash
make slides-build
grep -q "15-structure" dist/index.html
ls dist/lessons/15-structure/  # should exist
rm -rf dist
```

- [ ] **Step 8:** Binary smoke.

```bash
TMP=$(mktemp -d /tmp/lesson15-XXXXXX)

# cmd/expenses
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/t.json" add 2026-05-21 4.50 coffee
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/t.json" add 2026-05-21 12.00 lunch
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/t.json" list
go run ./lessons/15-structure/solutions/cmd/expenses -file="$TMP/t.json" summary

# cmd/expenses-import
echo "2026-05-22,80.00,groceries" | go run ./lessons/15-structure/solutions/cmd/expenses-import -file="$TMP/i.json"
cat "$TMP/i.json"

# warmup
go run ./lessons/15-structure/solutions/warmup

rm -rf "$TMP"
```

Expected: list shows two expenses; summary shows totals + biggest + bar; importer writes the JSON; warmup prints "€4.50".

- [ ] **Step 9:** Final code review subagent → push branch → create PR.

```bash
git push -u origin feature/plan-r-lesson-15-structure-capstone
gh pr create --title "feat(lesson-15): Project structure capstone — cmd/+internal/, golden files, benchmarks, fuzz" --body "..."
```

---

## Verification

After all 7 tasks:

```bash
make test                                                                              # green
go test -v ./lessons/15-structure/solutions/... 2>&1 | grep -E "PASS|FAIL" | wc -l    # 40+ PASS lines
go vet ./...                                                                           # clean
golangci-lint run ./...                                                                # 0 issues
gofmt -l lessons/15-structure/                                                         # empty
make slides-build && grep -q "15-structure" dist/index.html && rm -rf dist            # green
# Benchmark + fuzz sanity already covered above.
```

## Critical file paths

To be created:

- `lessons/15-structure/` (directory)
- `lessons/15-structure/README.md`
- `lessons/15-structure/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/15-structure/exercises/warmup/{main.go, internal/util/{format.go, format_test.go}}`
- `lessons/15-structure/exercises/cmd/expenses/{main.go, main_test.go, testdata/summary.golden}`
- `lessons/15-structure/exercises/cmd/expenses-import/{main.go, main_test.go}`
- `lessons/15-structure/exercises/internal/expense/{expense.go, expense_test.go}`
- `lessons/15-structure/exercises/internal/store/{store.go, store_test.go}`
- `lessons/15-structure/exercises/internal/csvimport/{csvimport.go, csvimport_test.go, csvimport_fuzz_test.go}`
- `lessons/15-structure/exercises/internal/summary/{summary.go, summary_test.go, summary_bench_test.go}`
- `lessons/15-structure/exercises/internal/testutil/golden.go`
- `lessons/15-structure/solutions/...` (mirrored)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-18-phase-2-idiomatic-go-design.md` (Lesson 15 section)
- Carry-forward sources listed in "Conventions" above
- `tools/build-index/main.go` (L15 slug `structure` already correct — no changes)
- `.golangci.yml` (no changes)
