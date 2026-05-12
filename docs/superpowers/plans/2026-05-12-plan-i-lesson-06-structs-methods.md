# Plan I — Lesson 06 (Composite types II: structs and methods) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the sixth lesson of Phase 1 — students learn struct definition and literal forms, methods with value receivers (the Phase 1 default), a light touch of pointer receivers (forward reference to lesson 09), and the lightest treatment of struct embedding and exported-vs-unexported fields. The main exercise replaces lesson 05's parallel-slice `TotalsByCategory` with a `[]Expense` version, plus two `Expense` methods (`Format` and `IsHigh`).

**Architecture:** Same per-lesson pattern as Plans D-H. Six tasks: scaffold, warm-up, main, slides, README, end-to-end verify. One warm-up type (`WarmupPoint`) with one method, plus one main type (`Expense`) with two methods + the standalone `TotalsByCategory` function. Skeleton tests across both warm-up and main (continuing the lesson 05 pattern). Heavy-explanatory slide deck with **four** concept blocks (struct definition → methods (value receivers) → pointer receivers (light) → embedding + exported/unexported).

**Tech Stack:** Go 1.23 stdlib only (`errors`, `fmt`, `math`, `testing`). Reveal.js 5.1.0 for the deck.

---

## Scope

Plan I produces lesson 06 only. After Plan I lands:

