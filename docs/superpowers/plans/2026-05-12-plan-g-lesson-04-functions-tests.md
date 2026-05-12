# Plan G — Lesson 04 (Functions & first tests) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the fourth lesson of Phase 1 — students learn Go's function shape (multi-return, named returns, variadic, `defer`), the `error` type with `errors.New` and the canonical `if err != nil { return …, err }` shape, and the testing discipline (filename convention, `func TestXxx(t *testing.T)`, `t.Errorf` vs `t.Fatalf`, table tests via `[]struct{}`). The main exercise is the **first time students author test code themselves**.

**Architecture:** Same per-lesson pattern as Plans D, E, and F. Six tasks: scaffold, warm-up, main, slides, README, end-to-end verify. Four functions across two exercises: `Add` + `MinMax` (warm-up) and `Categorise` + `FormatExpense` (main). Heavy-explanatory slide deck with **four** concept blocks (function shape → defer → errors → testing). The README mirrors the deck with the same Common-mistake examples and introduces `go vet ./...` to the tooling thread.

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `errors`, `testing`). No third-party dependencies. Reveal.js 5.1.0 for the deck.

---

## Scope

Plan G produces lesson 04 only. After Plan G lands:

- `lessons/04-functions/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make slides-dev LESSON=04-functions` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):** Lessons 05-08 (Plans H-K).

### Note on the deferred MinMax

Plan E's spec deviation moved `MinMax(xs ...int) (int, int, error)` from lesson 02 to lesson 04 — that's it landing here. `MinMax` exercises three things at once: variadic input (`xs ...int`), multi-return (two ints), and the `error` return for empty input. It's perfect lesson-04 material.

### Design decisions made during planning

These are choices within the spec's latitude. Easy to flip during plan review.

1. **Soft ramp for student-authored tests.** Per user choice: warm-up keeps the lesson 1-3 "tests pre-written, students implement" pattern; main exercise ships **skeleton test files** with the `cases := []struct{ … }{}` scaffold already in place — students fill in the case rows and the loop+assertion body. Easier first exposure than empty files. The README's "Exercise: main" section walks through the test shape step by step.

2. **Four concepts in the slide deck.** Grouped as: (1) Function shape — multi-return + named returns + variadic, (2) `defer`, (3) Errors — the `error` type + `errors.New` + the `if err != nil` shape, (4) Testing — `*_test.go`, `testing.T`, table tests, `t.Errorf` vs `t.Fatalf`. Lessons 02 and 03 both used four concepts; pacing matches.

3. **`defer` is taught but not exercised.** Spec lists `defer` as a concept but neither warm-up nor main has a natural use (no files / no concurrency yet). The slides show defer with a clean "hello/world LIFO" example plus a forward-looking `defer f.Close()` snippet labelled "you'll use this for real in lesson 13." No exercise; that's fine — students see the syntax and the LIFO rule, and meet defer for real when files arrive.

4. **`FormatExpense` format chosen.** The function signature is `FormatExpense(date string, amount float64, cat string) string`; the output format is `"YYYY-MM-DD  €AMOUNT  category"` with `%-7.2f` width on the amount (left-aligned, padded to 7 chars including the decimal). Concrete example: `FormatExpense("2026-05-12", 4.50, "coffee")` → `"2026-05-12  €4.50    coffee"`. This format mirrors lesson 02's `Summary` style and lesson 03's `runMain` per-line output.

5. **`error` is described without saying "interface."** The Phase 1 spec defers interfaces to lesson 10. The slides describe `error` as "a type with an `Error() string` method" without using the word interface explicitly. Forward reference: "We'll see what makes `error` special — and how you'd build your own — in lesson 10."

6. **`go vet ./...` joins the tooling thread.** Lesson 04 is where the spec says `go vet` becomes a routine. The README's "How to run" adds `go vet ./...` alongside `gofmt -w .`. The slides briefly mention it in the testing concept.

7. **`solutions/` is `package solutions`** (not `package main`). Lesson 03's `package main` was needed because of the interactive Scanln runner; lesson 04 has no entry point. Back to the standard layout.

---

## Plan F lessons-learned applied here

Issues from Plan F's implementation that get pre-emptively handled in Plan G:

1. **`.golangci.yml` exclusion already in place.** Plan F added an exclusion for `lessons/*/exercises/` that covers `staticcheck` + `unused`. Lesson 04's exercises don't have a `runMain`-style demo function, but the exclusion is still helpful: warm-up and main both ship stub functions that panic with TODOs. No new lint configuration needed.

2. **Lint step in every Go-touching task** — same as Plans D/E/F.

3. **`→` arrow consistency** — same.

4. **Common-mistake content in README** — every concept's README section includes the same common-mistake example shown in the slides. (4 total.)

5. **Slides are large prompts and timed out a subagent during Plan F.** Plan G's Task 4 (slides) breaks the slide content into two sequential edits — the first writes the title slide + concepts 1-2; the second appends concepts 3-4 + Practice/What-we-learned/Up-next. Avoids the single-write timeout risk.

6. **`gofmt -w .` mention** — continues; the README also adds `go vet ./...`.

---

## File Structure

After Plan G:

```
lessons/04-functions/                       (new — Plan G's deliverable)
├── README.md                               (Task 5)
├── slides/
│   ├── index.html                          (Task 1; unchanged)
│   ├── slides.md                           (Task 4 — two-step write)
│   └── assets/.gitkeep                     (Task 1; unchanged)
├── exercises/
│   ├── warmup.go                           (Task 2 — Add + MinMax with stubs)
│   ├── warmup_test.go                      (Task 2 — pre-written failing tests)
│   ├── main.go                             (Task 3 — Categorise + FormatExpense with stubs)
│   └── main_test.go                        (Task 3 — SKELETON: struct + empty case list + loop scaffold; students fill in cases)
└── solutions/
    ├── warmup.go                           (Task 2 — implementations)
    ├── warmup_test.go                      (Task 2 — same tests, package solutions)
    ├── main.go                             (Task 3 — implementations)
    └── main_test.go                        (Task 3 — full reference tests with completed table cases)
```

### Decomposition rationale

