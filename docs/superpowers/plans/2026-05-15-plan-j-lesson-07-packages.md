# Plan J — Lesson 07 (Packages and modules) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the seventh lesson of Phase 1 — students learn to split code into Go packages, encounter subfolders for the first time (the act of creating them is the exercise), formalise the exported-vs-unexported distinction first touched in lesson 06, see Go's module-path imports, and take a brief tour of the toolchain (`gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`). The main exercise refactors lesson 06's `Expense` + methods into a proper `expense` subpackage; the warm-up does the same shape with a tiny `greet` subpackage.

**Architecture:** This lesson **breaks the per-lesson layout convention** for the first time — instead of flat `exercises/{warmup,main}.go`, lesson 07's `exercises/` and `solutions/` contain a `package main` driver + two subpackages (`greet/`, `expense/`). Six tasks, same shape as plans D-I but Task 1 also restructures the scaffolder output. Skeleton tests across both subpackages (continuing the lesson 05/06 pattern). Heavy-explanatory slide deck with **four** concept blocks (packages → imports & module paths → exported/unexported (formal) → tooling tour).

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `strings`, `testing`). Reveal.js 5.1.0 for the deck.

---

## Scope

Plan J produces lesson 07 only. After Plan J lands:

- `lessons/07-packages/` is a complete teachable lesson with all four parts populated.
- The directory layout breaks from lessons 01-06's flat shape — `exercises/` and `solutions/` each contain a top-level `main.go` driver plus `greet/` and `expense/` subpackages with their own `*.go` and `*_test.go` files.
- `make test` passes; `make test-exercises` shows lesson 07's exercise tests passing vacuously (skeleton case slices empty).
- `make slides-dev LESSON=07-packages` serves the deck.
- `dist/index.html` lists lesson 07 alongside lessons 01-06.

**Out of scope (handled by future plans):** Lesson 08 (Plan K — capstone).

### Design decisions made during planning

1. **Simplified layout (per user choice).** Phase 1 design proposed `exercises/warmup/cmd/greet/main.go` (a tiny separate binary for the warm-up). Plan J consolidates: ONE top-level `main.go` per side imports BOTH subpackages (`greet` and `expense`) and drives both. Avoids the unannounced `cmd/` sneak-peek (formally introduced in lesson 08), and the single richer `main.go` is a better demonstration of "binary composing multiple packages" than two trivial ones.

2. **Layout per side:**
   ```
   exercises/                   (and solutions/)
   ├── main.go                  package main — imports greet + expense
   ├── greet/
   │   ├── greet.go             package greet — Greet(name) function
   │   └── greet_test.go        skeleton in exercises/, full in solutions/
   └── expense/
       ├── expense.go           package expense — Expense + Format + IsHigh + TotalsByCategory
       └── expense_test.go      skeleton in exercises/, full in solutions/
   ```
   No `main_test.go` at the top level — `main.go` is a demo binary; verification comes from running `go run` and from the subpackage tests. This is a small structural change from prior lessons (which always had a `main_test.go`); the README and slides flag it explicitly so students aren't confused.

3. **No `Warmup*` prefix on subpackage symbols.** The Phase 1 spec's prefix convention exists "to avoid collisions with the main exercise's identifiers" — that risk only applies inside a single package. Subpackages have their own namespace (`greet.Greet` vs `expense.Format`), so prefixes are unnecessary noise. The warm-up's function is plain `Greet(name string) string`.

4. **`TotalsByCategory` lives inside `expense`.** Lesson 06 had it at the package level. In lesson 07's subpackage layout, putting it in the `expense` package keeps cohesion — the function operates on `[]Expense`, so it belongs with the type. The top-level `main.go` calls `expense.TotalsByCategory(es)` to demonstrate the import-and-use pattern.

5. **Working `main.go` ships in both `exercises/` and `solutions/`** — students don't write the import boilerplate themselves; they study it. The lesson teaches package mechanics (subpackages + import paths + capitalisation), not how to write a `package main` from scratch (that's covered in lesson 01 already). With the subpackages stubbed, `go run ./lessons/07-packages/exercises` compiles and panics; once subpackages are filled in, it produces real output.

6. **Skeleton tests in subpackages.** Continuing the lesson 05/06 pattern: `greet/greet_test.go` and `expense/expense_test.go` ship with empty `cases` slices and `_ = tc` placeholders in `exercises/`. Solutions ship full reference tests. Total: 4 test functions across the lesson (TestGreet + TestExpenseFormat + TestExpenseIsHigh + TestTotalsByCategory).

7. **Spec deviation: `bufio` deferred to lesson 13.** Phase 1 spec lists `bufio` and `strings` as "new imports introduced" for lesson 07. Plan J introduces `strings` (used in `main.go` for output formatting via `strings.Repeat` for a separator line) but defers `bufio` to lesson 13 (Encoding & I/O), where it has a much more natural fit alongside `os.Open` and JSON streaming. Lesson 07's slide tooling-tour mentions `bufio` as a forward reference. Reverting this is a one-paragraph swap if you'd rather use bufio.NewScanner here for stdin reading.

8. **Four concepts in the slide deck.**
   1. **Splitting code into packages** — package declaration, file → package → directory mapping, why split at all.
   2. **Imports and module paths** — the `import "..."` syntax, the long `github.com/...` paths, the relationship between module-path and directory layout.
   3. **Exported vs unexported (formal)** — the capitalisation rule (touched in lesson 06, formalised here), with API-design examples.
   4. **Tooling tour** — `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`. One slide per tool, brief.

9. **Subpackage import paths are explicit.** `import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"`. Long but honest — Go uses full module-rooted paths for everything, no relative imports. Lesson 07 explains why module paths are so verbose.

10. **`package main` for the top-level driver in BOTH `exercises/` and `solutions/`.** Same as lesson 03's solutions/ (which also has a `main()` for the Scanln runner). Lesson 07's main.go is a binary entry point; it has to be `package main`. The README and slides note this.

---

## Plans F-I lessons-learned applied here

1. **No `Warmup*` prefix needed.** Subpackages provide namespacing, so `greet.Greet()` is clean.

2. **Lint exclusion already in place.** `.golangci.yml`'s `lessons/.*/exercises/` regex covers nested subdirs (`lessons/07-packages/exercises/greet/greet.go` matches). Verify in Task 1 by running `golangci-lint` on the scaffolded state.

3. **`→` arrow consistency** — same as previous plans. Doc-comment example arrows use `→` (Unicode), gofmt may reformat blocks but arrows must survive.

4. **Common-mistake content in README** — every slide concept's common-mistake example is mirrored in the README. 4 total.

5. **Slides + README written inline by controller** — Plans G/H/I established this. Plan J continues. Tasks 1-3 can be subagent-driven; Tasks 4-5 the controller writes directly.

6. **`gofmt -w .` and `go vet ./...` mentions** — README continues both habits. The tooling-tour slide concept doubles down on these.

