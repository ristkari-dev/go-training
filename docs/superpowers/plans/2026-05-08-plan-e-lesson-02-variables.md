# Plan E — Lesson 02 (Variables, types, operators) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the second lesson of Phase 1 — students learn Go's variable declaration forms, basic types and zero values, arithmetic operators, type conversions, and constants. The exercise extends the running expense-tracker theme to compute totals, averages, and tip amounts.

**Architecture:** Same per-lesson pattern as Plan D. Six tasks: scaffold, warm-up, main, slides, README, end-to-end verify. Pure functions in `package exercises` and `package solutions`; pre-written failing tests; heavy-explanatory slide deck (4 concepts); README that mirrors the deck including common-mistake examples per concept (a Plan D lessons-learned).

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `testing`). No third-party dependencies. Reveal.js 5.1.0 for the deck.

---

## Scope

Plan E produces lesson 02 only. After Plan E lands:

- `lessons/02-variables/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make slides-dev LESSON=02-variables` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):** Lessons 03-08 (Plans F-K).

### Note on a spec deviation

The Phase 1 design spec's lesson 02 main says: *"Given three hardcoded amounts, compute total / average / max-min using arithmetic and print a small summary table."* This plan replaces `max-min` with a 14% tip computation, producing `Total + Average + Tip-included Summary` instead.

**Why:** Computing min/max of three values requires `if` statements (or a `math.Min`/`math.Max` import — both out of scope for lesson 02). The spec's "using arithmetic" wording suggests no control flow, and lesson 03 (Plan F) introduces `if`/`for` formally. The 14% tip variant exercises the same skills (arithmetic, multi-line `Sprintf`) plus the lesson's `const` concept (`TipRate`), and foreshadows the capstone better.

The spec already places `MinMax(xs ...int) (int, int, error)` in lesson 04 (Functions & first tests) where it has a more natural home alongside variadics and error returns. So min/max isn't lost — it just lands one lesson later.

If you'd rather keep the spec literal and accept an early `if`-statement preview in lesson 02, that's a one-paragraph swap during plan review.

---

## Plan D lessons-learned applied here

Issues surfaced in Plan D's final review are baked into Plan E from the start:

1. **Lint step in every Go-touching task** — explicit `golangci-lint run ./...` and `gofmt -l ./lessons/...` checks before commit on tasks 2, 3.
2. **Common-mistake content in README** — every concept's README section includes the same common-mistake example shown in the slides.
3. **`→` arrow consistency** — all exercise/solution doc-comment example arrows use `→` (Unicode), not `=>`. Includes the explicit gofmt step so post-edit reformatting doesn't drift.
4. **`gofmt -w .` mention** — already established in lesson 01's slides and README; lesson 02's "How to run" continues the habit with a one-liner.

---

## File Structure

After Plan E:

```
lessons/02-variables/                       (new — Plan E's deliverable)
├── README.md                               (Task 5)
├── slides/
│   ├── index.html                          (Task 1; unchanged)
│   ├── slides.md                           (Task 4)
│   └── assets/.gitkeep                     (Task 1; unchanged)
├── exercises/
│   ├── warmup.go                           (Task 2 — WarmupZeroValues + WarmupConvert)
│   ├── warmup_test.go                      (Task 2 — failing tests)
│   ├── main.go                             (Task 3 — Total + Average + Summary + TipRate const)
│   └── main_test.go                        (Task 3 — failing tests)
└── solutions/
    ├── warmup.go                           (Task 2 — implementations)
    ├── warmup_test.go                      (Task 2 — same tests)
    ├── main.go                             (Task 3 — implementations)
    └── main_test.go                        (Task 3 — same tests)
```

### Decomposition rationale

