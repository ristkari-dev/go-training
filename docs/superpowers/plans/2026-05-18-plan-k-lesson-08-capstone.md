# Plan K — Lesson 08 (Phase 1 capstone — expense tracker CLI) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the eighth and final lesson of Phase 1 — the capstone. Students integrate everything they've learned (structs, methods, packages, modules, errors, tests) into a working expense-tracker CLI with three subcommands (`add`, `list`, `summary`), JSON-backed persistence via a provided `storage` package, and integration tests that drive the binary via `os/exec` against a temp JSON file. The lesson also formally introduces the `cmd/` convention (which we deliberately skipped in lesson 07).

**Architecture:** Same per-lesson pattern as plans D-J in spirit, but **10 tasks** instead of 6 because the capstone has substantially more files (24 vs the usual 12-14). Each task has one focused output: scaffold + restructure, warmup (clock + cmd/timestamp), expense subpackage, summary subpackage, storage subpackage (PROVIDED, fully implemented), cmd/expenses/main.go (orchestration), integration tests, slides, README, e2e verify. Skeleton tests in the warmup + expense + summary subpackages (continues lesson 05/06/07 pattern). Full integration tests in `cmd/expenses/main_test.go` (per user choice).

**Tech Stack:** Go 1.23 stdlib only (`encoding/json`, `errors`, `fmt`, `os`, `os/exec`, `path/filepath`, `sort`, `strings`, `testing`, `time`). Reveal.js 5.1.0 for the deck.

---

## Scope

Plan K produces lesson 08 only. After Plan K lands:

- `lessons/08-capstone/` is a complete teachable lesson with all four parts populated.
- The directory contains **subfolders within subfolders** for the first time — `exercises/cmd/expenses/`, `exercises/warmup/cmd/timestamp/`, plus three top-level subpackages (`expense/`, `summary/`, `storage/`).
- `make test` passes; `make test-exercises` shows lesson 08's exercise tests passing vacuously (skeleton case slices empty).
- `make slides-dev LESSON=08-capstone` serves the deck.
- `dist/index.html` lists lesson 08 alongside lessons 01-07.
- `go run ./lessons/08-capstone/solutions/cmd/expenses -file=/tmp/expenses.json add 2026-05-18 4.50 coffee` (and other subcommands) work end-to-end.

**Phase 1 COMPLETE after this plan ships.**

**Out of scope:** Phase 2 (lessons 09-15) and beyond.

### Design decisions made during planning

1. **Bar chart in summary output (per user choice).** `summary.Bar(value, max float64, width int) string` produces a string of Unicode block characters proportional to `value/max`. Visible payoff of the capstone — students see their expenses charted in their own terminal.

2. **Full integration tests (per user choice).** `cmd/expenses/main_test.go` uses `os/exec` + `t.TempDir()` to build and run the binary against a temp JSON file. Three integration tests: `add`, `add → list`, `add → summary`. Skeleton-style — the test scaffolding (build the binary once via `os/exec`, prepare a temp directory) ships pre-written; the per-subcommand assertions are TODOs students fill in.

3. **Warmup is spec-literal `clock` + `cmd/timestamp` (per user choice).** Tiny `clock` package with `Now() string`; tiny `cmd/timestamp/main.go` binary that prints it. The warmup mirrors the main exercise's structure (small package + `cmd/<name>/main.go`) — formal introduction of the `cmd/` convention before students see it in the bigger `cmd/expenses/main.go`.

4. **`storage` package is fully provided, 100% black-box.** Students don't write or modify `storage.go`. The file's top comment says "Treat as a black box for now — we'll see how this works in Phase 2 (lesson 13)." This is consistent with the Phase 1 spec's "stdlib only except encoding/json which is opaque" rule.

5. **`expense` package re-shipped (not imported from lesson 07).** Lesson 08's `exercises/expense/expense.go` ships the same stubs as lesson 07's. Students re-implement (or copy from their lesson 07 work). This keeps lesson 08 self-contained — no cross-lesson imports, no fragile dependence on what students did in earlier lessons. The README explicitly says "you may copy from your lesson 07 work".

6. **Three concepts in the slide deck.** Smaller than the usual 4 because lesson 08 is mostly applied — most slides are walking through the capstone's architecture, not teaching new language features. The three concepts:
   1. **The `cmd/` convention** — package main lives at `cmd/<binaryname>/main.go`; the binary name comes from the directory name; multiple binaries per module live as siblings under `cmd/`.
   2. **Composing multiple packages into a CLI** — orchestration code in main.go calls into focused packages. The "deps go one way" rule (cmd depends on summary which depends on expense; storage depends on expense).
   3. **End-to-end integration testing** — the `os/exec` + `t.TempDir()` pattern; testing a binary by running it; assertions on stdout / exit code / temp-file contents.

7. **Phase 1 capstone summary slide.** A standalone slide near the end of the deck that recaps what students have learned across all 8 lessons. Not a concept; a closing.

8. **No `flag` package.** Spec is explicit: `os.Args` is enough for three subcommands. `flag` introduces complexity (long options, defaults, help text) that isn't needed here and that lesson 26 covers properly.

9. **`-file=path` is the only CLI option.** Default is `~/.expenses.json` (resolved via `os.UserHomeDir()`). The option may appear before or after the subcommand: both `expenses -file=x.json add ...` and `expenses add ... -file=x.json` could work, but the spec doesn't require both. Plan K's main.go supports `-file=` as the FIRST argument only (before the subcommand) — simplest to parse without `flag`. README documents this limitation.

10. **Test fixtures use absolute dates.** `2026-05-18` (today), `2026-05-17` (yesterday), etc. — consistent with previous lessons. No `time.Now()`-relative dates in tests (those would be flaky).

---

## Plans F-J lessons-learned applied here

1. **Warmup* prefix unnecessary** — subpackages provide namespacing. `clock.Now()` and `expense.Expense` are clean.

2. **Lint exclusion covers nested subpackages** — `.golangci.yml`'s `lessons/.*/exercises/` regex matches `lessons/08-capstone/exercises/cmd/expenses/main.go` (path contains `/exercises/`). Verify in task 1 by running lint after the scaffold.

3. **`→` arrow consistency** — Unicode arrows in doc-comment examples; gofmt may reformat blocks but arrows must survive.

4. **Common-mistake content in README** — each slide concept's common-mistake example is mirrored in the README. 3 total (one per concept).

5. **Slides + README written inline by controller** — Plans G/H/I/J established this. Plan K continues. Tasks 1-7 can be subagent-driven; Tasks 8-9 the controller writes directly.

6. **`gofmt -w .` and `go vet ./...` mentions** — README continues both habits.

7. **Format-output verification** — Plan G/I/J's `Expense.Format()` byte-output is reused; tested cases are known-good.

8. **Empty case → vacuous pass.** Continues. Exercise tests pass vacuously until students add cases.

9. **Task count scales with file count.** Plans D-I had 6 tasks for 12 files. Plan J had 6 tasks for 14 files. Plan K has 10 tasks for 24 files — file count nearly doubled, task count not quite doubled (some files are tightly coupled and live in the same task).

---

## File Structure