7. **Format-output verification.** Lesson 07's `expense.Format()` reuses lesson 06's signature byte-for-byte (`"%s  €%-7.2f %s"`). Same expected outputs; no new format-check smoke test needed.

8. **Empty case → vacuous pass.** Established in Plan H, continued in I. Plan J's exercise subpackage tests pass vacuously until students add cases. README explicitly notes this.

---

## File Structure

After Plan J (14 files total):

```
lessons/07-packages/
├── README.md                                  (Task 5)
├── slides/
│   ├── index.html                             (Task 1; unchanged from scaffold)
│   ├── slides.md                              (Task 4 — controller writes inline)
│   └── assets/.gitkeep                        (Task 1; unchanged)
├── exercises/
│   ├── main.go                                (Task 3 — package main, imports both subpackages)
│   ├── greet/
│   │   ├── greet.go                           (Task 2 — package greet, Greet function stub)
│   │   └── greet_test.go                      (Task 2 — SKELETON test)
│   └── expense/
│       ├── expense.go                         (Task 3 — package expense, Expense + methods + TotalsByCategory stubs)
│       └── expense_test.go                    (Task 3 — SKELETON tests)
└── solutions/
    ├── main.go                                (Task 3 — full reference)
    ├── greet/
    │   ├── greet.go                           (Task 2 — implementation)
    │   └── greet_test.go                      (Task 2 — full reference test)
    └── expense/
        ├── expense.go                         (Task 3 — implementation)
        └── expense_test.go                    (Task 3 — full reference tests)
```

Files **deleted** from the scaffolder output:

```
exercises/warmup.go
exercises/warmup_test.go
exercises/main_test.go
solutions/warmup.go
solutions/warmup_test.go
solutions/main_test.go
```

Net file count: 14 (same as lessons 04-06's 12 + 2 extra from the subpackage split).

### Decomposition rationale

Same six-task pattern as plans D-I, but **Task 1 also performs the structural restructure** (deleting the unwanted scaffolder files and replacing the top-level `main.go` with a minimal shell). Task 2 creates the `greet/` subpackage tree; Task 3 creates the `expense/` subpackage tree AND fleshes out the top-level `main.go` driver in both `exercises/` and `solutions/`. Tasks 4-6 are slides, README, e2e verify as usual.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-j-lesson-07-packages` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold lesson 07 + restructure

The scaffolder produces the standard 12-file layout. Lesson 07 needs a different shape: 6 of those `.go` files get deleted, and the top-level `main.go` files get replaced with `package main` shells (so the directory still compiles after restructuring; Task 3 fleshes them out).

**Files:**
- Create via scaffolder: `lessons/07-packages/` (12 files)
- Delete: `lessons/07-packages/{exercises,solutions}/{warmup.go,warmup_test.go,main_test.go}` (6 files)
- Replace: `lessons/07-packages/{exercises,solutions}/main.go` with `package main` shells (2 files)

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-j-lesson-07-packages`, clean working tree, exactly one commit ahead of `main` (the Plan J doc).

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=07-packages
```

Expected: `created lesson 07-packages under lessons/`. 12 placeholder files produced.

- [ ] **Step 3: Delete the unwanted files**

```bash
rm lessons/07-packages/exercises/warmup.go \
   lessons/07-packages/exercises/warmup_test.go \
   lessons/07-packages/exercises/main_test.go \
   lessons/07-packages/solutions/warmup.go \
   lessons/07-packages/solutions/warmup_test.go \
   lessons/07-packages/solutions/main_test.go
```

Expected: 6 files removed without errors.

- [ ] **Step 4: Replace `lessons/07-packages/exercises/main.go`** with the minimal shell:

```go
// Package main is the lesson 07 driver — it imports the greet and expense
// subpackages and prints a small report. Task 3 fleshes this out.
package main

func main() {
}
```

- [ ] **Step 5: Replace `lessons/07-packages/solutions/main.go`** with the same minimal shell:

```go
// Package main is the lesson 07 driver (solutions reference) — fleshed out in Task 3.
package main

func main() {
}
```

- [ ] **Step 6: Verify the new tree shape**

```bash
find lessons/07-packages -type f | sort
```

Expected (6 files — slides + main.go shells, no subpackages yet, no warmup files):

```
lessons/07-packages/README.md
lessons/07-packages/exercises/main.go
lessons/07-packages/slides/assets/.gitkeep
lessons/07-packages/slides/index.html
lessons/07-packages/slides/slides.md
lessons/07-packages/solutions/main.go
```

- [ ] **Step 7: Verify gofmt and basic compilation**

```bash
gofmt -l lessons/07-packages/
go build ./lessons/07-packages/exercises ./lessons/07-packages/solutions
```

Expected: gofmt empty; both directories compile (the empty main() is valid Go).

- [ ] **Step 8: Verify make test still passes**

```bash
make test
```

Expected: every package passes. Lesson 07's `exercises` and `solutions` build into binaries with empty main(); `go test` reports `[no test files]` for both, which is fine.

- [ ] **Step 9: Commit**

```bash
git add lessons/07-packages/
git commit -m "feat(lessons): scaffold lesson 07-packages with subpackage layout"
```

- [ ] **Step 10: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch (the Plan J document + the scaffold).

---

## Task 2: Author the warm-up — `greet` subpackage

Create the `greet/` subpackage in both `exercises/` and `solutions/`. One function (`Greet`) and a test. Tests escalate to skeleton in `exercises/`.

**Files:**
- Create: `lessons/07-packages/exercises/greet/greet.go`
- Create: `lessons/07-packages/exercises/greet/greet_test.go`
- Create: `lessons/07-packages/solutions/greet/greet.go`
- Create: `lessons/07-packages/solutions/greet/greet_test.go`

- [ ] **Step 1: Create the subpackage directory**

```bash
mkdir -p lessons/07-packages/exercises/greet lessons/07-packages/solutions/greet
```

- [ ] **Step 2: Create `lessons/07-packages/exercises/greet/greet.go`** with:

```go
// Package greet provides a simple greeting helper for lesson 07's
// "split code into a subpackage" warm-up.
//
// We're putting this function in its own package — separate from main —
// to demonstrate Go's subpackage mechanics. The main.go in the parent
// directory imports this with:
//
//	import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
//
// and calls greet.Greet("Aki").
package greet

// Greet returns "Hello, <name>!".
//
// The function name is exported (capital G) so callers from other packages
// can use it. If we wrote `func greet(name string)` (lower-case g), the
// import would compile but `greet.greet("Aki")` would not — the symbol
// would be invisible from outside this package.
//
// Examples:
//
//	Greet("Aki")    → "Hello, Aki!"
//	Greet("World")  → "Hello, World!"
//	Greet("")       → "Hello, !"
//
// Hint: fmt.Sprintf("Hello, %s!", name) is the one-line implementation.
func Greet(name string) string {
	panic("TODO: return the greeting per the doc comment")
}
```

- [ ] **Step 3: Create `lessons/07-packages/exercises/greet/greet_test.go`** with (SKELETON):

```go
package greet

import "testing"

// TestGreet is a SKELETON. Fill in the cases and the t.Run body.
//
// Cases to cover: a normal name, the literal "World", an empty name,
// and at least one with non-ASCII characters (e.g. "Aki" or "Étienne").
func TestGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Greet(tc.in); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}
```

- [ ] **Step 4: Create `lessons/07-packages/solutions/greet/greet.go`** with:

```go
// Package greet is the reference implementation of lesson 07's warm-up.
package greet