Same pattern as plans D, E, F. Each task has one focused output and a clear verification step. The **shape of main_test.go differs between exercises and solutions** — exercises ships a skeleton (students complete it), solutions ships the full reference table tests. This is the first time the two diverge structurally; the README explains why.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-g-lesson-04-functions-tests` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)

---

## Task 1: Scaffold lesson 04

**Files:**
- Create: `lessons/04-functions/` (12 files via the scaffolder)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-g-lesson-04-functions-tests`, clean working tree, exactly one commit ahead of `main` (the Plan G doc). If you see something else, STOP and report.

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=04-functions
```

Expected: prints `created lesson 04-functions under lessons/`. The scaffolder produces 12 files with template placeholders. The slide title slide already uses the badge+phase format.

- [ ] **Step 3: Verify the 12-file tree**

```bash
find lessons/04-functions -type f | sort
```

Expected (12 lines):

```
lessons/04-functions/README.md
lessons/04-functions/exercises/main.go
lessons/04-functions/exercises/main_test.go
lessons/04-functions/exercises/warmup.go
lessons/04-functions/exercises/warmup_test.go
lessons/04-functions/slides/assets/.gitkeep
lessons/04-functions/slides/index.html
lessons/04-functions/slides/slides.md
lessons/04-functions/solutions/main.go
lessons/04-functions/solutions/main_test.go
lessons/04-functions/solutions/warmup.go
lessons/04-functions/solutions/warmup_test.go
```

- [ ] **Step 4: Commit the scaffolded skeleton**

```bash
git add lessons/04-functions/
git commit -m "feat(lessons): scaffold lesson 04-functions skeleton"
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: every package passes. Lesson 04 solutions pass (the scaffolded `Greet`/`WarmupGreet` work).

- [ ] **Step 6: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch.

---

## Task 2: Author the warm-up exercise

Replace the four warm-up files. `Add(a, b int) int` is the simplest possible function — a clean introduction to "implementation + table test." `MinMax(xs ...int) (int, int, error)` exercises variadic input, multi-return, and the `error` return for empty input.

**Files:**
- Replace: `lessons/04-functions/exercises/warmup.go`
- Replace: `lessons/04-functions/exercises/warmup_test.go`
- Replace: `lessons/04-functions/solutions/warmup.go`
- Replace: `lessons/04-functions/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/04-functions/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 04: Functions & first tests.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "function with multiple returns" and "variadic argument list,
// with an error return for empty input." Tests are pre-written — fix the
// implementations until they pass.
package exercises

import "errors"

// errEmptyMinMax is returned by MinMax when called with no values.
//
// Sentinel-style error variables are a common Go pattern: declare once with
// errors.New, return as-is. We'll see more sophisticated error styles in
// lesson 11.
var errEmptyMinMax = errors.New("MinMax: requires at least one value")

// Add returns the sum of two ints.
//
// Examples:
//   Add(2, 3)   → 5
//   Add(-1, 1)  → 0
func Add(a, b int) int {
	panic("TODO: return a + b")
}

// MinMax returns the smallest and largest of the given ints, plus an error.
//
// When called with at least one int, it returns (min, max, nil). When called
// with zero ints, it returns (0, 0, errEmptyMinMax).
//
// Examples:
//   MinMax(3, 1, 4, 1, 5, 9, 2, 6)  → (1, 9, nil)
//   MinMax(42)                       → (42, 42, nil)
//   MinMax()                         → (0, 0, error)
//
// Hint: variadic parameters arrive as a slice — `xs ...int` is `xs []int`
// inside the function body. Walk the slice with `for i, v := range xs` or
// `for _, v := range xs`. The first element is your initial min and max;
// fold the rest in.
func MinMax(xs ...int) (int, int, error) {
	panic("TODO: return min, max for non-empty input; return (0, 0, errEmptyMinMax) for empty input")
}
```

- [ ] **Step 2: Replace `lessons/04-functions/exercises/warmup_test.go`** with:

```go
package exercises

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name    string
		a, b    int
		want    int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	cases := []struct {
		name           string
		xs             []int
		wantMin, wantMax int
		wantErr        bool
	}{
		{"single", []int{42}, 42, 42, false},
		{"sorted", []int{1, 2, 3, 4, 5}, 1, 5, false},
		{"reverse", []int{5, 4, 3, 2, 1}, 1, 5, false},
		{"unsorted", []int{3, 1, 4, 1, 5, 9, 2, 6}, 1, 9, false},
		{"all-negative", []int{-3, -1, -7, -2}, -7, -1, false},
		{"mixed-signs", []int{-5, 0, 5}, -5, 5, false},
		{"empty", []int{}, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, gotMax, err := MinMax(tc.xs...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("MinMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyMinMax) {
					t.Errorf("MinMax(%v) error = %v, want errEmptyMinMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("MinMax(%v) unexpected error: %v", tc.xs, err)
			}
			if gotMin != tc.wantMin {
				t.Errorf("MinMax(%v) min = %d, want %d", tc.xs, gotMin, tc.wantMin)
			}
			if gotMax != tc.wantMax {
				t.Errorf("MinMax(%v) max = %d, want %d", tc.xs, gotMax, tc.wantMax)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/04-functions/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 04: Functions & first tests.
//
// This file holds the warm-up reference solution.
package solutions

import "errors"

// errEmptyMinMax is returned by MinMax when called with no values.
var errEmptyMinMax = errors.New("MinMax: requires at least one value")

// Add returns the sum of two ints.
func Add(a, b int) int {
	return a + b
}

// MinMax returns the smallest and largest of the given ints, plus an error
// for the empty-input case.
func MinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmptyMinMax
	}
	min, max := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, nil
}
```

- [ ] **Step 4: Replace `lessons/04-functions/solutions/warmup_test.go`** with the same content as `exercises/warmup_test.go`, but `package solutions`:

```go
package solutions

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	cases := []struct {
		name             string
		xs               []int
		wantMin, wantMax int
		wantErr          bool
	}{
		{"single", []int{42}, 42, 42, false},
		{"sorted", []int{1, 2, 3, 4, 5}, 1, 5, false},
		{"reverse", []int{5, 4, 3, 2, 1}, 1, 5, false},
		{"unsorted", []int{3, 1, 4, 1, 5, 9, 2, 6}, 1, 9, false},
		{"all-negative", []int{-3, -1, -7, -2}, -7, -1, false},
		{"mixed-signs", []int{-5, 0, 5}, -5, 5, false},
		{"empty", []int{}, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, gotMax, err := MinMax(tc.xs...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("MinMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyMinMax) {
					t.Errorf("MinMax(%v) error = %v, want errEmptyMinMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("MinMax(%v) unexpected error: %v", tc.xs, err)
			}
			if gotMin != tc.wantMin {
				t.Errorf("MinMax(%v) min = %d, want %d", tc.xs, gotMin, tc.wantMin)
			}
			if gotMax != tc.wantMax {
				t.Errorf("MinMax(%v) max = %d, want %d", tc.xs, gotMax, tc.wantMax)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/04-functions/
```

