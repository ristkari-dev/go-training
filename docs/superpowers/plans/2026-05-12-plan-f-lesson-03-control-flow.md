# Plan F — Lesson 03 (Control flow) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the third lesson of Phase 1 — students learn `if`/`else if`/`else`, early-return / guard-clause style, the three forms of `for`, and `switch` (both tagless and tagged). The exercise extends the running expense-tracker theme to classify amounts into `snack` / `regular` / `splurge` and tally counts across a sequence of amounts.

**Architecture:** Same per-lesson pattern as Plans D and E. Six tasks: scaffold, warm-up, main, slides, README, end-to-end verify. Pure functions in `package exercises` and `package solutions` (`Categorise`, `Tally`); pre-written failing tests; an *untested* `main()` that demos the interactive `fmt.Scanln` loop. Heavy-explanatory slide deck with **four** concept blocks (if/else → early returns → for → switch); README mirrors the deck and includes common-mistake examples per concept.

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `strconv`, `os`, `testing`). No third-party dependencies. Reveal.js 5.1.0 for the deck.

---

## Scope

Plan F produces lesson 03 only. After Plan F lands:

- `lessons/03-control-flow/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make slides-dev LESSON=03-control-flow` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):** Lessons 04-08 (Plans G-K).

### Design decisions made during planning

These are choices the planner makes within the spec's latitude. They're easy to flip during plan review if you'd rather:

1. **Pure functions for testability.** The spec's per-lesson example is "Read amounts in a loop with `fmt.Scanln`," which is hard to unit-test cleanly without `bufio`/`io.Reader` (both Phase 2 territory). Plan F resolves this by extracting two pure functions — `Categorise(amount float64) string` and `Tally(amounts []float64) (snack, regular, splurge int)` — that carry the lesson's testable surface, and leaves the `Scanln`-driven `main()` *untested* by `go test`. Students verify `main()` interactively from the terminal. The README and slide deck both call this out explicitly so it's not surprising.

2. **Four concepts in the slide deck (not three).** Spec lists "if/else, for, switch, early returns" as a concept set. Plan F splits them into four ordered blocks: (1) `if`/`else`/comparison, (2) early returns/guard clauses, (3) `for` in three forms, (4) `switch`. Same content as a three-concept reading; better paced and matches lesson 02's four-concept rhythm. The `Categorise` worked example appears twice — first as an if-else chain in concept 1, then refactored to a tagless `switch` in concept 4 (the "switch makes this nicer" payoff slide).

3. **`Tally` takes `[]float64`.** A small forward reference to slices (formally introduced in lesson 05). Phase 1 spec doesn't forbid it — lesson 02's slide deck already shows `[]byte` in the type-conversion section. The slide deck and README both note "we'll cover slices properly in lesson 5; here, treat `[]float64` as 'a sequence of floats.'" The alternative (no slice → no `Tally` function → main loop counts as it reads) leaves nothing testable beyond `Categorise`, which is too thin for a lesson-03 main exercise.

4. **`switch v.(type)` is mentioned, not shown.** Spec calls it out as "an out-of-scope reference." Plan F's switch slides include a single forward-reference bullet ("Go also supports `switch v.(type)` for runtime type checks — that requires interfaces, which we cover in lesson 10") and no executable example.

5. **Named return values appear quietly.** `Tally(...)  (snack, regular, splurge int)` uses named returns purely as documentation labels. They were already introduced quietly in lesson 02's `WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int)` signature. Lesson 04 (functions & first tests) explains them formally with `defer` and the implicit-return form.

6. **`fmt.Scanln` over `bufio.Scanner`.** Spec specifies `fmt.Scanln`. It's awkward for streaming input but works fine for the "read N then loop N times reading floats" pattern Plan F uses, and it keeps the import list to `fmt` + `os` + `strconv` (matching the spec's "new imports" list exactly).

---

## Plan E lessons-learned applied here

Issues from Plans D and E that bake into Plan F from the start:

1. **Lint step in every Go-touching task** — explicit `golangci-lint run ./...` and `gofmt -l ./lessons/...` before commit on tasks 2 and 3.
2. **Common-mistake content in README** — every concept's README section includes the same common-mistake example shown in the slides (4 total).
3. **`→` arrow consistency** — all exercise/solution doc-comment example arrows use `→` (Unicode), not `=>`. Includes the explicit gofmt step so post-edit reformatting doesn't drift the layout.
4. **`gofmt -w .` mention** — lesson 03's "How to run" continues the habit with a one-liner.
5. **New since plan E:** the lesson scaffolder now produces the title-slide-grid HTML (badge + phase label) by default. Task 4 (slides) keeps the scaffolder's title slide and only replaces the rest — the rendered title slide already has the "03 / Lesson / Phase 1 — Foundations / Control flow" structure.

---

## File Structure

After Plan F:

```
lessons/03-control-flow/                   (new — Plan F's deliverable)
├── README.md                              (Task 5)
├── slides/
│   ├── index.html                         (Task 1; unchanged from scaffold)
│   ├── slides.md                          (Task 4)
│   └── assets/.gitkeep                    (Task 1; unchanged)
├── exercises/
│   ├── warmup.go                          (Task 2 — WarmupClassify + WarmupFizzBuzz)
│   ├── warmup_test.go                     (Task 2 — failing tests)
│   ├── main.go                            (Task 3 — Categorise + Tally + main()/Scanln loop)
│   └── main_test.go                       (Task 3 — failing tests for the pure functions)
└── solutions/
    ├── warmup.go                          (Task 2 — implementations)
    ├── warmup_test.go                     (Task 2 — same tests)
    ├── main.go                            (Task 3 — implementations + working main())
    └── main_test.go                       (Task 3 — same tests as exercises/)
```

### Decomposition rationale

Same pattern as plans D and E. Each task has one focused output and a clear verification step. Tasks 2 and 3 both end with explicit lint checks. Task 3 also verifies that `main()` runs interactively via a `printf | go run` smoke test (no Go test needed — the smoke test runs in bash).

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-f-lesson-03-control-flow` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)

---

## Task 1: Scaffold lesson 03

**Files:**
- Create: `lessons/03-control-flow/` (12 files via the scaffolder)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --oneline main..HEAD
```