import "fmt"

// Greet returns "Hello, <name>!".
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
```

- [ ] **Step 5: Create `lessons/07-packages/solutions/greet/greet_test.go`** with the full reference tests:

```go
package greet

import "testing"

func TestGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"world", "World", "Hello, World!"},
		{"name", "Aki", "Hello, Aki!"},
		{"empty", "", "Hello, !"},
		{"unicode", "Étienne", "Hello, Étienne!"},
		{"with-space", "Go Course", "Hello, Go Course!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Greet(tc.in); got != tc.want {
				t.Errorf("Greet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 6: Run gofmt**

```bash
gofmt -l lessons/07-packages/
```

Expected: empty.

- [ ] **Step 7: Verify the warm-up exercise test "passes vacuously"**

```bash
go test ./lessons/07-packages/exercises/greet/... -v 2>&1 | tail -10
```

Expected: TestGreet PASS (0 sub-tests).

- [ ] **Step 8: Verify the warm-up solution test passes**

```bash
go test ./lessons/07-packages/solutions/greet/... -v 2>&1 | tail -15
```

Expected: TestGreet (5 sub-tests) PASS.

- [ ] **Step 9: `make test` and lint clean**

```bash
make test
golangci-lint run ./...
```

Expected: green, `0 issues.` (subpackages should be matched by the `lessons/.*/exercises/` exclusion pattern).

- [ ] **Step 10: Commit**

```bash
git add lessons/07-packages/exercises/greet/ lessons/07-packages/solutions/greet/
git commit -m "feat(lesson-07): warm-up — greet subpackage with skeleton test"
```

---

## Task 3: Author the main exercise — `expense` subpackage + driver `main.go`

Create the `expense/` subpackage AND flesh out the top-level `main.go` in both `exercises/` and `solutions/`. The expense package contains lesson 06's `Expense` + `Format` + `IsHigh` + `TotalsByCategory` (now living in their own package). The driver imports both `greet` and `expense` and prints a small report.

**Files:**
- Create: `lessons/07-packages/exercises/expense/expense.go`
- Create: `lessons/07-packages/exercises/expense/expense_test.go`
- Create: `lessons/07-packages/solutions/expense/expense.go`
- Create: `lessons/07-packages/solutions/expense/expense_test.go`
- Replace: `lessons/07-packages/exercises/main.go`
- Replace: `lessons/07-packages/solutions/main.go`

- [ ] **Step 1: Create the subpackage directories**

```bash
mkdir -p lessons/07-packages/exercises/expense lessons/07-packages/solutions/expense
```

- [ ] **Step 2: Create `lessons/07-packages/exercises/expense/expense.go`** with:

```go
// Package expense holds the Expense type + methods for lesson 07's main exercise.
//
// In lesson 06 these all lived at the package level alongside the test
// scaffolding. Here we extract them into their own subpackage so the
// top-level main.go imports `expense` and uses the type via its package
// qualifier (expense.Expense, expense.TotalsByCategory). That's how
// real-world Go code organises types you want to reuse across binaries.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// All three fields are exported (capitalised) because the parent package's
// main.go and the test file in this package construct Expense values with
// struct literals. If a field were lower-case, neither could set it.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04's FormatExpense
// and lesson 06's Expense.Format — now in its own package.
//
// Examples:
//
//	Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-15  €4.50    coffee"
//	Expense{Date: "2026-05-15", Amount: 12, Category: "lunch"}.Format()
//	  → "2026-05-15  €12.00   lunch"
//	Expense{Date: "2026-05-15", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-15  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Examples:
//
//	Expense{Amount: 50}.IsHigh()    → false   (boundary: 50 is NOT high)
//	Expense{Amount: 50.01}.IsHigh() → true
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// TotalsByCategory returns a map of category → summed amount, built from es.
// Empty/nil input returns a non-nil empty map.
//
// Examples:
//
//	es := []Expense{
//		{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-15", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]   (non-nil empty map)
//
// Hint: same idiom as lesson 06's TotalsByCategory — initialise totals
// as map[string]float64{}, range over es, accumulate.
func TotalsByCategory(es []Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into map[string]float64")
}
```

- [ ] **Step 3: Create `lessons/07-packages/exercises/expense/expense_test.go`** with (SKELETON):

```go
package expense

import (
	"reflect"
	"testing"
)

// TestExpenseFormat is a SKELETON. Fill in cases and the t.Run body.
// Use the doc-comment examples in expense.go as a starting point.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases. Cover small/medium/large amounts so
		// you exercise the %-7.2f width on different magnitudes.
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
		// TODO: at least 4 cases including 49.99, 50, 50.01.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Expense{Amount: tc.amount}.IsHigh()
			_ = tc
		})
	}
}

// TestTotalsByCategory is a SKELETON. Cover at least:
//   - distinct categories
//   - repeated categories (totals accumulate)
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

- [ ] **Step 4: Create `lessons/07-packages/solutions/expense/expense.go`** with the implementations:

```go
// Package expense is the reference implementation of lesson 07's main exercise.
package expense

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
// Empty/nil input returns a non-nil empty map.
func TotalsByCategory(es []Expense) map[string]float64 {
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}
	return totals
}
```

- [ ] **Step 5: Create `lessons/07-packages/solutions/expense/expense_test.go`** with the full reference tests:

```go
package expense

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
		{"coffee", Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}, "2026-05-15  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-15", Amount: 12, Category: "lunch"}, "2026-05-15  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-15", Amount: 999.99, Category: "rent"}, "2026-05-15  €999.99  rent"},
		{"zero-amount", Expense{Date: "2026-05-15", Amount: 0, Category: "free"}, "2026-05-15  €0.00    free"},
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
				{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-15", Amount: 12, Category: "lunch"},
				{Date: "2026-05-15", Amount: 75, Category: "rent"},
			},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
		},
		{
			"repeated-categories",
			[]Expense{
				{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-15", Amount: 12, Category: "lunch"},
				{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
			},
			map[string]float64{"coffee": 14.49, "lunch": 12},
		},
		{
			"single",
			[]Expense{{Date: "2026-05-15", Amount: 42, Category: "x"}},
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

- [ ] **Step 6: Replace `lessons/07-packages/exercises/main.go`** with the full driver (which will compile against the panic-stubs and panic at runtime — that's fine, students complete the subpackages first):

```go
// Package main is the lesson 07 driver. It imports the greet and expense
// subpackages and prints a small report.
//
// Run with:
//
//	go run ./lessons/07-packages/exercises
//
// Once you've implemented greet.Greet, expense.Expense.Format,
// expense.Expense.IsHigh, and expense.TotalsByCategory, the output is:
//
//	Hello, Aki!
//	================
//	2026-05-15  €4.50    coffee
//	2026-05-15  €12.00   lunch
//	2026-05-15  €75.00   rent     (high)
//	2026-05-15  €9.99    coffee
//	----------------
//	totals: coffee=14.49 lunch=12.00 rent=75.00
//
// Until then, main panics on the first stub it hits.
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
)

func main() {
	fmt.Println(greet.Greet("Aki"))
	fmt.Println(strings.Repeat("=", 16))

	es := []expense.Expense{
		{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-15", Amount: 12, Category: "lunch"},
		{Date: "2026-05-15", Amount: 75, Category: "rent"},
		{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
	}
	for _, e := range es {
		line := e.Format()
		if e.IsHigh() {
			line += "  (high)"
		}
		fmt.Println(line)
	}

	fmt.Println(strings.Repeat("-", 16))
	fmt.Print("totals:")

	totals := expense.TotalsByCategory(es)
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(" %s=%.2f", k, totals[k])
	}
	fmt.Println()
}
```

> Note: `sort.Strings` here is a small forward reference to lesson 14's `sort` package. We use it to give deterministic output (map iteration is randomised — see lesson 05). Comment-only mention; not a required learning point.

- [ ] **Step 7: Replace `lessons/07-packages/solutions/main.go`** with the same driver, but importing from the `solutions` subpackages:

```go
// Package main is the lesson 07 driver (solutions reference).
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/solutions/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/solutions/greet"
)

func main() {
	fmt.Println(greet.Greet("Aki"))
	fmt.Println(strings.Repeat("=", 16))

	es := []expense.Expense{
		{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-15", Amount: 12, Category: "lunch"},
		{Date: "2026-05-15", Amount: 75, Category: "rent"},
		{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
	}
	for _, e := range es {
		line := e.Format()
		if e.IsHigh() {
			line += "  (high)"
		}
		fmt.Println(line)
	}

	fmt.Println(strings.Repeat("-", 16))
	fmt.Print("totals:")

	totals := expense.TotalsByCategory(es)
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf(" %s=%.2f", k, totals[k])
	}
	fmt.Println()
}
```

- [ ] **Step 8: Format-output smoke test (Plan G/I lesson learned)**

The `Format()` output uses `%-7.2f` width. Verify the expected strings byte-match what `fmt.Sprintf` produces:

```bash
cat > /tmp/format-check.go <<'EOF'
package main
import "fmt"
func main() {
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-15", 4.50, "coffee"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-15", 12.00, "lunch"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-15", 999.99, "rent"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-15", 0.00, "free"))
}
EOF
go run /tmp/format-check.go
rm /tmp/format-check.go
```

Expected:

```
"2026-05-15  €4.50    coffee"
"2026-05-15  €12.00   lunch"
"2026-05-15  €999.99  rent"
"2026-05-15  €0.00    free"
```

If actual output differs from `solutions/expense/expense_test.go`'s `want` strings, fix the test cases (NOT the format).

- [ ] **Step 9: Run gofmt**

```bash
gofmt -l lessons/07-packages/
```

Expected: empty.

- [ ] **Step 10: Verify exercise tests pass vacuously**

```bash
go test ./lessons/07-packages/exercises/... -v 2>&1 | tail -20
```

Expected: TestGreet, TestExpenseFormat, TestExpenseIsHigh, TestTotalsByCategory all PASS with 0 sub-tests.

- [ ] **Step 11: Verify solution tests pass**

```bash
go test ./lessons/07-packages/solutions/... -v 2>&1 | tail -40
```

Expected: every test PASSes — TestGreet (5), TestExpenseFormat (4), TestExpenseIsHigh (6), TestTotalsByCategory (5). 20 sub-tests total.

- [ ] **Step 12: Run the solutions binary end-to-end**

```bash
go run ./lessons/07-packages/solutions
```

Expected output:

```
Hello, Aki!
================
2026-05-15  €4.50    coffee
2026-05-15  €12.00   lunch
2026-05-15  €75.00   rent     (high)
2026-05-15  €9.99    coffee
----------------
totals: coffee=14.49 lunch=12.00 rent=75.00
```

(The exact spacing in `(high)` depends on the format string — the `IsHigh` line just appends `  (high)` to whatever `Format` produced.)

- [ ] **Step 13: Verify exercises binary panics on the first stub**

```bash
go run ./lessons/07-packages/exercises 2>&1 | head -3
```

Expected: starts printing, then panics on `greet.Greet` → `panic: TODO: return the greeting per the doc comment`.

- [ ] **Step 14: Verify make test, lint, vet pass**

```bash
make test
golangci-lint run ./...
go vet ./...
```

Expected: green, `0 issues.`, clean vet.

- [ ] **Step 15: Commit**

```bash
git add lessons/07-packages/exercises/expense/ lessons/07-packages/solutions/expense/ \
        lessons/07-packages/exercises/main.go lessons/07-packages/solutions/main.go
git commit -m "feat(lesson-07): main — expense subpackage + driver main.go"
```

---

## Task 4: Author the slide deck

Replace `lessons/07-packages/slides/slides.md`. Four concepts: packages → imports & module paths → exported/unexported (formal) → tooling tour.

The controller writes this directly (Plans G-I pattern).

**Files:**
- Replace: `lessons/07-packages/slides/slides.md`

- [ ] **Step 1: Write `lessons/07-packages/slides/slides.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device. Use only three-backtick fences in the file. File starts with `<div class="title-slide-grid">`.

````markdown
<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">07</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Packages and modules</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Split your code into packages, import them via Go's full module-rooted paths, formalise the exported-vs-unexported rule, and tour the daily-use Go tools (<code>gofmt</code>, <code>go vet</code>, <code>go doc</code>, <code>go list</code>, <code>go mod tidy</code>).</p>
</div>
</div>
</div>

---

## What we'll cover

- **Packages** — the `package X` declaration, the file → package → directory mapping, why split code at all.
- **Imports and module paths** — `import "..."`, the long `github.com/...` paths, what `go.mod` does.
- **Exported vs unexported (formal)** — capitalised → public, lower-cased → private. Same rule everywhere.
- **Tooling tour** — `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`. One slide each.

---

## Concept 1: Splitting code into packages

### Motivation

Up to now we've put everything into the `exercises` (or `solutions`) package. That works for tiny lessons but doesn't scale — real Go programs split related code into focused **packages**, where each package is a single directory holding files that share a `package X` declaration. Lesson 07 takes the `Expense` type from lesson 06 and promotes it into its own `expense` package; the warm-up does the same with a tiny `greet` package.

---

### The basics

Three rules to internalise:

1. **One package per directory.** Every `.go` file in a directory must declare the same package.
2. **Package name = directory name (usually).** A directory called `greet/` contains files declaring `package greet`. There are exceptions (`main` packages can live in any directory) but the convention is uniform.
3. **The `package` line is the very first non-comment line in the file.**

```go
// File: lessons/07-packages/exercises/greet/greet.go
package greet

func Greet(name string) string {
	return "Hello, " + name + "!"
}
```

```go
// File: lessons/07-packages/exercises/expense/expense.go
package expense

type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func (e Expense) Format() string { /* ... */ }
```

```go
// File: lessons/07-packages/exercises/main.go
package main

import (
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
)

func main() {
	fmt.Println(greet.Greet("Aki"))
	e := expense.Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}
	fmt.Println(e.Format())
}
```

Three files, three packages, three directories. The `main` package lives at the root of the lesson; `greet` and `expense` are subpackages one level down. The relationship between directory layout and package name is direct: the path from the module root maps onto how you import it.

---

### A worked example

Lesson 07's full `exercises/` shape:

```
lessons/07-packages/exercises/
├── main.go                    # package main
├── greet/
│   ├── greet.go               # package greet
│   └── greet_test.go          # package greet
└── expense/
    ├── expense.go             # package expense
    └── expense_test.go        # package expense
```

`main.go` imports `greet` and `expense`. The two subpackages don't import each other — they're independent. Tests for each subpackage live alongside their code.

To run the binary: `go run ./lessons/07-packages/exercises`.
To run all tests in the lesson: `go test ./lessons/07-packages/exercises/...` (the `...` means "this directory and all subdirectories").

---

### Common mistake

Putting two different `package` declarations in the same directory:

```go
// File: greet/greet.go
package greet

// File: greet/hello.go
package hello   // wrong: must match the rest of the directory
```

Go complains:

```
found packages greet (greet.go) and hello (hello.go) in greet
```

Fix: make every `.go` file in `greet/` start with `package greet`. One directory, one package — no exceptions.

---

### Recap

- One package per directory; every file in a directory shares its `package X` declaration.
- Package name conventionally matches the directory name.
- A `main` package is a binary entry point; non-main packages are libraries.

---

## Concept 2: Imports and module paths

### Motivation

To use code from another package, you `import` it. Go's import paths are module-rooted strings — they look long, but that's because they're unambiguous: every package in the world has exactly one canonical import path, and Go's tooling never has to guess which "utils" package you meant.

---

### The basics

Import the standard library by short name:

```go
import "fmt"
import "strings"
import "errors"
```

Import your own subpackages by full module path:

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

Group multiple imports in parentheses (`gofmt` enforces this):

```go
import (
	"fmt"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
)
```

The blank line between stdlib imports and module-internal imports is convention — `gofmt` tolerates either form, but `goimports` enforces the split.

After import, refer to exported names through the package qualifier:

```go
greet.Greet("Aki")             // last segment of import path is the qualifier
expense.Expense{Date: "..."}
expense.TotalsByCategory(es)
```

The qualifier is the **last segment of the import path**, which (by convention) matches the package's `package X` declaration. So `import ".../greet"` gives you `greet.Greet` access.

---

### A worked example

What `go.mod` does — Go's module file at the repo root says:

```
module github.com/ristkari-dev/go-training

go 1.23
```

The `module` line declares the **import path prefix** for everything in this module. So when you write `import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"`, Go takes everything after the module-path prefix (`lessons/07-packages/exercises/greet`) as a directory path relative to the repo root, and finds the matching package there.

That's why imports look long but unambiguous: the module path is a globally unique prefix (typically a Git URL), and the rest is the repo's directory layout.

```bash
$ go list -m
github.com/ristkari-dev/go-training

$ go list ./lessons/07-packages/exercises/...
github.com/ristkari-dev/go-training/lessons/07-packages/exercises
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet
```

---

### Common mistake

Confusing relative paths (which Go does NOT support) with module paths:

```go
// Wrong — Go doesn't have relative imports.
import "./expense"
import "../shared/util"
```

Go complains:

```
"./expense" is relative, but relative import paths are not supported in module mode
```

Fix: always use the full module-rooted path, even for code in your own repo:

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

This sometimes feels verbose for a one-line change, but it's the same rule everywhere — there are no special cases for "stuff in the same module."

---

### Recap

- Imports are full module-rooted strings — even for your own subpackages.
- The package qualifier is the last segment of the import path: `import ".../greet"` → `greet.Greet`.
- `go.mod`'s `module` line declares the import-path prefix for everything in the module.
- No relative imports. Ever.

---

## Concept 3: Exported vs unexported (formal)

### Motivation

Lesson 06 introduced this rule lightly. Now we formalise it: **a name is exported if its first letter is uppercase, unexported otherwise**. The rule is uniform — fields, methods, functions, types, constants, variables, even `init` functions all follow it. Exported = visible from other packages; unexported = visible only within the declaring package.

---

### The basics

```go
package billing

// Exported — callers from other packages can use it.
type Invoice struct {
	Number int       // exported
	Amount float64   // exported
	notes  string    // unexported — field is package-private
}

// Exported method on Invoice.
func (i Invoice) Display() string {
	return fmt.Sprintf("Invoice #%d: €%.2f", i.Number, i.Amount)
}

// Unexported function — internal helper.
func formatNotes(s string) string {
	return strings.TrimSpace(s)
}

// Exported — the constructor is the API; the field is hidden.
func NewInvoice(number int, amount float64, notes string) Invoice {
	return Invoice{
		Number: number,
		Amount: amount,
		notes:  formatNotes(notes),
	}
}
```

From another package:

```go
import "github.com/example/myapp/billing"

inv := billing.NewInvoice(1, 12.50, "  groceries  ")  // works
inv.Display()                                          // works — exported method
inv.Number = 2                                         // works — exported field
inv.notes = "..."                                      // ERROR: unexported field
```

The capitalisation rule applies to:

- **Functions** — `Greet` exported, `greet` unexported (collides with the package name here, but that's fine).
- **Types** — `Expense` exported, `expense` unexported.
- **Methods** — `Format()` exported, `format()` unexported.
- **Fields** — `Date string` exported, `date string` unexported.
- **Constants and variables** — `MaxRetries = 3` exported, `maxRetries = 3` unexported.

This is the **only** access-control mechanism in Go. There's no `public` / `private` / `protected` keyword.

---

### A worked example

Compare two API designs for the `expense` package:

```go
// API design A — fields exported, no constructor.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Caller does:
//   e := expense.Expense{Date: "...", Amount: 4.50, Category: "coffee"}
```

```go
// API design B — fields unexported, constructor required.
type Expense struct {
	date     string
	amount   float64
	category string
}

func NewExpense(date string, amount float64, category string) (Expense, error) {
	// ... validate, normalise, etc ...
	return Expense{date: date, amount: amount, category: category}, nil
}

// Caller does:
//   e, err := expense.NewExpense("...", 4.50, "coffee")
```

Design A is what the lesson uses — fields exported, no validation, anyone can construct. Simple and direct; the caller has no guarantee that the fields are well-formed.

Design B is what real-world Go often does for types with invariants — the package controls construction, validates inputs, and exposes the data through methods. The fields stay hidden; the package's API is what the caller sees.

Phase 1 sticks with design A because we need our tests to construct values via struct literals. Production code often goes with B.

---

### Common mistake

Trying to access an unexported field from another package and getting a confusing-looking error:

```go
// In package mypkg:
type Counter struct {
	count int       // unexported
}

// In another package:
c := mypkg.Counter{count: 5}
```

Go says:

```
unknown field count in struct literal of type mypkg.Counter
```

The error says "unknown field" — but the field exists; it's just invisible from outside `mypkg`. Capitalise it (`Count`), or expose it through a method (`func (c Counter) Count() int { return c.count }`).

---

### Recap

- Capitalised first letter → exported (visible from other packages).
- Lower-case first letter → unexported (package-private).
- Same rule for fields, methods, functions, types, constants, variables.
- Go has no other access modifiers.

---

## Concept 4: Tooling tour

### Motivation

Go ships with a small set of tools you'll run every day. We've already met `go run`, `go build`, `go test`, `gofmt`, and `go vet`. Lesson 07 rounds out the daily-use set with `go doc`, `go list`, and `go mod tidy`, and reinforces the formatters/linters as part of every commit.

---

### The basics

**`gofmt`** — Go's canonical formatter. There's exactly one correct style and `gofmt` enforces it.

```bash
gofmt -l lessons/07-packages/    # list files that would be reformatted
gofmt -w lessons/07-packages/    # rewrite them in place
```

CI runs `gofmt -l` as a check. If the output is non-empty, the build fails.

**`go vet`** — static analyser that catches subtle bugs `go build` won't:

```bash
go vet ./...
```

Catches: wrong format verbs in `Printf` (`%d` with a string), shadowed variables, unreachable code, unused locks, missing struct field tags, and more. Run it locally; CI does too.

**`go doc`** — read the documentation for a package or symbol from the command line:

```bash
go doc fmt                        # the entire fmt package
go doc fmt.Sprintf                # one function
go doc github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense.Expense
```

The doc comments you write on exported names are what `go doc` shows. `go doc -all` includes unexported.

**`go list`** — enumerate packages or modules:

```bash
go list ./...                              # all packages in this module
go list ./lessons/07-packages/...          # packages under one directory
go list -m                                 # the current module's path
```

Useful in scripts (e.g., the project's `make test` does `go list ./... | grep -v '/exercises$'`).

**`go mod tidy`** — synchronise `go.mod` and `go.sum` with the imports actually used:

```bash
go mod tidy
```

After adding or removing imports, run this. It removes unused module dependencies and adds missing ones. CI in larger projects often runs `go mod tidy && git diff --exit-code go.mod` to catch un-tidied PRs.

---

### A worked example

A workflow you'll do dozens of times a day:

```bash
# Edit your code...
$ vim lessons/07-packages/exercises/expense/expense.go

# Format it.
$ gofmt -w lessons/07-packages/exercises/expense/

# Check static analysis.
$ go vet ./lessons/07-packages/...

# Run the tests.
$ go test ./lessons/07-packages/...

# All green — commit.
$ git add ...
$ git commit -m "..."
```

Make these four steps muscle memory. CI runs them all; if you run them locally first, you'll never push a broken build.

---

### Common mistake

Skipping `gofmt` and seeing a CI failure on a one-character indentation difference. The fix is to run `gofmt -w .` before every commit, and ideally to wire your editor to format on save.

A second common one: editing imports manually and getting "unused import" or "imported and not used" errors. `goimports` (a sibling of `gofmt`) auto-adjusts the import block; install it with:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

then run `goimports -w .` to add missing imports and remove unused ones.

---

### Recap

- `gofmt` — canonical formatter. CI enforces it.
- `go vet` — static analyser. Run alongside `go test`.
- `go doc` — package/symbol docs from the command line.
- `go list` — enumerate packages or the module path.
- `go mod tidy` — sync `go.mod` with what's actually imported.

---

## Practice

### Warm-up

In `exercises/greet/`:

- `greet/greet.go` — implement `Greet(name string) string` returning `"Hello, <name>!"`.
- `greet/greet_test.go` — fill in the skeleton `TestGreet` (cases + assertion body).

```bash
cd lessons/07-packages/exercises
go test ./greet/... -v
```

---

### Main

In `exercises/expense/`:

- `expense/expense.go` — implement `Expense.Format()`, `Expense.IsHigh()`, and `TotalsByCategory(es []Expense)`. Same logic as lesson 06; same byte-output for `Format`.
- `expense/expense_test.go` — fill in three skeleton tests.

The top-level `main.go` is already wired up — it imports both subpackages and prints a small report. Once the subpackages work, `go run ./lessons/07-packages/exercises` produces the expected output (see the doc comment on `main.go`).

```bash
cd lessons/07-packages/exercises
go test ./...    # warm-up + main
go run .         # the binary
```

Note:
For live: do the import-path tour live — type one of the long `import "github.com/..."` paths on the projector, then split it: "module path / repo subdirectory / package name". The realisation that "the import path is just the directory layout from the module root" is the lesson's lightbulb. Demo `go list ./lessons/07-packages/...` to show the discovered packages, and `go doc` on `expense.Expense` to show the doc comments rendering.

---

## What we learned

- One package per directory; package name conventionally matches the directory.
- Imports are full module-rooted paths (`github.com/.../subdir/pkg`) — no relative imports.
- The package qualifier is the last segment of the import path.
- Capitalised first letter → exported; lower-case → unexported. Uniform rule across fields, methods, functions, types, constants, variables.
- Daily Go tools: `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`. Run all of them.

---

## Up next

Lesson 08 — Phase 1 capstone. The expense tracker CLI: `add`, `list`, `summary` subcommands, JSON-backed persistence (via a provided `storage` package), `cmd/` convention, end-to-end CLI design.
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=07-packages &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/07-packages/slides/slides.md
curl -sS http://localhost:8000/lessons/07-packages/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/07-packages/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/07-packages/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `4`, `4`, `4`.

- [ ] **Step 3: Verify markdown structure**

```bash
grep -c '<div class="title-slide-grid">' lessons/07-packages/slides/slides.md
grep -c "<h1>Packages and modules</h1>" lessons/07-packages/slides/slides.md
grep -c "^## What we learned" lessons/07-packages/slides/slides.md
grep -c "^## Up next" lessons/07-packages/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/07-packages/slides/slides.md
git commit -m "feat(lesson-07): slides — Packages and modules (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/07-packages/README.md`. Same shape as Plans D-I, with extra emphasis on the subpackage layout and the import-path mechanics.

**Files:**
- Replace: `lessons/07-packages/README.md`

- [ ] **Step 1: Write `lessons/07-packages/README.md`** with:

> CRITICAL: Four-backtick wrapper is a documentation device; use three-backtick fences in the file. File starts with `# Lesson 07: Packages and modules`.

````markdown
# Lesson 07: Packages and modules

## Learning goals

- Split your code into packages — one package per directory, files share a `package X` declaration.
- Use Go's full module-rooted import paths to bring in your own subpackages.
- Apply the **formal** exported-vs-unexported rule: capitalised first letter → public, lower-case → private to the package. Same rule everywhere.
- Tour Go's daily-use tooling: `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`.

## Prerequisites

- Lessons 01-06. In particular, lesson 06's `Expense` + `Format`/`IsHigh` is what we're refactoring into a subpackage here.

## What's different about this lesson's layout

This is the first lesson where `exercises/` and `solutions/` are NOT flat. Each side contains a top-level `main.go` driver plus two subpackages:

```
lessons/07-packages/exercises/
├── main.go                    # package main — imports greet + expense
├── greet/
│   ├── greet.go               # package greet
│   └── greet_test.go
└── expense/
    ├── expense.go             # package expense
    └── expense_test.go
```

There's no `main_test.go` at the top level — `main.go` is a demo binary; verify it by running `go run ./lessons/07-packages/exercises`. Tests live in the subpackages.

## Concepts

### Splitting code into packages

Up to now we've put everything into the `exercises` (or `solutions`) package. Real Go programs split related code into focused **packages**, where each package is a single directory holding files that share a `package X` declaration.

Three rules:

1. **One package per directory.** Every `.go` file in a directory must declare the same package.
2. **Package name = directory name (usually).** A directory called `greet/` contains files that declare `package greet`.
3. **The `package` line is the very first non-comment line** of every `.go` file.

```go
// File: lessons/07-packages/exercises/greet/greet.go
package greet

func Greet(name string) string {
	return "Hello, " + name + "!"
}
```

```go
// File: lessons/07-packages/exercises/expense/expense.go
package expense

type Expense struct {
	Date     string
	Amount   float64
	Category string
}
```

To run the binary: `go run ./lessons/07-packages/exercises`.
To run all tests under the lesson: `go test ./lessons/07-packages/exercises/...` (the `...` means "this directory and all subdirectories").

**Common mistake.** Putting two different `package` declarations in the same directory:

```go
// File: greet/greet.go
package greet

// File: greet/hello.go
package hello   // wrong: must match the rest of the directory
```

Go complains: `found packages greet (greet.go) and hello (hello.go) in greet`. Fix: every `.go` file in a directory must share the same package declaration.

### Imports and module paths

Stdlib imports use short names:

```go
import "fmt"
import "strings"
```

Your own subpackages use the **full module-rooted path** (always — Go does not have relative imports):

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

Group imports in parentheses (`gofmt`'s convention; `goimports` enforces a blank line between stdlib and module-internal):

```go
import (
	"fmt"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
)
```

Use the imported package via its qualifier — the **last segment of the import path**:

```go
greet.Greet("Aki")
expense.Expense{Date: "..."}
expense.TotalsByCategory(es)
```

Why so verbose? Because `go.mod` at the repo root declares the module's import-path prefix:

```
module github.com/ristkari-dev/go-training

go 1.23
```

Anything imported under `github.com/ristkari-dev/go-training/...` is interpreted as a directory path relative to the repo root. Every package in the world has exactly one canonical import path, and Go's tooling never has to guess which one you meant.

```bash
$ go list ./lessons/07-packages/exercises/...
github.com/ristkari-dev/go-training/lessons/07-packages/exercises
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet
```

**Common mistake.** Trying relative imports:

```go
import "./expense"      // ERROR
import "../shared"      // ERROR
```

Go complains: `"./expense" is relative, but relative import paths are not supported in module mode`. Fix: use the full module-rooted path. Always.

### Exported vs unexported (formal)

Lesson 06 introduced this. Now we formalise: **a name is exported if its first letter is uppercase, unexported otherwise**. The rule is uniform across fields, methods, functions, types, constants, variables — Go has no other access modifiers.

```go
package billing

type Invoice struct {
	Number int       // exported — accessible from other packages
	Amount float64   // exported
	notes  string    // unexported — package-private
}

func (i Invoice) Display() string {     // exported method
	return fmt.Sprintf("...", i.notes)  // can read notes from within the package
}

func formatNotes(s string) string {     // unexported function — internal helper
	return strings.TrimSpace(s)
}
```

From another package:

```go
inv := billing.Invoice{Number: 1, Amount: 12.50}   // works
inv.notes = "..."                                   // ERROR: unexported field
```

Two API patterns to recognise:

```go
// Pattern A — fields exported, no constructor.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}
// Caller: e := expense.Expense{Date: "...", Amount: 4.50, Category: "coffee"}
```

```go
// Pattern B — fields unexported, constructor required.
type Expense struct {
	date, amount, category interface{}   // hidden
}
func NewExpense(date string, amount float64, cat string) (Expense, error) { /* ... */ }
// Caller: e, err := expense.NewExpense("...", 4.50, "coffee")
```

Pattern A is what we use this lesson — direct, simple, and our tests need to construct via struct literals. Pattern B is what production code often does for types with invariants — the constructor validates and the caller can't break the type's rules.

**Common mistake.** Lowercase = invisible from other packages, even though the field exists:

```go
// In package mypkg:
type Counter struct { count int }

// In another package:
c := mypkg.Counter{count: 5}
```

Go says: `unknown field count in struct literal of type mypkg.Counter` — even though `count` clearly exists. Fix: capitalise it (`Count`) or expose it through a method (`func (c Counter) Count() int { return c.count }`).

### Tooling tour

Five commands you'll run every day. We've met some already.

```bash
gofmt -l ./...              # list files needing reformat (CI fails if non-empty)
gofmt -w ./...              # rewrite in canonical format

go vet ./...                # static analysis — wrong Printf verbs, shadow vars, etc.

go doc fmt                  # docs for the fmt package
go doc fmt.Sprintf          # docs for one symbol
go doc -all fmt             # include unexported

go list ./...               # enumerate packages
go list -m                  # the module's path

go mod tidy                 # sync go.mod with imports actually used
```

The four-step rhythm before every commit:

```bash
gofmt -w ./...    # format
go vet ./...      # static check
go test ./...     # tests
git commit ...
```

Make it muscle memory. CI runs all four; running them locally first means you never push a broken build.

**Common mistake.** Skipping `gofmt` and seeing CI fail on indentation. Fix: wire your editor to format on save (`gofmt -w` on every save), or remember to run it before every commit.

A second one: editing imports manually and getting "unused import" or "imported and not used" errors. Install `goimports` (a sibling of `gofmt`):

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

Then `goimports -w .` adds missing imports and removes unused ones in one shot.

## Exercise: warm-up

In `exercises/greet/greet.go`:

- `func Greet(name string) string` — return `"Hello, <name>!"`. One-liner with `fmt.Sprintf`.

Then fill in the skeleton `TestGreet` in `exercises/greet/greet_test.go`. At least 4 cases: a normal name, the literal `"World"`, an empty name, and one with non-ASCII characters.

## Exercise: main

Three things in `exercises/expense/expense.go`:

- `func (e Expense) Format() string` — same as lesson 06's `Format`, byte-for-byte. Use `fmt.Sprintf("%s  €%-7.2f %s", ...)`.
- `func (e Expense) IsHigh() bool` — `e.Amount > 50`. The boundary value 50 is **not** high.
- `func TotalsByCategory(es []Expense) map[string]float64` — accumulate totals per category. Empty/nil input returns a non-nil empty map.

Fill in the three skeleton tests in `exercises/expense/expense_test.go`. Same case suggestions as lesson 06.

The top-level `main.go` is already wired — it imports both subpackages and runs them. Once the implementations are in, `go run ./lessons/07-packages/exercises` should produce:

```
Hello, Aki!
================
2026-05-15  €4.50    coffee
2026-05-15  €12.00   lunch
2026-05-15  €75.00   rent     (high)
2026-05-15  €9.99    coffee
----------------
totals: coffee=14.49 lunch=12.00 rent=75.00
```

> A note on `make test-lesson LESSON=07-packages`: lesson 07's exercise tests ship with empty `cases` slices, so the make output may show `ok` (vacuous pass) before you add cases. Don't be fooled — run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/07-packages/exercises
go test ./greet/... -v        # warm-up
go test ./expense/... -v      # main
go test ./...                  # both
go run .                       # the binary (panics until you implement the stubs)
```

Daily habits, now formalised:

```bash
gofmt -w ./...    # format
go vet ./...      # static analysis
```

Both run by CI. Add `goimports -w .` to your editor's save hook for the smoothest workflow.

## Going further

### Read

- [Effective Go — Names](https://go.dev/doc/effective_go#names) — package and identifier naming.
- [Go blog — Organizing Go code](https://go.dev/blog/organizing-go-code) — the canonical guide to package organisation.
- [Go module reference — Module paths](https://go.dev/ref/mod#module-path) — exhaustive reference for what makes a valid module path.

### Try

- **Add a third subpackage.** Create `exercises/greet/farewell/farewell.go` (yes, nested under greet — Go allows it) with a `Farewell(name string)` function. Import it from `main.go`. Note how the import path now reflects the deeper nesting.
- **Constructor pattern.** Convert `Expense` to Pattern B from the slides — make all fields unexported, write `NewExpense(date, amount, category)` as the only construction path. Update the test and `main.go` to use it. Observe what breaks (struct-literal construction is no longer valid from outside the `expense` package).
- **`go doc` on your code.** Run `go doc github.com/ristkari-dev/go-training/lessons/07-packages/solutions/expense.Expense` and read the output — it's your own doc comments rendered.
````

- [ ] **Step 2: Verify the README structure**

```bash
for h in "^# Lesson 07: Packages and modules$" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  c=$(grep -c "$h" lessons/07-packages/README.md)
  echo "  $h => $c"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/07-packages/README.md   # expect 4
grep -c "gofmt" lessons/07-packages/README.md                       # >= 2
grep -c "go vet" lessons/07-packages/README.md                      # >= 2
```

Expected: all section headings 1; Common-mistake count 4; gofmt and go vet ≥ 2.

- [ ] **Step 3: Commit**

```bash
git add lessons/07-packages/README.md
git commit -m "docs(lesson-07): README — Packages and modules self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes. Lesson 07 solutions pass; lessons 01-06 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — pass vacuously**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 07's TestGreet, TestExpenseFormat, TestExpenseIsHigh, TestTotalsByCategory all PASS with 0 sub-tests. Lessons 01-04 fail as before; lesson 05/06/07 pass vacuously. Make exits 0 because `-` ignores the failure.

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=07-packages 2>&1 | tail -30
```

Expected: exercise tests pass vacuously, solution tests pass fully (TestGreet 5, TestExpenseFormat 4, TestExpenseIsHigh 6, TestTotalsByCategory 5 = 20 sub-tests).

- [ ] **Step 4: Run the solutions binary end-to-end**

```bash
go run ./lessons/07-packages/solutions
```

Expected output:

```
Hello, Aki!
================
2026-05-15  €4.50    coffee
2026-05-15  €12.00   lunch
2026-05-15  €75.00   rent     (high)
2026-05-15  €9.99    coffee
----------------
totals: coffee=14.49 lunch=12.00 rent=75.00
```

- [ ] **Step 5: Verify exercises binary panics on first stub**

```bash
go run ./lessons/07-packages/exercises 2>&1 | head -5
```

Expected: panics with the `greet.Greet` TODO message (or the first stub the binary hits).

- [ ] **Step 6: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.` Confirm subpackage paths are matched by the exclusion (i.e. the staticcheck/unused ignore covers `lessons/07-packages/exercises/greet/` and `lessons/07-packages/exercises/expense/`).

- [ ] **Step 7: `go vet` clean**

```bash
go vet ./...
```

Expected: no output.

- [ ] **Step 8: Slides server smoke test**

```bash
make slides-dev LESSON=07-packages &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/07-packages/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/07-packages/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` or `302` for `/`, `200` or `301` for `index.html`, `200` for the others.

- [ ] **Step 9: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/07-packages/slides/slides.md && echo OK
grep -q "Packages" dist/index.html && echo "index lists lesson 07"
rm -rf dist
```

Expected: three success lines.

- [ ] **Step 10: Final repository sanity check**

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch.

This task makes no commit.

---

## Done definition

After Task 6:

- `lessons/07-packages/` contains 14 files in the subpackage layout.
- `make test` passes; `make test-exercises` shows lesson 07's tests passing vacuously.
- `make test-lesson LESSON=07-packages` runs both sides correctly.
- `go run ./lessons/07-packages/solutions` produces the expected report.
- `make slides-dev LESSON=07-packages` serves the deck.
- `make slides-build` produces `dist/` containing the lesson and lists it on the landing page.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- The git history is a clean sequence of small, conventional commits.

## What ships next

**Plan K — Lesson 08 (Phase 1 capstone — expense tracker CLI).** The end of Phase 1. Three subcommands (`add`/`list`/`summary`) backed by JSON file persistence via a provided `storage` package, the formal `cmd/` convention, a text bar chart in the summary output, and integration tests that drive the binary via `os/exec`. The existing `expense` package is carried forward (with stricter `Format` guarantees); a new `summary` package adds `BiggestCategory` and the bar-chart `Bar(value, max float64, width int) string` helper.