After Plan K (24 files total):

```
lessons/08-capstone/
├── README.md                                  (Task 9)
├── slides/
│   ├── index.html                             (Task 1; unchanged from scaffold)
│   ├── slides.md                              (Task 8 — controller writes inline)
│   └── assets/.gitkeep                        (Task 1)
├── exercises/
│   ├── warmup/
│   │   ├── clock/
│   │   │   ├── clock.go                       (Task 2 — Now() stub)
│   │   │   └── clock_test.go                  (Task 2 — SKELETON test)
│   │   └── cmd/
│   │       └── timestamp/
│   │           └── main.go                    (Task 2 — driver, working)
│   ├── expense/
│   │   ├── expense.go                         (Task 3 — Expense + Format + IsHigh stubs)
│   │   └── expense_test.go                    (Task 3 — SKELETON tests)
│   ├── summary/
│   │   ├── summary.go                         (Task 4 — TotalsByCategory + BiggestCategory + Bar stubs)
│   │   └── summary_test.go                    (Task 4 — SKELETON tests)
│   ├── storage/
│   │   └── storage.go                         (Task 5 — PROVIDED, full implementation)
│   └── cmd/
│       └── expenses/
│           ├── main.go                        (Task 6 — package main orchestration)
│           └── main_test.go                   (Task 7 — SKELETON integration tests)
└── solutions/
    └── (same structure, all implemented; main_test.go has full reference)
```

Files deleted from the scaffolder output (8 total):

```
exercises/main.go
exercises/main_test.go
exercises/warmup.go
exercises/warmup_test.go
solutions/main.go
solutions/main_test.go
solutions/warmup.go
solutions/warmup_test.go
```

The scaffolder produces 12 flat files; all 8 `.go` files at the `exercises`/`solutions` top level are deleted. Only `README.md`, `slides/*` survive from the scaffold. The 20 lesson `.go` files are created from scratch across tasks 2-7.

### Decomposition rationale

Same per-task focus as previous plans, but **10 tasks** to keep each subagent prompt bounded. The biggest tasks (Task 6 = cmd/expenses/main.go orchestration, Task 7 = integration tests) are still each focused on a single file per side.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-k-lesson-08-capstone` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold + restructure

The scaffolder produces 12 flat files. Lesson 08 needs a completely different shape — 8 of those files get deleted; the subpackage trees are created from scratch in tasks 2-7. Task 1 just runs the scaffolder and deletes the unwanted files, leaving README.md + slides/.

**Files:**
- Create via scaffolder: `lessons/08-capstone/` (12 files initially)
- Delete: 8 `.go` files at the flat `exercises`/`solutions` level

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-k-lesson-08-capstone`, clean working tree, exactly one commit ahead of `main` (the Plan K doc).

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=08-capstone
```

Expected: `created lesson 08-capstone under lessons/`. 12 placeholder files produced.

- [ ] **Step 3: Delete the unwanted scaffolder files**

```bash
rm lessons/08-capstone/exercises/main.go \
   lessons/08-capstone/exercises/main_test.go \
   lessons/08-capstone/exercises/warmup.go \
   lessons/08-capstone/exercises/warmup_test.go \
   lessons/08-capstone/solutions/main.go \
   lessons/08-capstone/solutions/main_test.go \
   lessons/08-capstone/solutions/warmup.go \
   lessons/08-capstone/solutions/warmup_test.go
```

Expected: 8 files removed.

- [ ] **Step 4: Verify the surviving tree**

```bash
find lessons/08-capstone -type f | sort
```

Expected (4 files):

```
lessons/08-capstone/README.md
lessons/08-capstone/slides/assets/.gitkeep
lessons/08-capstone/slides/index.html
lessons/08-capstone/slides/slides.md
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: every package passes. The lessons/08-capstone/{exercises,solutions} dirs no longer have any Go files; `go test` ignores them.

- [ ] **Step 6: Commit**

```bash
git add -A lessons/08-capstone/
git commit -m "feat(lessons): scaffold lesson 08-capstone with empty subpackage layout"
```

> Note: `git add -A` is needed (not just `git add`) because we deleted files; -A captures both creations and deletions.

- [ ] **Step 7: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch.

---

## Task 2: Author the warm-up — `clock` package + `cmd/timestamp` binary

The warmup is the lightest possible demonstration of the `cmd/` convention: a tiny library package (`clock`) with one function (`Now`), and a tiny binary (`cmd/timestamp/main.go`) that imports it and prints the result.

**Files:**
- Create: `lessons/08-capstone/exercises/warmup/clock/clock.go`
- Create: `lessons/08-capstone/exercises/warmup/clock/clock_test.go`
- Create: `lessons/08-capstone/exercises/warmup/cmd/timestamp/main.go`
- Create: `lessons/08-capstone/solutions/warmup/clock/clock.go`
- Create: `lessons/08-capstone/solutions/warmup/clock/clock_test.go`
- Create: `lessons/08-capstone/solutions/warmup/cmd/timestamp/main.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/08-capstone/exercises/warmup/clock \
         lessons/08-capstone/exercises/warmup/cmd/timestamp \
         lessons/08-capstone/solutions/warmup/clock \
         lessons/08-capstone/solutions/warmup/cmd/timestamp
```

- [ ] **Step 2: Create `lessons/08-capstone/exercises/warmup/clock/clock.go`** with:

```go
// Package clock provides a tiny time helper for the lesson 08 warm-up.
//
// This is the simplest possible "publish a function from a subpackage"
// demo. The matching cmd/timestamp/main.go imports clock and prints
// the result of Now().
package clock

// Now returns the current UTC time formatted as "2006-01-02 15:04:05".
//
// Examples:
//
//	clock.Now()  → "2026-05-18 14:30:00"   (whatever the current UTC time is)
//
// Hint: import "time", then time.Now().UTC().Format("2006-01-02 15:04:05").
// The string "2006-01-02 15:04:05" is Go's reference time — a quirky way
// to specify date/time formats. See pkg.go.dev/time#Time.Format if curious.
func Now() string {
	panic("TODO: return time.Now().UTC().Format(\"2006-01-02 15:04:05\")")
}
```

- [ ] **Step 3: Create `lessons/08-capstone/exercises/warmup/clock/clock_test.go`** with (SKELETON):

```go
package clock

import "testing"

// TestNow is a SKELETON. Fill in the assertion body.
//
// Time-dependent tests are tricky — Now() returns the current time,
// which differs every call. The simplest correctness check is the FORMAT:
//
//   - The returned string is exactly 19 characters long ("YYYY-MM-DD HH:MM:SS").
//   - Characters at positions 4 and 7 are '-'; position 10 is ' ';
//     positions 13 and 16 are ':'.
//
// We don't check the actual time — only that the format is right. A more
// rigorous test would inject a "now" function as a dependency (lesson 10's
// interfaces unlock that pattern); for Phase 1, format-only is enough.
func TestNow(t *testing.T) {
	got := Now()

	// TODO: assert len(got) == 19; t.Fatalf if not.
	// TODO: assert got[4] == '-' && got[7] == '-' && got[10] == ' ' &&
	//        got[13] == ':' && got[16] == ':'; t.Errorf for each.
	_ = got
}
```