Expected: no output. If anything is listed, `gofmt -w` and re-check. Doc-comment example arrows (`→`) must survive.

- [ ] **Step 6: Verify the warm-up exercise tests fail by design**

```bash
go test ./lessons/04-functions/exercises/... -run Warmup -v 2>&1 | tail -20
```

Expected: tests run but `Add`/`MinMax` panic with TODO messages → FAIL.

- [ ] **Step 7: Verify the warm-up solution tests pass**

```bash
go test ./lessons/04-functions/solutions/... -run Warmup -v 2>&1 | tail -30
```

Expected: `TestAdd` (4 sub-tests) and `TestMinMax` (7 sub-tests) PASS.

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: green, `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/04-functions/exercises/warmup.go lessons/04-functions/exercises/warmup_test.go \
        lessons/04-functions/solutions/warmup.go lessons/04-functions/solutions/warmup_test.go
git commit -m "feat(lesson-04): warm-up — Add + MinMax (variadic + error return)"
```

---

## Task 3: Author the main exercise

Replace the four main-exercise files. The **`exercises/main_test.go` ships as a SKELETON** — the table-test struct and the for-loop scaffold are there, but the `cases` slice is empty and the assertion body is a TODO. Students fill those in. The **`solutions/main_test.go` ships fully complete** as the reference.

This is the first lesson where exercises/ and solutions/ diverge in shape — exercises is intentionally incomplete in a way go test will still compile.

**Files:**
- Replace: `lessons/04-functions/exercises/main.go`
- Replace: `lessons/04-functions/exercises/main_test.go`
- Replace: `lessons/04-functions/solutions/main.go`
- Replace: `lessons/04-functions/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/04-functions/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 04: Functions & first tests.
//
// This file holds the MAIN exercise. Two functions to implement, AND the
// matching tests for them in main_test.go. The test file ships with a
// skeleton table test for each function — you fill in the cases and the
// loop body. See exercises/main_test.go and the README's "Exercise: main"
// section for the test-writing walk-through.
package exercises

// Categorise returns "snack" for amounts under €10, "regular" for €10
// (inclusive) up to €50 (inclusive), and "splurge" above €50.
//
// Same behaviour as lesson 03's Categorise — the focus here is on writing
// the *test* for it, in table form. Use whichever shape (if-chain or
// tagless switch) you prefer.
//
// Examples:
//   Categorise(4.50)  → "snack"
//   Categorise(50.00) → "regular"   (boundary: 50 is the upper edge of regular)
//   Categorise(75.00) → "splurge"
func Categorise(amount float64) string {
	panic("TODO: return \"snack\" / \"regular\" / \"splurge\" by amount")
}

// FormatExpense returns a single-line, column-aligned rendering of one
// expense.
//
// Format: "YYYY-MM-DD  €AMOUNT  category"
//   - the date is the caller's responsibility — it's printed as-is
//   - the amount is prefixed with €, has 2 decimal places, and is left-padded
//     to 7 chars total (the "%-7.2f" verb in fmt.Sprintf)
//   - the category follows after two spaces
//
// Examples:
//   FormatExpense("2026-05-12", 4.50, "coffee")   → "2026-05-12  €4.50    coffee"
//   FormatExpense("2026-05-12", 12.00, "lunch")   → "2026-05-12  €12.00   lunch"
//   FormatExpense("2026-05-12", 999.99, "rent")   → "2026-05-12  €999.99  rent"
//
// Hint: fmt.Sprintf with "%s  €%-7.2f %s" (note the two spaces, the leading
// €, the width specifier, and the single trailing space) produces the
// expected output for the examples above.
func FormatExpense(date string, amount float64, cat string) string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}
```

- [ ] **Step 2: Replace `lessons/04-functions/exercises/main_test.go`** with (skeleton — students complete it):

```go
package exercises

import "testing"

// TestCategorise is a SKELETON. The struct shape and the for-loop scaffold
// are here for you; your job is to fill in the cases slice (think about
// the boundaries: <10, exactly 10, between, exactly 50, >50) and the body
// of the t.Run block.
//
// When you're done, this test should fail (because Categorise panics with
// a TODO) — exactly like the warm-up tests. Once your Categorise
// implementation in main.go is correct, all sub-tests pass.
func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		// TODO: add at least 5 cases here. Cover the snack/regular/splurge
		// branches AND the boundary values (9.99, 10.00, 50.00, 50.01).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call Categorise(tc.amount), compare with tc.want,
			// use t.Errorf with a clear message if they differ.
			_ = tc
		})
	}
}

// TestFormatExpense is a SKELETON. Same shape as TestCategorise — fill in
// the cases and the t.Run body.
//
// Look at the doc comment on FormatExpense for the expected output strings
// (note the spacing: two spaces between fields, the €-prefixed amount
// padded to 7 chars, etc.). The "Hint" line in the doc comment gives away
// the exact fmt.Sprintf format string; building those expected outputs by
// hand is part of the exercise.
func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name   string
		date   string
		amount float64
		cat    string
		want   string
	}{
		// TODO: add at least 3 cases here. Use the examples in the
		// FormatExpense doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call FormatExpense, compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}
```

> Note: the `_ = tc` lines exist so the file compiles with empty case lists. Students replace them with real assertions. Without those underscores, Go would complain that `tc` is declared but not used.

- [ ] **Step 3: Replace `lessons/04-functions/solutions/main.go`** with:

```go
// Package solutions is the reference implementation for lesson 04: Functions & first tests.
package solutions

import "fmt"

// Categorise returns "snack" for amounts under €10, "regular" for €10..€50
// (inclusive on both ends), "splurge" above €50.
func Categorise(amount float64) string {
	switch {
	case amount < 10:
		return "snack"
	case amount <= 50:
		return "regular"
	default:
		return "splurge"
	}
}

// FormatExpense returns "YYYY-MM-DD  €AMOUNT  category" with the amount
// left-padded to 7 characters using the "%-7.2f" verb.
func FormatExpense(date string, amount float64, cat string) string {
	return fmt.Sprintf("%s  €%-7.2f %s", date, amount, cat)
}
```

- [ ] **Step 4: Replace `lessons/04-functions/solutions/main_test.go`** with the full reference table tests:

```go
package solutions

import "testing"

func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		{"zero-snack", 0, "snack"},
		{"small-snack", 4.50, "snack"},
		{"upper-snack-edge", 9.99, "snack"},
		{"lower-regular-edge", 10.00, "regular"},
		{"middle-regular", 12.00, "regular"},
		{"upper-regular-edge", 50.00, "regular"},
		{"just-above-regular", 50.01, "splurge"},
		{"large-splurge", 75.00, "splurge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Categorise(tc.amount); got != tc.want {
				t.Errorf("Categorise(%g) = %q, want %q", tc.amount, got, tc.want)
			}
		})
	}
}

func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name   string
		date   string
		amount float64
		cat    string
		want   string
	}{
		{"coffee", "2026-05-12", 4.50, "coffee", "2026-05-12  €4.50    coffee"},
		{"lunch", "2026-05-12", 12.00, "lunch", "2026-05-12  €12.00   lunch"},
		{"rent", "2026-05-12", 999.99, "rent", "2026-05-12  €999.99  rent"},
		{"zero-amount", "2026-05-12", 0, "free", "2026-05-12  €0.00    free"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExpense(tc.date, tc.amount, tc.cat)
			if got != tc.want {
				t.Errorf("FormatExpense(%q, %g, %q) =\n  %q\nwant\n  %q", tc.date, tc.amount, tc.cat, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/04-functions/
```

Expected: no output. `gofmt -w` if needed.

- [ ] **Step 6: Verify the main exercise tests "pass vacuously"**

Because the exercises test files have empty `cases` slices, `TestCategorise` and `TestFormatExpense` in `exercises/` will report **PASS with 0 sub-tests** — that's correct for the skeleton state. Students filling in cases will see them fail until their implementations are right.

```bash
go test ./lessons/04-functions/exercises/... -v 2>&1 | tail -20
```

Expected: `TestCategorise` and `TestFormatExpense` PASS with no sub-tests. Warm-up tests (TestAdd, TestMinMax) FAIL with panics from Task 2.

- [ ] **Step 7: Verify the main solution tests pass**

```bash
go test ./lessons/04-functions/solutions/... -v 2>&1 | tail -30
```

Expected: every test PASSes — TestAdd (4), TestMinMax (7), TestCategorise (8), TestFormatExpense (4).

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: green, `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/04-functions/exercises/main.go lessons/04-functions/exercises/main_test.go \
        lessons/04-functions/solutions/main.go lessons/04-functions/solutions/main_test.go
git commit -m "feat(lesson-04): main — Categorise + FormatExpense (with skeleton tests)"
```

---

## Task 4: Author the slide deck

Replace `lessons/04-functions/slides/slides.md` with the lesson 04 deck. Four concepts: function shape → defer → errors → testing.

**Files:**
- Replace: `lessons/04-functions/slides/slides.md` (the entire file)

**Implementation note:** Plan F's slide-deck single-write timed out a subagent. To avoid that, this task writes the deck in **two sequential edits**: Step 1 writes the title slide through Concept 2 (defer); Step 2 appends Concepts 3-4 plus the closing sections.

- [ ] **Step 1: Write `lessons/04-functions/slides/slides.md`** (title slide through concept 2 — defer):

> CRITICAL: Four-backtick wrapper is a documentation device. In the file use only three-backtick fences. File starts with `<div class="title-slide-grid">`.

````markdown
<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">04</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Functions &amp; first tests</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Write functions with multiple return values, named returns, and variadic arguments; use <code>defer</code> for cleanup; return and handle <code>error</code>s with <code>errors.New</code>; and write your first table tests with <code>testing.T</code>.</p>
</div>
</div>
</div>

---

## What we'll cover

- Function shape: multi-return values, named returns, variadic arguments (`xs ...int`).
- `defer` — running cleanup when a function returns, regardless of how it returns.
- The `error` return convention: returning `(value, error)`, the `if err != nil` shape, `errors.New`.
- Testing: `*_test.go` files, `func TestXxx(t *testing.T)`, table tests with `[]struct{}`, `t.Errorf` vs `t.Fatalf`.

---

## Concept 1: Function shape

### Motivation

You've used functions since lesson 01. Now we look at the parts of a function declaration that you haven't met yet: multiple return values, named returns, and variadic arguments. Multi-return is everywhere in Go's standard library — it's how `error` gets returned alongside a result.

---

### The basics

```go
package main

import "fmt"

// Multi-return — two values.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// Named returns — variables declared in the signature.
// The bare `return` at the end returns whatever's in q and r.
func divmodNamed(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

// Variadic — accepts zero or more ints.
// Inside the function, xs is a []int.
func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

func main() {
	q, r := divmod(17, 5)
	fmt.Println(q, r)         // 3 2

	q2, r2 := divmodNamed(17, 5)
	fmt.Println(q2, r2)       // 3 2

	fmt.Println(sum())             // 0
	fmt.Println(sum(1, 2, 3))      // 6
	fmt.Println(sum(1, 2, 3, 4))   // 10

	// Passing a slice to a variadic — note the trailing ... .
	xs := []int{10, 20, 30}
	fmt.Println(sum(xs...))         // 60
}
```

Three things:

- **Multi-return** uses parentheses around the return types. The caller destructures with `a, b := f()`.
- **Named returns** declare the variables in the signature. The bare `return` at the end is shorthand for "return all named returns." Use named returns to document what each return means; lesson 06 (structs) shows when they're worth the extra ceremony.
- **Variadic** — `xs ...T` is `xs []T` inside the function. Pass an existing slice with the `slice...` syntax (note the trailing `...`).

---

### A worked example

The expense theme — `MinMax(xs ...int) (int, int, error)` from your warm-up exercise:

```go
package main

import (
	"errors"
	"fmt"
)

var errEmpty = errors.New("MinMax: requires at least one value")

func MinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmpty
	}
	min, max := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, nil
}

func main() {
	lo, hi, err := MinMax(3, 1, 4, 1, 5, 9, 2, 6)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("min:", lo, "max:", hi)   // min: 1 max: 9
}
```

Three things show up in one function: variadic input, three-value multi-return, and the `error` return for the bad case. We'll dig into errors in concept 3.

---

### Common mistake

Forgetting the trailing `...` when passing a slice to a variadic:

```go
package main

import "fmt"

func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

func main() {
	nums := []int{1, 2, 3}
	fmt.Println(sum(nums))   // wrong: passes []int as a single arg
}
```

Go complains:

```
cannot use nums (variable of type []int) as int value in argument to sum
```

The fix is `sum(nums...)` — the three dots spread the slice into individual arguments. Easy to forget.

---

### Recap

- Multi-return: `func f() (T, U) { return … }`; caller does `a, b := f()`.
- Named returns: declare result variables in the signature; bare `return` returns them all.
- Variadic: `xs ...T` arrives as `[]T`. Spread an existing slice with `slice...`.

---

## Concept 2: `defer`

### Motivation

When a function returns, sometimes you want to run cleanup code regardless of *how* it returned — early return, normal return, even a panic. `defer` is Go's way to schedule that cleanup at the function's exit point. The classic use is closing files and unlocking mutexes; we'll meet both for real in later lessons.

---

### The basics

`defer expr` schedules `expr` to run when the surrounding function returns. Deferred calls run in **LIFO order** (last deferred, first to run):

```go
package main

import "fmt"

func main() {
	defer fmt.Println("world")  // scheduled — runs LAST
	defer fmt.Println("middle") // scheduled — runs before "world"
	fmt.Println("hello")        // runs first (no defer)
}
```

Output:

```
hello
middle
world
```

The deferred expression's *arguments* are evaluated at the `defer` line, but the *call* happens at function exit. That distinction trips people up sometimes (see the common mistake below).

---

### A worked example

A small file-close pattern you'll write hundreds of times once we get to I/O (lesson 13):

```go
// Roughly what real Go code looks like — we cover os.Open in lesson 13.
func loadAndPrint(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()  // runs when loadAndPrint returns, regardless of how

	// ... read from f ...
	return nil
}
```

Two important things:

- `defer f.Close()` is on the very next line after the error check — that's the conventional shape. You won't forget to close the file because Go guarantees it.
- If `loadAndPrint` returns early with an error after the defer, `f.Close()` still runs. If `loadAndPrint` returns normally, `f.Close()` still runs. If `loadAndPrint` panics, `f.Close()` still runs.

---

### Common mistake

Forgetting that argument evaluation happens at the `defer` line, not at function exit:

```go
package main

import "fmt"

func main() {
	i := 1
	defer fmt.Println("deferred i =", i)  // i evaluated to 1 here
	i = 2
	fmt.Println("current i =", i)
}
```

Output:

```
current i = 2
deferred i = 1
```

The defer captured `i`'s value at the defer statement, not at the function's exit. If you want the latter behaviour, wrap the call in a function literal: `defer func() { fmt.Println("deferred i =", i) }()`.

---

### Recap

- `defer expr` schedules `expr` for the function's exit.
- Multiple defers run in LIFO order.
- Arguments are evaluated at the `defer` line; the call itself runs later.
- Most common use: file close, mutex unlock — you'll meet both in later lessons.

---
````

- [ ] **Step 2: Append the rest of the deck** (concepts 3-4 plus closing sections) by **adding to the end of the file** — do NOT overwrite Step 1's content:

> Use the Edit tool with `old_string` set to the last line of Step 1's content (`---` followed by a blank line, but write the unique closing marker — the end of concept 2's recap section) and `new_string` set to the same closing marker followed by the new content. OR use a second Write with the full combined content. Whichever your tool ergonomics prefer; just make sure no content is lost.

The content to append (after concept 2):

````markdown
## Concept 3: Errors

### Motivation

In Go, functions that can fail return an `error` as their last result. Callers check it with `if err != nil { … }`. There are no exceptions, no try/catch — error handling is just another value you check. That's verbose but predictable; you always know exactly where errors can come from.

---

### The basics

The shape — function returns `(result, error)`, caller checks the error:

```go
package main

import (
	"errors"
	"fmt"
)

// Sentinel error declared as a package-level var.
var errEmpty = errors.New("MinMax: requires at least one value")

func MinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmpty
	}
	// ... normal-path implementation ...
	return xs[0], xs[0], nil
}

func main() {
	lo, hi, err := MinMax()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(lo, hi)
}
```

Output:

```
error: MinMax: requires at least one value
```

Two patterns to learn here:

1. **`errors.New("message")`** — the simplest way to make an error. Stash it in a package-level `var` if you want callers (or your own code) to compare against it. We call these "sentinel errors."
2. **`if err != nil { return ..., err }`** — the canonical guard clause. Echoes lesson 03's early-return pattern; you'll write this thousands of times.

`error` itself is a built-in type with one method: `Error() string`. That's all there is to it. (We'll see what makes it special — and how to build your own error types — in lesson 10 when interfaces arrive. For now, treat `error` as "a value that knows how to describe itself.")

---

### A worked example

Combining a multi-return function with the canonical error-check shape:

```go
package main

import (
	"errors"
	"fmt"
)

var errDivByZero = errors.New("division by zero")

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errDivByZero
	}
	return a / b, nil
}

func main() {
	q, err := divide(10, 3)
	if err != nil {
		fmt.Println("first error:", err)
		return
	}
	fmt.Println("10 / 3 =", q)

	q2, err := divide(10, 0)
	if err != nil {
		fmt.Println("second error:", err)
		return
	}
	fmt.Println("10 / 0 =", q2)   // never reached
}
```

Output:

```
10 / 3 = 3
second error: division by zero
```

Notice the early return after each error check. No `else` — the rest of the function only runs when there's no error.

---

### Common mistake

Ignoring the error by assigning it to `_`:

```go
q, _ := divide(10, 0)    // squashes the error
fmt.Println(q)           // prints 0 — but there was an error you ignored
```

