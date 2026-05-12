# Plan H — Lesson 05 (Composite types I: slices and maps) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the fifth lesson of Phase 1 — students learn slices (declaration shapes, `len`/`cap`, indexing, slicing, `append`, nil-vs-empty), the formal treatment of `range` (which they met informally in lesson 03), and maps (`map[K]V`, the `v, ok := m[k]` exists check, `delete`, range over a map). The main exercise composes both: given parallel `[]float64` amounts and `[]string` categories, return a `map[string]float64` of totals per category.

**Architecture:** Same per-lesson pattern as Plans D-G. Six tasks: scaffold, warm-up, main, slides, README, end-to-end verify. Three warm-up functions (`WarmupSum`, `WarmupMax`, `WarmupUnique`) plus one main function (`TotalsByCategory`). Skeleton tests across **both** warm-up and main (escalation from lesson 04 — confirmed with user). Heavy-explanatory slide deck with **four** concept blocks (slices basics → append & growth → range formally → maps).

**Tech Stack:** Go 1.23 stdlib only (`errors`, `fmt`, `testing`). No third-party dependencies. Reveal.js 5.1.0 for the deck.

---

## Scope

Plan H produces lesson 05 only. After Plan H lands:

- `lessons/05-slices-maps/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design (warmup) and passing vacuously (main, empty case slices).
- `make slides-dev LESSON=05-slices-maps` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):** Lessons 06-08 (Plans I-K).

### Design decisions made during planning

These are choices within the spec's latitude. Easy to flip during plan review.

1. **`WarmupMax` returns `(int, error)`.** Spec literal says `Max(xs []int) int`, but `Max([]int{})` is undefined — silently returning `0` would mask the error case. Per user choice: match lesson 04's `WarmupMinMax` pattern with `(int, error)` and a sentinel for empty input. Reinforces the lesson 04 idiom one lesson later.

2. **`WarmupSum` stays `(int)`-only.** Sum of an empty slice is well-defined as `0` (the additive identity). No error needed. Different from `Max` for a substantive reason — sum has a natural empty-input answer; max doesn't.

3. **`TotalsByCategory` checks two error conditions.** Spec says "Reject empty input with a clear error." Plan H adds **length mismatch** (`len(amounts) != len(categories)`) as a second error case — natural for parallel slices and a pedagogically useful "different errors for different bad inputs" example. Two sentinels: `errEmptyInputs` and `errLengthMismatch`. Single-error variants are easy to flip during plan review.

4. **Test scaffolding escalates: skeleton in BOTH warm-up and main.** Per user choice — Phase 1 spec says "Lessons 5-8: Exercises ship with skeleton tests that students must complete." Lesson 04 was the soft ramp (skeleton in main only); lesson 05 escalates. All four `Test*` functions in `exercises/` ship as skeletons (struct + `for ... t.Run` loop scaffold + `_ = tc` placeholder); students fill in `cases` and assertion bodies. `solutions/` ships full reference tests.

5. **Four concepts in the slide deck.** (1) Slices: declaration, indexing, slicing, len/cap; (2) `append` and growth — including the `xs = append(xs, v)` assignment requirement; (3) `range` — formal treatment now (lesson 03 introduced it informally); (4) Maps: `map[K]V`, zero value, `v, ok := m[k]`, `delete`, range. Matches the 4-concept rhythm.

6. **Nil-vs-empty distinction taught explicitly.** Slices: `var xs []int` (nil but len 0) vs `xs := []int{}` (non-nil empty). Maps: `var m map[K]V` (nil — assignment panics) vs `m := map[K]V{}` or `make(map[K]V)` (usable). This is the trip-everyone-up class of bug; the slides' common-mistake sections lean on it.

7. **`Warmup*` prefix maintained.** Lesson 04's post-review fix established that warm-up symbols use the `Warmup*` prefix even when there's no name collision with the main exercise. Plan H ships `WarmupSum`, `WarmupMax`, `WarmupUnique` from the start.

8. **`range` loop-variable semantics (Go 1.22+).** Go 1.22 changed `range` so the loop variable is fresh per iteration — the historical "value-is-shared-across-iterations" gotcha is gone in the repo's Go 1.23 build. The slides briefly mention this as historical context (so students reading older Go code aren't surprised); the lesson's actual code uses Go 1.23 semantics throughout.

---

## Plans F/G lessons-learned applied here

1. **Warmup* prefix from the start** — Plan G's literal `Add`/`MinMax` missed the Phase 1 spec's `Warmup*` convention; the fix commit `c2422f4` restored it. Plan H ships `WarmupSum`/`WarmupMax`/`WarmupUnique` from task 2.

2. **Lint step in every Go-touching task** — same as previous plans. `.golangci.yml`'s `lessons/*/exercises/` exclusion (Plan F) still applies, so panic-stubs and `_ = tc` placeholders don't trip lint.

3. **`→` arrow consistency** — Unicode arrow in doc-comment examples; gofmt may reformat the example block to canonical godoc form (blank line + tab indent), arrows must survive.

4. **Common-mistake content in README** — every concept's README section includes the same common-mistake example as the slides (4 total).

5. **Slides written inline by controller** — Plan F's slide-deck single-write timed out a subagent. Plan G handled this by having the controller write the slide deck directly. Plan H continues this pattern: Tasks 1-3 (and 6) can be subagent-driven; Tasks 4 (slides) and 5 (README) are large enough that the controller writes them directly to avoid timeout risk.

6. **`gofmt -w .` and `go vet ./...` mentions** — lesson 04 introduced `go vet` to the tooling thread. Lesson 05's README continues both.

7. **Verify FormatExpense-style alignment with a smoke test** — when a function's output is column-aligned (Plan G's `FormatExpense`), test cases need to byte-match what `fmt.Sprintf` actually produces. Plan H's `TotalsByCategory` returns a map (no column alignment), so this gotcha doesn't apply here — but flagged for future plans.

---

## File Structure

After Plan H:

```
lessons/05-slices-maps/                     (new — Plan H's deliverable)
├── README.md                               (Task 5)
├── slides/
│   ├── index.html                          (Task 1; unchanged)
│   ├── slides.md                           (Task 4 — written inline by controller)
│   └── assets/.gitkeep                     (Task 1; unchanged)
├── exercises/
│   ├── warmup.go                           (Task 2 — WarmupSum + WarmupMax + WarmupUnique stubs)
│   ├── warmup_test.go                      (Task 2 — SKELETON tests, students fill cases)
│   ├── main.go                             (Task 3 — TotalsByCategory stub)
│   └── main_test.go                        (Task 3 — SKELETON tests, students fill cases)
└── solutions/
    ├── warmup.go                           (Task 2 — implementations)
    ├── warmup_test.go                      (Task 2 — full reference tests)
    ├── main.go                             (Task 3 — implementation)
    └── main_test.go                        (Task 3 — full reference tests)
```

### Decomposition rationale

Same pattern as plans D, E, F, G. Each task has one focused output and a clear verification step. The exercises/solutions divergence (skeleton vs full reference tests) now applies to **both** warmup and main test files, not just main as in lesson 04. The README walks through what's expected of students for both.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-h-lesson-05-slices-maps` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)

---

## Task 1: Scaffold lesson 05

**Files:**
- Create: `lessons/05-slices-maps/` (12 files via the scaffolder)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-h-lesson-05-slices-maps`, clean working tree, exactly one commit ahead of `main` (the Plan H doc).

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=05-slices-maps
```

Expected: prints `created lesson 05-slices-maps under lessons/`. The scaffolder produces 12 files with template placeholders.

- [ ] **Step 3: Verify the 12-file tree**

```bash
find lessons/05-slices-maps -type f | sort
```

Expected (12 lines):

```
lessons/05-slices-maps/README.md
lessons/05-slices-maps/exercises/main.go
lessons/05-slices-maps/exercises/main_test.go
lessons/05-slices-maps/exercises/warmup.go
lessons/05-slices-maps/exercises/warmup_test.go
lessons/05-slices-maps/slides/assets/.gitkeep
lessons/05-slices-maps/slides/index.html
lessons/05-slices-maps/slides/slides.md
lessons/05-slices-maps/solutions/main.go
lessons/05-slices-maps/solutions/main_test.go
lessons/05-slices-maps/solutions/warmup.go
lessons/05-slices-maps/solutions/warmup_test.go
```

- [ ] **Step 4: Commit the scaffolded skeleton**

```bash
git add lessons/05-slices-maps/
git commit -m "feat(lessons): scaffold lesson 05-slices-maps skeleton"
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: all packages pass. Lesson 05 solutions pass (the scaffolded `Greet`/`WarmupGreet` work).

- [ ] **Step 6: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch.

---

## Task 2: Author the warm-up exercise

Three warm-up functions. `WarmupSum(xs []int) int` is the simplest (zero-value answer for empty). `WarmupMax(xs []int) (int, error)` mirrors lesson 04's `WarmupMinMax` shape with a sentinel for empty input. `WarmupUnique(xs []string) []string` preserves order and deduplicates — the natural use of a map-as-a-set.

Tests **escalate to skeleton** — all three test functions in `exercises/warmup_test.go` ship with the table-test struct and the `for ... t.Run` loop scaffold but empty `cases` and `_ = tc` body placeholders. Students fill them in. `solutions/warmup_test.go` ships the full reference.

**Files:**
- Replace: `lessons/05-slices-maps/exercises/warmup.go`
- Replace: `lessons/05-slices-maps/exercises/warmup_test.go`
- Replace: `lessons/05-slices-maps/solutions/warmup.go`
- Replace: `lessons/05-slices-maps/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/05-slices-maps/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 05: Slices and maps.
//
// This file holds the WARM-UP exercise. Three functions over slices that
// also let you write your first table tests of your own (warmup_test.go
// ships as a skeleton — fill in the cases and assertion bodies).
package exercises

import "errors"

// errEmptyWarmupMax is returned by WarmupMax when called with no values.
var errEmptyWarmupMax = errors.New("WarmupMax: requires at least one value")

// WarmupSum returns the sum of the integers in xs.
//
// An empty slice returns 0 — the additive identity is a well-defined answer.
//
// Examples:
//
//	WarmupSum([]int{1, 2, 3})  → 6
//	WarmupSum([]int{})         → 0
//	WarmupSum(nil)             → 0   (nil slice walks zero times under range)
//
// Hint: `for _, v := range xs { total += v }` is the canonical shape.
func WarmupSum(xs []int) int {
	panic("TODO: walk xs with range, accumulate into a total, return it")
}

// WarmupMax returns the largest int in xs, plus an error.
//
// Non-empty input returns (max, nil). Empty input returns (0, errEmptyWarmupMax).
// Unlike WarmupSum, "max of nothing" has no defensible default — hence the error.
//
// Examples:
//
//	WarmupMax([]int{3, 1, 4, 1, 5, 9, 2, 6})  → (9, nil)
//	WarmupMax([]int{-3, -1, -7})              → (-1, nil)
//	WarmupMax([]int{42})                      → (42, nil)
//	WarmupMax([]int{})                        → (0, error)
//
// Hint: guard the empty case first (early return). Then seed max with
// xs[0] and walk xs[1:] with range, updating when a larger value is found.
func WarmupMax(xs []int) (int, error) {
	panic("TODO: return error for empty; otherwise return the maximum")
}

// WarmupUnique returns the values of xs in their original order with
// duplicates removed. The original slice is not modified.
//
// Examples:
//
//	WarmupUnique([]string{"a", "b", "a", "c", "b"})  → ["a", "b", "c"]
//	WarmupUnique([]string{"x"})                      → ["x"]
//	WarmupUnique([]string{})                         → []   (non-nil empty slice)
//	WarmupUnique(nil)                                → []
//
// Hint: a `map[string]bool` makes a great "seen" set. For each value,
// check `seen[v]`; if not seen, append to the result and set `seen[v] = true`.
// Always return a non-nil slice — initialise with `result := []string{}`.
func WarmupUnique(xs []string) []string {
	panic("TODO: dedupe while preserving order; use a map[string]bool as a seen-set")
}
```

- [ ] **Step 2: Replace `lessons/05-slices-maps/exercises/warmup_test.go`** with (SKELETON):

```go
package exercises

import (
	"errors"
	"reflect"
	"testing"
)

// TestWarmupSum is a SKELETON. The struct and the loop scaffold are in
// place; you fill in the cases and the t.Run body.
//
// Cases to cover: a basic positive case, a single-element case, an
// empty slice, a nil slice, and a case mixing positive and negative.
func TestWarmupSum(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
		want int
	}{
		// TODO: add at least 4 cases. Use the examples in the WarmupSum
		// doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call WarmupSum(tc.xs), compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}

// TestWarmupMax is a SKELETON. Similar shape to TestWarmupSum, but the
// function returns (int, error) — your assertion needs to branch on the
// wantErr flag.
//
// Cases to cover: a basic case, a single-element case, all-negative,
// AND the empty-input case (which should return errEmptyWarmupMax).
func TestWarmupMax(t *testing.T) {
	cases := []struct {
		name    string
		xs      []int
		want    int
		wantErr bool
	}{
		// TODO: at least 4 cases. Include the empty case (wantErr: true).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: gotMax, err := WarmupMax(tc.xs)
			// If tc.wantErr: assert err is non-nil AND errors.Is(err, errEmptyWarmupMax).
			// Otherwise: assert err == nil AND gotMax == tc.want.
			_ = tc
			_ = errors.Is // satisfy the import while you write the test
		})
	}
}

// TestWarmupUnique is a SKELETON. Compare slices with reflect.DeepEqual.
// Be careful with nil vs empty: WarmupUnique always returns a non-nil slice.
func TestWarmupUnique(t *testing.T) {
	cases := []struct {
		name string
		xs   []string
		want []string
	}{
		// TODO: at least 4 cases. Cover duplicates-with-order-preserved,
		// a single-element case, an empty input, and a nil input.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := WarmupUnique(tc.xs)
			// Assert got != nil (it must be a non-nil empty slice for empty/nil input).
			// Assert reflect.DeepEqual(got, tc.want) is true.
			_ = tc
			_ = reflect.DeepEqual // satisfy the import while you write the test
		})
	}
}
```

> Note: the `_ = errors.Is` / `_ = reflect.DeepEqual` lines exist so the imports compile while the body is unwritten. Students delete these once they use the helpers in their assertions.

- [ ] **Step 3: Replace `lessons/05-slices-maps/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 05: Slices and maps.
//
// This file holds the warm-up reference solution.
package solutions

import "errors"

// errEmptyWarmupMax is returned by WarmupMax when called with no values.
var errEmptyWarmupMax = errors.New("WarmupMax: requires at least one value")

// WarmupSum returns the sum of the integers in xs. Empty/nil input → 0.
func WarmupSum(xs []int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

// WarmupMax returns the largest int in xs, plus an error for empty input.
func WarmupMax(xs []int) (int, error) {
	if len(xs) == 0 {
		return 0, errEmptyWarmupMax
	}
	max := xs[0]
	for _, v := range xs[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// WarmupUnique returns xs with duplicates removed, preserving original order.
// Always returns a non-nil slice.
func WarmupUnique(xs []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, v := range xs {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
```

- [ ] **Step 4: Replace `lessons/05-slices-maps/solutions/warmup_test.go`** with full reference tests (`package solutions`):

```go
package solutions

import (
	"errors"
	"reflect"
	"testing"
)

func TestWarmupSum(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
		want int
	}{
		{"basic", []int{1, 2, 3}, 6},
		{"single", []int{42}, 42},
		{"mixed-signs", []int{-3, 5, -7, 2}, -3},
		{"all-negative", []int{-1, -2, -3}, -6},
		{"empty", []int{}, 0},
		{"nil", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WarmupSum(tc.xs); got != tc.want {
				t.Errorf("WarmupSum(%v) = %d, want %d", tc.xs, got, tc.want)
			}
		})
	}
}

func TestWarmupMax(t *testing.T) {
	cases := []struct {
		name    string
		xs      []int
		want    int
		wantErr bool
	}{
		{"unsorted", []int{3, 1, 4, 1, 5, 9, 2, 6}, 9, false},
		{"sorted-ascending", []int{1, 2, 3, 4, 5}, 5, false},
		{"sorted-descending", []int{5, 4, 3, 2, 1}, 5, false},
		{"single", []int{42}, 42, false},
		{"all-negative", []int{-3, -1, -7}, -1, false},
		{"empty", []int{}, 0, true},
		{"nil", nil, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := WarmupMax(tc.xs)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("WarmupMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyWarmupMax) {
					t.Errorf("WarmupMax(%v) error = %v, want errEmptyWarmupMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("WarmupMax(%v) unexpected error: %v", tc.xs, err)
			}
			if got != tc.want {
				t.Errorf("WarmupMax(%v) = %d, want %d", tc.xs, got, tc.want)
			}
		})
	}
}

func TestWarmupUnique(t *testing.T) {
	cases := []struct {
		name string
		xs   []string
		want []string
	}{
		{"basic-dedupe", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"no-dupes", []string{"x", "y", "z"}, []string{"x", "y", "z"}},
		{"all-same", []string{"x", "x", "x"}, []string{"x"}},
		{"single", []string{"x"}, []string{"x"}},
		{"empty", []string{}, []string{}},
		{"nil", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupUnique(tc.xs)
			if got == nil {
				t.Fatalf("WarmupUnique(%v) returned nil; expected a non-nil slice", tc.xs)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("WarmupUnique(%v) = %v, want %v", tc.xs, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/05-slices-maps/
```

Expected: no output. `gofmt -w` if needed.

- [ ] **Step 6: Verify the warm-up exercise tests "pass vacuously"**

Because exercises ships skeleton tests with empty `cases`, the test functions report PASS with 0 sub-tests. Students filling in cases will see them fail until their implementations are right.

```bash
go test ./lessons/05-slices-maps/exercises/... -run Warmup -v 2>&1 | tail -20
```

Expected: TestWarmupSum, TestWarmupMax, TestWarmupUnique all PASS with no sub-tests.

- [ ] **Step 7: Verify the warm-up solution tests pass**

```bash
go test ./lessons/05-slices-maps/solutions/... -run Warmup -v 2>&1 | tail -30
```

Expected: TestWarmupSum (6 sub-tests), TestWarmupMax (7 sub-tests), TestWarmupUnique (6 sub-tests) PASS.

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: green, `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/05-slices-maps/exercises/warmup.go lessons/05-slices-maps/exercises/warmup_test.go \
        lessons/05-slices-maps/solutions/warmup.go lessons/05-slices-maps/solutions/warmup_test.go
git commit -m "feat(lesson-05): warm-up — WarmupSum + WarmupMax + WarmupUnique (skeleton tests)"
```

---

## Task 3: Author the main exercise

`TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error)` composes parallel slices into a per-category total map. Two error cases: empty input AND length mismatch.

**Files:**
- Replace: `lessons/05-slices-maps/exercises/main.go`
- Replace: `lessons/05-slices-maps/exercises/main_test.go`
- Replace: `lessons/05-slices-maps/solutions/main.go`
- Replace: `lessons/05-slices-maps/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/05-slices-maps/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 05: Slices and maps.
//
// This file holds the MAIN exercise. TotalsByCategory aggregates parallel
// expense slices into a per-category total map. The matching test file
// ships as a skeleton — fill in cases and assertions.
package exercises

import "errors"

// errEmptyInputs is returned by TotalsByCategory when either input slice
// is empty.
var errEmptyInputs = errors.New("TotalsByCategory: amounts and categories must be non-empty")

// errLengthMismatch is returned by TotalsByCategory when the two input
// slices have different lengths (they're meant to be parallel).
var errLengthMismatch = errors.New("TotalsByCategory: amounts and categories must have the same length")

// TotalsByCategory returns a map keyed by category whose values are the
// summed amounts for that category.
//
// Two error conditions, checked in order:
//  1. Empty inputs (either slice has length 0) → errEmptyInputs.
//  2. Length mismatch (len(amounts) != len(categories)) → errLengthMismatch.
//
// On success the map contains exactly the distinct categories from the input.
//
// Examples:
//
//	amounts    = []float64{4.50, 12, 75, 9.99}
//	categories = []string{"coffee", "lunch", "rent", "coffee"}
//	TotalsByCategory(amounts, categories) → map[coffee:14.49 lunch:12 rent:75], nil
//
//	TotalsByCategory(nil, nil)             → nil, errEmptyInputs
//	TotalsByCategory([]float64{1}, []string{"a", "b"}) → nil, errLengthMismatch
//
// Hint: guard the error cases first (early return). Then walk one of the
// slices with `for i, cat := range categories` (or amounts) and accumulate
// into a `map[string]float64`. Map zero-value is 0.0 for float64, so
// `totals[cat] += amounts[i]` works whether the key is present or not.
func TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error) {
	panic("TODO: validate inputs (empty + length mismatch); accumulate into map[string]float64")
}
```

- [ ] **Step 2: Replace `lessons/05-slices-maps/exercises/main_test.go`** with (SKELETON):

```go
package exercises

import (
	"errors"
	"reflect"
	"testing"
)

// TestTotalsByCategory is a SKELETON. Fill in the cases and the t.Run body.
//
// Cases to cover:
//   - a basic happy-path case (no duplicate categories)
//   - a case where one category appears multiple times (totals accumulate)
//   - a single-element case
//   - the two error cases: empty inputs AND length mismatch
//
// For the error cases: the wantErr tag tells you which sentinel to expect.
// Use errors.Is to compare. The `want` map should be nil for error cases.
type totalsCase struct {
	name       string
	amounts    []float64
	categories []string
	want       map[string]float64
	wantErr    error // nil, errEmptyInputs, or errLengthMismatch
}

func TestTotalsByCategory(t *testing.T) {
	cases := []totalsCase{
		// TODO: add at least 5 cases here:
		//   1. Basic distinct categories
		//   2. Repeated category (totals accumulate)
		//   3. Single element
		//   4. Empty inputs → wantErr: errEmptyInputs
		//   5. Length mismatch → wantErr: errLengthMismatch
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := TotalsByCategory(tc.amounts, tc.categories)
			//   If tc.wantErr != nil: assert err is non-nil AND errors.Is(err, tc.wantErr).
			//   Otherwise: assert err == nil AND reflect.DeepEqual(got, tc.want).
			_ = tc
			_ = errors.Is
			_ = reflect.DeepEqual
		})
	}
}
```

> Note: the `totalsCase` struct is hoisted outside `TestTotalsByCategory` so its `wantErr` field of type `error` is naturally typed; embedding the struct definition inside the test function would still work but the hoisted form reads more like real-world code. Students don't need to rearrange anything.

- [ ] **Step 3: Replace `lessons/05-slices-maps/solutions/main.go`** with:

```go
// Package solutions is the reference implementation for lesson 05: Slices and maps.
package solutions

import "errors"

// errEmptyInputs is returned by TotalsByCategory when either input slice
// is empty.
var errEmptyInputs = errors.New("TotalsByCategory: amounts and categories must be non-empty")

// errLengthMismatch is returned by TotalsByCategory when the two input
// slices have different lengths.
var errLengthMismatch = errors.New("TotalsByCategory: amounts and categories must have the same length")

// TotalsByCategory aggregates parallel amount/category slices into a per-
// category total map. See the doc comment in exercises/main.go for details.
func TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error) {
	if len(amounts) == 0 || len(categories) == 0 {
		return nil, errEmptyInputs
	}
	if len(amounts) != len(categories) {
		return nil, errLengthMismatch
	}
	totals := map[string]float64{}
	for i, cat := range categories {
		totals[cat] += amounts[i]
	}
	return totals, nil
}
```

- [ ] **Step 4: Replace `lessons/05-slices-maps/solutions/main_test.go`** with full reference tests:

```go
package solutions

import (
	"errors"
	"reflect"
	"testing"
)

type totalsCase struct {
	name       string
	amounts    []float64
	categories []string
	want       map[string]float64
	wantErr    error
}

func TestTotalsByCategory(t *testing.T) {
	cases := []totalsCase{
		{
			"distinct",
			[]float64{4.50, 12, 75},
			[]string{"coffee", "lunch", "rent"},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
			nil,
		},
		{
			"repeated-categories",
			[]float64{4.50, 12, 75, 9.99},
			[]string{"coffee", "lunch", "rent", "coffee"},
			map[string]float64{"coffee": 14.49, "lunch": 12, "rent": 75},
			nil,
		},
		{
			"all-same-category",
			[]float64{1, 2, 3},
			[]string{"x", "x", "x"},
			map[string]float64{"x": 6},
			nil,
		},
		{
			"single",
			[]float64{42},
			[]string{"a"},
			map[string]float64{"a": 42},
			nil,
		},
		{
			"empty-amounts",
			[]float64{},
			[]string{"a"},
			nil,
			errEmptyInputs,
		},
		{
			"empty-both",
			[]float64{},
			[]string{},
			nil,
			errEmptyInputs,
		},
		{
			"nil-both",
			nil,
			nil,
			nil,
			errEmptyInputs,
		},
		{
			"length-mismatch-short-categories",
			[]float64{1, 2, 3},
			[]string{"a", "b"},
			nil,
			errLengthMismatch,
		},
		{
			"length-mismatch-short-amounts",
			[]float64{1, 2},
			[]string{"a", "b", "c"},
			nil,
			errLengthMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TotalsByCategory(tc.amounts, tc.categories)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("TotalsByCategory: nil error, want %v", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("TotalsByCategory: err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("TotalsByCategory: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory: got %v, want %v", got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/05-slices-maps/
```

Expected: no output.

- [ ] **Step 6: Verify exercise main tests "pass vacuously"**

```bash
go test ./lessons/05-slices-maps/exercises/... -v 2>&1 | tail -20
```

Expected: TestTotalsByCategory PASS (0 sub-tests). All warm-up tests also pass vacuously.

- [ ] **Step 7: Verify solution main tests pass**

```bash
go test ./lessons/05-slices-maps/solutions/... -v 2>&1 | tail -40
```

Expected: every test PASSes — TestWarmupSum (6), TestWarmupMax (7), TestWarmupUnique (6), TestTotalsByCategory (9 sub-tests).

- [ ] **Step 8: Verify `make test`, lint, vet pass**

```bash
make test
golangci-lint run ./...
go vet ./...
```

Expected: green, `0 issues.`, clean vet.

- [ ] **Step 9: Commit**

```bash
git add lessons/05-slices-maps/exercises/main.go lessons/05-slices-maps/exercises/main_test.go \
        lessons/05-slices-maps/solutions/main.go lessons/05-slices-maps/solutions/main_test.go
git commit -m "feat(lesson-05): main — TotalsByCategory (parallel slices → map, two error cases)"
```

---

## Task 4: Author the slide deck

Replace `lessons/05-slices-maps/slides/slides.md` with the lesson 05 deck. Four concepts: slices basics → append & growth → range formally → maps.

**Files:**
- Replace: `lessons/05-slices-maps/slides/slides.md`

Note: the controller writes this directly (Plan F's slide-deck single-write timed out a subagent; Plan G and onwards skip the subagent for slides and READMEs).

- [ ] **Step 1: Write `lessons/05-slices-maps/slides/slides.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device. In the file use only three-backtick fences. File starts with `<div class="title-slide-grid">`.

````markdown
<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">05</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Slices and maps</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Declare and index slices, grow them with <code>append</code>, walk them with <code>range</code> properly, and use <code>map[K]V</code> with the <code>v, ok := m[k]</code> exists check and <code>delete</code>.</p>
</div>
</div>
</div>

---

## What we'll cover

- Slices: declaration shapes (`var xs []T`, `xs := []T{...}`, `make([]T, n)`), indexing, slicing (`xs[i:j]`), `len`/`cap`, nil vs empty.
- `append` — growing a slice, the canonical `xs = append(xs, v)` shape, how capacity grows.
- `range` formally — index+value, index-only, value-only; the value-is-a-copy rule.
- Maps — `map[K]V`, zero value (nil maps panic on write), the `v, ok := m[k]` exists check, `delete`, ranging over a map (and randomised iteration order).

---

## Concept 1: Slices — the basics

### Motivation

A slice is Go's go-to "sequence of values" type. Strings, function arguments, JSON arrays, file lines — almost everything that's "a list of things" in Go is a slice. Three declaration shapes cover all the cases. Slices look simple but they share an underlying array under the hood, which is the source of the most common slice surprises.

---

### The basics

```go
package main

import "fmt"

func main() {
	// 1. Zero-value declaration — a nil slice. len/cap both 0; safe to range.
	var xs []int
	fmt.Println(xs, len(xs), cap(xs)) // [] 0 0

	// 2. Slice literal — a non-nil slice with the given values.
	ys := []int{1, 2, 3}
	fmt.Println(ys, len(ys), cap(ys)) // [1 2 3] 3 3

	// 3. make — pre-allocates with the given length (and optionally capacity).
	zs := make([]int, 3)             // len 3, cap 3, zero-valued
	fmt.Println(zs)                  // [0 0 0]
	zsBig := make([]int, 0, 10)      // len 0, cap 10
	fmt.Println(zsBig, len(zsBig), cap(zsBig)) // [] 0 10

	// Indexing and slicing.
	fmt.Println(ys[0])   // 1
	fmt.Println(ys[1:3]) // [2 3]   (half-open: includes 1, excludes 3)
	fmt.Println(ys[:2])  // [1 2]   (from start)
	fmt.Println(ys[1:])  // [2 3]   (to end)
}
```

Three things to internalise:

- **Nil slice vs empty slice.** `var xs []int` is `nil` but its length is `0`. `xs := []int{}` is non-nil and length 0. For most operations they behave the same: you can `len()`, `range`, and `append` either one. The difference shows up if you compare against `nil`, or marshal to JSON (nil → `null`, empty → `[]`).
- **Half-open slicing.** `xs[i:j]` includes index `i`, excludes `j`. Length of the result is `j - i`. Same convention as Python.
- **`len` vs `cap`.** Length is how many elements are there; capacity is how many the backing array can hold before a re-allocation. `cap` matters when you `append`.

---

### A worked example

A small "build a list as we go" pattern — collect even numbers from 1 to 10:

```go
package main

import "fmt"

func main() {
	var evens []int
	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			evens = append(evens, i)
		}
	}
	fmt.Println(evens) // [2 4 6 8 10]
}
```

Notice we never initialised `evens` to `[]int{}` — `var evens []int` declared a nil slice, and `append(nil, v)` returns a non-nil slice with `v` in it. Both work; pick whichever reads better in context.

---

### Common mistake

Forgetting that `len()` and `cap()` are different — building a slice with `make([]T, n)` when you wanted `make([]T, 0, n)`:

```go
package main

import "fmt"

func main() {
	xs := make([]int, 5)         // len 5, cap 5 — already has five zeros!
	xs = append(xs, 1, 2, 3)
	fmt.Println(xs)               // [0 0 0 0 0 1 2 3]   (probably not what you wanted)
}
```

Fix: `make([]int, 0, 5)` if you want an empty slice with room to grow. The two-argument form pre-fills the slice with `n` zero-valued elements.

---

### Recap

- Three declaration shapes: `var xs []T` (nil), `xs := []T{...}` (literal), `make([]T, n, cap)` (pre-allocated).
- Half-open slicing: `xs[i:j]` includes `i`, excludes `j`.
- Nil slice and empty slice are usually interchangeable.
- `len(xs)` is what's there; `cap(xs)` is how much room there is.

---

## Concept 2: `append` and growth

### Motivation

Slices grow with `append`. The function takes a slice and one or more values and returns a (possibly new) slice. The "possibly new" part is where surprises hide.

---

### The basics

`append` returns the slice — you MUST assign it back:

```go
package main

import "fmt"

func main() {
	xs := []int{1, 2, 3}
	xs = append(xs, 4)        // canonical shape — capture the return
	xs = append(xs, 5, 6, 7)  // multiple values at once
	fmt.Println(xs, cap(xs))  // [1 2 3 4 5 6 7] capacity at least 7

	// Spreading another slice — note the trailing ...
	more := []int{8, 9, 10}
	xs = append(xs, more...)
	fmt.Println(xs)           // [1 2 3 4 5 6 7 8 9 10]
}
```

How growth works under the hood: when `append` runs out of capacity, Go allocates a new backing array (usually doubling the capacity), copies the existing elements over, adds the new one, and returns the new slice. If there's still capacity available, it modifies the existing array in place.

You almost never have to think about this — `xs = append(xs, v)` Just Works. But understanding it explains why "always capture the return" matters: the slice header you got back may point at a *different* underlying array than the one you passed in.

---

### A worked example

Lesson 04's pattern — accumulating a result with `append`:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12, 75, 9.99}
	var snacks []float64
	for _, a := range amounts {
		if a < 10 {
			snacks = append(snacks, a)
		}
	}
	fmt.Println(snacks) // [4.5 9.99]
}
```

Familiar shape — declare a nil slice, append into it, return.

---

### Common mistake

Calling `append` without using its return value:

```go
package main

import "fmt"

func main() {
	xs := []int{1, 2, 3}
	append(xs, 4)             // wrong: return value discarded
	fmt.Println(xs)           // [1 2 3]   (4 went into the void)
}
```

Go's compiler flags this with `append must be used`. Capture it: `xs = append(xs, 4)`.

A subtler variant: appending inside a function with the caller's slice. Because `append` may allocate a new backing array, modifying the slice header in the called function doesn't affect the caller. We'll see this pattern when we introduce pointers in lesson 09 — for now, just always return the new slice from your function and let the caller reassign.

---

### Recap

- `xs = append(xs, v)` — always capture the return value.
- Multiple values at once: `append(xs, v1, v2, v3)`.
- Spread another slice: `append(xs, more...)` (note the trailing `...`).
- `append` may allocate a new backing array — that's why you assign the return.

---

## Concept 3: `range` formally

### Motivation

You met `for ... range` in lesson 03 as the "walk this slice" shape. Now we formalise: range works on slices, arrays, strings, maps, and channels — and the value it produces is always a **copy**.

---

### The basics

Range on a slice — two values per iteration, both optional:

```go
package main

import "fmt"

func main() {
	xs := []string{"a", "b", "c"}

	// Index + value (the common case).
	for i, v := range xs {
		fmt.Printf("%d:%s ", i, v)
	}
	fmt.Println()              // 0:a 1:b 2:c

	// Index only — drop the value with _.
	for i := range xs {
		fmt.Println(i)
	}

	// Value only — drop the index with _.
	for _, v := range xs {
		fmt.Println(v)
	}
}
```

Important: `v` is a **copy** of the element. Mutating `v` does not mutate `xs[i]`:

```go
xs := []int{1, 2, 3}
for _, v := range xs {
	v *= 2          // modifies the local copy, NOT xs
}
fmt.Println(xs)     // [1 2 3]   (unchanged)
```

To modify the slice in place, use the index:

```go
for i := range xs {
	xs[i] *= 2
}
fmt.Println(xs)     // [2 4 6]
```

Range on a string yields `(byte-index, rune)` — Unicode code points, not bytes:

```go
for i, r := range "héllo" {
	fmt.Printf("%d:%c ", i, r)  // 0:h 1:é 3:l 4:l 5:o  (é is 2 bytes in UTF-8)
}
```

---

### A worked example

Walking parallel slices with `range`:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12, 75, 9.99}
	categories := []string{"coffee", "lunch", "rent", "coffee"}

	for i, cat := range categories {
		fmt.Printf("%s: €%.2f\n", cat, amounts[i])
	}
}
```

You can range over either slice and index into the other — they're parallel, so any of `range amounts`, `range categories`, or `range len(amounts)` works.

---

### Common mistake

Trying to mutate the slice via the loop variable:

```go
package main

import "fmt"

func main() {
	xs := []int{1, 2, 3}
	for _, v := range xs {
		v += 10        // works on the copy, not on xs
	}
	fmt.Println(xs)   // [1 2 3] — unchanged
}
```

Fix: use the index.

```go
for i := range xs {
	xs[i] += 10
}
fmt.Println(xs)       // [11 12 13]
```

Historical footnote: before Go 1.22, the *loop variable itself* was shared across iterations. Capturing it in a closure or a goroutine produced the famous "all goroutines see the last value" bug. Go 1.22 (and our 1.23) fixed this by making the loop variable fresh per iteration. If you read old Go articles about loops + closures, that's the bug they're warning you about — it's gone.

---

### Recap

- Range produces `(index, value)`; drop either with `_`.
- The value is a copy — to mutate the slice, range over the index and write through `xs[i]`.
- Range on a string yields `(byte-index, rune)`.
- Go 1.22+ gives you a fresh loop variable per iteration; the old "closure captures the last value" bug is fixed.

---

## Concept 4: Maps

### Motivation

A map is Go's hash-map / dictionary type: key/value pairs with O(1) average lookup. The type is written `map[K]V` where `K` is the key type and `V` is the value type. Maps look simple but have three traps that bite everyone the first time: nil maps, zero values on missing keys, and randomised iteration order.

---

### The basics

```go
package main

import "fmt"

func main() {
	// Three ways to make a map:
	var m1 map[string]int          // nil map — reads OK, writes panic
	m2 := map[string]int{}         // empty, non-nil
	m3 := make(map[string]int)     // empty, non-nil
	m4 := map[string]int{"a": 1, "b": 2} // literal with initial values

	_ = m1
	fmt.Println(m2, m3, m4)        // map[] map[] map[a:1 b:2]

	// Read returns the zero value if the key isn't there.
	count := m4["c"]               // 0 (the zero value of int)
	fmt.Println(count)

	// The exists check — comma-ok form.
	if v, ok := m4["c"]; ok {
		fmt.Println("c =", v)
	} else {
		fmt.Println("c missing")
	}

	// Write — must be a non-nil map.
	m4["d"] = 4
	fmt.Println(m4)                // map[a:1 b:2 d:4]

	// Delete.
	delete(m4, "a")
	fmt.Println(m4)                // map[b:2 d:4]
}
```

Four idioms:

- **Make it non-nil before you write.** `var m map[K]V` is nil. Writing to a nil map panics. `make(map[K]V)` or `map[K]V{}` gives you an empty, writable map.
- **Reads always succeed.** `m[k]` returns the value's zero value if `k` isn't present. No error, no panic — just zero. That's why `totals[cat] += amount` works for a key that hasn't been seen before.
- **The `v, ok := m[k]` form** is how you tell "key was missing" apart from "key was present with zero value." Use it whenever the distinction matters.
- **`delete(m, k)`** removes a key. Idempotent — deleting a missing key is a no-op.

---

### A worked example

Counting categories with a map — uses the "zero value on missing key" idiom:

```go
package main

import "fmt"

func main() {
	categories := []string{"coffee", "lunch", "rent", "coffee", "coffee"}

	counts := map[string]int{}
	for _, cat := range categories {
		counts[cat]++          // works whether cat is present or not
	}
	fmt.Println(counts)         // map[coffee:3 lunch:1 rent:1]
}
```

Five lines. The whole "is the key there yet?" dance disappears because `counts[cat]` returns `0` for a missing key, and incrementing zero gives one. The same trick powers your main exercise's `TotalsByCategory`.

Ranging over a map gives you key/value pairs, but **in a randomised order** — Go intentionally randomises it so you don't accidentally depend on insertion order:

```go
for k, v := range counts {
	fmt.Printf("%s=%d ", k, v)
}
// Possible output: coffee=3 rent=1 lunch=1
// Or:              lunch=1 coffee=3 rent=1
// You can't predict the order.
```

If you need deterministic order, sort the keys first (lesson 14 introduces `sort.Strings`).

---

### Common mistake

Writing to a nil map:

```go
package main

func main() {
	var m map[string]int   // nil
	m["a"] = 1             // panic: assignment to entry in nil map
}
```

Fix: `m := map[string]int{}` or `m := make(map[string]int)`.

A subtler variant — function return values:

```go
func makeMap() map[string]int {
	var m map[string]int
	// ... forget to assign m = make(...) ...
	return m
}

m := makeMap()
m["a"] = 1   // panic
```

If your function returns a map, make sure you actually instantiate it before returning. The compiler can't catch this — `nil` is a valid value for a map type.

---

### Recap

- `map[K]V` — Go's hash map. Make it with `map[K]V{}` or `make(map[K]V)`. `var m map[K]V` is nil and panics on write.
- Reads return the zero value on missing keys — that's why `m[k]++` works.
- `v, ok := m[k]` is the "was the key there?" check.
- `delete(m, k)` removes a key; ranging gives you `k, v` in a randomised order.

---

## Practice

### Warm-up

In `exercises/warmup.go` — three functions to implement, with skeleton tests to fill in:

- `WarmupSum(xs []int) int` — sum the slice. Empty/nil returns 0.
- `WarmupMax(xs []int) (int, error)` — largest value, or `errEmptyWarmupMax` for empty. Same shape as lesson 04's `WarmupMinMax`.
- `WarmupUnique(xs []string) []string` — dedupe while preserving order. Use a `map[string]bool` as a seen-set. Always returns a non-nil slice.

Tests in `exercises/warmup_test.go` ship as skeletons — fill in cases and assertion bodies as you did for lesson 04's main exercise. The README walks through the shape.

```bash
cd lessons/05-slices-maps/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error)` — parallel slices in, totals-per-category map out. Two error cases:
  - Empty inputs (either slice has length 0) → `errEmptyInputs`
  - Length mismatch (`len(amounts) != len(categories)`) → `errLengthMismatch`

Tests in `exercises/main_test.go` ship as a skeleton — fill in cases (cover the happy path AND both error cases).

```bash
cd lessons/05-slices-maps/exercises
go test -v
```

Note:
For live: the "writing to a nil map" panic is the highest-payoff live demo of the lesson. Type it in front of the class, run it, watch the panic, fix with `make(...)`. Also pay off the deferred slices forward-reference from lesson 03 — show the parallel-slice walk in `TotalsByCategory` and tie it back to lesson 03's `Tally` (which also walked a slice but only needed counts, not key/value totals).

---

## What we learned

- Slices: three declaration shapes; half-open slicing; nil vs empty; `len`/`cap`.
- `append` returns the slice — always assign it back. May allocate a new backing array.
- `range` produces an index and a *copy* of the value. Modify-in-place via the index. Loop variable is fresh per iteration in Go 1.22+.
- Maps: `map[K]V`; nil maps panic on write; reads return the zero value on missing keys; `v, ok := m[k]` is the exists check; `delete(m, k)` removes; iteration order is randomised.
- The `counts[k]++` / `totals[cat] += amount` idiom rides on "zero value for missing key" — clean and idiomatic.

---

## Up next

Lesson 06 — Composite types II: structs and methods (`type T struct { ... }`, value vs pointer receivers, struct embedding light).
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=05-slices-maps &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/05-slices-maps/slides/slides.md
curl -sS http://localhost:8000/lessons/05-slices-maps/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/05-slices-maps/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/05-slices-maps/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 3: Verify markdown structure**

```bash
grep -c '<div class="title-slide-grid">' lessons/05-slices-maps/slides/slides.md
grep -c "<h1>Slices and maps</h1>" lessons/05-slices-maps/slides/slides.md
grep -c "^## What we learned" lessons/05-slices-maps/slides/slides.md
grep -c "^## Up next" lessons/05-slices-maps/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/05-slices-maps/slides/slides.md
git commit -m "feat(lesson-05): slides — Slices and maps (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/05-slices-maps/README.md` with the full lesson 05 self-study prose. Mirrors the slide narrative, includes Common-mistake examples per concept, walks through what students must do for the skeleton tests.

**Files:**
- Replace: `lessons/05-slices-maps/README.md`

- [ ] **Step 1: Write `lessons/05-slices-maps/README.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device; in the file use only three-backtick fences. File starts with `# Lesson 05: Slices and maps`.

````markdown
# Lesson 05: Slices and maps

## Learning goals

- Declare and index slices in Go's three shapes (`var xs []T`, `xs := []T{...}`, `make([]T, n)`); slice with `xs[i:j]`; understand `len` vs `cap`.
- Grow a slice with `append` and the canonical `xs = append(xs, v)` shape.
- Walk a collection with `range` formally — index+value, index-only, value-only; understand the "value is a copy" rule.
- Use `map[K]V`: declare safely (don't write to a nil map!), use `v, ok := m[k]` for the exists check, `delete` keys, range over key/value pairs.

## Prerequisites

- Lessons 01-04. In particular, lesson 03 introduced `for ... range` informally, and lesson 04 introduced multi-return values + the `(value, error)` shape that the warm-up uses again here.

## Concepts

### Slices — the basics

A slice is Go's "sequence of values" — string lines, function arguments, JSON arrays, anything list-shaped. Three ways to declare:

```go
var xs []int                 // 1. nil slice — len/cap 0; safe to range
ys := []int{1, 2, 3}         // 2. slice literal
zs := make([]int, 3)         // 3. make — pre-allocates len 3, cap 3 (all zeros)
zsBig := make([]int, 0, 10)  //    make with length 0 and capacity 10
```

Index with `xs[i]`. Slice with `xs[i:j]` (half-open — includes `i`, excludes `j`):

```go
ys := []int{1, 2, 3}
ys[0]    // 1
ys[1:3]  // [2 3]
ys[:2]   // [1 2]
ys[1:]   // [2 3]
```

Three things to internalise:

- **Nil slice vs empty slice.** `var xs []int` is nil but length 0. `xs := []int{}` is non-nil and length 0. Both are safe to `range` and `append`. They differ for JSON marshalling (`null` vs `[]`) and `== nil` checks.
- **Half-open slicing.** Length of `xs[i:j]` is always `j - i`. Same as Python.
- **`len` vs `cap`.** `len` is what's there; `cap` is how much the backing array can hold before a re-allocation. `cap` only matters when you `append`.

**Common mistake.** Confusing `make([]T, n)` (n zero-valued elements) with `make([]T, 0, n)` (empty with capacity n):

```go
xs := make([]int, 5)         // [0 0 0 0 0]
xs = append(xs, 1, 2, 3)
// xs is now [0 0 0 0 0 1 2 3] — probably not what you wanted.
```

Fix: `make([]int, 0, 5)` for an empty slice with room to grow.

### `append` and growth

`append` returns the (possibly new) slice — you MUST assign the return:

```go
xs := []int{1, 2, 3}
xs = append(xs, 4)
xs = append(xs, 5, 6, 7)     // multiple values at once

more := []int{8, 9, 10}
xs = append(xs, more...)     // spread another slice
```

How growth works: when the backing array runs out of room, Go allocates a new one (usually with double the capacity), copies, adds the new value, returns the new slice. If there's room, it appends in place. You almost never have to think about this — but you do have to capture the return value, because the slice header you get back may point at a different array than the one you passed in.

**Common mistake.** Discarding `append`'s return value:

```go
xs := []int{1, 2, 3}
append(xs, 4)                // compiler error: "append must be used"
```

Fix: `xs = append(xs, 4)`.

### `range` formally

Range produces an index and a *copy* of the value:

```go
xs := []string{"a", "b", "c"}

for i, v := range xs {       // both
	fmt.Printf("%d:%s ", i, v)
}

for i := range xs {           // index only
	fmt.Println(i)
}

for _, v := range xs {        // value only
	fmt.Println(v)
}
```

Because `v` is a copy, you can't mutate the slice through it:

```go
xs := []int{1, 2, 3}
for _, v := range xs {
	v *= 2                    // modifies local copy, not xs
}
fmt.Println(xs)               // [1 2 3]
```

To mutate the slice, write through the index:

```go
for i := range xs {
	xs[i] *= 2
}
fmt.Println(xs)               // [2 4 6]
```

Range on a string yields `(byte-index, rune)` — the rune is a Unicode code point, not a byte:

```go
for i, r := range "héllo" {
	fmt.Printf("%d:%c ", i, r)  // 0:h 1:é 3:l 4:l 5:o
}
```

`é` is two bytes in UTF-8, hence the byte-index jump from `1` to `3`.

**Common mistake.** Trying to mutate the slice via the loop variable, as shown above. Fix: range over the index.

Historical footnote: before Go 1.22, the loop variable itself was shared across iterations. Closures and goroutines that captured the loop variable saw the *last* value, not the per-iteration value. Go 1.22 (and our 1.23) fixed this — the loop variable is now fresh per iteration. If you read old articles warning about "loop variable capture", that's the bug; it's gone.

### Maps

A map is Go's hash-map / dictionary. Three ways to make one:

```go
var m1 map[string]int                // nil — reads OK, writes panic
m2 := map[string]int{}               // empty, non-nil
m3 := make(map[string]int)           // empty, non-nil
m4 := map[string]int{"a": 1, "b": 2} // literal with initial values
```

Four idioms:

- **Make it non-nil before you write.** Writing to a nil map panics at runtime.
- **Reads always succeed.** `m[k]` returns the value's zero value if `k` isn't present. No error, no panic.
- **`v, ok := m[k]`** distinguishes "key missing" from "key present with zero value." Use it when the distinction matters.
- **`delete(m, k)`** removes a key. Idempotent — deleting a missing key is a no-op.

```go
m := map[string]int{"a": 1, "b": 2}

m["c"]                       // 0 — zero value for missing key

if v, ok := m["c"]; ok {     // exists check
	fmt.Println("c =", v)
} else {
	fmt.Println("c missing")
}

m["d"] = 4                   // write — needs a non-nil map
delete(m, "a")               // remove
```

Ranging over a map gives `(key, value)` — **but in randomised order**:

```go
counts := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range counts {
	fmt.Printf("%s=%d ", k, v)
}
// Could print "a=1 b=2 c=3" or "c=3 a=1 b=2" or any other permutation.
```

Go randomises iteration order on purpose, so code can't accidentally rely on a stable order. If you need deterministic output, sort the keys first (lesson 14 covers `sort.Strings`).

The "zero value on missing key" rule powers a beautiful idiom — counting or summing without an exists check:

```go
counts := map[string]int{}
for _, cat := range categories {
	counts[cat]++            // works whether cat is present or not
}

totals := map[string]float64{}
for i, cat := range categories {
	totals[cat] += amounts[i]
}
```

That's exactly what your main exercise's `TotalsByCategory` does.

**Common mistake.** Writing to a nil map:

```go
var m map[string]int
m["a"] = 1                   // panic: assignment to entry in nil map
```

Fix: `m := map[string]int{}` or `m := make(map[string]int)`.

A subtler variant — function return values:

```go
func makeMap() map[string]int {
	var m map[string]int
	// ... forget to assign m = make(...) ...
	return m
}

m := makeMap()
m["a"] = 1                   // panic
```

The compiler can't catch this — `nil` is a valid value for a map type. Always initialise before returning.

## Exercise: warm-up

Three functions in `exercises/warmup.go`, three skeleton test functions in `exercises/warmup_test.go`. Same shape as lesson 04's main exercise:

- **Implement** `WarmupSum`, `WarmupMax`, `WarmupUnique` per their doc comments.
- **Fill in the skeleton tests** — for each test function, add at least 4 cases to the `cases` slice and write the `t.Run` body to call the function and assert.

The functions:

- `WarmupSum(xs []int) int` — sum the slice. Empty/nil → 0 (the additive identity).
- `WarmupMax(xs []int) (int, error)` — largest value, or `errEmptyWarmupMax` for empty input. Mirror lesson 04's `WarmupMinMax` shape: branch on the `wantErr` flag, use `errors.Is(err, errEmptyWarmupMax)`.
- `WarmupUnique(xs []string) []string` — dedupe while preserving order. Use a `map[string]bool` as a seen-set. Always returns a non-nil slice (initialise with `[]string{}`).

For the test cases: think about what each function *should* do, not what your stub returns today. Cover positive, single-element, edge, and nil cases.

## Exercise: main

In `exercises/main.go`:

- `TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error)` — parallel slices in, totals-per-category map out.

Two error conditions, **checked in this order**:

1. Either slice is empty → `errEmptyInputs`.
2. Slices have different lengths → `errLengthMismatch`.

If both inputs pass validation, walk one slice with `range` and accumulate into a `map[string]float64`. The "zero value on missing key" idiom makes this a one-liner inside the loop:

```go
totals := map[string]float64{}
for i, cat := range categories {
	totals[cat] += amounts[i]
}
return totals, nil
```

Fill in the skeleton tests in `exercises/main_test.go`. Cases to include:

1. A happy-path case with all distinct categories.
2. A case where one category appears multiple times — totals must accumulate, not overwrite.
3. A single-element case.
4. The empty-inputs error case — assert via `errors.Is(err, errEmptyInputs)`.
5. The length-mismatch error case — assert via `errors.Is(err, errLengthMismatch)`.

The test struct is hoisted (`type totalsCase struct {...}` above the test function) so the `wantErr error` field is naturally typed. You'll branch on `tc.wantErr != nil` to pick the assertion path.

Look at `solutions/main_test.go` if you want to compare to the reference — but try the cases yourself first.

## How to run

```bash
cd lessons/05-slices-maps/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main
```

Keep the lesson 04 habits going:

```bash
gofmt -w .       # reformat to canonical Go style
go vet ./...     # catch the subtle bugs go test won't
```

CI enforces both.

Once both exercises pass, compare your code (especially your tests) against `solutions/`.

## Going further

### Read

- [Go blog — Go slices: usage and internals](https://go.dev/blog/slices-intro) — explains the slice header (pointer + len + cap) and why `append` may allocate.
- [Effective Go — Maps](https://go.dev/doc/effective_go#maps) — short canonical reference for map idioms.
- [Go blog — Strings, bytes, runes and characters](https://go.dev/blog/strings) — why `range "héllo"` jumps byte indices.

### Try

- **Reverse a slice in place.** Write `Reverse(xs []int)` that reverses in place using `for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1`. No reference solution.
- **Find the second-largest.** Extend `WarmupMax` to a `SecondMax(xs []int) (int, error)` that returns the second-largest (or an error if there's no clear second-largest, e.g. fewer than two elements or all elements are equal).
- **Count word frequencies.** Read a string of space-separated words and return a `map[string]int` of word → count. (No file I/O yet — pass the string as an argument.)
````

- [ ] **Step 2: Verify the README structure**

```bash
for h in "^# Lesson 05: Slices and maps$" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  c=$(grep -c "$h" lessons/05-slices-maps/README.md)
  echo "  $h => $c"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/05-slices-maps/README.md   # expect 4
grep -c "gofmt" lessons/05-slices-maps/README.md                       # expect >= 1
grep -c "go vet" lessons/05-slices-maps/README.md                      # expect >= 1
```

Expected: all section headings 1; Common mistake count 4; gofmt and go vet counts ≥ 1.

- [ ] **Step 3: Commit**

```bash
git add lessons/05-slices-maps/README.md
git commit -m "docs(lesson-05): README — Slices and maps self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises). Lesson 05 solutions pass; lessons 01-04 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — warm-up + main both pass vacuously**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 05's `TestWarmupSum`, `TestWarmupMax`, `TestWarmupUnique`, `TestTotalsByCategory` all PASS with 0 sub-tests (skeleton case slices are empty). Lessons 01-04 fail as before (their exercises are still unimplemented). Make exits 0 because `-` ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=05-slices-maps 2>&1 | tail -30
```

Expected: exercise tests pass vacuously (ignored), solution tests pass fully (TestWarmupSum 6 sub-tests, TestWarmupMax 7, TestWarmupUnique 6, TestTotalsByCategory 9).

- [ ] **Step 4: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 5: `go vet` clean**

```bash
go vet ./...
```

Expected: no output.

- [ ] **Step 6: Slides server smoke test**

```bash
make slides-dev LESSON=05-slices-maps &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/05-slices-maps/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/05-slices-maps/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 7: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/05-slices-maps/slides/slides.md && echo OK
grep -q "Slices and maps" dist/index.html && echo "index lists lesson 05"
rm -rf dist
```

Expected: three success lines. The landing page should now list lessons 01-05.

- [ ] **Step 8: Final repository sanity check**

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch (plan + scaffold + warm-up + main + slides + README).

This task makes no commit — verification only.

---

## Done definition

After Task 6:

- `lessons/05-slices-maps/` contains 12 files with full lesson content.
- `make test` passes; `make test-exercises` shows lesson 05's exercise tests passing vacuously (empty case slices, students will fill them in).
- `make test-lesson LESSON=05-slices-maps` runs both sides correctly.
- `make slides-dev LESSON=05-slices-maps` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan I — Lesson 06 (Composite types II: structs and methods).** Same per-lesson pattern. Lesson 06 covers `type T struct { ... }`, struct literals, value vs pointer receivers (favouring value receivers in Phase 1), light struct embedding, and exported vs unexported fields. The main exercise refactors lesson 05's parallel-slice logic into a `[]Expense` based design — `type Expense struct { Date string; Amount float64; Category string }` with `Format()` and `IsHigh()` methods, plus a `TotalsByCategory(es []Expense) map[string]float64` revisiting lesson 05's main with a richer input type.