- `lessons/06-structs/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows lesson 06's tests passing vacuously (skeleton case slices empty).
- `make slides-dev LESSON=06-structs` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):** Lessons 07-08 (Plans J-K).

### Design decisions made during planning

1. **`TotalsByCategory(es []Expense) map[string]float64`** — plain map return (no error), per user choice. Spec literal. Empty/nil input → empty map. Lesson 05 returned `(map[string]float64, error)` because it had to validate two parallel slices for both empty-input AND length-mismatch; with a single `[]Expense` neither failure mode applies, and an empty map is a useful answer. Trusts the caller. Reduces noise; keeps lesson 06's focus on structs/methods, not error idioms.

2. **`math.Sqrt` for `DistanceFromOrigin`** — minor spec deviation. The Phase 1 design lists "New imports introduced: none" for lesson 06, but the spec also literally names `DistanceFromOrigin() float64` as the warm-up method, which requires `math.Sqrt` (or `math.Hypot`). `math` is a small stdlib import with no external dependencies — Phase 1's "stdlib only" rule is the hard constraint, and that still holds. The README's "Imports" section is one line. Easy to flip during plan review if you'd rather rename to `DistanceSquared()` and avoid the import.

3. **Warm-up type name: `WarmupPoint`** — the `Warmup*` prefix convention (established in Plan G) extends to types here. Alternatives like `Point` + `WarmupDistanceFromOrigin()` would put the prefix on the method and leave `Point` exposed in the warm-up package; that crosses the prefix-the-warm-up-identifiers convention. `WarmupPoint` keeps the method name clean (`DistanceFromOrigin`, not the awkward `WarmupDistanceFromOrigin`).

4. **One warm-up method, not two.** Spec literally says "Define `Point{X, Y float64}` and `DistanceFromOrigin() float64`; tests." A single method matches the spec; lessons 04-05 had multi-function warmups because the spec asked for multiple functions. Lesson 06's warm-up is light by design — most of the lesson's surface area is in the main exercise (Expense + Format + IsHigh + TotalsByCategory).

5. **Skeleton tests in BOTH warm-up and main.** Continuing the lesson 05 pattern. All four test functions (`TestWarmupPointDistanceFromOrigin`, `TestExpenseFormat`, `TestExpenseIsHigh`, `TestTotalsByCategory`) in `exercises/` ship as skeletons. Solutions ships full reference tests.

6. **Four concepts in the slide deck.**
   1. Struct definition and literals — `type T struct { ... }`, named-field / positional / partial-named literal forms, zero-value structs.
   2. Methods (value receivers) — `(t T) Method()`, calling, method set.
   3. Pointer receivers — light treatment. Show the syntax `(t *T) Method()` and the "when to mutate" rule, but defer the deep treatment of pointers to lesson 09. The slides explicitly say "we'll see pointers proper next module."
   4. Embedding + exported/unexported — light. Show the `type Inner struct { … }; type Outer struct { Inner; Extra string }` shape and the capitalisation rule for exported fields. Lesson 07 (packages) gives the full treatment of exporting.

7. **Phase 1 prefers value receivers.** Methods on `Expense` and `WarmupPoint` use value receivers throughout. The slides' concept 3 covers pointer receivers for awareness (so students reading other Go code aren't surprised), but the lesson's own code never uses them. The "Going further" section invites students to try a pointer-receiver variant.

8. **`Expense.Format()` matches lesson 04's `FormatExpense` output byte-for-byte.** Same format string (`"%s  €%-7.2f %s"`), same expected outputs. Lesson 06's value here is the method-receiver shape — same logic, different surface.

---

## Plans F/G/H lessons-learned applied here

1. **Warmup* prefix from the start.** `WarmupPoint` type + `WarmupPoint.DistanceFromOrigin()` method. No post-review fix needed.

2. **Lint exclusion already in place.** `.golangci.yml`'s `lessons/*/exercises/` exclusion covers `staticcheck` + `unused`. Lesson 06's exercises don't introduce new sentinel error variables (no error returns this lesson), so the false-positive "unused" warnings from the IDE LSP that bit lessons 04-05 don't appear here.

3. **`→` arrow consistency** — same as previous plans.

4. **Common-mistake content in README** — every concept's README section includes the same common-mistake example as the slides. 4 total.

5. **Slides and README written inline by controller** — Plans G/H established this. Plan I continues. Tasks 1-3 (and 6) can be subagent-driven; Tasks 4-5 (slides + README) the controller writes directly to avoid timeout risk.

6. **`gofmt -w .` and `go vet ./...` mentions** — README continues the lesson 04 habit.

7. **Format-output verification.** Plan G learned that column-aligned `fmt.Sprintf` output needs byte-level verification before committing test cases. Plan I's `Expense.Format()` uses the same format string as Plan G's `FormatExpense` (`"%s  €%-7.2f %s"`) and the same reference outputs. Verify once in Task 3 to confirm.

8. **Empty case → vacuous pass.** Plan H's "test-lesson exercises now pass vacuously instead of failing by panic" tradeoff is the new normal. README's exercise sections explicitly call this out so students aren't fooled by the green checkmark.

---

## File Structure

After Plan I:

```
lessons/06-structs/                         (new — Plan I's deliverable)
├── README.md                               (Task 5)
├── slides/
│   ├── index.html                          (Task 1; unchanged)
│   ├── slides.md                           (Task 4 — controller writes inline)
│   └── assets/.gitkeep                     (Task 1; unchanged)
├── exercises/
│   ├── warmup.go                           (Task 2 — WarmupPoint + DistanceFromOrigin stub)
│   ├── warmup_test.go                      (Task 2 — SKELETON test)
│   ├── main.go                             (Task 3 — Expense + Format + IsHigh + TotalsByCategory stubs)
│   └── main_test.go                        (Task 3 — SKELETON tests for all three)
└── solutions/
    ├── warmup.go                           (Task 2 — implementation)
    ├── warmup_test.go                      (Task 2 — full reference test)
    ├── main.go                             (Task 3 — implementations)
    └── main_test.go                        (Task 3 — full reference tests)
```

### Decomposition rationale

Same pattern as plans D-H. Each task has one focused output and a clear verification step. The exercises/solutions skeleton-vs-reference divergence applies to all four test files (warmup + main).

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-i-lesson-06-structs-methods` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold lesson 06

**Files:**
- Create: `lessons/06-structs/` (12 files via the scaffolder)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-i-lesson-06-structs-methods`, clean working tree, exactly one commit ahead of `main` (the Plan I doc).

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=06-structs
```

Expected: `created lesson 06-structs under lessons/`. 12 placeholder files produced.

- [ ] **Step 3: Verify the 12-file tree**

```bash
find lessons/06-structs -type f | sort
```

Expected (12 lines):

```
lessons/06-structs/README.md
lessons/06-structs/exercises/main.go
lessons/06-structs/exercises/main_test.go
lessons/06-structs/exercises/warmup.go
lessons/06-structs/exercises/warmup_test.go
lessons/06-structs/slides/assets/.gitkeep
lessons/06-structs/slides/index.html
lessons/06-structs/slides/slides.md
lessons/06-structs/solutions/main.go
lessons/06-structs/solutions/main_test.go
lessons/06-structs/solutions/warmup.go
lessons/06-structs/solutions/warmup_test.go
```

- [ ] **Step 4: Commit**

```bash
git add lessons/06-structs/
git commit -m "feat(lessons): scaffold lesson 06-structs skeleton"
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: all packages pass. Lesson 06 solutions pass with the scaffolded `Greet`/`WarmupGreet`.

- [ ] **Step 6: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch.

---

## Task 2: Author the warm-up exercise

Define `WarmupPoint{X, Y float64}` and `(p WarmupPoint) DistanceFromOrigin() float64` (returns `sqrt(x*x + y*y)`). Tests ship as a skeleton.

**Files:**
- Replace: `lessons/06-structs/exercises/warmup.go`
- Replace: `lessons/06-structs/exercises/warmup_test.go`
- Replace: `lessons/06-structs/solutions/warmup.go`
- Replace: `lessons/06-structs/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/06-structs/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 06: Structs and methods.
//
// This file holds the WARM-UP exercise: a small struct with one method,
// to build muscle memory for "type T struct { ... }" and "func (t T) M()".
// The matching test ships as a skeleton — fill in the cases and assertion body.
package exercises

import "math"

// WarmupPoint is a 2D point with float64 coordinates.
//
// We use the `Warmup` prefix to keep this type out of the way of the
// main exercise's Expense struct. Real Go code would just call this Point.
type WarmupPoint struct {
	X float64
	Y float64
}

// DistanceFromOrigin returns the Euclidean distance from the origin (0, 0)
// to p — i.e. sqrt(X*X + Y*Y).
//
// Examples:
//
//	WarmupPoint{X: 3, Y: 4}.DistanceFromOrigin()  → 5
//	WarmupPoint{X: 0, Y: 0}.DistanceFromOrigin()  → 0
//	WarmupPoint{X: 1, Y: 1}.DistanceFromOrigin()  → 1.4142135…
//
// Hint: use math.Sqrt. p.X * p.X + p.Y * p.Y is the squared distance.
func (p WarmupPoint) DistanceFromOrigin() float64 {
	_ = math.Sqrt // keep the import compiling until you use it
	panic("TODO: return math.Sqrt(p.X*p.X + p.Y*p.Y)")
}
```

- [ ] **Step 2: Replace `lessons/06-structs/exercises/warmup_test.go`** with (SKELETON):

```go
package exercises

import "testing"

// TestWarmupPointDistanceFromOrigin is a SKELETON. Fill in the cases and
// the t.Run body. Distances on floats need a tolerance — see the doc-comment
// hint about absolute-difference comparison.
//
// Cases to cover: a 3-4-5 right triangle (exact), the origin (0), at least
// one case with both negative coordinates, and a small case where the
// answer is irrational (use math.Abs(got - want) < 1e-9 for the comparison).
func TestWarmupPointDistanceFromOrigin(t *testing.T) {
	cases := []struct {
		name string
		p    WarmupPoint
		want float64
	}{
		// TODO: add at least 4 cases. Use the examples in the
		// DistanceFromOrigin doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got := tc.p.DistanceFromOrigin()
			//   diff := got - tc.want
			//   if diff < 0 { diff = -diff }    // math.Abs without the import
			//   if diff > 1e-9 { t.Errorf("...") }
			//
			// Or import math and use math.Abs.
			_ = tc
		})
	}
}
```

> Note: floating-point comparison needs a tolerance. Reasonable values are `1e-9` for the irrational cases here. The doc-comment hint guides students through the inline approach; importing `math` and using `math.Abs` is equally fine.

- [ ] **Step 3: Replace `lessons/06-structs/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 06: Structs and methods.
//
// This file holds the warm-up reference solution.
package solutions

import "math"

// WarmupPoint is a 2D point with float64 coordinates.
type WarmupPoint struct {
	X float64
	Y float64
}

// DistanceFromOrigin returns sqrt(X*X + Y*Y).
func (p WarmupPoint) DistanceFromOrigin() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}
```

- [ ] **Step 4: Replace `lessons/06-structs/solutions/warmup_test.go`** with the full reference tests (`package solutions`):

```go
package solutions

import (
	"math"
	"testing"
)

func TestWarmupPointDistanceFromOrigin(t *testing.T) {
	cases := []struct {
		name string
		p    WarmupPoint
		want float64
	}{
		{"3-4-5", WarmupPoint{X: 3, Y: 4}, 5},
		{"origin", WarmupPoint{X: 0, Y: 0}, 0},
		{"unit-x", WarmupPoint{X: 1, Y: 0}, 1},
		{"unit-y", WarmupPoint{X: 0, Y: 1}, 1},
		{"diagonal", WarmupPoint{X: 1, Y: 1}, math.Sqrt2},
		{"both-negative", WarmupPoint{X: -3, Y: -4}, 5},
		{"large", WarmupPoint{X: 300, Y: 400}, 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.p.DistanceFromOrigin()
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("WarmupPoint{%g, %g}.DistanceFromOrigin() = %g, want %g (tolerance 1e-9)",
					tc.p.X, tc.p.Y, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/06-structs/
```

Expected: empty. `→` arrows must survive.

- [ ] **Step 6: Verify the warm-up exercise test "passes vacuously"**

```bash
go test ./lessons/06-structs/exercises/... -run Warmup -v 2>&1 | tail -10
```

Expected: TestWarmupPointDistanceFromOrigin PASS (0 sub-tests).

- [ ] **Step 7: Verify the warm-up solution tests pass**

```bash
go test ./lessons/06-structs/solutions/... -run Warmup -v 2>&1 | tail -15
```

Expected: TestWarmupPointDistanceFromOrigin (7 sub-tests) PASS.

- [ ] **Step 8: `make test` and lint clean**

```bash
make test
golangci-lint run ./...
```

Expected: green, `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/06-structs/exercises/warmup.go lessons/06-structs/exercises/warmup_test.go \
        lessons/06-structs/solutions/warmup.go lessons/06-structs/solutions/warmup_test.go
git commit -m "feat(lesson-06): warm-up — WarmupPoint + DistanceFromOrigin (skeleton test)"
```

---

## Task 3: Author the main exercise

Define `Expense{Date, Amount, Category}` with `Format() string` and `IsHigh() bool` methods, plus the standalone `TotalsByCategory(es []Expense) map[string]float64`. Tests for all three ship as skeletons.

**Files:**
- Replace: `lessons/06-structs/exercises/main.go`
- Replace: `lessons/06-structs/exercises/main_test.go`
- Replace: `lessons/06-structs/solutions/main.go`
- Replace: `lessons/06-structs/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/06-structs/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 06: Structs and methods.
//
// This file holds the MAIN exercise. Define the Expense struct, give it
// two methods (Format and IsHigh), and write the TotalsByCategory function
// that aggregates a []Expense by category.
package exercises

// Expense is one row of the expense tracker — a single charge with a date,
// amount, and category.
//
// All three fields are exported (capitalised) because callers in the test
// file need to construct Expense values directly using struct literals.
// Lesson 07 covers exported vs unexported formally.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same format as lesson 04's FormatExpense — now bound to
// the Expense receiver.
//
// Examples:
//
//	Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-12  €4.50    coffee"
//	Expense{Date: "2026-05-12", Amount: 12, Category: "lunch"}.Format()
//	  → "2026-05-12  €12.00   lunch"
//	Expense{Date: "2026-05-12", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-12  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Examples:
//
//	Expense{Amount: 4.50}.IsHigh()  → false
//	Expense{Amount: 50}.IsHigh()    → false   (boundary: 50 is NOT high)
//	Expense{Amount: 50.01}.IsHigh() → true
//	Expense{Amount: 999}.IsHigh()   → true
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// TotalsByCategory returns a map of category → summed amount, built from es.
//
// Empty or nil input returns an empty (non-nil) map — not an error. Callers
// can range over it normally.
//
// Examples:
//
//	es := []Expense{
//		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-12", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]   (non-nil empty map)
//
// Hint: initialise `totals := map[string]float64{}`, then range over es
// accumulating totals[e.Category] += e.Amount. Lesson 05's idiom works here.
func TotalsByCategory(es []Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into a map[string]float64")
}
```

- [ ] **Step 2: Replace `lessons/06-structs/exercises/main_test.go`** with (SKELETON):

```go
package exercises

import (
	"reflect"
	"testing"
)

// TestExpenseFormat is a SKELETON. Fill in cases and the t.Run body.
//
// Build the expected output strings by hand using "%s  €%-7.2f %s".
// Use the examples in the Format doc comment as a starting point.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases. Include the small/medium/large amounts
		// from the doc comment so you exercise the %-7.2f width on values
		// of different magnitude.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover the snack/regular/splurge-like
// boundaries: amount well under 50, exactly 50 (NOT high), just over 50,
// and a large value.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases including the boundary at 50.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Expense{Amount: tc.amount}.IsHigh()
			//   compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}

// TestTotalsByCategory is a SKELETON. Cover at least:
//   - distinct categories (no overlap)
//   - repeated categories (totals must accumulate)
//   - a single-element slice
//   - an empty slice (result is a non-nil empty map)
//   - nil input (also returns an empty map)
func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []Expense
		want map[string]float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := TotalsByCategory(tc.es)
			//   assert got != nil (always returns a non-nil map)
			//   assert reflect.DeepEqual(got, tc.want)
			_ = tc
			_ = reflect.DeepEqual
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/06-structs/solutions/main.go`** with:

```go
// Package solutions is the reference implementation for lesson 06: Structs and methods.
package solutions

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

// TotalsByCategory returns a map of category → summed amount.
// Empty/nil input returns an empty (non-nil) map.
func TotalsByCategory(es []Expense) map[string]float64 {
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}
	return totals
}
```

- [ ] **Step 4: Replace `lessons/06-structs/solutions/main_test.go`** with full reference tests:

```go
package solutions

import (
	"reflect"
	"testing"
)

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"}, "2026-05-12  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-12", Amount: 12, Category: "lunch"}, "2026-05-12  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-12", Amount: 999.99, Category: "rent"}, "2026-05-12  €999.99  rent"},
		{"zero-amount", Expense{Date: "2026-05-12", Amount: 0, Category: "free"}, "2026-05-12  €0.00    free"},
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

func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []Expense
		want map[string]float64
	}{
		{
			"distinct",
			[]Expense{
				{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-12", Amount: 12, Category: "lunch"},
				{Date: "2026-05-12", Amount: 75, Category: "rent"},
			},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
		},
		{
			"repeated-categories",
			[]Expense{
				{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-12", Amount: 12, Category: "lunch"},
				{Date: "2026-05-12", Amount: 9.99, Category: "coffee"},
			},
			map[string]float64{"coffee": 14.49, "lunch": 12},
		},
		{
			"single",
			[]Expense{{Date: "2026-05-12", Amount: 42, Category: "x"}},
			map[string]float64{"x": 42},
		},
		{"empty", []Expense{}, map[string]float64{}},
		{"nil", nil, map[string]float64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TotalsByCategory(tc.es)
			if got == nil {
				t.Fatalf("TotalsByCategory(%v) returned nil; expected a non-nil map", tc.es)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory(%v) = %v, want %v", tc.es, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Format-output smoke test (Plan G lesson learned)**

The `Format()` output uses `%-7.2f` width — verify the expected strings byte-match what `fmt.Sprintf` produces:

```bash
cat > /tmp/format-check.go <<'EOF'
package main
import "fmt"
func main() {
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-12", 4.50, "coffee"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-12", 12.00, "lunch"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-12", 999.99, "rent"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-12", 0.00, "free"))
}
EOF
go run /tmp/format-check.go
rm /tmp/format-check.go
```

Expected output (must match the test's `want` strings exactly):

```
"2026-05-12  €4.50    coffee"
"2026-05-12  €12.00   lunch"
"2026-05-12  €999.99  rent"
"2026-05-12  €0.00    free"
```

If the actual output differs, fix the test cases (NOT the format string).

- [ ] **Step 6: Run gofmt**

```bash
gofmt -l lessons/06-structs/
```

Expected: empty.

- [ ] **Step 7: Verify exercise main tests pass vacuously**

```bash
go test ./lessons/06-structs/exercises/... -v 2>&1 | tail -20
```

Expected: TestExpenseFormat, TestExpenseIsHigh, TestTotalsByCategory all PASS with 0 sub-tests. Warm-up also passes vacuously.

- [ ] **Step 8: Verify solution main tests pass**

```bash
go test ./lessons/06-structs/solutions/... -v 2>&1 | tail -40
```

Expected: every test PASSes — TestWarmupPointDistanceFromOrigin (7), TestExpenseFormat (4), TestExpenseIsHigh (6), TestTotalsByCategory (5).

- [ ] **Step 9: Verify make test, lint, vet pass**

```bash
make test
golangci-lint run ./...
go vet ./...
```

Expected: green, `0 issues.`, clean vet.

- [ ] **Step 10: Commit**

```bash
git add lessons/06-structs/exercises/main.go lessons/06-structs/exercises/main_test.go \
        lessons/06-structs/solutions/main.go lessons/06-structs/solutions/main_test.go
git commit -m "feat(lesson-06): main — Expense{Format,IsHigh} + TotalsByCategory(es []Expense)"
```

---

## Task 4: Author the slide deck

Replace `lessons/06-structs/slides/slides.md`. Four concepts: struct definition → methods (value receivers) → pointer receivers (light) → embedding + exported/unexported (light).

**Files:**
- Replace: `lessons/06-structs/slides/slides.md`

The controller writes this directly (Plans G/H pattern).

- [ ] **Step 1: Write `lessons/06-structs/slides/slides.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device. Use only three-backtick fences in the file. File starts with `<div class="title-slide-grid">`.

````markdown
<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">06</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Structs and methods</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Define your own types with <code>struct</code>, attach behaviour with methods, prefer value receivers (the Phase 1 default) and understand when pointer receivers come in, plus a light touch of embedding and exported-vs-unexported fields.</p>
</div>
</div>
</div>

---

## What we'll cover

- `type T struct { ... }` — struct definitions, literal forms, zero-value structs.
- Methods with value receivers: `func (t T) M() ...` — calling, method set.
- Pointer receivers (light): `func (t *T) M() ...` — when to mutate, full treatment in lesson 09.
- Embedding + exported/unexported fields — light, full treatment in lesson 07.

---

## Concept 1: Struct definition and literals

### Motivation

Structs are how you bundle related data into a single value. Three named fields beat three parallel slices any day — `Expense{Date, Amount, Category}` carries its own structure, while `(amounts []float64, categories []string)` is just two slices the programmer has to keep in sync. Phase 1 has been edging toward this — lesson 05's `TotalsByCategory` took parallel slices; lesson 06's takes a `[]Expense`.

---

### The basics

```go
package main

import "fmt"

// Type declaration — once per type, at package level.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func main() {
	// 1. Named-field literal (the only style you should write).
	a := Expense{
		Date:     "2026-05-12",
		Amount:   4.50,
		Category: "coffee",
	}

	// 2. Positional literal — fields in declaration order.
	// Avoid this in real code — it breaks the moment someone adds a field.
	b := Expense{"2026-05-12", 12, "lunch"}

	// 3. Partial named literal — omitted fields get their zero value.
	c := Expense{Date: "2026-05-12"}     // Amount: 0, Category: ""

	// 4. Zero-value struct — every field at its type's zero value.
	var d Expense                          // {Date: "", Amount: 0, Category: ""}

	fmt.Printf("%+v\n", a)  // {Date:2026-05-12 Amount:4.5 Category:coffee}
	fmt.Printf("%+v\n", b)
	fmt.Printf("%+v\n", c)
	fmt.Printf("%+v\n", d)

	// Field access.
	fmt.Println(a.Amount, a.Category)

	// Field mutation (when the variable is addressable).
	a.Amount = 5.00
	fmt.Println(a.Amount)
}
```

Three idioms:

- **Always use named-field literals.** `Expense{Date: ..., Amount: ..., Category: ...}` survives field reordering and additions. Positional literals look concise but break under maintenance.
- **Every field starts at its zero value.** Same rule as lesson 02 for plain variables — string `""`, numeric `0`, bool `false`.
- **`%+v` prints field names** in `fmt.Printf` — useful for debugging.

---

### A worked example

The lesson's main exercise — `Expense` with all three fields exported, ready for tests to construct values directly:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func main() {
	es := []Expense{
		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
		{Date: "2026-05-12", Amount: 75, Category: "rent"},
	}
	for _, e := range es {
		fmt.Printf("%s  €%.2f  %s\n", e.Date, e.Amount, e.Category)
	}
}
```

A slice of structs is the bread-and-butter "rows of data" shape in Go. We'll come back to this in concept 2 once we add methods.

---

### Common mistake

Positional literals + adding a field = silent bug:

```go
// Original code.
type Expense struct {
	Date   string
	Amount float64
}

e := Expense{"2026-05-12", 4.50}  // works
```

Later, someone adds `Category` to the struct:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string   // new field
}

e := Expense{"2026-05-12", 4.50}  // ERROR: too few values in struct literal
```

Now every positional literal in the codebase fails to compile. With named literals, this isn't an issue — the new field just gets its zero value.

```go
e := Expense{Date: "2026-05-12", Amount: 4.50}   // Category: "" implicitly
```

Use named-field literals from the start.

---

### Recap

- `type T struct { Field Type; ... }` defines a struct type at package level.
- Use named-field literals: `T{Field: value, ...}`. Avoid positional literals.
- Every field gets its zero value when omitted.
- `%+v` in `Printf` prints field names — handy for debugging.

---

## Concept 2: Methods (value receivers)

### Motivation

A method is a function attached to a type. Behavioural code that operates on a value of type T gets attached to T as a method, so you can call `t.Method()` instead of `Method(t)`. Same logic, better organisation.

---

### The basics

A method declaration looks like a function with one extra bit — the **receiver** before the method name:

```go
package main

import "fmt"

type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Method on Expense. The (e Expense) is the receiver — it makes this function
// a method of type Expense.
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// Another method on the same type.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

func main() {
	e := Expense{Date: "2026-05-12", Amount: 75, Category: "rent"}
	fmt.Println(e.Format())   // 2026-05-12  €75.00   rent
	fmt.Println(e.IsHigh())   // true
}
```

Three things:

- **Receiver syntax: `(e Expense)`** — like an extra parameter that comes before the method name. By convention, the receiver variable is one short letter or two — `e Expense`, not `expense Expense`.
- **Call site: `e.Format()`** — looks just like calling a method in any OO language. Under the hood, this is `Format(e)` with `e` passed as the receiver.
- **Methods live on the same package as the type.** You can only define methods on types declared in your own package — you can't add methods to `int` or `string`.

---

### A worked example

The main exercise — slice of Expense, format each, total by category:

```go
package main

import "fmt"

func main() {
	es := []Expense{
		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
		{Date: "2026-05-12", Amount: 75, Category: "rent"},
	}
	for _, e := range es {
		marker := ""
		if e.IsHigh() {
			marker = "  (!)"
		}
		fmt.Println(e.Format() + marker)
	}
}
```

`Format` and `IsHigh` are methods, so we call them on each element naturally. Same logic, much better organised than a top-level `FormatExpense(date, amount, category)` function with positional arguments.

---

### Common mistake

Forgetting the receiver type — writing a plain function and trying to call it as a method:

```go
// Wrong — this is just a function, not a method.
func Format(e Expense) string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func main() {
	e := Expense{...}
	fmt.Println(e.Format())   // error: e.Format undefined (type Expense has no field or method Format)
}
```

The fix is the receiver — turn `func Format(e Expense)` into `func (e Expense) Format()`. The change is small but it's what makes the function a method.

---

### Recap

- Methods are functions with a receiver — `func (e Expense) Format() string`.
- Call sites look like OO: `e.Format()`.
- You can only define methods on types declared in your own package.
- Convention: receiver names are short (`e Expense`, not `expense Expense`).

---

## Concept 3: Pointer receivers (light)

### Motivation

Value receivers (`(t T)`) get a *copy* of the receiver. That's fine for read-only methods, but if you want to mutate the value, you need a **pointer receiver** (`(t *T)`). Pointer receivers also avoid copying large structs. Pointers are introduced properly in lesson 09; here we just show the syntax and the rule, so you recognise it when you see it.

---

### The basics

```go
package main

import "fmt"

type Counter struct {
	n int
}

// Value receiver — copies the Counter. Mutations don't affect the caller.
func (c Counter) IncrementByValue() {
	c.n++   // mutates the local copy only
}

// Pointer receiver — receives a pointer to the Counter. Mutations affect the caller.
func (c *Counter) IncrementByPointer() {
	c.n++   // mutates the original
}

func main() {
	c := Counter{n: 0}
	c.IncrementByValue()
	fmt.Println(c.n)   // 0 — IncrementByValue had no effect

	c.IncrementByPointer()
	fmt.Println(c.n)   // 1 — IncrementByPointer mutated the original
}
```

Two rules of thumb (lesson 09 explains *why*):

- **Use a value receiver when the method just reads.** `Format()` and `IsHigh()` in our `Expense` are value receivers — they don't change the struct. Most of Phase 1's methods are value receivers.
- **Use a pointer receiver when the method mutates the receiver** — `IncrementByPointer` modifies `c.n`. Or when the struct is large and copying it is expensive.

Tiny stylistic rule: be consistent. If *one* method on a type needs a pointer receiver, the convention is to make *all* methods on that type pointer receivers. You don't have to follow this in Phase 1 since we're using value receivers, but you'll see it in real code.

---

### A worked example

You won't write a pointer receiver in this lesson — the main exercise stays with value receivers throughout. Here's a preview of when you'd reach for one:

```go
type Cart struct {
	Items []string
}

// Pointer receiver: Add() needs to mutate Cart.Items.
func (c *Cart) Add(item string) {
	c.Items = append(c.Items, item)
}

func main() {
	cart := Cart{}
	cart.Add("apple")
	cart.Add("banana")
	fmt.Println(cart.Items)   // [apple banana]
}
```

If `Add` had a value receiver, the appends would happen on a local copy and `cart.Items` would stay empty.

---

### Common mistake

A value-receiver method "trying" to mutate the struct:

```go
type Expense struct {
	Amount float64
}

// Doesn't work — value receiver mutates the copy.
func (e Expense) Discount(rate float64) {
	e.Amount *= (1 - rate)
}

func main() {
	e := Expense{Amount: 100}
	e.Discount(0.10)
	fmt.Println(e.Amount)   // 100 — discount didn't stick
}
```

Two fixes:

- **Pointer receiver:** `func (e *Expense) Discount(rate float64)`. The method now mutates the original.
- **Return a new value:** `func (e Expense) Discounted(rate float64) Expense` returning the discounted version. Caller writes `e = e.Discounted(0.10)`. This is actually the more idiomatic shape — immutability by default, mutate-by-assignment.

Phase 1 leans on the "return a new value" pattern. Pointer receivers come into their own in lesson 09.

---

### Recap

- Value receiver `(t T)` — gets a copy. Read-only.
- Pointer receiver `(t *T)` — gets a pointer. Can mutate.
- Phase 1 prefers value receivers; lesson 09 covers pointers properly.
- The "return a new value" alternative often reads cleaner than a mutating method.

---

## Concept 4: Embedding + exported/unexported (light)

### Motivation

Two short Go-isms to round out the lesson. **Embedding** is Go's answer to "I want type B to include all of type A's fields and methods" — close to inheritance but actually composition. **Exported vs unexported** is Go's encapsulation rule: capitalised names are public, lower-cased names are private to the package. Both get fuller treatment in lesson 07.

---

### The basics

**Embedding** — declare one struct type inside another with no field name:

```go
package main

import "fmt"

type Address struct {
	City    string
	Country string
}

type Customer struct {
	Address          // embedded — no field name
	Name string
}

func (a Address) Format() string {
	return fmt.Sprintf("%s, %s", a.City, a.Country)
}

func main() {
	c := Customer{
		Address: Address{City: "Helsinki", Country: "Finland"},
		Name:    "Aki",
	}
	fmt.Println(c.Name)                // Aki
	fmt.Println(c.City)                // Helsinki — promoted from Address
	fmt.Println(c.Address.City)        // Helsinki — explicit path also works
	fmt.Println(c.Format())            // Helsinki, Finland — method promoted too
}
```

Two things to notice:

- `c.City` works because the field is *promoted* from the embedded `Address` — there's no separate `City` field on `Customer`.
- `c.Format()` works similarly — `Address`'s method is also promoted.

This isn't inheritance — it's composition. Under the hood, `Customer` *has-a* `Address`; the dot notation just hides the indirection when there's no ambiguity. Lesson 07 shows how this maps to "compose a thin orchestrator on top of focused sub-types."

**Exported vs unexported** — pure case-of-first-letter:

```go
type Expense struct {
	Date     string   // exported (capital D)
	Amount   float64  // exported
	Category string   // exported
	notes    string   // unexported — accessible only within this package
}

func (e Expense) AddNote(note string) Expense {   // exported method
	e.notes = note   // assign within the package; outside callers can't read this
	return e
}

func (e Expense) note() string {   // unexported method
	return e.notes
}
```

From another package, `e.Date` is accessible, `e.notes` is not. `e.AddNote(...)` is accessible, `e.note()` is not. The rule is uniform across fields, methods, functions, types, constants, and variables: capitalised first letter → exported, otherwise → package-private.

We've been declaring all our exercise/solution fields as exported (capitalised) because the tests construct values via struct literals — they need to set every field. Lesson 07 shows how to design APIs that hide implementation details behind unexported fields.

---

### A worked example

Our `Expense` type is the simplest possible case — three exported fields, no embedding. Even so, the capitalisation matters: rename `Date` to `date` in `lessons/06-structs/solutions/main.go` and `solutions/main_test.go` would stop compiling, because `Expense{Date: "..."}` in the test would refer to a non-existent field.

A toy embedding example (you won't write this in the exercise but worth seeing):

```go
type Timestamped struct {
	CreatedAt string
}

type TimestampedExpense struct {
	Expense        // embed Expense
	Timestamped    // embed Timestamped
}

func main() {
	te := TimestampedExpense{
		Expense:     Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		Timestamped: Timestamped{CreatedAt: "2026-05-12T08:30:00Z"},
	}
	fmt.Println(te.Format(), te.CreatedAt)   // both promoted from embedded types
}
```

---

### Common mistake

Forgetting that lowercase = unexported, and being surprised when an external test can't see the field:

```go
// In package mypkg:
type Counter struct {
	count int   // unexported
}

// In another package's test:
c := mypkg.Counter{count: 5}   // ERROR: cannot refer to unexported field
```

Fix: either capitalise the field (`Count`), or expose it through an exported method (`func (c Counter) Count() int { return c.count }`).

The whole-package convention to follow: if a field's value is part of the type's public behaviour, export it. If it's an implementation detail, keep it unexported. Lesson 07 explores this more rigorously.

---

### Recap

- **Embedding:** `type Outer struct { Inner; ... }` (no field name) — `Inner`'s fields and methods are promoted onto `Outer`. Composition, not inheritance.
- **Exported vs unexported:** capitalised first letter → visible to other packages, lowercased → package-private. Applies to fields, methods, functions, types, etc.
- Both topics get fuller treatment in lesson 07 (packages and modules).

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `type WarmupPoint struct { X, Y float64 }` — already defined.
- `func (p WarmupPoint) DistanceFromOrigin() float64` — implement. Return `math.Sqrt(p.X*p.X + p.Y*p.Y)`.

The skeleton test is in `exercises/warmup_test.go`. Floats need a tolerance — use `math.Abs(got - want) < 1e-9` (and import `math`), or compute the absolute difference inline.

```bash
cd lessons/06-structs/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `type Expense struct { Date string; Amount float64; Category string }` — already defined.
- `func (e Expense) Format() string` — return `"%s  €%-7.2f %s"` using `fmt.Sprintf`. Same as lesson 04's `FormatExpense`, but as a method.
- `func (e Expense) IsHigh() bool` — return `e.Amount > 50`. Boundary value 50 is NOT high.
- `func TotalsByCategory(es []Expense) map[string]float64` — walk `es` with `range`, accumulate into a `map[string]float64`. Empty/nil input returns a non-nil empty map.

Skeleton tests in `exercises/main_test.go` — fill in cases for all three.

```bash
cd lessons/06-structs/exercises
go test -v
```

Note:
For live: walk through the field-promotion magic of embedding on the projector — define Timestamped + TimestampedExpense, range over a `[]TimestampedExpense` calling `.Format()` and `.CreatedAt` to show both promoted. The "promoted = composition, not inheritance" point lands well as a contrast for students with OO backgrounds.

---

## What we learned

- `type T struct { ... }` defines a struct; always use named-field literals.
- Methods attach behaviour to a type via the receiver: `func (e Expense) Format()`.
- Value receivers (`(t T)`) get a copy; pointer receivers (`(t *T)`) can mutate. Phase 1 prefers value receivers + "return a new value" patterns; lesson 09 covers pointers properly.
- Embedding (`type B struct { A; ... }`) promotes `A`'s fields and methods onto `B` — composition, not inheritance.
- Exported (capitalised) vs unexported (lowercased) names. Both fields and methods.

---

## Up next

Lesson 07 — Packages and modules (splitting code into packages, formal exported/unexported, `gofmt`/`go vet`/`go doc` tour).
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=06-structs &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/06-structs/slides/slides.md
curl -sS http://localhost:8000/lessons/06-structs/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/06-structs/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/06-structs/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 3: Verify markdown structure**

```bash
grep -c '<div class="title-slide-grid">' lessons/06-structs/slides/slides.md
grep -c "<h1>Structs and methods</h1>" lessons/06-structs/slides/slides.md
grep -c "^## What we learned" lessons/06-structs/slides/slides.md
grep -c "^## Up next" lessons/06-structs/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/06-structs/slides/slides.md
git commit -m "feat(lesson-06): slides — Structs and methods (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/06-structs/README.md` with the lesson 06 self-study prose. Same shape as Plans D-H.

**Files:**
- Replace: `lessons/06-structs/README.md`

- [ ] **Step 1: Write `lessons/06-structs/README.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device; use three-backtick fences in the file. File starts with `# Lesson 06: Structs and methods`.

````markdown
# Lesson 06: Structs and methods

## Learning goals

- Define your own types with `type T struct { ... }` and construct values with named-field literals.
- Attach behaviour to a type via methods (`func (t T) M() ...`); know when to use a value receiver.
- Recognise pointer receivers (`func (t *T) M() ...`) and the value-vs-pointer distinction — Phase 1 prefers value receivers; lesson 09 covers pointers fully.
- See struct embedding (`type B struct { A; ... }`) for composition, and exported vs unexported names for visibility — both get the full treatment in lesson 07.

## Prerequisites

- Lessons 01-05. In particular, lesson 05's `TotalsByCategory(amounts, categories)` is refactored here into `TotalsByCategory(es []Expense)` — go back and skim lesson 05 if the map-accumulation idiom isn't fresh.

## Concepts

### Struct definition and literals

A struct bundles related fields into one named type:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}
```

Four ways to build a value:

```go
a := Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"}    // named-field
b := Expense{"2026-05-12", 12, "lunch"}                                 // positional — avoid
c := Expense{Date: "2026-05-12"}                                        // partial — others zero
var d Expense                                                           // zero-value
```

Three idioms:

- **Always use named-field literals.** Positional literals look concise but break the moment someone adds a field.
- **Every field starts at its zero value.** `""`, `0`, `false`, `nil` (for slices/maps/pointers/interfaces).
- **`%+v`** in `Printf` prints field names — useful for debugging.

**Common mistake.** Positional literal + new field = compile error:

```go
type Expense struct {
	Date   string
	Amount float64
}

e := Expense{"2026-05-12", 4.50}     // works

// Later, someone adds Category:
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

e := Expense{"2026-05-12", 4.50}     // ERROR: too few values
```

Fix: use named-field literals from the start. `Expense{Date: "...", Amount: ...}` survives field additions — the new field just gets its zero value.

### Methods (value receivers)

A method is a function with a **receiver** before the name:

```go
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func (e Expense) IsHigh() bool {
	return e.Amount > 50
}
```

Call sites look like methods in any OO language:

```go
e := Expense{Date: "2026-05-12", Amount: 75, Category: "rent"}
e.Format()   // "2026-05-12  €75.00   rent"
e.IsHigh()   // true
```

Three things:

- **Receiver syntax** `(e Expense)` is like an extra parameter that comes before the method name. Convention: receiver names are short (`e`, not `expense`).
- **`e.Format()` is shorthand for `Format(e)`** under the hood. The dot is just sugar.
- **Methods live in the same package as the type.** You can't add methods to `int`, `string`, or anything from another package — only to types you declared yourself.

**Common mistake.** Writing a function and trying to call it as a method:

```go
func Format(e Expense) string { ... }   // plain function, not a method

e := Expense{...}
e.Format()                                // error: undefined
```

Fix: add the receiver. Change `func Format(e Expense)` to `func (e Expense) Format()`.

### Pointer receivers (light)

Value receivers `(t T)` get a *copy* of the receiver. Mutations don't reach the caller:

```go
type Counter struct {
	n int
}

func (c Counter) IncrementByValue() {
	c.n++   // mutates the local copy only
}

c := Counter{n: 0}
c.IncrementByValue()
fmt.Println(c.n)   // 0 — IncrementByValue had no effect
```

Pointer receivers `(t *T)` get a pointer to the value; mutations stick:

```go
func (c *Counter) IncrementByPointer() {
	c.n++
}

c.IncrementByPointer()
fmt.Println(c.n)   // 1 — mutation stuck
```

Two rules of thumb (lesson 09 explains *why*):

- **Value receiver when the method just reads.** Most of Phase 1's methods are value receivers.
- **Pointer receiver when the method mutates the receiver** — `IncrementByPointer` above. Or when the struct is large and copying would be expensive.

Convention: if one method on a type needs a pointer receiver, make *all* methods on that type pointer receivers. Be consistent.

Phase 1 leans on a different pattern altogether: **return a new value**, immutability by default.

```go
// Instead of mutating, return a new Expense.
func (e Expense) Discounted(rate float64) Expense {
	e.Amount *= (1 - rate)
	return e
}

// Caller reassigns:
e = e.Discounted(0.10)
```

This sidesteps the pointer-receiver question entirely and is often more readable. Use it when you can.

**Common mistake.** A value-receiver method trying to mutate the struct:

```go
func (e Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)   // mutates the copy, not the caller's value
}

e := Expense{Amount: 100}
e.ApplyDiscount(0.10)
fmt.Println(e.Amount)   // 100 — discount didn't stick
```

Two fixes: `*Expense` pointer receiver, or return a new value (`Discounted` above).

### Embedding + exported/unexported (light)

**Embedding** is Go's "type B has-a type A" shape, with field promotion:

```go
type Address struct {
	City    string
	Country string
}

type Customer struct {
	Address          // embedded — no field name
	Name string
}

func (a Address) Format() string {
	return fmt.Sprintf("%s, %s", a.City, a.Country)
}

c := Customer{
	Address: Address{City: "Helsinki", Country: "Finland"},
	Name:    "Aki",
}
c.City                // "Helsinki" — promoted from Address
c.Address.City        // also "Helsinki" — explicit path
c.Format()            // "Helsinki, Finland" — method also promoted
```

`Address` fields and methods are *promoted* onto `Customer` — `c.City` looks identical to a real field on `Customer`. It isn't inheritance; it's composition with a dot-notation shortcut.

**Exported vs unexported** is pure case-of-first-letter:

```go
type Expense struct {
	Date     string   // exported (capital D) — visible from other packages
	Amount   float64  // exported
	Category string   // exported
	notes    string   // unexported (lower-case n) — package-private
}
```

The rule applies to everything: fields, methods, functions, types, constants, variables. Capitalised → public; lower-case → private to the declaring package.

We've been declaring all exercise/solution fields as exported because the tests construct values with struct literals. Lesson 07 shows how to design APIs that hide internals behind unexported names.

**Common mistake.** Lowercase = unexported = invisible to other packages:

```go
// In package mypkg:
type Counter struct {
	count int
}

// In another package:
c := mypkg.Counter{count: 5}   // ERROR: cannot refer to unexported field
```

Fix: capitalise the field (`Count`), or expose it through a method (`func (c Counter) Count() int { return c.count }`).

## Exercise: warm-up

Two things in `exercises/warmup.go`:

- `type WarmupPoint struct { X, Y float64 }` — already defined.
- `func (p WarmupPoint) DistanceFromOrigin() float64` — implement. Return `math.Sqrt(p.X*p.X + p.Y*p.Y)`.

Then fill in the skeleton `TestWarmupPointDistanceFromOrigin` in `exercises/warmup_test.go`. Floats need a tolerance — use `math.Abs(got - want) < 1e-9` (import `math` if you want) or compute the absolute difference inline.

## Exercise: main

Three things in `exercises/main.go`:

- `type Expense struct { Date string; Amount float64; Category string }` — already defined.
- `func (e Expense) Format() string` — use `fmt.Sprintf` with `"%s  €%-7.2f %s"`. Same byte-for-byte output as lesson 04's `FormatExpense`. The doc comment has the three expected output strings.
- `func (e Expense) IsHigh() bool` — return `e.Amount > 50`. Note: 50 itself is **not** high.
- `func TotalsByCategory(es []Expense) map[string]float64` — walk the slice with `range`, accumulate into a `map[string]float64`. Empty/nil input must return a non-nil empty map.

Fill in the three skeleton tests in `exercises/main_test.go`:

- `TestExpenseFormat` — at least 3 cases. The `%-7.2f` width matters — build the expected strings carefully.
- `TestExpenseIsHigh` — at least 4 cases, including the boundary value 50 (not high) and just over 50 (high).
- `TestTotalsByCategory` — at least 4 cases. Cover distinct categories, repeated category (totals accumulate), single-element slice, and empty input (non-nil empty map).

> A note on `make test-lesson LESSON=06-structs`: lesson 06's exercise tests now ship with empty `cases` slices, so the make output may show `ok` (vacuous pass) before you add cases. Don't be fooled — run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/06-structs/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main
```

Keep the lesson 04 habits:

```bash
gofmt -w .       # reformat to canonical Go style
go vet ./...     # subtle bug catcher
```

CI enforces both.

## Going further

### Read

- [Effective Go — Structs](https://go.dev/doc/effective_go#composite_literals) — short canonical reference.
- [Effective Go — Embedding](https://go.dev/doc/effective_go#embedding) — the "composition not inheritance" case study.
- [Go FAQ — Why doesn't Go have inheritance?](https://go.dev/doc/faq#inheritance) — the design rationale; useful counterpoint if you're coming from an OO background.

### Try

- **A pointer-receiver variant.** Rewrite `Expense.IsHigh()` with a pointer receiver and observe that nothing changes for the caller (because `IsHigh` doesn't mutate). Then write a mutating method — `Expense.Bump(amount float64)` that adds to `Amount` — and observe that you need a pointer receiver for it to stick.
- **Embedding in action.** Add a `Timestamped` type with a `CreatedAt string` field, then define `type TimestampedExpense struct { Expense; Timestamped }` and verify that `te.Format()` and `te.CreatedAt` both work without explicit field paths.
- **An unexported field.** Add a private `note string` field to `Expense` and an `AddNote(string)` method (value receiver — returns a new `Expense`). Run `go test` from `solutions/` — the test should still pass because it doesn't touch the unexported field. Then try to access `e.note` from a scratch file in `exercises/` — it works, because both files are in the same package. Move that scratch file to a different package (a `cmd/play/main.go` would do it) and watch the compiler reject it.
````

- [ ] **Step 2: Verify the README structure**

```bash
for h in "^# Lesson 06: Structs and methods$" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  c=$(grep -c "$h" lessons/06-structs/README.md)
  echo "  $h => $c"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/06-structs/README.md   # expect 4
grep -c "gofmt" lessons/06-structs/README.md
grep -c "go vet" lessons/06-structs/README.md
```

Expected: all section headings 1; Common-mistake count 4; gofmt and go vet counts ≥ 1.

- [ ] **Step 3: Commit**

```bash
git add lessons/06-structs/README.md
git commit -m "docs(lesson-06): README — Structs and methods self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes. Lesson 06 solutions pass; lessons 01-05 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — warm-up + main both pass vacuously**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 06's TestWarmupPointDistanceFromOrigin, TestExpenseFormat, TestExpenseIsHigh, TestTotalsByCategory all PASS with 0 sub-tests (skeleton case slices empty). Lessons 01-04 fail as before (panic-stubs); lesson 05 also passes vacuously (skeleton); make exits 0 because `-` ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=06-structs 2>&1 | tail -30
```

Expected: exercise tests pass vacuously (0 sub-tests, ignored), solution tests pass fully (22 sub-tests: TestWarmupPointDistanceFromOrigin 7, TestExpenseFormat 4, TestExpenseIsHigh 6, TestTotalsByCategory 5).

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
make slides-dev LESSON=06-structs &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/06-structs/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/06-structs/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 7: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/06-structs/slides/slides.md && echo OK
grep -q "Structs" dist/index.html && echo "index lists lesson 06"
rm -rf dist
```

Expected: three success lines. Landing page lists lessons 01-06.

- [ ] **Step 8: Final repository sanity check**

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch (plan + scaffold + warm-up + main + slides + README).

This task makes no commit.

---

## Done definition

After Task 6:

- `lessons/06-structs/` contains 12 files with full lesson content.
- `make test` passes; `make test-exercises` shows lesson 06's tests passing vacuously.
- `make test-lesson LESSON=06-structs` runs both sides correctly.
- `make slides-dev LESSON=06-structs` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan J — Lesson 07 (Packages and modules).** Same per-lesson pattern. Lesson 07 is the first lesson with **subfolders** — the act of creating an `expense/` subpackage is the exercise. The Phase 1 spec also gives lesson 07 the formal "exported vs unexported" treatment (this lesson covered it lightly), the `gofmt`/`go vet`/`go doc` tour, and the first appearance of stdlib `bufio`/`strings`. The main exercise refactors lesson 06's `Expense` + `Format`/`IsHigh` into a proper `expense` subpackage that `exercises/main.go` imports.