Go's compiler doesn't flag this (`_` means "I deliberately don't want this"), but `golangci-lint`'s `errcheck` does. Almost always you want to check the error and handle it. If you really do want to ignore it, leave a comment explaining why — your future self will thank you.

---

### Recap

- Functions that can fail return `(result, error)`. The error is always the last return.
- `errors.New("msg")` creates a simple error. Stash sentinels in package-level `var`s if you want comparisons.
- `if err != nil { return ..., err }` is *the* Go shape for error propagation.
- Don't silently swallow errors with `_` unless you have a clear reason.

---

## Concept 4: Testing with `testing.T`

### Motivation

Go ships with a testing framework in the standard library — no third-party libraries required. Tests live in `*_test.go` files alongside your code. The `go test` command finds and runs them. Three patterns will carry you through Phase 1: the basic `func TestXxx(t *testing.T)` shape, the table-test pattern with `[]struct{}`, and the difference between `t.Errorf` (report and continue) and `t.Fatalf` (report and stop).

---

### The basics

A minimal test:

```go
package mypkg

import "testing"

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d, want %d", got, want)
	}
}
```

Three rules:

1. **Filename ends in `_test.go`.** `go build` ignores these; `go test` runs them.
2. **Function name starts with `Test`** and takes `*testing.T`. Any other signature is invisible to `go test`.
3. **Report failures with `t.Errorf`/`t.Fatalf`.** The message you pass becomes the test failure output — make it useful (input + got + want).

Run them with `go test ./...` from anywhere in the module, or `go test ./path/to/package` to scope.

---

### A worked example — the table test pattern

When you have many cases to check, write one test that iterates over a slice of structs:

```go
package mypkg

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
```

What `t.Run(name, func)` buys you:

- Each case is reported as a separate sub-test (`TestAdd/positive`, `TestAdd/zero`).
- One case failing doesn't stop the others — the loop keeps going.
- You can run a single case with `go test -run TestAdd/positive`.

This is the shape used in your warm-up tests AND the shape you'll fill in for the main exercise.

`t.Errorf` vs `t.Fatalf`:

- **`t.Errorf("...")`** — record the failure and **keep running** the current test function. Use it for assertions where checking the rest still gives useful information.
- **`t.Fatalf("...")`** — record the failure and **stop** the current test function. Use it when continuing would crash (e.g., dereferencing a nil pointer) or produce noise.

---

### A note on `go vet`

Alongside `go test`, get into the habit of running `go vet ./...`. It's a sibling stdlib tool that catches subtle bugs — wrong format verbs in `Printf`, shadowed variables, unreachable code, missing struct field tags. CI runs it, and from this lesson on, the lesson README's "How to run" mentions it explicitly.

```bash
go vet ./...
```

---

### Common mistake

Using `t.Errorf` when you should use `t.Fatalf`:

```go
result, err := mightFail()
if err != nil {
	t.Errorf("unexpected error: %v", err)   // wrong: keeps going
}
result.DoSomething()                          // crashes if result is nil
```

Fix: `t.Fatalf` when the next line would crash, `t.Errorf` when it wouldn't:

```go
result, err := mightFail()
if err != nil {
	t.Fatalf("unexpected error: %v", err)   // stops the test cleanly
}
result.DoSomething()
```

---

### Recap

- Tests live in `*_test.go` files; `go test ./...` runs them.
- `func TestXxx(t *testing.T)` is the only test signature `go test` recognises.
- Table tests: a `[]struct{}` of cases, a `for ... t.Run` loop. Sub-tests are individually addressable and one failure doesn't block the rest.
- `t.Errorf` → record and continue. `t.Fatalf` → record and stop.
- `go vet ./...` catches the bugs `go test` won't.

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `Add(a, b int) int` — the simplest function in the course. Return `a + b`.
- `MinMax(xs ...int) (int, int, error)` — variadic input, multi-return, error for empty input. Return `(0, 0, errEmptyMinMax)` when `len(xs) == 0`; otherwise return the smallest, the largest, and `nil`.

Tests are pre-written.

```bash
cd lessons/04-functions/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `Categorise(amount float64) string` — same signature and behaviour as lesson 03's. Use whichever shape (if-chain or tagless switch) you prefer.
- `FormatExpense(date string, amount float64, cat string) string` — return `"YYYY-MM-DD  €AMOUNT  category"` using `fmt.Sprintf` with `%s  €%-7.2f %s`.

For the **first time, you write the tests yourself** — `exercises/main_test.go` ships as a skeleton with the struct shape and the loop scaffold; you fill in the case rows and the assertion body. The README walks through the test shape step by step. Cases to include:

- For Categorise — at least 5 cases including the boundary values (9.99, 10, 50, 50.01).
- For FormatExpense — at least 3 cases. Build the expected output strings by hand using the format.

```bash
cd lessons/04-functions/exercises
go test -v
```

Note:
For live: emphasise the test-authoring jump. Walk through writing one Categorise case live on the projector — name the case, fill in the input, type out the expected string, run the test, watch it fail with a useful message, fix the implementation, watch it pass. That round-trip is the single most valuable habit Phase 1 hands students.

---

## What we learned

- Multi-return values are everywhere in Go (especially `(result, error)`).
- Variadic args (`xs ...T`) arrive as `[]T`. Spread a slice with `slice...`.
- Named returns are sugar for documenting result variables.
- `defer` schedules cleanup at function exit; arguments evaluate at the defer line, the call runs at return time.
- `error` is Go's failure convention. `errors.New` makes a sentinel; `if err != nil { return ..., err }` is the canonical propagation shape.
- Tests live in `*_test.go`; table tests with `[]struct{}` and `t.Run` are *the* pattern for parameterised testing. `t.Errorf` vs `t.Fatalf` is the recovery-vs-stop distinction.
- `go vet ./...` — run it alongside `go test`.

---

## Up next

Lesson 05 — Composite types I: arrays, slices, maps (`[]T`, `len`/`cap`, `append`, `range`, `map[K]V`).
````

- [ ] **Step 3: Verify the deck renders correctly**

```bash
make slides-dev LESSON=04-functions &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/04-functions/slides/slides.md
curl -sS http://localhost:8000/lessons/04-functions/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/04-functions/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/04-functions/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 4: Verify markdown structure**