Expected: `## feature/plan-f-lesson-03-control-flow`, clean working tree, exactly one commit ahead of `main` ("Add Plan F: Lesson 03 …"). If you see something else, STOP and report — do not try to recreate the branch.

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=03-control-flow
```

Expected: prints `created lesson 03-control-flow under lessons/`. The scaffolder produces 12 files under `lessons/03-control-flow/` with template placeholder content (TODOs, `WarmupGreet`, `Greet`). The slide deck's title slide is already wired with the new badge+phase format.

- [ ] **Step 3: Verify the 12-file tree exists**

```bash
find lessons/03-control-flow -type f | sort
```

Expected (12 lines):

```
lessons/03-control-flow/README.md
lessons/03-control-flow/exercises/main.go
lessons/03-control-flow/exercises/main_test.go
lessons/03-control-flow/exercises/warmup.go
lessons/03-control-flow/exercises/warmup_test.go
lessons/03-control-flow/slides/assets/.gitkeep
lessons/03-control-flow/slides/index.html
lessons/03-control-flow/slides/slides.md
lessons/03-control-flow/solutions/main.go
lessons/03-control-flow/solutions/main_test.go
lessons/03-control-flow/solutions/warmup.go
lessons/03-control-flow/solutions/warmup_test.go
```

- [ ] **Step 4: Commit the scaffolded skeleton**

```bash
git add lessons/03-control-flow/
git commit -m "feat(lessons): scaffold lesson 03-control-flow skeleton"
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: every package passes (excluding exercises). The lesson 03 solutions pass because the scaffolded solution implements `WarmupGreet` and `Greet` already.

- [ ] **Step 6: Sanity check git status**

```bash
git status
git log --oneline main..HEAD
```

Expected: clean working tree, two commits on the branch (the Plan F document + the scaffold).

---

## Task 2: Author the warm-up exercise

Replace the four warm-up files with lesson 03's content: `WarmupClassify` returns `"positive"` / `"negative"` / `"zero"` for an int (exercises `if`/`else if`/`else`); `WarmupFizzBuzz(n int) []string` returns the FizzBuzz sequence for 1..n (exercises `for` in C-style + `switch`).

**Files:**
- Replace: `lessons/03-control-flow/exercises/warmup.go`
- Replace: `lessons/03-control-flow/exercises/warmup_test.go`
- Replace: `lessons/03-control-flow/solutions/warmup.go`
- Replace: `lessons/03-control-flow/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/03-control-flow/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 03: Control flow.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "branch on a value with if" and "loop a fixed number of times
// with for." Make the failing tests in warmup_test.go pass.
package exercises

import "strconv"

// WarmupClassify returns "positive", "negative", or "zero" for the given int.
//
// Examples:
//   WarmupClassify(7)   → "positive"
//   WarmupClassify(-3)  → "negative"
//   WarmupClassify(0)   → "zero"
//
// Hint: an if / else if / else chain is the most natural shape. Compare with
// 0 using >, <, and ==.
func WarmupClassify(n int) string {
	panic("TODO: return \"positive\" / \"negative\" / \"zero\" using if/else if/else")
}

// WarmupFizzBuzz returns the classic FizzBuzz sequence for the integers 1..n.
//
// For each integer i in 1..n (inclusive):
//   - if i is divisible by 15: "FizzBuzz"
//   - else if i is divisible by 3: "Fizz"
//   - else if i is divisible by 5: "Buzz"
//   - else: the integer formatted with strconv.Itoa(i)
//
// If n < 1, return an empty (non-nil) slice.
//
// Examples:
//   WarmupFizzBuzz(5)   → ["1", "2", "Fizz", "4", "Buzz"]
//   WarmupFizzBuzz(15)  → [..., "Fizz", "13", "14", "FizzBuzz"]
//   WarmupFizzBuzz(0)   → []
//
// Hint: a C-style for loop (`for i := 1; i <= n; i++`) plus an if/else if
// chain inside is the simplest implementation. Use strconv.Itoa to convert
// the int to its decimal string. Append to the result slice as you go.
func WarmupFizzBuzz(n int) []string {
	_ = strconv.Itoa // keep the import compiling until you use it
	panic("TODO: build the FizzBuzz slice with a for loop and an if/else if chain")
}
```

- [ ] **Step 2: Replace `lessons/03-control-flow/exercises/warmup_test.go`** with:

```go
package exercises

import (
	"reflect"
	"testing"
)

func TestWarmupClassify(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want string
	}{
		{"positive-small", 7, "positive"},
		{"negative-small", -3, "negative"},
		{"zero", 0, "zero"},
		{"positive-large", 1_000_000, "positive"},
		{"negative-large", -1_000_000, "negative"},
		{"positive-one", 1, "positive"},
		{"negative-one", -1, "negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WarmupClassify(tc.n); got != tc.want {
				t.Errorf("WarmupClassify(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}

func TestWarmupFizzBuzz(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want []string
	}{
		{"five", 5, []string{"1", "2", "Fizz", "4", "Buzz"}},
		{"fifteen", 15, []string{
			"1", "2", "Fizz", "4", "Buzz",
			"Fizz", "7", "8", "Fizz", "Buzz",
			"11", "Fizz", "13", "14", "FizzBuzz",
		}},
		{"one", 1, []string{"1"}},
		{"three", 3, []string{"1", "2", "Fizz"}},
		{"zero-empty-slice", 0, []string{}},
		{"negative-empty-slice", -5, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupFizzBuzz(tc.n)
			if got == nil {
				t.Fatalf("WarmupFizzBuzz(%d) returned a nil slice; expected an empty slice", tc.n)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("WarmupFizzBuzz(%d) = %v, want %v", tc.n, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/03-control-flow/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 03: Control flow.
//
// This file holds the warm-up reference solution.
package solutions

import "strconv"

// WarmupClassify returns "positive", "negative", or "zero" for the given int.
func WarmupClassify(n int) string {
	if n > 0 {
		return "positive"
	} else if n < 0 {
		return "negative"
	}
	return "zero"
}

// WarmupFizzBuzz returns the FizzBuzz sequence for the integers 1..n.
func WarmupFizzBuzz(n int) []string {
	out := []string{}
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			out = append(out, "FizzBuzz")
		case i%3 == 0:
			out = append(out, "Fizz")
		case i%5 == 0:
			out = append(out, "Buzz")
		default:
			out = append(out, strconv.Itoa(i))
		}
	}
	return out
}
```

- [ ] **Step 4: Replace `lessons/03-control-flow/solutions/warmup_test.go`** with the same content as `exercises/warmup_test.go`, but `package solutions`:

```go
package solutions

import (
	"reflect"
	"testing"
)

func TestWarmupClassify(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want string
	}{
		{"positive-small", 7, "positive"},
		{"negative-small", -3, "negative"},
		{"zero", 0, "zero"},
		{"positive-large", 1_000_000, "positive"},
		{"negative-large", -1_000_000, "negative"},
		{"positive-one", 1, "positive"},
		{"negative-one", -1, "negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WarmupClassify(tc.n); got != tc.want {
				t.Errorf("WarmupClassify(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}

func TestWarmupFizzBuzz(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want []string
	}{
		{"five", 5, []string{"1", "2", "Fizz", "4", "Buzz"}},
		{"fifteen", 15, []string{
			"1", "2", "Fizz", "4", "Buzz",
			"Fizz", "7", "8", "Fizz", "Buzz",
			"11", "Fizz", "13", "14", "FizzBuzz",
		}},
		{"one", 1, []string{"1"}},
		{"three", 3, []string{"1", "2", "Fizz"}},
		{"zero-empty-slice", 0, []string{}},
		{"negative-empty-slice", -5, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupFizzBuzz(tc.n)
			if got == nil {
				t.Fatalf("WarmupFizzBuzz(%d) returned a nil slice; expected an empty slice", tc.n)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("WarmupFizzBuzz(%d) = %v, want %v", tc.n, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Run gofmt to catch any indentation drift**

```bash
gofmt -l lessons/03-control-flow/
```

Expected: no output. If anything is listed, run `gofmt -w` on those files and re-check.

- [ ] **Step 6: Verify the warm-up exercise tests fail by design**

```bash
go test ./lessons/03-control-flow/exercises/... -run Warmup -v 2>&1 | tail -30
```

Expected: both `TestWarmupClassify` and `TestWarmupFizzBuzz` fail with panic messages mentioning the TODOs.

- [ ] **Step 7: Verify the warm-up solution tests pass**

```bash
go test ./lessons/03-control-flow/solutions/... -run Warmup -v
```

Expected: both `TestWarmupClassify` (7 sub-tests) and `TestWarmupFizzBuzz` (6 sub-tests) PASS.

- [ ] **Step 8: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: all packages green; lint reports `0 issues.`

- [ ] **Step 9: Commit**

```bash
git add lessons/03-control-flow/exercises/warmup.go lessons/03-control-flow/exercises/warmup_test.go \
        lessons/03-control-flow/solutions/warmup.go lessons/03-control-flow/solutions/warmup_test.go
git commit -m "feat(lesson-03): warm-up — WarmupClassify + WarmupFizzBuzz"
```

---

## Task 3: Author the main exercise

Replace the four main-exercise files with lesson 03's content: `Categorise(amount float64) string` (snack/regular/splurge via `switch`), `Tally(amounts []float64) (snack, regular, splurge int)` (counts via `for ... range`), and a `main()` that demos the interactive `fmt.Scanln` loop and is *not* covered by `go test`.

**Files:**
- Replace: `lessons/03-control-flow/exercises/main.go`
- Replace: `lessons/03-control-flow/exercises/main_test.go`
- Replace: `lessons/03-control-flow/solutions/main.go`
- Replace: `lessons/03-control-flow/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/03-control-flow/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 03: Control flow.
//
// This file holds the MAIN exercise. Implement Categorise and Tally so the
// failing tests in main_test.go pass. The runMain function below shows the
// interactive Scanln loop — it's *not* tested by `go test`; verify it from
// the terminal once your pure functions pass.
package exercises

import (
	"fmt"
	"os"
)

// Categorise returns "snack" for amounts under €10, "regular" for €10
// (inclusive) up to €50 (inclusive), and "splurge" for amounts above €50.
//
// Examples:
//   Categorise(4.50)  → "snack"
//   Categorise(12.00) → "regular"
//   Categorise(50.00) → "regular"   (boundary: 50 is the upper edge of regular)
//   Categorise(75.00) → "splurge"
//   Categorise(0)     → "snack"
//
// Hint: a tagless switch (`switch { case amount < 10: ... }`) reads cleanly
// here. An if/else if/else chain works just as well — try both and pick the
// shape you prefer.
func Categorise(amount float64) string {
	panic("TODO: return \"snack\" / \"regular\" / \"splurge\" by amount")
}

// Tally counts how many entries of amounts fall in each category.
//
// Returns three named ints — the snack count, the regular count, and the
// splurge count — in that order. An empty input returns (0, 0, 0).
//
// Examples:
//   Tally([]float64{4.50, 12, 75, 9.99}) → (snack=2, regular=1, splurge=1)
//   Tally([]float64{})                   → (0, 0, 0)
//
// Hint: use `for _, a := range amounts` to walk the slice. Inside, switch
// on Categorise(a) and bump the matching counter.
func Tally(amounts []float64) (snack, regular, splurge int) {
	panic("TODO: walk amounts with for ... range, classify each with Categorise, count")
}

// runMain demos the interactive flow: it reads N from stdin, then reads N
// amounts, classifies and tallies them, and prints a summary. Run it from
// `solutions/main.go`'s main() — exercises don't ship a main() of their own
// because Categorise/Tally must be implemented first.
//
// runMain is *not* covered by `go test`; verify it interactively with:
//
//   printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
//
// Expected:
//   Reading 3 amounts...
//   €4.50  → snack
//   €12.00 → regular
//   €75.00 → splurge
//   Tally: 1 snack, 1 regular, 1 splurge
func runMain() {
	var n int
	if _, err := fmt.Scanln(&n); err != nil {
		fmt.Fprintln(os.Stderr, "could not read count:", err)
		return
	}
	if n <= 0 {
		fmt.Println("nothing to do (n must be > 0)")
		return
	}
	fmt.Printf("Reading %d amounts...\n", n)

	amounts := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		var a float64
		if _, err := fmt.Scanln(&a); err != nil {
			fmt.Fprintf(os.Stderr, "could not read amount %d: %v\n", i+1, err)
			return
		}
		amounts = append(amounts, a)
		fmt.Printf("€%-6.2f → %s\n", a, Categorise(a))
	}
	snack, regular, splurge := Tally(amounts)
	fmt.Printf("Tally: %d snack, %d regular, %d splurge\n", snack, regular, splurge)
}
```

- [ ] **Step 2: Replace `lessons/03-control-flow/exercises/main_test.go`** with:

```go
package exercises

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
		{"very-large-splurge", 999.99, "splurge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Categorise(tc.amount); got != tc.want {
				t.Errorf("Categorise(%g) = %q, want %q", tc.amount, got, tc.want)
			}
		})
	}
}