Same pattern as Plan D. Each task has one focused output and a clear verification step. Tasks 2 and 3 both end with explicit lint checks (Plan D's Task 2 missed lint and we paid for it later — fixed here).

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-e-lesson-02` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)

---

## Task 1: Scaffold lesson 02

**Files:**
- Create: `lessons/02-variables/` (12 files via the scaffolder)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --oneline main..HEAD
```

Expected: `## feature/plan-e-lesson-02`, clean working tree, exactly one commit ahead of `main` ("Add Plan E: Lesson 02 …"). If you see something else, STOP and report — do not try to recreate the branch.

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=02-variables
```

Expected: prints `created lesson 02-variables under lessons/`. The scaffolder produces 12 files under `lessons/02-variables/` with template placeholder content (TODOs, `WarmupGreet`, `Greet`).

- [ ] **Step 3: Verify the 12-file tree exists**

```bash
find lessons/02-variables -type f | sort
```

Expected (12 lines):

```
lessons/02-variables/README.md
lessons/02-variables/exercises/main.go
lessons/02-variables/exercises/main_test.go
lessons/02-variables/exercises/warmup.go
lessons/02-variables/exercises/warmup_test.go
lessons/02-variables/slides/assets/.gitkeep
lessons/02-variables/slides/index.html
lessons/02-variables/slides/slides.md
lessons/02-variables/solutions/main.go
lessons/02-variables/solutions/main_test.go
lessons/02-variables/solutions/warmup.go
lessons/02-variables/solutions/warmup_test.go
```

- [ ] **Step 4: Commit the scaffolded skeleton**

```bash
git add lessons/02-variables/
git commit -m "feat(lessons): scaffold lesson 02-variables skeleton"
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: every package passes (excluding exercises). The lesson 02 solutions pass because the scaffolded solution implements `WarmupGreet` and `Greet` already.

- [ ] **Step 6: Sanity check git status**

```bash
git status
git log --oneline main..HEAD
```

Expected: clean working tree, two commits on the branch (the Plan E document + the scaffold).

---

## Task 2: Author the warm-up exercise

Replace the four warm-up files with lesson 02's content: `WarmupZeroValues` returns the zero values of int/float64/string/bool; `WarmupConvert` performs simple type conversions.

**Files:**
- Replace: `lessons/02-variables/exercises/warmup.go`
- Replace: `lessons/02-variables/exercises/warmup_test.go`
- Replace: `lessons/02-variables/solutions/warmup.go`
- Replace: `lessons/02-variables/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/02-variables/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 02: Variables, types, operators.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "declare a variable" and "convert between types." Make the
// failing tests in warmup_test.go pass.
package exercises

// WarmupZeroValues returns the zero values of int, float64, string, and bool
// in that order.
//
// In Go, every variable has a value. If you declare it without assigning, the
// variable holds the type's "zero value." This function demonstrates the four
// most common ones.
//
// Examples:
//   WarmupZeroValues() → (0, 0.0, "", false)
//
// Hint: declare four variables with `var name type` and return them. You don't
// need to assign anything — the zero values are already there.
func WarmupZeroValues() (int, float64, string, bool) {
	panic("TODO: declare an int, a float64, a string, and a bool with var, then return them")
}

// WarmupConvert performs two simple numeric type conversions and returns both.
//
// Given an int and a float64:
// - return the int as a float64 using float64(intVal)
// - return the float64 as an int using int(floatVal); the fractional part is truncated toward zero
//
// Examples:
//   WarmupConvert(5, 3.7)   → (5.0, 3)
//   WarmupConvert(0, -2.9)  → (0.0, -2)
//
// Hint: Go is strict about types. To "promote" an int to float64 you write
// float64(x); to truncate a float64 to int you write int(y).
func WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int) {
	panic("TODO: return float64(intVal) and int(floatVal)")
}
```

- [ ] **Step 2: Replace `lessons/02-variables/exercises/warmup_test.go`** with:

```go
package exercises

import "testing"

func TestWarmupZeroValues(t *testing.T) {
	i, f, s, b := WarmupZeroValues()
	if i != 0 {
		t.Errorf("int zero value: got %d, want 0", i)
	}
	if f != 0.0 {
		t.Errorf("float64 zero value: got %g, want 0", f)
	}
	if s != "" {
		t.Errorf("string zero value: got %q, want %q", s, "")
	}
	if b != false {
		t.Errorf("bool zero value: got %v, want false", b)
	}
}

func TestWarmupConvert(t *testing.T) {
	cases := []struct {
		name      string
		intVal    int
		floatVal  float64
		wantFloat float64
		wantInt   int
	}{
		{"positive", 5, 3.7, 5.0, 3},
		{"zero", 0, 0.0, 0.0, 0},
		{"negative-truncates-toward-zero", 0, -2.9, 0.0, -2},
		{"large", 1000, 999.99, 1000.0, 999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFloat, gotInt := WarmupConvert(tc.intVal, tc.floatVal)
			if gotFloat != tc.wantFloat {
				t.Errorf("WarmupConvert(%d, %g) float: got %g, want %g",
					tc.intVal, tc.floatVal, gotFloat, tc.wantFloat)
			}
			if gotInt != tc.wantInt {
				t.Errorf("WarmupConvert(%d, %g) int: got %d, want %d",
					tc.intVal, tc.floatVal, gotInt, tc.wantInt)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/02-variables/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 02: Variables, types, operators.
//
// This file holds the warm-up reference solution.
package solutions

// WarmupZeroValues returns the zero values of int, float64, string, and bool.
func WarmupZeroValues() (int, float64, string, bool) {
	var i int
	var f float64
	var s string
	var b bool
	return i, f, s, b
}

// WarmupConvert performs simple numeric type conversions.
func WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int) {
	return float64(intVal), int(floatVal)
}
```

- [ ] **Step 4: Replace `lessons/02-variables/solutions/warmup_test.go`** with the same content as `exercises/warmup_test.go`, but `package solutions`:

```go
package solutions

import "testing"

func TestWarmupZeroValues(t *testing.T) {
	i, f, s, b := WarmupZeroValues()
	if i != 0 {
		t.Errorf("int zero value: got %d, want 0", i)
	}
	if f != 0.0 {
		t.Errorf("float64 zero value: got %g, want 0", f)
	}
	if s != "" {
		t.Errorf("string zero value: got %q, want %q", s, "")
	}
	if b != false {
		t.Errorf("bool zero value: got %v, want false", b)
	}
}

func TestWarmupConvert(t *testing.T) {
	cases := []struct {
		name      string
		intVal    int
		floatVal  float64
		wantFloat float64
		wantInt   int
	}{
		{"positive", 5, 3.7, 5.0, 3},
		{"zero", 0, 0.0, 0.0, 0},
		{"negative-truncates-toward-zero", 0, -2.9, 0.0, -2},
		{"large", 1000, 999.99, 1000.0, 999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFloat, gotInt := WarmupConvert(tc.intVal, tc.floatVal)
			if gotFloat != tc.wantFloat {
				t.Errorf("WarmupConvert(%d, %g) float: got %g, want %g",
					tc.intVal, tc.floatVal, gotFloat, tc.wantFloat)
			}
			if gotInt != tc.wantInt {
				t.Errorf("WarmupConvert(%d, %g) int: got %d, want %d",
					tc.intVal, tc.floatVal, gotInt, tc.wantInt)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt to catch any indentation drift**

```bash
gofmt -l lessons/02-variables/
```

Expected: no output (no files need reformatting). If anything is listed, run `gofmt -w` on those files and re-check.

- [ ] **Step 6: Verify the warm-up exercise tests fail by design**

```bash
go test ./lessons/02-variables/exercises/... -run Warmup -v 2>&1 | tail -20
```

Expected: both `TestWarmupZeroValues` and `TestWarmupConvert` fail with panic messages mentioning the TODOs.

- [ ] **Step 7: Verify the warm-up solution tests pass**

```bash
go test ./lessons/02-variables/solutions/... -run Warmup -v
```

Expected: both `TestWarmupZeroValues` and `TestWarmupConvert` (with all 4 sub-tests) PASS.

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: all packages green; lint reports `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/02-variables/exercises/warmup.go lessons/02-variables/exercises/warmup_test.go \
        lessons/02-variables/solutions/warmup.go lessons/02-variables/solutions/warmup_test.go
git commit -m "feat(lesson-02): warm-up — WarmupZeroValues + WarmupConvert"
```

---

## Task 3: Author the main exercise

Replace the four main-exercise files with lesson 02's content: `Total`, `Average`, `Summary` plus a `TipRate` constant. The exercise computes a multi-line summary including a 14% tip — exercising arithmetic, formatting, and the `const` keyword.

**Files:**
- Replace: `lessons/02-variables/exercises/main.go`
- Replace: `lessons/02-variables/exercises/main_test.go`
- Replace: `lessons/02-variables/solutions/main.go`
- Replace: `lessons/02-variables/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/02-variables/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 02: Variables, types, operators.
//
// This file holds the MAIN exercise. Implement Total, Average, and Summary
// so the failing tests in main_test.go pass.
package exercises

// TipRate is the tip percentage applied by Summary.
//
// Constants in Go declare values that never change. Compare with `var` for
// mutable variables.
const TipRate = 0.14

// Total returns the sum of three amounts.
//
// Examples:
//   Total(4.50, 12.00, 23.50) → 40.0
//   Total(0, 0, 0)            → 0.0
func Total(a, b, c float64) float64 {
	panic("TODO: return a + b + c")
}

// Average returns the average of three amounts.
//
// Examples:
//   Average(10, 20, 30)         → 20.0
//   Average(4.50, 12.00, 23.50) → 13.333… (exact value depends on float arithmetic)
//
// Hint: divide by 3.0 (a float64), not 3 (an int), to keep float64 arithmetic.
func Average(a, b, c float64) float64 {
	panic("TODO: return (a + b + c) / 3.0")
}

// Summary returns a multi-line summary of three amounts including a 14% tip.
//
// The format is:
//   3 expenses
//   Total: €T.TT
//   Average: €A.AA
//   Tip (14%): €P.PP
//   Total with tip: €X.XX
//
// where each amount is printed with two decimal places.
//
// Examples:
//   Summary(4.50, 12.00, 23.50) →
//     "3 expenses\nTotal: €40.00\nAverage: €13.33\nTip (14%): €5.60\nTotal with tip: €45.60"
//
// Hint: use Total and Average above, plus the TipRate constant. Format with
// fmt.Sprintf using %.2f for each amount; separate lines with \n.
func Summary(a, b, c float64) string {
	panic("TODO: implement Summary using Total, Average, and TipRate")
}
```

- [ ] **Step 2: Replace `lessons/02-variables/exercises/main_test.go`** with:

```go
package exercises

import "testing"

func TestTipRate(t *testing.T) {
	if TipRate != 0.14 {
		t.Errorf("TipRate = %g, want 0.14", TipRate)
	}
}

func TestTotal(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"three-expenses", 4.50, 12.00, 23.50, 40.0},
		{"zero", 0, 0, 0, 0},
		{"single-large", 1000.0, 0, 0, 1000.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Total(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Total(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestAverage(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"clean-divide", 10.0, 20.0, 30.0, 20.0},
		{"zero", 0, 0, 0, 0},
		{"three-expenses-avg", 4.50, 12.00, 23.50, 40.0 / 3.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Average(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Average(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	got := Summary(4.50, 12.00, 23.50)
	want := "3 expenses\nTotal: €40.00\nAverage: €13.33\nTip (14%): €5.60\nTotal with tip: €45.60"
	if got != want {
		t.Errorf("Summary(4.50, 12.00, 23.50) =\n%q\nwant\n%q", got, want)
	}
}
```

- [ ] **Step 3: Replace `lessons/02-variables/solutions/main.go`** with:

```go
// Package solutions is the reference implementation for lesson 02: Variables, types, operators.
package solutions

import "fmt"

// TipRate is the tip percentage applied by Summary.
const TipRate = 0.14

// Total returns the sum of three amounts.
func Total(a, b, c float64) float64 {
	return a + b + c
}

// Average returns the average of three amounts.
func Average(a, b, c float64) float64 {
	return (a + b + c) / 3.0
}

// Summary returns a multi-line summary of three amounts including a 14% tip.
func Summary(a, b, c float64) string {
	total := Total(a, b, c)
	avg := Average(a, b, c)
	tip := total * TipRate
	totalWithTip := total + tip
	return fmt.Sprintf(
		"3 expenses\nTotal: €%.2f\nAverage: €%.2f\nTip (14%%): €%.2f\nTotal with tip: €%.2f",
		total, avg, tip, totalWithTip,
	)
}
```

- [ ] **Step 4: Replace `lessons/02-variables/solutions/main_test.go`** with the same content as `exercises/main_test.go`, but `package solutions`:

```go
package solutions

import "testing"

func TestTipRate(t *testing.T) {
	if TipRate != 0.14 {
		t.Errorf("TipRate = %g, want 0.14", TipRate)
	}
}

func TestTotal(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"three-expenses", 4.50, 12.00, 23.50, 40.0},
		{"zero", 0, 0, 0, 0},
		{"single-large", 1000.0, 0, 0, 1000.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Total(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Total(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestAverage(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"clean-divide", 10.0, 20.0, 30.0, 20.0},
		{"zero", 0, 0, 0, 0},
		{"three-expenses-avg", 4.50, 12.00, 23.50, 40.0 / 3.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Average(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Average(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	got := Summary(4.50, 12.00, 23.50)
	want := "3 expenses\nTotal: €40.00\nAverage: €13.33\nTip (14%): €5.60\nTotal with tip: €45.60"
	if got != want {
		t.Errorf("Summary(4.50, 12.00, 23.50) =\n%q\nwant\n%q", got, want)
	}
}
```

- [ ] **Step 5: Run gofmt**

```bash
gofmt -l lessons/02-variables/
```

Expected: no output. If any files are listed, `gofmt -w` them and re-check.

- [ ] **Step 6: Verify the main exercise tests fail by design**

```bash
go test ./lessons/02-variables/exercises/... -v 2>&1 | tail -30
```

Expected: `TestTipRate` PASSes (the constant is set), `TestTotal`/`TestAverage`/`TestSummary` FAIL with panic messages. Warm-up tests also still fail (from Task 2's stubs).

- [ ] **Step 7: Verify the main solution tests pass**

```bash
go test ./lessons/02-variables/solutions/... -v
```

Expected: every test PASSes — `TestTipRate`, `TestTotal` (3 sub-tests), `TestAverage` (3 sub-tests), `TestSummary`, `TestWarmupZeroValues`, `TestWarmupConvert` (4 sub-tests).

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: all packages green; lint reports `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/02-variables/exercises/main.go lessons/02-variables/exercises/main_test.go \
        lessons/02-variables/solutions/main.go lessons/02-variables/solutions/main_test.go
git commit -m "feat(lesson-02): main — Total + Average + Summary + TipRate"
```

---

## Task 4: Author the slide deck

Replace `lessons/02-variables/slides/slides.md` with the lesson 02 deck. Four concept blocks (Variables and constants → Basic types and zero values → Operators and arithmetic → Type conversions), each with the heavy-explanatory pattern.

**Files:**
- Replace: `lessons/02-variables/slides/slides.md`

- [ ] **Step 1: Replace `lessons/02-variables/slides/slides.md`** with the markdown below.

> CRITICAL: The block below is wrapped in four-backtick fences as a documentation device. When you write the actual file, use only three-backtick fences inside. The file should start with `## Lesson 02` (no leading fence).

````markdown
## Lesson 02

# Variables, types, operators

Learning goal: declare typed variables in Go's three forms, recognise basic types and their zero values, do arithmetic with operators, and convert between numeric types.

---

## What we'll cover

- Three ways to declare a variable: `var x int = 5`, `var x = 5`, `x := 5`.
- Basic types — `int`, `float64`, `string`, `bool` — and their zero values.
- Arithmetic operators (`+`, `-`, `*`, `/`, `%`) and a tip-calculation example.
- Explicit type conversions (`float64(x)`, `int(y)`) and why Go is strict.
- Constants with `const`.

---

## Concept 1: Variables and constants

### Motivation

Variables hold values. We need a way to introduce a name, optionally give it a type, and either assign a starting value or rely on Go's default. Go has three syntactic forms — they're equivalent, but you'll see them in different places, so it's worth knowing all three.

---

### The basics

```go
package main

import "fmt"

func main() {
	var x int = 5      // explicit: type and initial value
	var y = 5          // type inferred from the literal (int)
	z := 5             // short form: only inside functions

	const Pi = 3.14    // constant: value never changes

	fmt.Println(x, y, z, Pi)
}
```

The three variable forms are equivalent in this example. `:=` is the most common in practice — it's shorter and Go infers the type for you. The explicit `var x int = 5` form is more common at package level (outside functions) where `:=` isn't allowed.

`const` declares a value that the compiler refuses to let you change after the declaration.

---

### A worked example

The expense theme — declare an expense as three typed variables:

```go
package main

import "fmt"

func main() {
	date := "2026-05-08"
	amount := 4.50
	category := "coffee"

	fmt.Printf("%s  €%.2f  %s\n", date, amount, category)
}
```

Three short-form declarations, three different types (`string`, `float64`, `string`), all inferred from their literal initialisers. The output is the same as lesson 01's `FormatExpense` — except now the values are real variables you can compute with.

---

### Common mistake

Trying to use `:=` outside a function:

```go
package main

import "fmt"

z := 5  // wrong: := only allowed inside functions

func main() {
	fmt.Println(z)
}
```

Go complains:

```
syntax error: non-declaration statement outside function body
```

At package level, you must use `var` (or `const`). Inside a function, both forms work.

---

### Recap

- Three variable forms: `var x int = 5` (explicit), `var x = 5` (inferred), `x := 5` (short).
- `:=` only works inside functions.
- `const` declares a compile-time-fixed value.

---

## Concept 2: Basic types and zero values

### Motivation

Every variable in Go has a type. The type determines what values the variable can hold and what operations you can do with it. Even more important: every variable always has a value, even before you assign one — Go fills in the type's "zero value" so there's no such thing as an "uninitialised variable" you can read.

---

### The basics

```go
package main

import "fmt"

func main() {
	var i int       // 0
	var f float64   // 0.0
	var s string    // ""
	var b bool      // false

	fmt.Println(i, f, s, b)
}
```

Output:

```
0 0  false
```

The four most common types you'll meet first:

- `int` — whole numbers. Size depends on the platform (usually 64-bit). Zero value: `0`.
- `float64` — 64-bit floating-point numbers. Zero value: `0.0`.
- `string` — immutable sequence of bytes. Zero value: `""` (empty).
- `bool` — `true` or `false`. Zero value: `false`.

---

### A worked example

You can declare and assign in one go, or split them. Both produce the same final state:

```go
package main

import "fmt"

func main() {
	// declare-and-assign
	count := 3
	average := 13.33
	currency := "€"
	settled := true

	// or: declare-then-assign (verbose)
	var anotherCount int
	anotherCount = 3

	fmt.Println(count, average, currency, settled)
	fmt.Println("declared then assigned:", anotherCount)
}
```

The split form is occasionally useful when the assignment depends on a condition you'll write in the next few lines (lesson 03 territory).

---

### Common mistake

Assuming Go has `nil` for numeric types:

```go
var x int
if x == nil {  // wrong: int is never nil
	// unreachable
}
```

Go complains:

```
cannot convert nil to type int
```

`nil` exists in Go but only for pointers, slices, maps, channels, functions, and interfaces — not for basic types. An uninitialised `int` is `0`, not `nil`. The same goes for `float64` (0.0), `string` (""), and `bool` (false).

---

### Recap

- Four basic types you'll use first: `int`, `float64`, `string`, `bool`.
- Every variable has a value — the zero value if you didn't assign one.
- Zero values: `0`, `0.0`, `""`, `false`. No nil for basic types.

---

## Concept 3: Operators and arithmetic

### Motivation

You've seen variables. Now we want to compute new values from them. Arithmetic operators (`+`, `-`, `*`, `/`, `%`) work on numbers; comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) work on values of the same type and produce booleans; logical operators (`&&`, `||`, `!`) combine booleans.

---

### The basics

```go
package main

import "fmt"

func main() {
	a := 10
	b := 3

	fmt.Println(a + b)   // 13
	fmt.Println(a - b)   // 7
	fmt.Println(a * b)   // 30
	fmt.Println(a / b)   // 3   (integer division — fractional part dropped)
	fmt.Println(a % b)   // 1   (remainder)

	fmt.Println(a == b)  // false
	fmt.Println(a > b)   // true
}
```

`+`, `-`, `*`, `/`, `%` are the five basic arithmetic operators. `%` is the remainder operator, only valid for integers.

Note `a / b` is `3`, not `3.333…`. When both operands are integers, Go does *integer division* — the fractional part is dropped. To get a float result, convert at least one operand: `float64(a) / float64(b)` would give `3.333…`.

---

### A worked example

The expense theme — compute a 14% tip:

```go
package main

import "fmt"

func main() {
	const TipRate = 0.14

	amount := 40.00
	tip := amount * TipRate
	totalWithTip := amount + tip

	fmt.Printf("Subtotal: €%.2f\n", amount)
	fmt.Printf("Tip (14%%): €%.2f\n", tip)
	fmt.Printf("Total: €%.2f\n", totalWithTip)
}
```

Output:

```
Subtotal: €40.00
Tip (14%): €5.60
Total: €45.60
```

Notice `%%` in the format string — that's how you print a literal `%` with `Printf`.

---

### Common mistake

Integer division silently drops the fractional part:

```go
package main

import "fmt"

func main() {
	total := 7
	count := 2
	avg := total / count
	fmt.Println(avg)  // 3, not 3.5
}
```

If you wanted `3.5`, both operands need to be floats:

```go
avg := float64(total) / float64(count)  // 3.5
```

This bites everyone the first time. The compiler doesn't warn — it just gives the integer answer.

---

### Recap

- Arithmetic: `+`, `-`, `*`, `/`, `%`. `%` is remainder; only on integers.
- Integer division drops the fractional part — convert to float for fractional results.
- Comparison (`==`, `<`, …) yields a `bool`.
- Print a literal `%` in `Printf` with `%%`.

---

## Concept 4: Type conversions

### Motivation

Go is strict about types. You can't add an `int` to a `float64`, or compare a `string` to a `[]byte`, without an explicit conversion. This feels strict at first but it prevents whole categories of bugs that other languages let through. The syntax for conversion is `T(value)` — read it as "cast to T."

---

### The basics

```go
package main

import "fmt"

func main() {
	count := 3            // int
	average := 13.33      // float64

	// average + count  // this would fail to compile

	asFloat := float64(count)
	asInt := int(average)

	fmt.Println(asFloat, asInt)  // 3 13
}
```

`float64(count)` converts the int to float64 (no information lost). `int(average)` converts the float64 to int — but the fractional part is *truncated toward zero*, so `int(13.33)` is `13`, not `14`. There's no rounding.

---

### A worked example

The expense theme — say cents are stored as integers, but we want to display them as euros:

```go
package main

import "fmt"

func main() {
	amountCents := 450  // €4.50 in cents
	amountEuros := float64(amountCents) / 100.0
	fmt.Printf("€%.2f\n", amountEuros)
}
```

Output:

```
€4.50
```

Without the conversion, `amountCents / 100` would be integer division (`4`, not `4.5`). The `float64(amountCents)` promotes one side; the literal `100.0` is already a float64, so the division produces a float64.

---

### Common mistake

Truncation, not rounding, when converting float64 to int:

```go
package main

import "fmt"

func main() {
	amount := 3.9
	asInt := int(amount)
	fmt.Println(asInt)  // 3, not 4
}
```

`int(3.9)` is `3`. `int(-2.9)` is `-2` — truncates *toward zero*, not toward negative infinity. If you need rounding, you'd use `math.Round` (lesson 14 territory) — but for simple display purposes, `%.2f` formatting is usually what you actually want.

---

### Recap

- Conversion syntax: `T(value)`.
- Numeric conversions are explicit — Go won't promote types implicitly.
- `int(floatValue)` truncates toward zero — no rounding.
- Use conversions to escape integer division when you want fractional results.

---

## Practice

### Warm-up

Two small functions in `exercises/warmup.go`:

- `WarmupZeroValues()` returns the zero values of `int`, `float64`, `string`, `bool`.
- `WarmupConvert(intVal, floatVal)` returns `float64(intVal)` and `int(floatVal)`.

```bash
cd lessons/02-variables/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `TipRate` is already defined as a `const`.
- `Total(a, b, c)` and `Average(a, b, c)` are simple arithmetic.
- `Summary(a, b, c)` uses Total, Average, and TipRate to produce a multi-line string with `fmt.Sprintf` and `\n`.

```bash
cd lessons/02-variables/exercises
go test -v
```

Note:
For live: walk through the integer-division gotcha on the projector (`7/2 = 3`, not `3.5`). Also the `int(3.9) = 3` truncation. Both produce surprised faces and lock the lesson in. Skip the worked example in concept 2's "split assignment" form if running short — the exercises don't require it.

---

## What we learned

- Three variable forms: `var x int = 5`, `var x = 5`, `x := 5`. Use `:=` inside functions; `var` at package level.
- Basic types: `int`, `float64`, `string`, `bool`. Zero values: `0`, `0.0`, `""`, `false`.
- Arithmetic: `+`, `-`, `*`, `/`, `%`. Integer division drops fractions — convert to float to get fractional results.
- Conversions are explicit (`float64(x)`, `int(y)`); `int(float)` truncates toward zero.
- `const` declares values that never change.
- Run `gofmt -w .` to keep your code formatted.

---

## Up next

Lesson 03 — Control flow (`if`, `for`, `switch`).
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=02-variables &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/02-variables/slides/slides.md
curl -sS http://localhost:8000/lessons/02-variables/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/02-variables/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/02-variables/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 3: Verify markdown structure**

```bash
grep -c "^## Lesson 02" lessons/02-variables/slides/slides.md
grep -c "^# Variables, types, operators" lessons/02-variables/slides/slides.md
grep -c "^## What we learned" lessons/02-variables/slides/slides.md
grep -c "^## Up next" lessons/02-variables/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/02-variables/slides/slides.md
git commit -m "feat(lesson-02): slides — Variables, types, operators (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/02-variables/README.md` with the full lesson 02 self-study prose. Mirrors the slide narrative, includes common-mistake examples for every concept (Plan D lesson learned), and ends with the `gofmt -w .` mention.

**Files:**
- Replace: `lessons/02-variables/README.md`

- [ ] **Step 1: Replace `lessons/02-variables/README.md`** with the markdown below.

> CRITICAL: Same convention. Four-backtick wrapper is a documentation device; in the file use only three-backtick fences. The file starts with `# Lesson 02: Variables, types, operators`.

````markdown
# Lesson 02: Variables, types, operators

## Learning goals

- Declare variables in Go's three forms (`var x int = 5`, `var x = 5`, `x := 5`) and pick the right one for the context.
- Recognise the four basic types you'll see most often (`int`, `float64`, `string`, `bool`) and their zero values.
- Use arithmetic operators (`+`, `-`, `*`, `/`, `%`) and understand integer-vs-float division.
- Convert between numeric types explicitly with `T(value)` syntax.
- Declare values that never change with `const`.

## Prerequisites

- Lesson 01 (Hello, Go) — you should be comfortable running a Go file with `go run` and using `fmt.Println` / `fmt.Printf`.

## Concepts

### Variables and constants

Variables hold values. Go has three equivalent ways to declare one:

```go
var x int = 5      // explicit: type and initial value
var y = 5          // type inferred from the literal (int)
z := 5             // short form: only allowed inside functions
```

All three produce an `int` variable holding `5`. The short form `z := 5` is the most common in practice — it's compact, and Go figures out the type from the expression on the right. The explicit `var x int = 5` form is essential at package level (outside any function), where `:=` isn't allowed.

For values that should never change after declaration, use `const`:

```go
const Pi = 3.14
const TipRate = 0.14
```

The compiler will reject any attempt to assign a new value to a constant.

**Common mistake.** Trying to use `:=` outside a function:

```go
package main

import "fmt"

z := 5  // wrong: := only allowed inside functions

func main() {
	fmt.Println(z)
}
```

Go complains:

```
syntax error: non-declaration statement outside function body
```

At package level, you must use `var` (or `const`). Inside a function, both forms work.

### Basic types and zero values

Every variable in Go has a type, and the type determines what values it can hold and what operations are allowed. The four types you'll meet first:

- `int` — whole numbers. Size is platform-dependent (almost always 64-bit on modern machines). Zero value: `0`.
- `float64` — 64-bit floating-point numbers. Zero value: `0.0`.
- `string` — an immutable sequence of bytes (typically interpreted as UTF-8 text). Zero value: `""` (empty string).
- `bool` — `true` or `false`. Zero value: `false`.

Even more important than the type list: every variable always has a value. If you declare `var x int` without assigning, `x` is `0`. There's no such thing in Go as reading an uninitialised variable — the language guarantees that wouldn't happen.

```go
var i int       // 0
var f float64   // 0.0
var s string    // ""
var b bool      // false
```

**Common mistake.** Assuming Go has `nil` for numeric types:

```go
var x int
if x == nil {  // wrong: int is never nil
	// unreachable
}
```

Go complains:

```
cannot convert nil to type int
```

`nil` exists in Go but only for pointers, slices, maps, channels, functions, and interfaces — not for basic types. An uninitialised `int` is `0`, not `nil`.

### Operators and arithmetic

Arithmetic operators do what you'd expect:

```go
a := 10
b := 3

a + b   // 13
a - b   // 7
a * b   // 30
a / b   // 3   (integer division — fractional part dropped)
a % b   // 1   (remainder)
```

Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) work on values of the same type and produce a `bool`. Logical operators (`&&`, `||`, `!`) combine booleans.

The expense theme — a 14% tip:

```go
const TipRate = 0.14

amount := 40.00
tip := amount * TipRate         // 5.60
totalWithTip := amount + tip    // 45.60

fmt.Printf("Tip (14%%): €%.2f\n", tip)
```

Notice `%%` in the format string — that's how you print a literal `%` with `Printf`.

**Common mistake.** Integer division silently drops the fractional part:

```go
total := 7
count := 2
avg := total / count   // 3, not 3.5
```

If you wanted `3.5`, at least one operand needs to be a float:

```go
avg := float64(total) / float64(count)   // 3.5
```

This trips up everyone the first time. The compiler doesn't warn — it just gives the integer answer.

### Type conversions

Go is strict about types. You can't add an `int` to a `float64`, or compare a `string` to a `[]byte`, without an explicit conversion. This feels strict at first but prevents whole categories of bugs that other languages let through.

The syntax is `T(value)` — read it as "cast to T":

```go
count := 3            // int
average := 13.33      // float64

asFloat := float64(count)    // 3.0
asInt := int(average)        // 13 (fractional part truncated)
```

`float64(count)` converts the int to float64 with no loss. `int(average)` converts the float64 to int by *truncating toward zero* — `int(13.33)` is `13`, `int(-2.9)` is `-2`. There's no rounding.

A cents-to-euros example:

```go
amountCents := 450
amountEuros := float64(amountCents) / 100.0   // 4.5
```

Without the conversion, `amountCents / 100` would be integer division (`4`, not `4.5`).

**Common mistake.** Truncation, not rounding, when converting float64 to int:

```go
amount := 3.9
asInt := int(amount)
fmt.Println(asInt)   // 3, not 4
```

`int(3.9)` is `3`, not `4`. Go truncates toward zero — it doesn't round to the nearest integer. If you need rounding, you'd use `math.Round` (lesson 14 territory) — but for simple display, `%.2f` formatting is usually what you actually want.

## Exercise: warm-up

Open `exercises/warmup.go`. Two functions to implement:

- `WarmupZeroValues() (int, float64, string, bool)` — return the zero values of these four types. The simplest way is to declare four variables with `var` and return them.
- `WarmupConvert(intVal int, floatVal float64) (float64, int)` — return `float64(intVal)` and `int(floatVal)` (the float-to-int conversion truncates toward zero).

The tests cover positive, zero, negative, and large values.

## Exercise: main

Open `exercises/main.go`. The `TipRate` constant is already declared at the top of the file — use it in your `Summary`.

- `Total(a, b, c float64) float64` — return `a + b + c`.
- `Average(a, b, c float64) float64` — return `(a + b + c) / 3.0`. Note the `3.0` (not `3`) to keep float64 arithmetic.
- `Summary(a, b, c float64) string` — return a multi-line summary using `Total`, `Average`, and `TipRate`. Use `fmt.Sprintf` with `%.2f` for each amount and `\n` between lines. The expected format is:

```
3 expenses
Total: €40.00
Average: €13.33
Tip (14%): €5.60
Total with tip: €45.60
```

Notice `(14%)` in the output — that's `%%` in your format string (escape for `%`).

## How to run

```bash
cd lessons/02-variables/exercises
go test -run Warmup -v   # warm-up only
go test -v                # everything
```

After you finish (or while iterating), run `gofmt -w .` from the lesson folder to keep your code in canonical Go style. Make this a habit.

Once both exercises pass, take a look at `solutions/` to compare your code with the reference.

## Going further

### Read

- [A Tour of Go — Variables](https://go.dev/tour/basics/8) — the official tour's variable section, with playgrounds.
- [Effective Go — Constants](https://go.dev/doc/effective_go#constants) — short explanation of typed and untyped constants.
- [The Go specification — Numeric types](https://go.dev/ref/spec#Numeric_types) — exhaustive reference for Go's numeric types (skim it; you don't need to memorise it).

### Try

- **`iota` in action.** Add a constant block with `iota` to `solutions/main.go`: `const (Cents = iota; Euros; Dollars)` gives Cents=0, Euros=1, Dollars=2. Try it in a small program. No reference solution provided — `iota` is an "extra credit" topic for lesson 02.
- **Format width specifiers.** Go's `Printf` supports width specifiers: `%8.2f` pads the float to width 8, right-aligned. Modify your `Summary` to right-align the amounts in a clean column. Different from the spec — but it's a useful display trick worth knowing.
- **Integer overflow.** Try `var x int8 = 127; x = x + 1` and print `x`. The result is surprising. Look up "integer overflow" if you want to understand why.
````

- [ ] **Step 2: Verify the README has the expected structure**

```bash
grep -c "^# Lesson 02: Variables, types, operators$" lessons/02-variables/README.md
grep -c "^## Learning goals$" lessons/02-variables/README.md
grep -c "^## Prerequisites$" lessons/02-variables/README.md
grep -c "^## Concepts$" lessons/02-variables/README.md
grep -c "^## Exercise: warm-up$" lessons/02-variables/README.md
grep -c "^## Exercise: main$" lessons/02-variables/README.md
grep -c "^## How to run$" lessons/02-variables/README.md
grep -c "^## Going further$" lessons/02-variables/README.md
grep -c "^### Read$" lessons/02-variables/README.md
grep -c "^### Try$" lessons/02-variables/README.md
grep -c "^\*\*Common mistake\.\*\*" lessons/02-variables/README.md
grep -c "gofmt" lessons/02-variables/README.md
```

Expected:
- All section heading counts: `1` each.
- `Common mistake.` count: `4` (one per concept).
- `gofmt` count: ≥ 1.

- [ ] **Step 3: Commit**

```bash
git add lessons/02-variables/README.md
git commit -m "docs(lesson-02): README — Variables, types, operators self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises). The lesson 02 solutions pass; lesson 01 still passes; tools still pass.

- [ ] **Step 2: Run exercise tests — must fail by design**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 02's `TestWarmupZeroValues`, `TestWarmupConvert`, `TestTotal`, `TestAverage`, `TestSummary` all fail with panic messages. `TestTipRate` passes (the constant is declared in the exercise file as the lesson's "given"). Lesson 01's `TestGreet` also still passes; everything else fails (the exercises). Make exits 0 because `-` ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=02-variables 2>&1 | tail -30
```

Expected: exercise tests fail (ignored), solution tests pass.

- [ ] **Step 4: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 5: Slides server smoke test**

```bash
make slides-dev LESSON=02-variables &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/02-variables/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/02-variables/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 6: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/02-variables/slides/slides.md && echo OK
grep -q "Variables, types, operators" dist/index.html && echo "index lists lesson 02"
rm -rf dist
```

Expected: three success lines. The landing page should now list both lesson 01 (Hello, Go) and lesson 02.

- [ ] **Step 7: Final repository sanity check**

```bash
git status
make test
git log --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch (plan + scaffold + warm-up + main + slides + README).

This task makes no commit — it is verification only.

---

## Done definition

After Task 6:

- `lessons/02-variables/` contains 12 files with full lesson content.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make test-lesson LESSON=02-variables` runs both sides correctly.
- `make slides-dev LESSON=02-variables` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan F — Lesson 03 (Control flow).** Same pattern: `if`, `for`, `switch` plus `os.Args` / `fmt.Scanln` for stdin reading. The lesson 03 main exercise will revisit the categorisation theme (`under €10 → snack`, `€10-50 → regular`, `over €50 → splurge`) using control flow.