- [ ] **Step 4: Create `lessons/08-capstone/exercises/warmup/cmd/timestamp/main.go`** with (working, no panic-stub):

```go
// Package main is a tiny binary that prints the current time via the
// clock package. Demonstrates the cmd/<binaryname>/main.go convention.
//
// Run with:
//
//	go run ./lessons/08-capstone/exercises/warmup/cmd/timestamp
//
// Once clock.Now() is implemented, output is the current UTC time formatted
// as "2026-05-18 14:30:00" (whatever the current time is).
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/warmup/clock"
)

func main() {
	fmt.Println(clock.Now())
}
```

- [ ] **Step 5: Create `lessons/08-capstone/solutions/warmup/clock/clock.go`** with the implementation:

```go
// Package clock is the lesson 08 warm-up reference implementation.
package clock

import "time"

// Now returns the current UTC time formatted as "2006-01-02 15:04:05".
func Now() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}
```

- [ ] **Step 6: Create `lessons/08-capstone/solutions/warmup/clock/clock_test.go`** with the full reference test:

```go
package clock

import "testing"

func TestNow(t *testing.T) {
	got := Now()

	if len(got) != 19 {
		t.Fatalf("Now() = %q (len %d), want a 19-char string", got, len(got))
	}
	checks := []struct {
		pos int
		ch  byte
	}{
		{4, '-'},
		{7, '-'},
		{10, ' '},
		{13, ':'},
		{16, ':'},
	}
	for _, c := range checks {
		if got[c.pos] != c.ch {
			t.Errorf("Now()[%d] = %q, want %q (full output: %q)", c.pos, got[c.pos], c.ch, got)
		}
	}
}
```

- [ ] **Step 7: Create `lessons/08-capstone/solutions/warmup/cmd/timestamp/main.go`** with (same as exercises, importing from solutions):

```go
// Package main is the timestamp binary (solutions reference).
package main

import (
	"fmt"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/warmup/clock"
)

func main() {
	fmt.Println(clock.Now())
}
```

- [ ] **Step 8: Run gofmt**

```bash
gofmt -l lessons/08-capstone/
```

Expected: empty.

- [ ] **Step 9: Verify warm-up tests**

```bash
go test ./lessons/08-capstone/exercises/warmup/... -v 2>&1 | tail -10
go test ./lessons/08-capstone/solutions/warmup/... -v 2>&1 | tail -10
```

Expected: exercises TestNow PASS vacuously (the body just has `_ = got` which compiles); solutions TestNow PASS (validates the format).

- [ ] **Step 10: Run the solutions timestamp binary**

```bash
go run ./lessons/08-capstone/solutions/warmup/cmd/timestamp
```

Expected: a single line like `2026-05-18 14:30:00` (current UTC time).

- [ ] **Step 11: `make test`, lint, vet clean**

```bash
make test
golangci-lint run ./...
go vet ./...
```

Expected: green, `0 issues.`, clean vet.

- [ ] **Step 12: Commit**

```bash
git add lessons/08-capstone/exercises/warmup/ lessons/08-capstone/solutions/warmup/
git commit -m "feat(lesson-08): warmup — clock package + cmd/timestamp binary"
```

---

## Task 3: Author the `expense` subpackage

Carry forward lesson 07's expense package. The stub shape is identical to lesson 07's; students copy from their lesson 07 work (or re-implement).

**Files:**
- Create: `lessons/08-capstone/exercises/expense/expense.go`
- Create: `lessons/08-capstone/exercises/expense/expense_test.go`
- Create: `lessons/08-capstone/solutions/expense/expense.go`
- Create: `lessons/08-capstone/solutions/expense/expense_test.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/08-capstone/exercises/expense lessons/08-capstone/solutions/expense
```

- [ ] **Step 2: Create `lessons/08-capstone/exercises/expense/expense.go`** with:

```go
// Package expense holds the Expense type + methods for lesson 08's capstone.
//
// This is the same code you wrote in lesson 07 — you may copy from
// lessons/07-packages/exercises/expense/ or re-implement here. Lesson 08's
// summary, storage, and cmd/expenses packages all import expense.Expense.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// All three fields are exported because the test file, the summary and
// storage subpackages, and the CLI binary all construct Expense values
// with struct literals or read them from JSON.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04/06/07's Format.
//
// Examples:
//
//	Expense{Date: "2026-05-18", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-18  €4.50    coffee"
//	Expense{Date: "2026-05-18", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-18  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Boundary: amount of exactly 50 is NOT high; 50.01 IS high.
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}
```

> Note: the JSON struct tags (`` `json:"date"` `` etc) are needed for the `storage` package's `encoding/json` calls. They map struct fields to lowercase JSON keys. We don't dive into struct tags in Phase 1 — treat them as "magic strings that make JSON work."

- [ ] **Step 3: Create `lessons/08-capstone/exercises/expense/expense_test.go`** with (SKELETON):

```go
package expense

import "testing"

// TestExpenseFormat is a SKELETON. Same shape as lesson 07's test —
// fill in cases for small/medium/large amounts.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases. Use lesson 07's solutions/expense/expense_test.go
		// as a reference if needed.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover the boundary at 50 explicitly.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases including 49.99, 50 (false), 50.01 (true).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO
			_ = tc
		})
	}
}
```

- [ ] **Step 4: Create `lessons/08-capstone/solutions/expense/expense.go`** with:

```go
// Package expense is the lesson 08 capstone reference implementation of Expense.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}
```

- [ ] **Step 5: Create `lessons/08-capstone/solutions/expense/expense_test.go`** with full reference tests:

```go
package expense

import "testing"

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-18", Amount: 4.50, Category: "coffee"}, "2026-05-18  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-18", Amount: 12, Category: "lunch"}, "2026-05-18  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-18", Amount: 999.99, Category: "rent"}, "2026-05-18  €999.99  rent"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.Format(); got != tc.want {
				t.Errorf("%+v.Format() =\n  %q\nwant\n  %q", tc.e, got, tc.want)
			}
		})
	}
}

func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		{"zero", 0, false},
		{"small", 4.50, false},
		{"just-under-50", 49.99, false},
		{"exactly-50", 50, false},
		{"just-over-50", 50.01, true},
		{"large", 999, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := (Expense{Amount: tc.amount}).IsHigh(); got != tc.want {
				t.Errorf("Expense{Amount: %g}.IsHigh() = %v, want %v", tc.amount, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 6: gofmt, run tests, lint clean**

```bash
gofmt -l lessons/08-capstone/
go test ./lessons/08-capstone/exercises/expense/... -v 2>&1 | tail -10
go test ./lessons/08-capstone/solutions/expense/... -v 2>&1 | tail -20
make test
golangci-lint run ./...
```

Expected: gofmt empty; exercises pass vacuously; solutions pass 9 sub-tests (TestExpenseFormat 3 + TestExpenseIsHigh 6); make test green; lint 0 issues.

- [ ] **Step 7: Commit**

```bash
git add lessons/08-capstone/exercises/expense/ lessons/08-capstone/solutions/expense/
git commit -m "feat(lesson-08): expense subpackage (forward-port from lesson 07)"
```

---

## Task 4: Author the `summary` subpackage

Three pure functions: `TotalsByCategory` (same as lesson 06/07 but takes `[]expense.Expense`), `BiggestCategory`, and `Bar` (the visual text bar chart).

**Files:**
- Create: `lessons/08-capstone/exercises/summary/summary.go`
- Create: `lessons/08-capstone/exercises/summary/summary_test.go`
- Create: `lessons/08-capstone/solutions/summary/summary.go`
- Create: `lessons/08-capstone/solutions/summary/summary_test.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/08-capstone/exercises/summary lessons/08-capstone/solutions/summary
```

- [ ] **Step 2: Create `lessons/08-capstone/exercises/summary/summary.go`** with:

```go
// Package summary provides aggregation and visualisation helpers for
// the lesson 08 capstone CLI. Three pure functions, no I/O.
package exercises_summary