func TestTally(t *testing.T) {
	cases := []struct {
		name                            string
		amounts                         []float64
		wantSnack, wantRegular, wantSpl int
	}{
		{"empty", []float64{}, 0, 0, 0},
		{"all-snack", []float64{1, 2, 9.99}, 3, 0, 0},
		{"all-regular", []float64{10, 25, 50}, 0, 3, 0},
		{"all-splurge", []float64{75, 100, 999}, 0, 0, 3},
		{"mixed", []float64{4.50, 12, 75, 9.99}, 2, 1, 1},
		{"boundary-mix", []float64{9.99, 10, 50, 50.01}, 1, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snack, regular, splurge := Tally(tc.amounts)
			if snack != tc.wantSnack || regular != tc.wantRegular || splurge != tc.wantSpl {
				t.Errorf("Tally(%v) = (%d, %d, %d), want (%d, %d, %d)",
					tc.amounts, snack, regular, splurge,
					tc.wantSnack, tc.wantRegular, tc.wantSpl)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/03-control-flow/solutions/main.go`** with:

```go
// Package main is the reference implementation for lesson 03: Control flow.
//
// Both `Categorise` and `Tally` are pure functions covered by main_test.go.
// `main()` demos the interactive Scanln loop using them — it is NOT covered
// by `go test`; verify it interactively with:
//
//   printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
package main

import (
	"fmt"
	"os"
)

// Categorise returns "snack" for amounts under €10, "regular" for €10..€50
// (inclusive on both ends), and "splurge" above €50.
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

// Tally counts how many entries of amounts fall in each category.
func Tally(amounts []float64) (snack, regular, splurge int) {
	for _, a := range amounts {
		switch Categorise(a) {
		case "snack":
			snack++
		case "regular":
			regular++
		case "splurge":
			splurge++
		}
	}
	return snack, regular, splurge
}

func main() {
	var n int
	if _, err := fmt.Scanln(&n); err != nil {
		fmt.Fprintln(os.Stderr, "could not read count:", err)
		return
	}
	if n <= 0 {
		fmt.Println("nothing to do (n must be > 0)")
		return
	}
	fmt.Printf("Reading %d amounts...\n", n)

	amounts := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		var a float64
		if _, err := fmt.Scanln(&a); err != nil {
			fmt.Fprintf(os.Stderr, "could not read amount %d: %v\n", i+1, err)
			return
		}
		amounts = append(amounts, a)
		fmt.Printf("€%-6.2f → %s\n", a, Categorise(a))
	}
	snack, regular, splurge := Tally(amounts)
	fmt.Printf("Tally: %d snack, %d regular, %d splurge\n", snack, regular, splurge)
}
```

> Note: the solutions package is `package main` so `go run ./lessons/03-control-flow/solutions` actually runs the binary. The exercises package stays as `package exercises` because the main() flow there depends on Categorise/Tally which the student hasn't implemented yet — students verify their own code by writing a tiny scratch main if they want to play with it interactively. The slide deck shows them how.

- [ ] **Step 4: Replace `lessons/03-control-flow/solutions/main_test.go`** with the same content as `exercises/main_test.go`, but `package main`:

```go
package main

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
		{"very-large-splurge", 999.99, "splurge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Categorise(tc.amount); got != tc.want {
				t.Errorf("Categorise(%g) = %q, want %q", tc.amount, got, tc.want)
			}
		})
	}
}