```bash
grep -c '<div class="title-slide-grid">' lessons/04-functions/slides/slides.md
grep -c "<h1>Functions" lessons/04-functions/slides/slides.md
grep -c "^## What we learned" lessons/04-functions/slides/slides.md
grep -c "^## Up next" lessons/04-functions/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 5: Commit**

```bash
git add lessons/04-functions/slides/slides.md
git commit -m "feat(lesson-04): slides — Functions & first tests (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/04-functions/README.md` with the full lesson 04 self-study prose. Mirrors the slide narrative, includes common-mistake examples for every concept, walks through the test-writing workflow for the main exercise, and adds `go vet ./...` to the tooling thread.

**Files:**
- Replace: `lessons/04-functions/README.md`

- [ ] **Step 1: Replace `lessons/04-functions/README.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device; in the file use only three-backtick fences. File starts with `# Lesson 04: Functions & first tests`.

````markdown
# Lesson 04: Functions & first tests

## Learning goals

- Write functions with multiple return values, named returns, and variadic arguments.
- Use `defer` to schedule cleanup at function exit.
- Return and handle `error`s with the canonical `if err != nil { return …, err }` shape.
- Write your first tests using `testing.T`, the table-test pattern, and `t.Errorf` vs `t.Fatalf`.
- Get into the habit of running `go vet ./...` alongside `go test`.

## Prerequisites

- Lesson 01 — `go run`, `package main`, `fmt`.
- Lesson 02 — variable forms, types, arithmetic.
- Lesson 03 — `if`/`else`, `for`, `switch`, early returns.

## Concepts

### Function shape: multi-return, named returns, variadic

You've used functions since lesson 01. Three additional shapes round out the toolkit:

```go
// Multi-return — two values.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// Named returns — variables declared in the signature.
// The bare `return` returns whatever's in q and r.
func divmodNamed(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

// Variadic — accepts zero or more ints.
// Inside the function, xs is []int.
func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}
```

Calling them:

```go
q, r := divmod(17, 5)       // 3 2
q2, r2 := divmodNamed(17, 5) // 3 2

sum()                        // 0
sum(1, 2, 3)                 // 6
sum(1, 2, 3, 4)              // 10

xs := []int{10, 20, 30}
sum(xs...)                   // 60 — note the trailing ...
```

Three points:

- **Multi-return** is everywhere in the stdlib — most notably for `(result, error)` pairs (concept 3).
- **Named returns** declare result variables in the signature. The bare `return` returns them all. Useful when the return values are non-obvious from the types alone.
- **Variadic** — `xs ...T` is `xs []T` inside the function. Spread an existing slice with `slice...`.

**Common mistake.** Forgetting the trailing `...` when passing a slice to a variadic:

```go
nums := []int{1, 2, 3}
fmt.Println(sum(nums))    // compile error: cannot use nums as int
```

Fix: `sum(nums...)`.

### `defer`

`defer expr` schedules `expr` to run when the surrounding function returns. Cleanup that needs to happen regardless of return path (early return, normal return, panic) goes here.

```go
func main() {
	defer fmt.Println("world")  // runs last
	defer fmt.Println("middle") // runs second
	fmt.Println("hello")        // runs first
}
// Output:
// hello
// middle
// world
```

Deferred calls run in **LIFO order** (last deferred, first to run).

The most common use is paired with resource acquisition — open a file, immediately defer the close:

```go
// You'll write this exact shape hundreds of times once we hit lesson 13's I/O.
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close()
// ... use f ...
```

`f.Close()` will run when the surrounding function returns, no matter how. You can't forget.

**Common mistake.** Argument evaluation happens at the `defer` line, not at function exit:

```go
i := 1
defer fmt.Println("deferred i =", i)  // i evaluated to 1 right now
i = 2
fmt.Println("current i =", i)
```

Output:

```
current i = 2
deferred i = 1
```

If you want late binding, wrap the call in a function literal: `defer func() { fmt.Println(i) }()`.

### Errors

Go functions that can fail return an `error` as their last result. Callers check it with `if err != nil`. There's no `try`/`catch`; error handling is just another value to inspect.

```go
import (
	"errors"
	"fmt"
)

var errDivByZero = errors.New("division by zero")

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errDivByZero
	}
	return a / b, nil
}

func main() {
	q, err := divide(10, 3)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("10 / 3 =", q)
}
```

Two patterns to internalise:

- **`errors.New("message")`** — the simplest way to make an error. Stash recurring ones in package-level `var`s as "sentinels" (e.g. `var errDivByZero = errors.New(...)`) — that lets callers compare or check via `errors.Is` (lesson 11 covers `errors.Is` formally; for now, `==` works on sentinels).
- **`if err != nil { return ..., err }`** — *the* canonical guard. You'll write it thousands of times.

`error` itself is a built-in type with one method: `Error() string`. (What makes that work is interfaces — lesson 10.)

**Common mistake.** Silently ignoring errors with `_`:

```go
q, _ := divide(10, 0)
fmt.Println(q)           // prints 0 — but there was an error
```

`golangci-lint`'s `errcheck` flags it. If you really do want to discard the error, leave a comment explaining why.

### Testing with `testing.T`

Tests live in `*_test.go` files alongside your code. `go test ./...` finds and runs them.

```go
package mypkg

import "testing"

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d, want %d", got, want)
	}
}
```

Three rules:

1. Filename ends in `_test.go`.
2. Function name starts with `Test`, takes `*testing.T`.
3. Report failures with `t.Errorf` (record and continue) or `t.Fatalf` (record and stop).

The **table-test pattern** scales the basic shape to many cases:

```go
func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
```

`t.Run(name, func)` makes each case a separately-named sub-test. Sub-tests are individually addressable (`go test -run TestAdd/positive`), and one failing case doesn't block the rest.

`t.Errorf` vs `t.Fatalf`:

- **`t.Errorf`** — record the failure, keep running. Use when the rest of the test still gives useful information.
- **`t.Fatalf`** — record the failure, stop the test function. Use when continuing would crash or produce noise.

**Common mistake.** Using `t.Errorf` when the next line would crash:

```go
result, err := mightFail()
if err != nil {
	t.Errorf("unexpected error: %v", err)   // wrong: keeps going
}
result.DoSomething()                          // crashes if result is nil
```

Fix: `t.Fatalf` when continuing would crash.

## Exercise: warm-up