import (
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// TotalsByCategory returns a map of category → summed amount, built from es.
// Empty/nil input returns a non-nil empty map.
//
// Same behaviour as lesson 06/07's TotalsByCategory — repeated here as
// it's the natural fit for the capstone's "summary" command.
//
// Examples:
//
//	es := []expense.Expense{
//		{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-18", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-18", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]
//
// Hint: same as before — initialise totals := map[string]float64{}, range
// over es, accumulate.
func TotalsByCategory(es []expense.Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into a map[string]float64")
}

// BiggestCategory returns the category with the largest total.
//
// If totals is empty or nil, BiggestCategory returns ("", 0) — there's no
// "biggest" of nothing.
//
// If two categories tie for biggest, BiggestCategory returns the one
// whose name sorts alphabetically first (so the output is deterministic
// regardless of map iteration order).
//
// Examples:
//
//	BiggestCategory(map[string]float64{"coffee": 14.49, "rent": 75})
//	  → ("rent", 75)
//	BiggestCategory(map[string]float64{"a": 10, "b": 10})
//	  → ("a", 10)  (alphabetical tie-break)
//	BiggestCategory(map[string]float64{})    → ("", 0)
//	BiggestCategory(nil)                      → ("", 0)
//
// Hint: range over the map keys, sort them to make iteration deterministic,
// then walk sorted keys tracking the running maximum.
func BiggestCategory(totals map[string]float64) (string, float64) {
	panic("TODO: return the category with the largest total; alphabetical tie-break; (\"\", 0) for empty")
}

// Bar returns a string of up to `width` Unicode block characters
// proportional to value/max. Uses '█' (U+2588 FULL BLOCK) for full
// units and '▌' (U+258C LEFT HALF BLOCK) for half units.
//
// The result has between 0 and `width` chars. A half-block appears when
// the fractional remainder is >= 0.5 of a unit.
//
// Returns "" when:
//   - max <= 0 (can't compute a ratio)
//   - value <= 0 (no bar to draw)
//   - width <= 0 (no room)
//
// Examples (width=8):
//
//	Bar(0, 10, 8)       → ""
//	Bar(10, 10, 8)      → "████████"        (full)
//	Bar(5, 10, 8)       → "████"            (half → 4 full chars)
//	Bar(2.5, 10, 8)     → "██"              (quarter → 2 full)
//	Bar(0.5, 8, 8)      → "▌"               (just half)
//	Bar(1, 8, 8)        → "█"               (1 full)
//
// Hint: units := value / max * float64(width); full := int(units);
// half := units - float64(full) >= 0.5;
// result := strings.Repeat("█", full); if half { result += "▌" }.
func Bar(value, max float64, width int) string {
	_ = strings.Repeat // keep the import compiling until you use it
	panic("TODO: build the bar string per the doc comment")
}
```

> Note: the package declaration is `package exercises_summary`, not `package summary`. This is a small workaround for a Go limitation — the directory is named `summary/` but if we named the package `summary`, it would conflict with the solutions/summary package when both are loaded together. The trick is: each side uses a unique package name (`exercises_summary` vs `solutions_summary`), but both live in directories named `summary`. The cmd/expenses/main.go in each side imports the matching one. Lesson 07 didn't hit this because `greet` and `expense` subpackages had unique imports per side; the cmd binaries imported only their own side. Same here, technically — let me reconsider.
>
> Actually: Go doesn't require the package name to match the directory name (just convention). And since cmd/expenses/main.go in exercises/ imports `.../exercises/summary` and the one in solutions/ imports `.../solutions/summary`, the package names *can* both be `summary` — they're never loaded into the same binary. The previous note is wrong — use `package summary` for both. **Revised below in Step 2 of solutions; please use `package summary` in exercises too.**

> **CORRECTION: use `package summary` (not `exercises_summary`) — both sides use the same package name since they're imported by different binaries. Same as lesson 07's `greet` and `expense` packages.**

Replace the `package exercises_summary` line above with `package summary`.

- [ ] **Step 3: Create `lessons/08-capstone/exercises/summary/summary_test.go`** with (SKELETON):

```go
package summary

import (
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// TestTotalsByCategory is a SKELETON. Same shape as lesson 05/06/07's
// TotalsByCategory test.
func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []expense.Expense
		want map[string]float64
	}{
		// TODO: at least 4 cases (distinct, repeated, single, empty).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := TotalsByCategory(tc.es)
			//   assert got != nil; reflect.DeepEqual(got, tc.want)
			_ = tc
			_ = reflect.DeepEqual
		})
	}
}

// TestBiggestCategory is a SKELETON. Cover at least:
//   - one category (trivial maximum)
//   - several categories with a clear maximum
//   - a tie (assert alphabetical tie-break)
//   - empty map (returns ("", 0))
//   - nil map (returns ("", 0))
func TestBiggestCategory(t *testing.T) {
	cases := []struct {
		name      string
		totals    map[string]float64
		wantName  string
		wantTotal float64
	}{
		// TODO: at least 5 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: gotName, gotTotal := BiggestCategory(tc.totals)
			//   assert gotName == tc.wantName && gotTotal == tc.wantTotal
			_ = tc
		})
	}
}

// TestBar is a SKELETON. Cover at least:
//   - full bar (value == max)
//   - empty bar (value == 0)
//   - half-block on the trailing position
//   - max <= 0 returns ""
//   - width <= 0 returns ""
func TestBar(t *testing.T) {
	cases := []struct {
		name   string
		value  float64
		max    float64
		width  int
		want   string
	}{
		// TODO: at least 5 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Bar(tc.value, tc.max, tc.width)
			//   assert got == tc.want
			_ = tc
		})
	}
}
```

- [ ] **Step 4: Create `lessons/08-capstone/solutions/summary/summary.go`** with:

```go
// Package summary is the lesson 08 capstone reference implementation.
package summary

import (
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
)

// TotalsByCategory returns a map of category → summed amount.
// Empty/nil input returns a non-nil empty map.
func TotalsByCategory(es []expense.Expense) map[string]float64 {
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}
	return totals
}