func TestTally(t *testing.T) {
	cases := []struct {
		name                            string
		amounts                         []float64
		wantSnack, wantRegular, wantSpl int
	}{
		{"empty", []float64{}, 0, 0, 0},
		{"all-snack", []float64{1, 2, 9.99}, 3, 0, 0},
		{"all-regular", []float64{10, 25, 50}, 0, 3, 0},
		{"all-splurge", []float64{75, 100, 999}, 0, 0, 3},
		{"mixed", []float64{4.50, 12, 75, 9.99}, 2, 1, 1},
		{"boundary-mix", []float64{9.99, 10, 50, 50.01}, 1, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snack, regular, splurge := Tally(tc.amounts)
			if snack != tc.wantSnack || regular != tc.wantRegular || splurge != tc.wantSpl {
				t.Errorf("Tally(%v) = (%d, %d, %d), want (%d, %d, %d)",
					tc.amounts, snack, regular, splurge,
					tc.wantSnack, tc.wantRegular, tc.wantSpl)
			}
		})
	}
}
```

- [ ] **Step 5: Solutions package rename — fix the build-index expectation if needed**

Lessons 01 and 02 use `package solutions`; lesson 03's solutions is `package main` because it needs an entry point. Confirm `tools/build-index` doesn't assume a `solutions` package by running:

```bash
go run ./tools/build-index -lessons lessons -shared shared/reveal -out /tmp/index-check
test -f /tmp/index-check/index.html && echo OK
grep -q "Control flow" /tmp/index-check/index.html && echo "lesson 03 listed"
rm -rf /tmp/index-check
```

Expected: `OK` and `lesson 03 listed`. The build-index walks `lessons/*/slides/`, not `solutions/`, so the package change is invisible to it.

- [ ] **Step 6: Run gofmt**

```bash
gofmt -l lessons/03-control-flow/
```

Expected: no output. If any files are listed, `gofmt -w` them and re-check.

- [ ] **Step 7: Verify the main exercise tests fail by design**

```bash
go test ./lessons/03-control-flow/exercises/... -v 2>&1 | tail -30
```

Expected: `TestCategorise` and `TestTally` FAIL with panic messages. Warm-up tests also still fail (from Task 2's stubs).

- [ ] **Step 8: Verify the main solution tests pass**

```bash
go test ./lessons/03-control-flow/solutions/... -v
```

Expected: every test PASSes — `TestCategorise` (9 sub-tests), `TestTally` (6 sub-tests), `TestWarmupClassify` (7 sub-tests), `TestWarmupFizzBuzz` (6 sub-tests).

- [ ] **Step 9: Verify the interactive `main()` works end-to-end**

```bash
printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
```

Expected output (note the spacing comes from `%-6.2f`):

```
Reading 3 amounts...
€4.50   → snack
€12.00  → regular
€75.00  → splurge
Tally: 1 snack, 1 regular, 1 splurge
```

Try the edge cases too:

```bash
printf "0\n" | go run ./lessons/03-control-flow/solutions    # nothing to do
printf "1\n50\n" | go run ./lessons/03-control-flow/solutions # 50 is regular (boundary)
```

- [ ] **Step 10: Verify `make test` and lint pass**

```bash
make test
golangci-lint run ./...
```

Expected: all packages green; lint reports `0 issues.`

- [ ] **Step 11: Commit**

```bash
git add lessons/03-control-flow/exercises/main.go lessons/03-control-flow/exercises/main_test.go \
        lessons/03-control-flow/solutions/main.go lessons/03-control-flow/solutions/main_test.go
git commit -m "feat(lesson-03): main — Categorise + Tally + interactive Scanln runner"
```

---

## Task 4: Author the slide deck

Replace `lessons/03-control-flow/slides/slides.md` with the lesson 03 deck. **Keep the scaffolder's title slide** (the badge + phase label HTML at the top) and replace everything from `## What we'll cover` onward. Four concept blocks (if/else → early returns → for → switch), each with the heavy-explanatory pattern.

**Files:**
- Replace: `lessons/03-control-flow/slides/slides.md` (entire file, including the scaffolder's title slide — easier than surgical edits)

- [ ] **Step 1: Replace `lessons/03-control-flow/slides/slides.md`** with the markdown below.

> CRITICAL: The block below is wrapped in four-backtick fences as a documentation device. When you write the actual file, use only three-backtick fences inside. The file should start with `<div class="title-slide-grid">` (no leading fence).

````markdown
<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">03</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Control flow</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Branch on values with <code>if</code>/<code>else</code>, prefer early returns to deep nesting, write the three forms of <code>for</code> loop Go provides, and pick between <code>if</code>-chains and <code>switch</code> for multi-way branches.</p>
</div>
</div>
</div>

---

## What we'll cover

- `if` / `else if` / `else` and Go's comparison and logical operators.
- Early returns / guard clauses — why deeply nested code is a smell.
- `for` in three forms: infinite, condition-only, and C-style. Plus a peek at `for ... range`.
- `switch`, both tagless (`switch { case … }`) and tagged (`switch x { case … }`). Brief note on `switch v.(type)` (Phase 2).

---

## Concept 1: `if`, `else`, and comparison

### Motivation

Programs need to behave differently depending on the values they're working with. Go's `if` is the basic branching construct — it takes a boolean expression and a body, and optionally chains with `else if` and `else`. Unlike many languages, Go does not put parentheses around the condition.

---

### The basics

```go
package main

import "fmt"

func main() {
	x := 7

	if x > 0 {
		fmt.Println("positive")
	} else if x < 0 {
		fmt.Println("negative")
	} else {
		fmt.Println("zero")
	}
}
```

Three things to notice:

- **No parentheses around the condition.** `if x > 0` not `if (x > 0)`. The braces are required, however — even for a one-line body.
- **`else if` and `else` chain on the same line as the closing brace** of the previous block. `gofmt` enforces this.
- **The condition must be a `bool`.** Go won't accept `if x { … }` for a non-bool `x`.

Comparison operators: `==`, `!=`, `<`, `>`, `<=`, `>=`. Logical operators: `&&` (and), `||` (or), `!` (not). Both groups produce `bool`.

---

### A worked example

The expense theme — a first version of the `Categorise` function with an `if`/`else if`/`else` chain:

```go
package main

import "fmt"

// Categorise — first version, using an if/else if/else chain.
func Categorise(amount float64) string {
	if amount < 10 {
		return "snack"
	} else if amount <= 50 {
		return "regular"
	}
	return "splurge"
}

func main() {
	fmt.Println(Categorise(4.50))   // snack
	fmt.Println(Categorise(12.00))  // regular
	fmt.Println(Categorise(75.00))  // splurge
}
```

We'll come back to this in concept 4 and refactor it to a tagless `switch`. Same behaviour, slightly cleaner shape.

---

### Common mistake

Forgetting that the condition must be `bool` and trying C-style truthy ints:

```go
package main

func main() {
	x := 7
	if x {           // wrong: Go won't coerce int to bool
		// ...
	}
}
```

Go complains:

```
non-boolean condition in if statement
```

Write `if x != 0 { … }` or `if x > 0 { … }` — be explicit about what "truthy" means.

---

### Recap

- `if` / `else if` / `else` — conditions are bare booleans, no parentheses, braces required.
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`. Logical: `&&`, `||`, `!`.
- Go does not coerce non-booleans to `bool` — be explicit.

---

## Concept 2: Early returns and guard clauses

### Motivation

Deeply nested `if` statements are hard to read. Go culture leans hard toward returning early — handle the bad / unusual cases first, then let the happy path flow through the bottom of the function with no indentation. The pattern is sometimes called "guard clauses."

---

### The basics

A function with one input check, written two ways:

```go
// Nested style — discouraged.
func discountNested(amount float64, member bool) float64 {
	if amount > 0 {
		if member {
			return amount * 0.9
		} else {
			return amount
		}
	} else {
		return 0
	}
}

// Early-return style — preferred.
func discountEarly(amount float64, member bool) float64 {
	if amount <= 0 {
		return 0
	}
	if !member {
		return amount
	}
	return amount * 0.9
}
```

Both functions compute the same thing. The second one is easier to read because each guard clause sits on the left margin and you can scan straight down to the happy path at the bottom.

This is *the* idiomatic Go shape for any function that has prerequisites — it gets used heavily once we introduce error handling in lesson 04 (`if err != nil { return err }`).

---

### A worked example

A small validator that rejects empty input or weird amounts:

```go
package main

import "fmt"

func describe(category string, amount float64) string {
	if category == "" {
		return "(no category)"
	}
	if amount < 0 {
		return "(negative amount)"
	}
	if amount == 0 {
		return fmt.Sprintf("%s: free", category)
	}
	return fmt.Sprintf("%s: €%.2f", category, amount)
}

func main() {
	fmt.Println(describe("coffee", 4.50))   // coffee: €4.50
	fmt.Println(describe("", 12))           // (no category)
	fmt.Println(describe("coffee", -1))     // (negative amount)
	fmt.Println(describe("free-sample", 0)) // free-sample: free
}
```

Three guards, then the happy path. No `else` anywhere — each guard ends with `return`, so the rest of the function can assume the prior guards passed.

---

### Common mistake

The "useless else" — keeping an `else` after a guard that already returns:

```go
// Style smell — flagged by some linters.
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	} else {
		return fmt.Sprintf("€%.2f", amount)
	}
}
```

The `else` adds no information — if `amount < 0` returned, control already left the function. Drop the `else`:

```go
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	}
	return fmt.Sprintf("€%.2f", amount)
}
```

Reads identically; one less indentation level.

---

### Recap

- Prefer early returns / guard clauses over deeply nested `if`.
- After a guard with `return`, drop the redundant `else`.
- This pattern is *the* shape Go's stdlib uses everywhere — get comfortable with it now; lesson 04's error handling leans on it heavily.

---

## Concept 3: `for` — Go's only loop

### Motivation

Go has *one* loop keyword: `for`. There's no `while`, no `do…while`, no `loop`. The single `for` keyword takes three syntactic forms which between them cover every iteration pattern other languages spread across multiple keywords.

---

### The basics

The three forms:

```go
package main

import "fmt"

func main() {
	// 1. Infinite for — like `while true` in other languages.
	count := 0
	for {
		if count >= 3 {
			break
		}
		fmt.Println("infinite", count)
		count++
	}

	// 2. Condition-only for — like a classical `while`.
	x := 1
	for x < 10 {
		x *= 2
	}
	fmt.Println("doubled to", x) // 16

	// 3. C-style for — init, condition, post.
	for i := 0; i < 3; i++ {
		fmt.Println("c-style", i)
	}
}
```

Each form is just `for` with a different number of clauses. The C-style form is the one you'll write most often. `break` exits the loop; `continue` skips to the next iteration.

---

### A worked example

Walking a slice with `for ... range` — a fourth form, not technically a "shape" but worth showing now since we'll use it in the main exercise:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12.00, 75.00}
	for i, a := range amounts {
		fmt.Printf("%d: €%.2f\n", i, a)
	}

	// Index-only — drop the value with _.
	for i := range amounts {
		fmt.Println(i)
	}

	// Value-only — drop the index with _.
	for _, a := range amounts {
		fmt.Println(a)
	}
}
```

`range` works on slices, arrays, strings, maps, and channels. We're using it on a slice here. We cover slices properly in lesson 05; for now, treat `[]float64` as "a sequence of float64s" you can `range` over.

---

### Common mistake

C-style for with the wrong loop bound — off by one:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12.00, 75.00}
	for i := 0; i <= len(amounts); i++ {  // wrong: <= overruns
		fmt.Println(amounts[i])
	}
}
```

Go panics at runtime:

```
panic: runtime error: index out of range [3] with length 3
```

Slices and arrays are zero-indexed: valid indices are `0` to `len-1`. The loop condition should be `i < len(amounts)`, not `<=`. (Better still: use `for i, a := range amounts` and let the compiler handle the bounds.)

---

### Recap

- One keyword (`for`), three syntactic forms: infinite, condition-only, C-style.
- `break` / `continue` work as you'd expect.
- `for ... range` walks a collection — convenient and bounds-safe.
- Always check loop bounds: indices run `0` to `len-1`.

---

## Concept 4: `switch`

### Motivation

A long `if`/`else if` chain comparing one value against many possibilities is a "switch in disguise." Go's `switch` makes that intent explicit and reads more cleanly. Go's `switch` is also slightly more powerful than C/Java's: it can switch on a tag (`switch x { case 1: ... }`) or stay tagless (`switch { case x > 0: ... }`).

Cases do **not** fall through by default — the opposite of C. You don't need `break`.

---

### The basics

Tagged switch — branch on the value of an expression:

```go
package main

import "fmt"

func main() {
	day := "Tue"
	switch day {
	case "Mon", "Tue", "Wed", "Thu", "Fri":
		fmt.Println("weekday")
	case "Sat", "Sun":
		fmt.Println("weekend")
	default:
		fmt.Println("unknown day:", day)
	}
}
```

Tagless switch — each case is an arbitrary boolean expression:

```go
package main

import "fmt"

func main() {
	amount := 42.0
	switch {
	case amount < 10:
		fmt.Println("snack")
	case amount <= 50:
		fmt.Println("regular")
	default:
		fmt.Println("splurge")
	}
}
```

Two things to remember:

- **No fall-through by default.** Once a case body runs, the switch is done. (You can opt in with `fallthrough`, but it's rare.)
- **Multiple values per case** — comma-separated, as in the weekday example.

---

### A worked example

The `Categorise` refactor: take concept 1's if-else version and turn it into a tagless switch. Same behaviour, less ceremony:

```go
// Categorise — second version, tagless switch.
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
```

This is the version you'll implement in `exercises/main.go`. It also pairs nicely with `Tally`, where the inner switch dispatches on the string the outer `Categorise` returns:

```go
for _, a := range amounts {
	switch Categorise(a) {
	case "snack":
		snack++
	case "regular":
		regular++
	case "splurge":
		splurge++
	}
}
```

---

### Common mistake

Adding a `break` because muscle memory — and being surprised it does nothing useful:

```go
switch amount {
case 0:
	fmt.Println("zero")
	break  // unnecessary — switch doesn't fall through anyway
case 1:
	fmt.Println("one")
}
```

The `break` doesn't break out of any enclosing loop — it just exits the (already exiting) switch. `gofmt` won't remove it; `golangci-lint`'s `staticcheck` will flag it.

If you genuinely want fall-through from one case into the next, use the `fallthrough` keyword on the line *before* the next case — but you almost never want this. The default no-fallthrough behaviour is what makes `switch` safe in Go.

---

### A note for later

Go has a third form — `switch v.(type)` — for runtime type checks. It needs interfaces, which we cover in lesson 10. You won't write or read one in Phase 1.

---

### Recap

- Tagged `switch x { case … }` and tagless `switch { case bool-expr }`.
- No fall-through by default (the safe default).
- Multiple values per case with commas.
- Use `switch` over an `if`/`else if` chain when you're branching on the same value/condition.

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `WarmupClassify(n int) string` — return `"positive"` / `"negative"` / `"zero"` using `if`/`else if`/`else`.
- `WarmupFizzBuzz(n int) []string` — build the FizzBuzz sequence for `1..n` with a `for` loop and an `if`/`else if` (or `switch`) chain inside. Return an empty (non-nil) slice when `n < 1`.

```bash
cd lessons/03-control-flow/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `Categorise(amount float64) string` — return `"snack"` / `"regular"` / `"splurge"` by amount. Use a tagless `switch`.
- `Tally(amounts []float64) (snack, regular, splurge int)` — walk the slice with `for ... range` and count per category.
- `runMain()` — already provided. It demos the interactive `fmt.Scanln` loop and is *not* covered by `go test`. Try it from the `solutions/` package once your code compiles:

```bash
printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
```

```bash
cd lessons/03-control-flow/exercises
go test -v
```

Note:
For live: walk through the if-vs-switch refactor in concept 4 on the projector — same code shape, side by side. The "no fall-through" point lands well as a contrast with C/Java if students have that background. Demo `runMain` interactively at the end if there's time; the off-by-one slide also pays for itself when a student inevitably writes `i <= len(s)`.

---

## What we learned

- `if`, `else if`, `else` — bare booleans (no parentheses), braces required, no implicit truthiness.
- Early returns / guard clauses keep code flat — use them; drop redundant `else`s.
- `for` is Go's only loop. Three forms cover everything; `for ... range` walks collections.
- `switch` — tagged or tagless, no fall-through by default; clearer than long if-chains for multi-way branches.
- A tagless `switch { case bool-expr }` is just an if-chain in nicer clothes — pick whichever reads better in context.

---

## Up next

Lesson 04 — Functions & first tests (multi-return values, the `error` type, table tests, `t.Errorf` vs `t.Fatalf`).
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=03-control-flow &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/03-control-flow/slides/slides.md
curl -sS http://localhost:8000/lessons/03-control-flow/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/03-control-flow/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/03-control-flow/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 3: Verify markdown structure**

```bash
grep -c '<div class="title-slide-grid">' lessons/03-control-flow/slides/slides.md
grep -c "<h1>Control flow</h1>" lessons/03-control-flow/slides/slides.md
grep -c "^## What we learned" lessons/03-control-flow/slides/slides.md
grep -c "^## Up next" lessons/03-control-flow/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/03-control-flow/slides/slides.md
git commit -m "feat(lesson-03): slides — Control flow (4 concepts: if/else, early returns, for, switch)"
```

---

## Task 5: Author the README

Replace `lessons/03-control-flow/README.md` with the full lesson 03 self-study prose. Mirrors the slide narrative, includes common-mistake examples for every concept, and ends with the `gofmt -w .` mention plus the interactive `runMain` smoke test.

**Files:**
- Replace: `lessons/03-control-flow/README.md`

- [ ] **Step 1: Replace `lessons/03-control-flow/README.md`** with the markdown below.

> CRITICAL: Same convention. Four-backtick wrapper is a documentation device; in the file use only three-backtick fences. The file starts with `# Lesson 03: Control flow`.

````markdown
# Lesson 03: Control flow

## Learning goals

- Branch on a condition with `if` / `else if` / `else`, using Go's bare-boolean (no parentheses) form.
- Prefer guard clauses / early returns to deeply nested code.
- Write the three forms of Go's single `for` loop and walk a collection with `for ... range`.
- Pick between a `switch` and an `if`/`else if` chain when branching on a single value or condition.

## Prerequisites

- Lesson 01 (Hello, Go) — `go run`, `package main`, basic `Printf`.
- Lesson 02 (Variables, types, operators) — typed declarations, comparison and arithmetic operators, type conversions.

## Concepts

### `if`, `else`, and comparison

Go's `if` takes a bare boolean condition (no parentheses) and a braced body:

```go
if x > 0 {
	fmt.Println("positive")
} else if x < 0 {
	fmt.Println("negative")
} else {
	fmt.Println("zero")
}
```

Three ground rules:

- **No parens around the condition.** `if x > 0`, not `if (x > 0)`.
- **Braces are required**, even for a one-line body.
- **The condition must be `bool`.** Go won't coerce ints, strings, or pointers to truthy / falsy. Be explicit: `if x != 0`, `if s != ""`, `if p != nil`.

Comparison operators: `==`, `!=`, `<`, `>`, `<=`, `>=`. Logical operators: `&&` (and), `||` (or), `!` (not). Both groups produce a `bool`.

**Common mistake.** Trying C-style truthy ints:

```go
x := 7
if x {           // wrong: Go won't coerce int to bool
	// ...
}
```

Go complains:

```
non-boolean condition in if statement
```

Write `if x != 0` or `if x > 0` — be explicit about what "truthy" means.

### Early returns and guard clauses

Deeply nested `if` statements are hard to read. Go's idiomatic shape returns early — handle the bad / unusual cases at the top, then let the happy path flow through the bottom of the function with no indentation.

```go
// Nested — discouraged.
func discountNested(amount float64, member bool) float64 {
	if amount > 0 {
		if member {
			return amount * 0.9
		} else {
			return amount
		}
	} else {
		return 0
	}
}

// Early return — preferred.
func discountEarly(amount float64, member bool) float64 {
	if amount <= 0 {
		return 0
	}
	if !member {
		return amount
	}
	return amount * 0.9
}
```

Both compute the same thing. The second one is easier to read because each guard sits on the left margin and you can scan straight down to the happy path at the bottom.

This pattern dominates idiomatic Go — especially around error handling, which lesson 04 introduces with `if err != nil { return err }`.

**Common mistake.** Keeping a useless `else` after a guard that already returns:

```go
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	} else {
		return fmt.Sprintf("€%.2f", amount)
	}
}
```

The `else` adds no information — if `amount < 0` returned, control already left the function. Drop it:

```go
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	}
	return fmt.Sprintf("€%.2f", amount)
}
```

Reads identically; one less indentation level.

### `for` — Go's only loop

Go has *one* loop keyword: `for`. There's no `while`, no `do…while`. The single `for` keyword takes three syntactic forms:

```go
// 1. Infinite — like `while true` elsewhere.
for {
	if done {
		break
	}
	// ...
}

// 2. Condition-only — like a classical `while`.
for x < 10 {
	x *= 2
}

// 3. C-style — init, condition, post.
for i := 0; i < n; i++ {
	// ...
}
```

`break` exits the loop; `continue` skips to the next iteration.

A fourth, very common pattern is `for ... range`, which walks a collection:

```go
amounts := []float64{4.50, 12.00, 75.00}
for i, a := range amounts {
	fmt.Printf("%d: €%.2f\n", i, a)
}
for i := range amounts {  // index-only
	fmt.Println(i)
}
for _, a := range amounts {  // value-only
	fmt.Println(a)
}
```

`range` works on slices, arrays, strings, maps, and channels. We cover slices properly in lesson 05; here, treat `[]float64` as "a sequence of float64s" you can `range` over.

**Common mistake.** Off-by-one in a C-style for loop:

```go
amounts := []float64{4.50, 12.00, 75.00}
for i := 0; i <= len(amounts); i++ {  // wrong: <= overruns
	fmt.Println(amounts[i])
}
```

Panics at runtime:

```
panic: runtime error: index out of range [3] with length 3
```

Indices are `0` to `len-1`. Use `<`, not `<=`. Better still, use `for i, a := range amounts` and let the compiler handle the bounds.

### `switch`

A long `if`/`else if` chain comparing one value against many possibilities is a "switch in disguise." Go's `switch` makes that intent explicit:

```go
// Tagged — branches on the value of a single expression.
switch day {
case "Mon", "Tue", "Wed", "Thu", "Fri":
	fmt.Println("weekday")
case "Sat", "Sun":
	fmt.Println("weekend")
default:
	fmt.Println("unknown day:", day)
}

// Tagless — each case is an arbitrary boolean expression.
switch {
case amount < 10:
	fmt.Println("snack")
case amount <= 50:
	fmt.Println("regular")
default:
	fmt.Println("splurge")
}
```

Two things to remember:

- **No fall-through by default** — once a case body runs, the switch is done. The opposite of C/Java. (You can opt in with the `fallthrough` keyword, but it's rare.)
- **Multiple values per case** — comma-separated, as in the weekday example.

The lesson's `Categorise` function uses a tagless switch — same behaviour as an if-else chain, slightly cleaner shape.

> Go also supports `switch v.(type)` for runtime type checks. It needs interfaces, which we cover in lesson 10. You won't write or read one in Phase 1.

**Common mistake.** Adding a `break` because muscle memory — and being surprised it does nothing useful:

```go
switch amount {
case 0:
	fmt.Println("zero")
	break  // unnecessary — switch doesn't fall through anyway
case 1:
	fmt.Println("one")
}
```

The `break` doesn't break out of any enclosing loop — it just exits the (already exiting) switch. `staticcheck` (run by `golangci-lint`) flags it.

## Exercise: warm-up

Open `exercises/warmup.go`. Two functions to implement:

- `WarmupClassify(n int) string` — return `"positive"` for `n > 0`, `"negative"` for `n < 0`, `"zero"` otherwise.
- `WarmupFizzBuzz(n int) []string` — return the FizzBuzz sequence for the integers 1..n. For each `i`: divisible by 15 → `"FizzBuzz"`; divisible by 3 → `"Fizz"`; divisible by 5 → `"Buzz"`; otherwise the integer formatted with `strconv.Itoa(i)`. If `n < 1`, return an empty (non-nil) slice.

The tests cover positive, negative, zero, boundary, and "n < 1" cases.

## Exercise: main

Open `exercises/main.go`. Three things:

- `Categorise(amount float64) string` — return `"snack"` for `amount < 10`, `"regular"` for `10 <= amount <= 50`, `"splurge"` for `amount > 50`. The test suite checks the boundary values (9.99, 10, 50, 50.01) explicitly.
- `Tally(amounts []float64) (snack, regular, splurge int)` — walk the slice with `for ... range`, classify each amount with `Categorise`, and bump the matching counter. Empty input returns `(0, 0, 0)`.
- `runMain()` — already provided; do not modify. It demos the interactive `fmt.Scanln` loop using `Categorise` and `Tally`. It is **not** covered by `go test` — verify it interactively from the `solutions/` package once your own code compiles (see "How to run" below).

## How to run

```bash
cd lessons/03-control-flow/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main pure functions
```

To verify the interactive `runMain` flow end to end (using the reference `solutions/` package, not your `exercises/` code — `solutions` is `package main` so it has an entry point):

```bash
printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
```

Expected output:

```
Reading 3 amounts...
€4.50   → snack
€12.00  → regular
€75.00  → splurge
Tally: 1 snack, 1 regular, 1 splurge
```

After you finish (or while iterating), run `gofmt -w .` from the lesson folder to keep your code in canonical Go style. Make this a habit.

Once both exercises pass, take a look at `solutions/` to compare your code with the reference.

## Going further

### Read

- [A Tour of Go — Flow control statements](https://go.dev/tour/flowcontrol/1) — the official tour's control-flow section, with playgrounds.
- [Effective Go — Control structures](https://go.dev/doc/effective_go#control-structures) — short explanations of Go's `if`, `for`, `switch`, and the early-return style.
- [Go FAQ — Why does Go not have the ternary operator?](https://go.dev/doc/faq#Does_Go_have_a_ternary_form) — a useful look at how Go's design favours plain `if` over more compact alternatives.

### Try

- **`break` and `continue` with labels.** Go supports labelled `break LOOP` / `continue LOOP` to break out of a *specific* enclosing loop. Look up the syntax and write a small example. No reference solution provided.
- **Switch on the result of a function call.** Refactor `Tally` so the inner switch tags the call: `switch Categorise(a) { case "snack": ... }`. (The reference solution already does this.)
- **`for` with a `range` on a string.** Try `for i, r := range "héllo"`. The variable `r` is a `rune` (int32), not a byte. Why? (Hint: UTF-8.) No reference solution.
````

- [ ] **Step 2: Verify the README has the expected structure**

```bash
grep -c "^# Lesson 03: Control flow$" lessons/03-control-flow/README.md
grep -c "^## Learning goals$" lessons/03-control-flow/README.md
grep -c "^## Prerequisites$" lessons/03-control-flow/README.md
grep -c "^## Concepts$" lessons/03-control-flow/README.md
grep -c "^## Exercise: warm-up$" lessons/03-control-flow/README.md
grep -c "^## Exercise: main$" lessons/03-control-flow/README.md
grep -c "^## How to run$" lessons/03-control-flow/README.md
grep -c "^## Going further$" lessons/03-control-flow/README.md
grep -c "^### Read$" lessons/03-control-flow/README.md
grep -c "^### Try$" lessons/03-control-flow/README.md
grep -c "^\*\*Common mistake\.\*\*" lessons/03-control-flow/README.md
grep -c "gofmt" lessons/03-control-flow/README.md
```

Expected:
- All section heading counts: `1` each.
- `Common mistake.` count: `4` (one per concept).
- `gofmt` count: ≥ 1.

- [ ] **Step 3: Commit**

```bash
git add lessons/03-control-flow/README.md
git commit -m "docs(lesson-03): README — Control flow self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises). The lesson 03 solutions pass; lessons 01-02 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — must fail by design**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 03's `TestWarmupClassify`, `TestWarmupFizzBuzz`, `TestCategorise`, `TestTally` all fail with panic messages. Make exits 0 because the `-` prefix ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=03-control-flow 2>&1 | tail -30
```

Expected: exercise tests fail (ignored), solution tests pass.

- [ ] **Step 4: Run the interactive `main()` smoke test once more**

```bash
printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
```

Expected:

```
Reading 3 amounts...
€4.50   → snack
€12.00  → regular
€75.00  → splurge
Tally: 1 snack, 1 regular, 1 splurge
```

- [ ] **Step 5: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 6: Slides server smoke test**

```bash
make slides-dev LESSON=03-control-flow &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/03-control-flow/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/03-control-flow/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 7: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/03-control-flow/slides/slides.md && echo OK
grep -q "Control flow" dist/index.html && echo "index lists lesson 03"
rm -rf dist
```

Expected: three success lines. The landing page should now list lessons 01, 02, and 03.

- [ ] **Step 8: Final repository sanity check**

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

- `lessons/03-control-flow/` contains 12 files with full lesson content.
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make test-lesson LESSON=03-control-flow` runs both sides correctly.
- `make slides-dev LESSON=03-control-flow` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- The interactive `printf | go run` smoke test produces the expected output for the documented happy path.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan G — Lesson 04 (Functions & first tests).** Same per-lesson pattern. Lesson 04 introduces multi-return values formally, named returns and `defer`, the `error` type, table tests as a structural pattern, and `t.Errorf` vs `t.Fatalf`. The main exercise extracts `Categorise` and a new `FormatExpense` from this lesson's code, and adds `Add` / `MinMax` (the previously deferred `MinMax` from lesson 02) to the warm-up. Lesson 04 is also where the **tooling thread** introduces `go vet ./...` formally.