Open `exercises/warmup.go`. Two functions:

- `Add(a, b int) int` — return `a + b`.
- `MinMax(xs ...int) (int, int, error)` — return the smallest and largest, plus `nil`. For empty input, return `(0, 0, errEmptyMinMax)` (the sentinel is already declared at the top of the file).

Tests are pre-written. Implement until they pass.

## Exercise: main

**This is the first lesson where you write tests yourself.** Open `exercises/main_test.go` — it ships as a *skeleton*:

```go
func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		// TODO: add at least 5 cases here...
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call Categorise(tc.amount), compare with tc.want...
			_ = tc
		})
	}
}
```

Your job is to fill in:

1. **The `cases` slice** — at least 5 entries for `TestCategorise`, at least 3 for `TestFormatExpense`. Cover the snack/regular/splurge branches and the boundary amounts (9.99, 10, 50, 50.01) for Categorise; for FormatExpense, build the expected output strings by hand using the format string in the doc comment.
2. **The `t.Run` body** — call the function, compare against `tc.want`, report a failure with `t.Errorf` if they differ. Replace the `_ = tc` line — it's just there so the empty skeleton compiles.

Look at `exercises/warmup_test.go` (and `solutions/main_test.go` if you want to compare to the reference) for the shape.

The two implementations in `exercises/main.go`:

- `Categorise(amount float64) string` — return `"snack"` (`amount < 10`), `"regular"` (`10 <= amount <= 50`), or `"splurge"` (`amount > 50`). Same as lesson 03.
- `FormatExpense(date string, amount float64, cat string) string` — return `"YYYY-MM-DD  €AMOUNT  category"` using `fmt.Sprintf` with `"%s  €%-7.2f %s"`. Note the spacing in the expected outputs in the doc comment.

## How to run

```bash
cd lessons/04-functions/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main
```

This lesson introduces two new habits — get into both of them now and you won't have to come back:

```bash
gofmt -w .       # reformat your code to canonical Go style
go vet ./...     # find subtle bugs go test won't catch
```

`gofmt` is non-negotiable (CI enforces it). `go vet` is run by CI too; running it locally catches the surprises before you commit. From this lesson onwards, every lesson README mentions it.

Once both exercises pass, compare your code (especially your tests) against `solutions/`.

## Going further

### Read

- [A Tour of Go — Functions](https://go.dev/tour/basics/4) — multi-return values and named returns with playgrounds.
- [Effective Go — Defer](https://go.dev/doc/effective_go#defer) — short explanation of `defer`'s semantics and why Go's compiler likes resource pairs.
- [Go blog — Error handling and Go](https://go.dev/blog/error-handling-and-go) — the canonical introduction to the `error` type and the `if err != nil` shape (skip the `errors.Is`/`errors.As` bits for now — lesson 11).
- [`testing` package docs](https://pkg.go.dev/testing) — the stdlib reference. You'll meet `t.Helper`, `t.TempDir`, and subtests-beyond-table-tests in lesson 15.

### Try

- **A third return for divmod.** Add `divmodChecked(a, b int) (q, r int, err error)` that returns `errDivByZero` when `b == 0`. Write the test for it.
- **Run a single sub-test.** Pick one of your `TestCategorise/...` sub-tests and run only that one: `go test -run "TestCategorise/upper-snack-edge" -v`. Useful when you want to focus on one failing case.
- **`go vet` finds a real bug.** Try writing `fmt.Printf("%d", "hello")` in a small program and run `go vet ./...` — it'll catch the type mismatch before you run the program. Format-string bugs are the most common thing `go vet` finds.
````

- [ ] **Step 2: Verify the README structure**

```bash
for h in "^# Lesson 04: Functions & first tests$" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  c=$(grep -c "$h" lessons/04-functions/README.md)
  echo "  $h => $c"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/04-functions/README.md   # expect 4
grep -c "gofmt" lessons/04-functions/README.md                       # expect >= 1
grep -c "go vet" lessons/04-functions/README.md                      # expect >= 2 (added to tooling thread)
```

Expected: all section headings 1; Common mistake count 4; gofmt count ≥ 1; go vet count ≥ 2.

- [ ] **Step 3: Commit**

```bash
git add lessons/04-functions/README.md
git commit -m "docs(lesson-04): README — Functions & first tests self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises). Lesson 04 solutions pass; lessons 01-03 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — warm-up must fail; main must "pass vacuously"**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 04's `TestAdd` and `TestMinMax` FAIL with panics; `TestCategorise` and `TestFormatExpense` PASS (with 0 sub-tests because the skeleton case lists are empty). Make exits 0 because the `-` prefix ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=04-functions 2>&1 | tail -30
```

Expected: exercise warmup tests fail (ignored), exercise main tests pass vacuously, solution tests pass fully.

- [ ] **Step 4: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 5: go vet clean (lesson 04 introduces it formally)**

```bash
go vet ./...
```

Expected: no output (clean exit code).

- [ ] **Step 6: Slides server smoke test**

```bash
make slides-dev LESSON=04-functions &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/04-functions/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/04-functions/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 7: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/04-functions/slides/slides.md && echo OK
grep -q "Functions" dist/index.html && echo "index lists lesson 04"
rm -rf dist
```

Expected: three success lines. The landing page should now list lessons 01-04.

- [ ] **Step 8: Final repository sanity check**

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch (plan + scaffold + warm-up + main + slides + README).

This task makes no commit — it is verification only.

---

## Done definition

After Task 6:

- `lessons/04-functions/` contains 12 files with full lesson content.
- `make test` passes; `make test-exercises` shows the warm-up exercise tests failing by design and the main exercise tests passing vacuously (empty case lists).
- `make test-lesson LESSON=04-functions` runs both sides correctly.
- `make slides-dev LESSON=04-functions` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan H — Lesson 05 (Composite types I: slices and maps).** Same per-lesson pattern. Lesson 05 covers `[]T` properly (len/cap, append, slice headers), `range` on slices and maps, and `map[K]V` (zero value, lookup with `_, ok := m[k]`, deletion). Warm-up: `Sum` / `Max` / `Unique`. Main: `TotalsByCategory(amounts []float64, categories []string) map[string]float64` — the first lesson where the main exercise produces a `map`. Students continue writing their own tests (no more skeleton scaffolding from lesson 05 onwards).