// BiggestCategory returns the category with the largest total. Alphabetical
// tie-break. ("", 0) for empty/nil input.
func BiggestCategory(totals map[string]float64) (string, float64) {
	if len(totals) == 0 {
		return "", 0
	}
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	bestName := keys[0]
	bestTotal := totals[bestName]
	for _, k := range keys[1:] {
		if totals[k] > bestTotal {
			bestName = k
			bestTotal = totals[k]
		}
	}
	return bestName, bestTotal
}

// Bar returns a string of up to `width` Unicode block characters
// proportional to value/max.
func Bar(value, max float64, width int) string {
	if max <= 0 || value <= 0 || width <= 0 {
		return ""
	}
	units := value / max * float64(width)
	full := int(units)
	half := units-float64(full) >= 0.5
	result := strings.Repeat("█", full)
	if half {
		result += "▌"
	}
	return result
}
```

- [ ] **Step 5: Create `lessons/08-capstone/solutions/summary/summary_test.go`** with full reference tests:

```go
package summary

import (
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
)

func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []expense.Expense
		want map[string]float64
	}{
		{
			"distinct",
			[]expense.Expense{
				{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-18", Amount: 12, Category: "lunch"},
				{Date: "2026-05-18", Amount: 75, Category: "rent"},
			},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
		},
		{
			"repeated",
			[]expense.Expense{
				{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-18", Amount: 12, Category: "lunch"},
				{Date: "2026-05-18", Amount: 9.99, Category: "coffee"},
			},
			map[string]float64{"coffee": 14.49, "lunch": 12},
		},
		{"single", []expense.Expense{{Amount: 42, Category: "x"}}, map[string]float64{"x": 42}},
		{"empty", []expense.Expense{}, map[string]float64{}},
		{"nil", nil, map[string]float64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TotalsByCategory(tc.es)
			if got == nil {
				t.Fatalf("TotalsByCategory returned nil; expected non-nil map")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBiggestCategory(t *testing.T) {
	cases := []struct {
		name      string
		totals    map[string]float64
		wantName  string
		wantTotal float64
	}{
		{"single", map[string]float64{"a": 10}, "a", 10},
		{"clear-max", map[string]float64{"coffee": 14.49, "rent": 75, "lunch": 12}, "rent", 75},
		{"tie-alphabetical", map[string]float64{"a": 10, "b": 10}, "a", 10},
		{"tie-three-way", map[string]float64{"c": 5, "b": 5, "a": 5}, "a", 5},
		{"empty", map[string]float64{}, "", 0},
		{"nil", nil, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotTotal := BiggestCategory(tc.totals)
			if gotName != tc.wantName || gotTotal != tc.wantTotal {
				t.Errorf("BiggestCategory(%v) = (%q, %g), want (%q, %g)",
					tc.totals, gotName, gotTotal, tc.wantName, tc.wantTotal)
			}
		})
	}
}

func TestBar(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		max   float64
		width int
		want  string
	}{
		{"full", 10, 10, 8, "████████"},
		{"empty-zero-value", 0, 10, 8, ""},
		{"empty-zero-max", 5, 0, 8, ""},
		{"empty-zero-width", 5, 10, 0, ""},
		{"negative-value", -1, 10, 8, ""},
		{"half-units-rounds-down", 5, 10, 8, "████"},
		{"quarter-units", 2.5, 10, 8, "██"},
		{"half-block-only", 0.5, 8, 8, "▌"},
		{"one-full-block", 1, 8, 8, "█"},
		{"one-and-a-half", 1.5, 8, 8, "█▌"},
		{"seven-and-a-half", 7.5, 8, 8, "███████▌"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Bar(tc.value, tc.max, tc.width); got != tc.want {
				t.Errorf("Bar(%g, %g, %d) = %q, want %q", tc.value, tc.max, tc.width, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 6: gofmt, run tests, lint clean**

```bash
gofmt -l lessons/08-capstone/
go test ./lessons/08-capstone/exercises/summary/... -v 2>&1 | tail -10
go test ./lessons/08-capstone/solutions/summary/... -v 2>&1 | tail -30
make test
golangci-lint run ./...
```

Expected: gofmt empty; exercises pass vacuously; solutions pass all sub-tests (TestTotalsByCategory 5 + TestBiggestCategory 6 + TestBar 11 = 22); make test green; lint 0 issues.

- [ ] **Step 7: Commit**

```bash
git add lessons/08-capstone/exercises/summary/ lessons/08-capstone/solutions/summary/
git commit -m "feat(lesson-08): summary subpackage — TotalsByCategory + BiggestCategory + Bar"
```

---

## Task 5: Provide the `storage` subpackage (black-box)

The `storage` package is fully implemented in both `exercises/` and `solutions/` — students don't write it. It's identical content in both sides (just different package import paths). The top comment tells students "treat as a black box".

**Files:**
- Create: `lessons/08-capstone/exercises/storage/storage.go`
- Create: `lessons/08-capstone/solutions/storage/storage.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/08-capstone/exercises/storage lessons/08-capstone/solutions/storage
```

- [ ] **Step 2: Create `lessons/08-capstone/exercises/storage/storage.go`** with:

```go
// Package storage handles JSON-backed persistence of Expense slices.
//
// This package is PROVIDED — you do not write or modify it for this lesson.
// We'll see how it works in Phase 2 (lesson 13 covers encoding/json
// properly). For now, treat it as a black box that loads and saves a
// []expense.Expense to a JSON file on disk.
//
// API:
//   - LoadExpenses(path) → ([]expense.Expense, error)
//     Returns an empty slice + nil error when the file doesn't exist
//     (so first-time use is friendly).
//   - SaveExpenses(path, es) → error
//     Writes pretty-printed JSON via json.MarshalIndent. Creates the
//     parent directory if needed.
package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// LoadExpenses reads the JSON file at path and returns its contents.
//
// If the file doesn't exist, returns ([]expense.Expense{}, nil) — a clean
// empty slice, no error. This makes first-time use (no file yet) friendly:
// the caller can always range over the result.
//
// If the file exists but is unreadable or malformed, returns nil and an
// error wrapping the underlying cause.
func LoadExpenses(path string) ([]expense.Expense, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return []expense.Expense{}, nil
	}
	if err != nil {
		return nil, err
	}
	var es []expense.Expense
	if err := json.Unmarshal(data, &es); err != nil {
		return nil, err
	}
	if es == nil {
		es = []expense.Expense{}
	}
	return es, nil
}

// SaveExpenses writes es to path as pretty-printed JSON.
//
// Creates the parent directory (with MkdirAll) if it doesn't exist.
// Overwrites the file if it does.
func SaveExpenses(path string, es []expense.Expense) error {
	data, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}
```

- [ ] **Step 3: Create `lessons/08-capstone/solutions/storage/storage.go`** with the same content but importing from solutions/expense:

```go
// Package storage handles JSON-backed persistence of Expense slices.
//
// (Solutions reference — identical to exercises/storage/storage.go except
// for the expense import path.)
package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
)

// LoadExpenses reads the JSON file at path and returns its contents.
// Returns ([]expense.Expense{}, nil) if the file doesn't exist.
func LoadExpenses(path string) ([]expense.Expense, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return []expense.Expense{}, nil
	}
	if err != nil {
		return nil, err
	}
	var es []expense.Expense
	if err := json.Unmarshal(data, &es); err != nil {
		return nil, err
	}
	if es == nil {
		es = []expense.Expense{}
	}
	return es, nil
}

// SaveExpenses writes es to path as pretty-printed JSON.
func SaveExpenses(path string, es []expense.Expense) error {
	data, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}
```

- [ ] **Step 4: gofmt, build, lint clean**

```bash
gofmt -l lessons/08-capstone/
go build ./lessons/08-capstone/exercises/storage ./lessons/08-capstone/solutions/storage
make test
golangci-lint run ./...
```

Expected: gofmt empty; both packages build; make test green; lint 0 issues. (No test files for storage — tested indirectly via cmd/expenses integration tests in Task 7.)

- [ ] **Step 5: Commit**

```bash
git add lessons/08-capstone/exercises/storage/ lessons/08-capstone/solutions/storage/
git commit -m "feat(lesson-08): storage subpackage (PROVIDED, black-box JSON persistence)"
```

---

## Task 6: Author the `cmd/expenses/main.go` CLI orchestration

The capstone binary. Parses `os.Args`, dispatches on subcommand (`add`, `list`, `summary`), wires storage + expense + summary.

**Files:**
- Create: `lessons/08-capstone/exercises/cmd/expenses/main.go`
- Create: `lessons/08-capstone/solutions/cmd/expenses/main.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/08-capstone/exercises/cmd/expenses lessons/08-capstone/solutions/cmd/expenses
```

- [ ] **Step 2: Create `lessons/08-capstone/exercises/cmd/expenses/main.go`** with the full stub (compiles, but every subcommand panics — students implement each branch):

```go
// Package main is the lesson 08 capstone — an expense tracker CLI.
//
// Usage:
//
//	expenses [-file=path] <subcommand> [args...]
//
// Subcommands:
//
//	add DATE AMOUNT CATEGORY    Append an expense to the JSON file.
//	list                         Print all expenses, one per line.
//	summary                      Print per-category totals and a bar chart.
//
// The -file=path option (must appear BEFORE the subcommand) sets the JSON
// file path. Default is $HOME/.expenses.json.
//
// Run with:
//
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json add 2026-05-18 4.50 coffee
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json list
//	go run ./lessons/08-capstone/exercises/cmd/expenses -file=/tmp/x.json summary
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/storage"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/summary"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. It returns an error instead of calling
// os.Exit so integration tests can drive it without process teardown.
func run(args []string) error {
	// Parse -file=path option (must come first if present).
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(path, args[1:])
	case "list":
		return cmdList(path)
	case "summary":
		return cmdSummary(path)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

// parseFileOption pulls -file=path off the front of args (if present) and
// returns the resolved path plus the remaining args. If -file= isn't given,
// the path defaults to $HOME/.expenses.json.
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

// cmdAdd parses DATE AMOUNT CATEGORY from args, appends an expense to the
// JSON file, and prints a confirmation line.
//
// Expected output:
//   added: 2026-05-18  €4.50    coffee
func cmdAdd(path string, args []string) error {
	panic("TODO: parse DATE AMOUNT CATEGORY; storage.LoadExpenses; append; storage.SaveExpenses; print confirmation via e.Format()")
	// Hints:
	// - if len(args) != 3 → return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	// - amount, err := strconv.ParseFloat(args[1], 64); handle err
	// - es, err := storage.LoadExpenses(path); handle err
	// - e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	// - es = append(es, e)
	// - if err := storage.SaveExpenses(path, es); ...
	// - fmt.Println("added:", e.Format())
	_ = strconv.ParseFloat
}

// cmdList loads the JSON file and prints each expense on its own line via
// Format(). Prints nothing for an empty file (no header).
func cmdList(path string) error {
	panic("TODO: storage.LoadExpenses; for each e: fmt.Println(e.Format())")
}

// cmdSummary loads the JSON file, computes per-category totals, prints
// the totals + bar chart, and the biggest category.
//
// Expected output (for 3 expenses totalling 40 across coffee/lunch/rent):
//
//   3 expenses, total €40.00
//   by category:
//     coffee     €4.50   ▌
//     lunch      €12.00  ████
//     rent       €23.50  ████████
//   biggest category: rent (€23.50)
//
// Empty file output:
//
//   0 expenses, total €0.00
//
// (No "by category" or "biggest" sections when empty.)
func cmdSummary(path string) error {
	panic("TODO: storage.LoadExpenses; compute total (sum of e.Amount); print header; if non-empty: print per-category totals with summary.Bar; print summary.BiggestCategory")
	// Hints for the body once you've loaded es:
	// - if len(es) == 0 → print "0 expenses, total €0.00" and return nil
	// - var total float64; for _, e := range es { total += e.Amount }
	// - fmt.Printf("%d expenses, total €%.2f\n", len(es), total)
	// - fmt.Println("by category:")
	// - totals := summary.TotalsByCategory(es)
	// - sort the keys for deterministic output (sort.Strings on a keys slice)
	// - find maxTotal for bar scaling: for _, t := range totals { if t > maxTotal { maxTotal = t } }
	// - for each sorted key: fmt.Printf("  %-10s €%-7.2f %s\n", k, totals[k], summary.Bar(totals[k], maxTotal, 8))
	// - name, biggestTotal := summary.BiggestCategory(totals)
	// - fmt.Printf("biggest category: %s (€%.2f)\n", name, biggestTotal)
	_ = sort.Strings
}
```

- [ ] **Step 3: Create `lessons/08-capstone/solutions/cmd/expenses/main.go`** with the full reference implementation:

```go
// Package main is the lesson 08 capstone CLI (solutions reference).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/storage"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/summary"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(path, args[1:])
	case "list":
		return cmdList(path)
	case "summary":
		return cmdSummary(path)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
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

func cmdAdd(path string, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := storage.SaveExpenses(path, es); err != nil {
		return err
	}
	fmt.Println("added:", e.Format())
	return nil
}

func cmdList(path string) error {
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(path string) error {
	es, err := storage.LoadExpenses(path)
	if err != nil {
		return err
	}

	var total float64
	for _, e := range es {
		total += e.Amount
	}
	fmt.Printf("%d expenses, total €%.2f\n", len(es), total)
	if len(es) == 0 {
		return nil
	}

	fmt.Println("by category:")
	totals := summary.TotalsByCategory(es)

	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var maxTotal float64
	for _, t := range totals {
		if t > maxTotal {
			maxTotal = t
		}
	}

	for _, k := range keys {
		fmt.Printf("  %-10s €%-7.2f %s\n", k, totals[k], summary.Bar(totals[k], maxTotal, 8))
	}

	name, biggestTotal := summary.BiggestCategory(totals)
	fmt.Printf("biggest category: %s (€%.2f)\n", name, biggestTotal)
	return nil
}
```

- [ ] **Step 4: gofmt + build + lint**

```bash
gofmt -l lessons/08-capstone/
go build ./lessons/08-capstone/exercises/cmd/expenses ./lessons/08-capstone/solutions/cmd/expenses
golangci-lint run ./...
```

Expected: gofmt empty; both binaries build; lint 0 issues.

- [ ] **Step 5: Smoke test the solutions binary end-to-end**

```bash
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP    # delete so storage sees "file not found" and starts fresh

go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 4.50 coffee
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 12.00 lunch
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 23.50 groceries

echo "--- list ---"
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP list

echo "--- summary ---"
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP summary

rm $TMP
```

Expected (the bar widths depend on totals; with the three above, groceries is max at 23.50):

```
added: 2026-05-18  €4.50    coffee
added: 2026-05-18  €12.00   lunch
added: 2026-05-18  €23.50   groceries
--- list ---
2026-05-18  €4.50    coffee
2026-05-18  €12.00   lunch
2026-05-18  €23.50   groceries
--- summary ---
3 expenses, total €40.00
by category:
  coffee     €4.50   █▌
  groceries  €23.50  ████████
  lunch      €12.00  ████
biggest category: groceries (€23.50)
```

(Coffee bar: 4.50/23.50 × 8 = 1.53 → "█▌". Lunch: 12.00/23.50 × 8 = 4.08 → "████". Groceries: 23.50/23.50 × 8 = 8.0 → "████████".)

- [ ] **Step 6: `make test` clean**

```bash
make test
```

Expected: all packages green.

- [ ] **Step 7: Commit**

```bash
git add lessons/08-capstone/exercises/cmd/ lessons/08-capstone/solutions/cmd/
git commit -m "feat(lesson-08): cmd/expenses/main.go orchestration (add/list/summary)"
```

---

## Task 7: Author the integration tests

`cmd/expenses/main_test.go` exercises the binary via `os/exec` against a temp JSON file. Skeleton in exercises (test scaffolding is provided; per-subcommand assertions are TODOs); full reference in solutions.

**Files:**
- Create: `lessons/08-capstone/exercises/cmd/expenses/main_test.go`
- Create: `lessons/08-capstone/solutions/cmd/expenses/main_test.go`

- [ ] **Step 1: Create `lessons/08-capstone/exercises/cmd/expenses/main_test.go`** with (SKELETON):

```go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runBinary builds the expenses binary into a temp location and returns a
// helper that runs it with the given args, returning stdout, stderr, and
// any error from the process (non-zero exit codes show up as *exec.ExitError).
//
// This is the canonical "test a Go binary by building it once and running
// it many times" pattern. The build cost is paid once; each call to the
// returned function is fast (just exec).
type runFn func(args ...string) (stdout, stderr string, err error)

func buildAndRun(t *testing.T) runFn {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "expenses")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return func(args ...string) (string, string, error) {
		cmd := exec.Command(binPath, args...)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
}

// TestAddSingle is a SKELETON. Test that `add 2026-05-18 4.50 coffee`:
//   - exits with code 0
//   - prints "added: 2026-05-18  €4.50    coffee" to stdout
//   - creates the JSON file at the -file path
//   - the JSON file contains exactly one expense
func TestAddSingle(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO:
	//   stdout, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee")
	//   if err != nil → t.Fatalf with stderr
	//   assert stdout contains "added: 2026-05-18  €4.50    coffee"
	//   assert os.ReadFile(tmpFile) succeeds and the JSON parses as 1 entry
	_ = run
	_ = tmpFile
	_ = os.ReadFile
}

// TestAddThenList is a SKELETON. Test that after adding three expenses,
// `list` prints them in insertion order, one per line.
func TestAddThenList(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO: add three expenses, then `list`, assert all three lines present
	// in stdout in insertion order. Use strings.Contains or split by "\n".
	_ = run
	_ = tmpFile
}

// TestAddThenSummary is a SKELETON. Test that after adding three expenses
// (coffee 4.50, lunch 12, groceries 23.50), `summary` prints the expected
// totals + the "biggest category: groceries (€23.50)" line.
func TestAddThenSummary(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// TODO: add 3, then `summary`, assert key lines present:
	//   "3 expenses, total €40.00"
	//   "biggest category: groceries (€23.50)"
	// Don't assert the bar-chart whitespace — it's locale/font-fragile.
	_ = run
	_ = tmpFile
}
```

- [ ] **Step 2: Create `lessons/08-capstone/solutions/cmd/expenses/main_test.go`** with full reference tests:

```go
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type runFn func(args ...string) (stdout, stderr string, err error)

func buildAndRun(t *testing.T) runFn {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "expenses")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return func(args ...string) (string, string, error) {
		cmd := exec.Command(binPath, args...)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
}

func TestAddSingle(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	stdout, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee")
	if err != nil {
		t.Fatalf("add failed: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "added: 2026-05-18  €4.50    coffee") {
		t.Errorf("stdout missing expected add confirmation, got: %q", stdout)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("JSON contains %d entries, want 1", len(got))
	}
}

func TestAddThenList(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee"); err != nil {
		t.Fatalf("add 1: %v\n%s", err, stderr)
	}
	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "12.00", "lunch"); err != nil {
		t.Fatalf("add 2: %v\n%s", err, stderr)
	}
	if _, stderr, err := run("-file="+tmpFile, "add", "2026-05-18", "23.50", "groceries"); err != nil {
		t.Fatalf("add 3: %v\n%s", err, stderr)
	}

	stdout, stderr, err := run("-file="+tmpFile, "list")
	if err != nil {
		t.Fatalf("list failed: %v\nstderr: %s", err, stderr)
	}
	wantLines := []string{
		"2026-05-18  €4.50    coffee",
		"2026-05-18  €12.00   lunch",
		"2026-05-18  €23.50   groceries",
	}
	for _, w := range wantLines {
		if !strings.Contains(stdout, w) {
			t.Errorf("list output missing line %q\nfull stdout:\n%s", w, stdout)
		}
	}
}

func TestAddThenSummary(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	for _, args := range [][]string{
		{"add", "2026-05-18", "4.50", "coffee"},
		{"add", "2026-05-18", "12.00", "lunch"},
		{"add", "2026-05-18", "23.50", "groceries"},
	} {
		if _, stderr, err := run(append([]string{"-file=" + tmpFile}, args...)...); err != nil {
			t.Fatalf("setup add failed: %v\n%s", err, stderr)
		}
	}

	stdout, stderr, err := run("-file="+tmpFile, "summary")
	if err != nil {
		t.Fatalf("summary failed: %v\nstderr: %s", err, stderr)
	}
	want := []string{
		"3 expenses, total €40.00",
		"by category:",
		"biggest category: groceries (€23.50)",
	}
	for _, w := range want {
		if !strings.Contains(stdout, w) {
			t.Errorf("summary output missing %q\nfull stdout:\n%s", w, stdout)
		}
	}
}
```

- [ ] **Step 3: gofmt, run tests, lint clean**

```bash
gofmt -l lessons/08-capstone/
go test ./lessons/08-capstone/exercises/cmd/expenses/... -v 2>&1 | tail -20
go test ./lessons/08-capstone/solutions/cmd/expenses/... -v 2>&1 | tail -40
make test
golangci-lint run ./...
```

Expected: gofmt empty; exercises pass vacuously (3 test functions with `_ = run` placeholders); solutions pass all 3 sub-tests (each takes ~1-2 seconds because of the `go build` step). make test green; lint 0 issues.

> Note: solutions integration tests take ~3-6 seconds total because each `buildAndRun(t)` invokes `go build`. That's acceptable for this lesson; if it becomes a CI nuisance later, the build can be hoisted into `TestMain` to build once for the whole package.

- [ ] **Step 4: Commit**

```bash
git add lessons/08-capstone/exercises/cmd/expenses/main_test.go lessons/08-capstone/solutions/cmd/expenses/main_test.go
git commit -m "feat(lesson-08): cmd/expenses integration tests (os/exec + t.TempDir)"
```

---

## Task 8: Author the slide deck

3 concepts (not the usual 4) — lesson 08 is more applied than conceptual, so most of the deck is walking through the capstone's architecture. The controller writes this inline.

**Files:**
- Replace: `lessons/08-capstone/slides/slides.md`

- [ ] **Step 1: Write `lessons/08-capstone/slides/slides.md`**

The deck content is long (~400 lines). The controller writes it inline rather than dispatching a subagent (Plan F/G/H/I/J established this pattern for slide decks).

Title slide → What we'll cover → Concept 1 (cmd/ convention) → Concept 2 (composing multiple packages) → Concept 3 (integration testing with os/exec) → Phase 1 capstone summary slide → Practice (warmup + main) → What we learned → Up next (Phase 2 preview).

Verification:

```bash
make slides-build 2>&1 | tail -3
grep -c '<div class="title-slide-grid">' lessons/08-capstone/slides/slides.md
grep -c "<h1>Phase 1 capstone</h1>" lessons/08-capstone/slides/slides.md
grep -c "^## Concept " lessons/08-capstone/slides/slides.md     # expect 3
grep -c "^### Motivation$" lessons/08-capstone/slides/slides.md # expect 3
grep -c "^### Common mistake$" lessons/08-capstone/slides/slides.md # expect 3
grep -q "Phase 1 capstone" dist/index.html && echo "lists capstone"
rm -rf dist
```

- [ ] **Step 2: Commit**

```bash
git add lessons/08-capstone/slides/slides.md
git commit -m "feat(lesson-08): slides — Phase 1 capstone (3 concepts + closing summary)"
```

---

## Task 9: Author the README

The README walks through the capstone's architecture, the package layout, how to run the binary, and what each subpackage does. Includes a final "Phase 1 complete!" closing.

**Files:**
- Replace: `lessons/08-capstone/README.md`

- [ ] **Step 1: Write `lessons/08-capstone/README.md`**

Sections (same shape as previous lessons + a closing Phase 1 recap):

- `# Lesson 08: Phase 1 capstone — expense tracker CLI`
- `## Learning goals`
- `## Prerequisites`
- `## What's different about this lesson`
- `## Concepts`
  - `### The cmd/ convention`
  - `### Composing multiple packages into a CLI`
  - `### End-to-end integration testing`
- `## Exercise: warm-up`
- `## Exercise: main`
- `## How to run`
- `## Phase 1 — complete!` (closing recap of what students learned across all 8 lessons)
- `## Going further`
  - `### Read`
  - `### Try`

Verification:

```bash
for h in "^# Lesson 08: Phase 1 capstone" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Phase 1 — complete!$" "^## Going further$" "^### Read$" "^### Try$"; do
  echo "  $h => $(grep -c "$h" lessons/08-capstone/README.md)"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/08-capstone/README.md   # expect 3
grep -c "gofmt" lessons/08-capstone/README.md
grep -c "go vet" lessons/08-capstone/README.md
```

Expected: all section headings 1; Common-mistake count 3; gofmt and go vet counts ≥ 1.

- [ ] **Step 2: Commit**

```bash
git add lessons/08-capstone/README.md
git commit -m "docs(lesson-08): README — Phase 1 capstone self-study + Phase 1 recap"
```

---

## Task 10: End-to-end verification

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package green.

- [ ] **Step 2: Run exercise tests — pass vacuously**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 08's TestNow, TestExpenseFormat, TestExpenseIsHigh, TestTotalsByCategory, TestBiggestCategory, TestBar, TestAddSingle, TestAddThenList, TestAddThenSummary all PASS with 0 sub-tests. (TestNow's body uses `_ = got` so even though it has no `t.Run` it counts as "passed with no assertions").

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=08-capstone 2>&1 | tail -30
```

Expected: exercises pass vacuously; solutions pass fully. Solution sub-test totals:
- warmup/clock TestNow: 1
- expense TestFormat: 3, TestIsHigh: 6
- summary TestTotalsByCategory: 5, TestBiggestCategory: 6, TestBar: 11
- cmd/expenses TestAddSingle, TestAddThenList, TestAddThenSummary: 3
Total: 35 sub-tests/assertions in solutions.

- [ ] **Step 4: Run the solutions binary end-to-end (smoke test)**

```bash
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 4.50 coffee
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 12.00 lunch
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP add 2026-05-18 23.50 groceries
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP list
go run ./lessons/08-capstone/solutions/cmd/expenses -file=$TMP summary
rm $TMP
```

Expected: three "added: ..." lines, three list lines, and a summary block with bar chart + biggest-category line.

- [ ] **Step 5: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 6: `go vet` clean**

```bash
go vet ./...
```

Expected: no output.

- [ ] **Step 7: Slides server smoke test**

```bash
make slides-dev LESSON=08-capstone &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/08-capstone/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/08-capstone/slides/slides.md
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`/`301`/`302` codes.

- [ ] **Step 8: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/08-capstone/slides/slides.md && echo OK
grep -q "capstone" dist/index.html && echo "index lists lesson 08"
rm -rf dist
```

- [ ] **Step 9: Final repository sanity check**

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 10 commits on the branch (plan + scaffold + 6 implementation tasks + slides + README — actually let me recount: plan + scaffold + warmup + expense + summary + storage + cmd + integration tests + slides + README = 10 commits).

This task makes no commit.

---

## Done definition

After Task 10:

- `lessons/08-capstone/` contains 24 files in the subpackage layout.
- `make test` passes; `make test-exercises` shows lesson 08's tests passing vacuously.
- `make test-lesson LESSON=08-capstone` runs both sides correctly.
- `go run ./lessons/08-capstone/solutions/cmd/expenses ...` works end-to-end for add/list/summary.
- `make slides-dev LESSON=08-capstone` serves the deck.
- `make slides-build` produces `dist/` containing the lesson.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- 10 commits on the branch.

## Phase 1 — complete!

After Plan K, the eight Phase 1 lessons are all shipped:
- Lesson 01 — Hello, Go
- Lesson 02 — Variables, types, operators
- Lesson 03 — Control flow
- Lesson 04 — Functions & first tests
- Lesson 05 — Composite types I: slices and maps
- Lesson 06 — Composite types II: structs and methods
- Lesson 07 — Packages and modules
- Lesson 08 — Phase 1 capstone (this plan)

The next plan (Plan L) starts Phase 2 — Idiomatic Go. Lesson 09 introduces pointers properly.
