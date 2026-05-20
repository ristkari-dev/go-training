# Plan L — Lesson 09 (Pointers) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the first lesson of Phase 2 — students learn pointer mechanics (`&` and `*`), pointer receivers as a deep dive (lesson 06 was a light touch), the "value receiver copies; pointer receiver shares" rule, when to reach for pointers (mutation, large structs), and the nil-pointer panic that's the most common runtime crash in Go. The main exercise extends lesson 08's `Expense` with two mutation methods using pointer receivers.

**Architecture:** Same per-lesson pattern as Plans D-K. Six tasks: scaffold + restructure, warmup, main, slides, README, end-to-end verify. Subpackages + no `cmd/` (lesson 09 has no binary — tests are the verification surface). Skeleton tests in both warmup and main (Phase 1's L05-L08 ending shape continues into Phase 2 per the design). Heavy-explanatory slide deck with **four** concept blocks (pointer basics → pointer receivers deep dive → when to use pointers → nil-pointer pitfalls).

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `testing`). Reveal.js 5.1.0 for the deck.

---

## Scope

Plan L produces lesson 09 only. After Plan L lands:

- `lessons/09-pointers/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows lesson 09's tests passing vacuously (skeleton case slices empty).
- `make slides-dev LESSON=09-pointers` serves the deck.
- The lesson is included in `dist/index.html` produced by `make slides-build`.
- **Phase 2 first lesson lands.**

**Out of scope (handled by future plans):** Lessons 10-15 (Plans M-R).

### Design decisions made during planning

1. **`Warmup*` prefix retired (per user choice).** Subpackages provide their own namespace; no collision risk. Lesson 09's warm-up type is `counter.Counter` (not `WarmupCounter`); future Phase 2 warmups follow suit. The Phase 2 design's open item is now answered.

2. **No `cmd/` binary in this lesson.** The Phase 2 design says "subpackages + cmd/ from L09 onward" as the default; lesson 09 is the first exception. Pointer mechanics are best learned via tests that demonstrate "the value stuck (or didn't)" — a binary would add ceremony without teaching anything new. Future lessons (L10 onward) include `cmd/` binaries where they fit naturally.

3. **Two subpackages: `warmup/counter/` and `expense/`.** No top-level `main.go`. Lesson 09's test surface is the per-subpackage `*_test.go` files. Students run `go test ./lessons/09-pointers/exercises/...` to verify their work.

4. **`Expense` carries forward from lesson 08.** Same three exported fields (`Date`, `Amount`, `Category`) with JSON tags (the tags don't hurt and L13 will rely on them). Existing value-receiver methods (`Format`, `IsHigh`) stay; the new value of this lesson is the two **pointer-receiver mutation methods** (`ApplyDiscount`, `Bump`).

5. **`Counter` warmup is minimal — one type, two methods.** `Counter` struct with one unexported field `n int`, plus `Inc()` (pointer receiver — mutates) and `Value() int` (value receiver — reads). The point is the receiver-mechanic contrast, not building a sophisticated counter API. Pedagogically aligned with Phase 1's smallest-possible warmups (e.g., lesson 04's `Add`).

6. **Four concepts in the slide deck.**
   1. **Pointers basics** — `&` (address-of), `*` (dereference), what a pointer IS (memory address), value vs reference semantics, the nil pointer.
   2. **Pointer receivers — deep dive** — `(c *Counter) Inc()` vs `(c Counter) Value()`, method set rules (which methods are callable on `T` vs `*T`), the consistency rule.
   3. **When to use pointers** — mutation (the primary reason), large structs (the secondary reason), and the **don't-reach-for-pointers-by-default** rule. Brief escape-analysis intuition (no internals).
   4. **Common nil-pointer pitfalls** — uninitialized pointer dereference, nil pointer in struct field, methods on nil receivers (sometimes valid, often a foot-gun).

7. **No `cmp.Diff`, no third-party deps.** Phase 2 stays stdlib-only; pointer-test assertions use plain `==` on the struct fields after the mutating call.

8. **`*Expense` mutation methods stick on the caller.** The lesson's primary "aha" — calling `e.ApplyDiscount(0.10)` on a value `e` mutates `e` *in the caller*. Tests verify this by reading `e.Amount` after the call. Demonstrates the receiver-mechanic contrast against `Format`/`IsHigh` which don't mutate.

9. **`Counter.n` is unexported.** First Phase 2 use of the lesson 07 "exported vs unexported" formal rule. The exported method `Value() int` is the only way to read `n` from outside the `counter` package. Discussed briefly in the slides' concept 2 worked example.

---

## Plans F-K lessons-learned applied here

1. **Subpackages have their own namespace** — no `Warmup*` prefix needed (formalised this lesson; see decision 1).

2. **Lint exclusion covers nested subpackages** — `.golangci.yml`'s `lessons/.*/exercises/` regex matches `lessons/09-pointers/exercises/warmup/counter/counter.go`. Verify in Task 1 after the scaffold restructure.

3. **`→` arrow consistency** — Unicode arrows in doc-comment examples; gofmt may reformat blocks; arrows survive.

4. **Common-mistake content in README** — every slide concept's common-mistake example is mirrored in the README. 4 total.

5. **Slides + README written inline by controller** — Plans G-K established this pattern; Plan L continues. Tasks 1-3 can be subagent-driven; Tasks 4-5 the controller writes directly.

6. **`gofmt -w .` and `go vet ./...` mentions** — README continues both habits from Phase 1.

7. **Format-output verification (Plan G/I/J/K lesson learned)** — lesson 09's `Expense.Format()` reuses lesson 08's byte-output (same format string `"%s  €%-7.2f %s"`). Test cases are known-good.

8. **Empty case → vacuous pass.** Skeleton tests pass vacuously until students add cases. README warns explicitly.

9. **Scaffold + restructure** — Phase 2's first scaffold-then-restructure (same pattern as Plans J, K). The scaffolder produces flat `exercises/{main,warmup}.go`; Plan L deletes those and creates the subpackage tree.

---

## File Structure

After Plan L (12 files total):

```
lessons/09-pointers/
├── README.md                                  (Task 5)
├── slides/
│   ├── index.html                             (Task 1; unchanged from scaffold)
│   ├── slides.md                              (Task 4 — controller writes inline)
│   └── assets/.gitkeep                        (Task 1)
├── exercises/
│   ├── warmup/
│   │   └── counter/
│   │       ├── counter.go                     (Task 2 — Counter + Inc + Value stubs)
│   │       └── counter_test.go                (Task 2 — SKELETON test)
│   └── expense/
│       ├── expense.go                         (Task 3 — Expense + Format + IsHigh + ApplyDiscount + Bump stubs)
│       └── expense_test.go                    (Task 3 — SKELETON tests)
└── solutions/
    ├── warmup/
    │   └── counter/
    │       ├── counter.go                     (Task 2 — implementation)
    │       └── counter_test.go                (Task 2 — full reference test)
    └── expense/
        ├── expense.go                         (Task 3 — implementation)
        └── expense_test.go                    (Task 3 — full reference tests)
```

Files deleted from the scaffolder output (8):

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

### Decomposition rationale

Same 6-task pattern as plans D-J (lesson 08 was 10 tasks; that was the capstone scale). Lesson 09 is more like lessons 05/06/07 — focused on one set of concepts, 12 files. Task 1 also performs the structural restructure (delete unwanted scaffolder files, create empty subpackage tree placeholders).

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-l-lesson-09-pointers` (already created by the controller; the plan doc is the only commit ahead of `main`)
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold lesson 09 + restructure

The scaffolder produces 12 flat files; we delete the 8 `.go` files at the top level and leave the README + slides shell. Subpackage trees are created from scratch in tasks 2-3.

**Files:**
- Create via scaffolder: `lessons/09-pointers/` (12 files initially)
- Delete: 8 `.go` files

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --no-show-signature --oneline main..HEAD
```

Expected: `## feature/plan-l-lesson-09-pointers`, clean working tree, exactly one commit ahead of `main` (the Plan L doc).

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=09-pointers
```

Expected: `created lesson 09-pointers under lessons/`. 12 placeholder files produced.

- [ ] **Step 3: Delete the unwanted files**

```bash
rm lessons/09-pointers/exercises/main.go \
   lessons/09-pointers/exercises/main_test.go \
   lessons/09-pointers/exercises/warmup.go \
   lessons/09-pointers/exercises/warmup_test.go \
   lessons/09-pointers/solutions/main.go \
   lessons/09-pointers/solutions/main_test.go \
   lessons/09-pointers/solutions/warmup.go \
   lessons/09-pointers/solutions/warmup_test.go
```

Expected: 8 files removed.

- [ ] **Step 4: Verify the surviving tree**

```bash
find lessons/09-pointers -type f | sort
```

Expected (4 files):

```
lessons/09-pointers/README.md
lessons/09-pointers/slides/assets/.gitkeep
lessons/09-pointers/slides/index.html
lessons/09-pointers/slides/slides.md
```

- [ ] **Step 5: Verify make test still passes**

```bash
make test
```

Expected: every package green. The lesson 09 dirs have no Go files yet; `go test` ignores them.

- [ ] **Step 6: Commit**

```bash
git add -A lessons/09-pointers/
git commit -m "feat(lessons): scaffold lesson 09-pointers with empty subpackage layout"
```

- [ ] **Step 7: Sanity check**

```bash
git status
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, two commits on the branch.

---

## Task 2: Author the warm-up — `counter` subpackage

A `Counter` struct with `Inc()` (pointer receiver — mutates) and `Value() int` (value receiver — reads). The smallest possible "pointer receiver mutates; value receiver doesn't" demo.

**Files:**
- Create: `lessons/09-pointers/exercises/warmup/counter/counter.go`
- Create: `lessons/09-pointers/exercises/warmup/counter/counter_test.go`
- Create: `lessons/09-pointers/solutions/warmup/counter/counter.go`
- Create: `lessons/09-pointers/solutions/warmup/counter/counter_test.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/09-pointers/exercises/warmup/counter \
         lessons/09-pointers/solutions/warmup/counter
```

- [ ] **Step 2: Create `lessons/09-pointers/exercises/warmup/counter/counter.go`** with:

```go
// Package counter is the lesson 09 warm-up: a tiny demo of value-receiver
// vs pointer-receiver semantics.
//
// Counter has one unexported field `n` (the running count). It exposes:
//   - Inc()         — POINTER receiver. Mutates the underlying value.
//   - Value() int   — VALUE receiver. Reads the count; can't mutate.
//
// The whole lesson is: "if you want a method to mutate the receiver,
// declare it with a pointer receiver. Value receivers get a COPY."
package counter

// Counter holds an int count. The field is unexported — the only way to
// read it from outside this package is via Value().
type Counter struct {
	n int
}

// Inc increments the counter by 1.
//
// Pointer receiver — the call site's Counter is mutated. After
// `c.Inc()`, the caller's `c.Value()` will return 1 more than before.
//
// Examples:
//
//	var c Counter
//	c.Inc()
//	c.Inc()
//	c.Value()   → 2
//
// Hint: this is a one-liner. `c.n++`.
func (c *Counter) Inc() {
	panic("TODO: increment c.n by 1")
}

// Value returns the current count.
//
// Value receiver — it can READ c.n but mutations to c here would be
// invisible to the caller (because c is a copy). For a read-only method,
// value receivers are the conventional choice.
//
// Examples:
//
//	var c Counter
//	c.Value()   → 0
//	c.Inc()
//	c.Value()   → 1
//
// Hint: one-liner — return c.n.
func (c Counter) Value() int {
	panic("TODO: return c.n")
}
```

- [ ] **Step 3: Create `lessons/09-pointers/exercises/warmup/counter/counter_test.go`** with (SKELETON):

```go
package counter

import "testing"

// TestCounter is a SKELETON. Fill in the cases and the t.Run body.
//
// Each case describes a sequence of Inc() calls then asserts the final
// Value(). Cases to cover: zero (no Inc calls), one Inc, several Inc calls,
// and a sanity check that two separate Counters don't interfere.
//
// Hint: the cleanest shape is a table of (name, incCount, want):
//
//	cases := []struct {
//		name     string
//		incCount int
//		want     int
//	}{ ... }
//
// Then in the loop: declare a fresh `var c Counter`, call c.Inc() in a for
// loop incCount times, assert c.Value() == want.
func TestCounter(t *testing.T) {
	cases := []struct {
		name     string
		incCount int
		want     int
	}{
		// TODO: at least 4 cases. zero, one, many, large-n.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: var c Counter
			//   for i := 0; i < tc.incCount; i++ { c.Inc() }
			//   if got := c.Value(); got != tc.want { t.Errorf(...) }
			_ = tc
		})
	}
}

// TestCountersDoNotInterfere is a SKELETON. Verify that two separate
// Counter values are independent — incrementing one does not affect the
// other. This catches a class of bug where Counter was implemented as a
// pointer-typed field (or used a package-level var).
func TestCountersDoNotInterfere(t *testing.T) {
	// TODO:
	//   var a, b Counter
	//   a.Inc(); a.Inc(); a.Inc()
	//   b.Inc()
	//   if got := a.Value(); got != 3 { t.Errorf(...) }
	//   if got := b.Value(); got != 1 { t.Errorf(...) }
	_ = t
}
```

- [ ] **Step 4: Create `lessons/09-pointers/solutions/warmup/counter/counter.go`** with:

```go
// Package counter is the lesson 09 warm-up reference implementation.
package counter

// Counter holds an int count.
type Counter struct {
	n int
}

// Inc increments the counter by 1. Pointer receiver — mutates the caller's value.
func (c *Counter) Inc() {
	c.n++
}

// Value returns the current count. Value receiver — read-only.
func (c Counter) Value() int {
	return c.n
}
```

- [ ] **Step 5: Create `lessons/09-pointers/solutions/warmup/counter/counter_test.go`** with the full reference tests:

```go
package counter

import "testing"

func TestCounter(t *testing.T) {
	cases := []struct {
		name     string
		incCount int
		want     int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"three", 3, 3},
		{"hundred", 100, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Counter
			for i := 0; i < tc.incCount; i++ {
				c.Inc()
			}
			if got := c.Value(); got != tc.want {
				t.Errorf("after %d Inc() calls, Value() = %d, want %d", tc.incCount, got, tc.want)
			}
		})
	}
}

func TestCountersDoNotInterfere(t *testing.T) {
	var a, b Counter
	a.Inc()
	a.Inc()
	a.Inc()
	b.Inc()
	if got := a.Value(); got != 3 {
		t.Errorf("a.Value() = %d, want 3", got)
	}
	if got := b.Value(); got != 1 {
		t.Errorf("b.Value() = %d, want 1", got)
	}
}
```

- [ ] **Step 6: gofmt + tests + lint + vet**

```bash
gofmt -l lessons/09-pointers/
go test ./lessons/09-pointers/exercises/warmup/counter/... -v 2>&1 | tail -10
go test ./lessons/09-pointers/solutions/warmup/counter/... -v 2>&1 | tail -15
make test
golangci-lint run ./...
go vet ./...
```

Expected: gofmt empty; exercises TestCounter + TestCountersDoNotInterfere pass vacuously (bodies are TODO placeholders); solutions TestCounter (4 sub-tests) + TestCountersDoNotInterfere pass; make test green; lint 0 issues; vet clean.

- [ ] **Step 7: Commit**

```bash
git add lessons/09-pointers/exercises/warmup/ lessons/09-pointers/solutions/warmup/
git commit -m "feat(lesson-09): warmup — counter subpackage (pointer vs value receiver demo)"
```

---

## Task 3: Author the main exercise — `expense` subpackage

Extend lesson 08's `Expense` with two pointer-receiver mutation methods: `ApplyDiscount(rate float64)` reduces `Amount` by `rate` (e.g. `0.10` = 10% off); `Bump(amount float64)` adds `amount` to `Amount`. Existing value-receiver methods (`Format`, `IsHigh`) stay.

**Files:**
- Create: `lessons/09-pointers/exercises/expense/expense.go`
- Create: `lessons/09-pointers/exercises/expense/expense_test.go`
- Create: `lessons/09-pointers/solutions/expense/expense.go`
- Create: `lessons/09-pointers/solutions/expense/expense_test.go`

- [ ] **Step 1: Create directories**

```bash
mkdir -p lessons/09-pointers/exercises/expense lessons/09-pointers/solutions/expense
```

- [ ] **Step 2: Create `lessons/09-pointers/exercises/expense/expense.go`** with:

```go
// Package expense is the lesson 09 main exercise — adds pointer-receiver
// mutation methods to the Expense type from lesson 08.
//
// Existing value-receiver methods (Format, IsHigh) stay — they read
// e but don't change it. The two new methods (ApplyDiscount, Bump) use
// pointer receivers because they mutate e.Amount.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// Same fields and JSON tags as lesson 08. Mutation methods (ApplyDiscount,
// Bump) operate on a pointer to the value; read-only methods (Format,
// IsHigh) take a value.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" via fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04/06/07/08.
//
// Value receiver — reads e, doesn't mutate.
//
// Hint: same as lesson 08.
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf(\"%s  €%-7.2f %s\", e.Date, e.Amount, e.Category)")
}

// IsHigh reports whether the expense is over €50. Value receiver.
//
// Boundary: amount of exactly 50 is NOT high; 50.01 IS high.
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// ApplyDiscount reduces e.Amount by the given rate (0.0 to 1.0).
//
// Pointer receiver — the call mutates the caller's Expense. After
// `e.ApplyDiscount(0.10)`, e.Amount is 90% of what it was.
//
// rate < 0 or rate > 1 is undefined behaviour for this exercise (no
// validation; trust the caller). Lesson 11 covers error handling for
// these "trust boundary" cases properly.
//
// Examples:
//
//	e := Expense{Amount: 100}
//	e.ApplyDiscount(0.10)
//	e.Amount  → 90.0
//
//	e := Expense{Amount: 50}
//	e.ApplyDiscount(0)
//	e.Amount  → 50.0      (no change for zero rate)
//
//	e := Expense{Amount: 100}
//	e.ApplyDiscount(1)
//	e.Amount  → 0.0       (100% discount → free)
//
// Hint: e.Amount *= (1 - rate)
func (e *Expense) ApplyDiscount(rate float64) {
	panic("TODO: multiply e.Amount by (1 - rate)")
}

// Bump adds amount to e.Amount.
//
// Pointer receiver — mutates the caller. `amount` can be positive
// (increase) or negative (refund / correction). Use Bump for adjustments
// that aren't a proportional discount.
//
// Examples:
//
//	e := Expense{Amount: 10}
//	e.Bump(5)
//	e.Amount  → 15.0
//
//	e := Expense{Amount: 10}
//	e.Bump(-3)
//	e.Amount  → 7.0
//
//	e := Expense{Amount: 10}
//	e.Bump(0)
//	e.Amount  → 10.0
//
// Hint: e.Amount += amount
func (e *Expense) Bump(amount float64) {
	panic("TODO: add amount to e.Amount")
}
```

- [ ] **Step 3: Create `lessons/09-pointers/exercises/expense/expense_test.go`** with (SKELETON):

```go
package expense

import "testing"

// TestExpenseFormat is a SKELETON. Same shape as lesson 06/07/08 — fill in
// cases for small/medium/large amounts and assert against the expected
// "%s  €%-7.2f %s" output strings.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover boundary 50 explicitly.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases incl. 49.99, 50 (false), 50.01 (true).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO
			_ = tc
		})
	}
}

// TestExpenseApplyDiscount is a SKELETON. Test that the discount STICKS
// on the caller's Expense — that's the lesson's main point (pointer
// receiver mutates).
//
// Cases to cover: 10% off a round number, 50% off, 0 rate (no-op),
// 1.0 rate (full discount → 0), 100% discount on zero-amount expense.
func TestExpenseApplyDiscount(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		rate        float64
		wantAmount  float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   e := Expense{Amount: tc.startAmount}
			//   e.ApplyDiscount(tc.rate)
			//   if e.Amount != tc.wantAmount { t.Errorf(...) }
			_ = tc
		})
	}
}

// TestExpenseBump is a SKELETON. Same shape — verify the bump sticks.
//
// Cases: positive bump, negative bump (refund/correction), zero bump
// (no-op), bump on zero-amount expense.
func TestExpenseBump(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		bump        float64
		wantAmount  float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   e := Expense{Amount: tc.startAmount}
			//   e.Bump(tc.bump)
			//   if e.Amount != tc.wantAmount { t.Errorf(...) }
			_ = tc
		})
	}
}
```

- [ ] **Step 4: Create `lessons/09-pointers/solutions/expense/expense.go`** with:

```go
// Package expense is the lesson 09 main exercise reference implementation.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" with "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

// ApplyDiscount reduces e.Amount by the given rate (0.0 to 1.0).
// Pointer receiver — mutates the caller.
func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

// Bump adds amount to e.Amount. Pointer receiver — mutates the caller.
func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
```

- [ ] **Step 5: Create `lessons/09-pointers/solutions/expense/expense_test.go`** with full reference tests:

```go
package expense

import "testing"

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-19", Amount: 4.50, Category: "coffee"}, "2026-05-19  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-19", Amount: 12, Category: "lunch"}, "2026-05-19  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-19", Amount: 999.99, Category: "rent"}, "2026-05-19  €999.99  rent"},
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

func TestExpenseApplyDiscount(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		rate        float64
		wantAmount  float64
	}{
		{"10-percent-off-100", 100, 0.10, 90},
		{"50-percent-off-50", 50, 0.50, 25},
		{"zero-rate-no-op", 50, 0, 50},
		{"full-discount", 100, 1.0, 0},
		{"discount-on-zero", 0, 0.10, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Expense{Amount: tc.startAmount}
			e.ApplyDiscount(tc.rate)
			if e.Amount != tc.wantAmount {
				t.Errorf("Expense{Amount: %g}.ApplyDiscount(%g) → Amount %g, want %g",
					tc.startAmount, tc.rate, e.Amount, tc.wantAmount)
			}
		})
	}
}

func TestExpenseBump(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		bump        float64
		wantAmount  float64
	}{
		{"positive-bump", 10, 5, 15},
		{"negative-bump", 10, -3, 7},
		{"zero-bump", 10, 0, 10},
		{"bump-on-zero", 0, 5, 5},
		{"bump-to-negative", 5, -10, -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Expense{Amount: tc.startAmount}
			e.Bump(tc.bump)
			if e.Amount != tc.wantAmount {
				t.Errorf("Expense{Amount: %g}.Bump(%g) → Amount %g, want %g",
					tc.startAmount, tc.bump, e.Amount, tc.wantAmount)
			}
		})
	}
}
```

- [ ] **Step 6: Format-output smoke test (Plan G/I/J/K lesson learned)**

```bash
cat > /tmp/format-check.go <<'EOF'
package main
import "fmt"
func main() {
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-19", 4.50, "coffee"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-19", 12.00, "lunch"))
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-19", 999.99, "rent"))
}
EOF
go run /tmp/format-check.go
rm /tmp/format-check.go
```

Expected (byte-match against solution test cases):

```
"2026-05-19  €4.50    coffee"
"2026-05-19  €12.00   lunch"
"2026-05-19  €999.99  rent"
```

- [ ] **Step 7: gofmt + tests + lint + vet**

```bash
gofmt -l lessons/09-pointers/
go test ./lessons/09-pointers/exercises/expense/... -v 2>&1 | tail -15
go test ./lessons/09-pointers/solutions/expense/... -v 2>&1 | tail -30
make test
golangci-lint run ./...
go vet ./...
```

Expected: gofmt empty; exercises pass vacuously (4 test functions); solutions pass all sub-tests: TestExpenseFormat (3) + TestExpenseIsHigh (6) + TestExpenseApplyDiscount (5) + TestExpenseBump (5) = 19; make test green; lint 0 issues; vet clean.

- [ ] **Step 8: Commit**

```bash
git add lessons/09-pointers/exercises/expense/ lessons/09-pointers/solutions/expense/
git commit -m "feat(lesson-09): main — Expense + ApplyDiscount + Bump (pointer receivers)"
```

---

## Task 4: Author the slide deck

4 concepts: pointer basics → pointer receivers (deep dive) → when to use pointers → common nil-pointer pitfalls.

The controller writes this directly inline (Plans G-K pattern).

**Files:**
- Replace: `lessons/09-pointers/slides/slides.md`

Structure check after writing:

```bash
grep -c '<div class="title-slide-grid">' lessons/09-pointers/slides/slides.md     # 1
grep -c "<h1>Pointers</h1>" lessons/09-pointers/slides/slides.md                  # 1
grep -c "^## Concept " lessons/09-pointers/slides/slides.md                        # 4
grep -c "^### Motivation$" lessons/09-pointers/slides/slides.md                    # 4
grep -c "^### Common mistake$" lessons/09-pointers/slides/slides.md                # 4
grep -c "^## What we learned" lessons/09-pointers/slides/slides.md                 # 1
grep -c "^## Up next" lessons/09-pointers/slides/slides.md                          # 1
```

Commit:

```bash
git add lessons/09-pointers/slides/slides.md
git commit -m "feat(lesson-09): slides — Pointers (4 concepts)"
```

---

## Task 5: Author the README

Sections: Learning goals, Prerequisites, What's different (Phase 2 begins — pointer mechanics formal; subpackages without cmd/), Concepts (mirroring 4 slide concepts with the Common-mistake examples), Exercise: warm-up, Exercise: main, How to run, Going further (Read + Try).

The controller writes this directly inline.

**Files:**
- Replace: `lessons/09-pointers/README.md`

Structure check after writing:

```bash
for h in "^# Lesson 09: Pointers$" "^## Learning goals$" "^## Prerequisites$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  echo "  $h => $(grep -c "$h" lessons/09-pointers/README.md)"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/09-pointers/README.md    # expect 4
grep -c "gofmt" lessons/09-pointers/README.md                        # >= 1
grep -c "go vet" lessons/09-pointers/README.md                       # >= 1
```

Expected: all section headings 1; Common-mistake count 4; gofmt + go vet ≥ 1.

Commit:

```bash
git add lessons/09-pointers/README.md
git commit -m "docs(lesson-09): README — Pointers self-study"
```

---

## Task 6: End-to-end verification

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package green. Lesson 09 solutions pass; lessons 01-08 still pass; tools still pass.

- [ ] **Step 2: Run exercise tests — pass vacuously**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: lesson 09's TestCounter, TestCountersDoNotInterfere, TestExpenseFormat, TestExpenseIsHigh, TestExpenseApplyDiscount, TestExpenseBump all PASS with 0 sub-tests (skeleton case slices empty + `_ = tc` placeholders).

- [ ] **Step 3: Run lesson-specific tests**

```bash
make test-lesson LESSON=09-pointers 2>&1 | tail -30
```

Expected: exercises pass vacuously; solutions pass fully. Solution sub-test totals:
- warmup/counter TestCounter (4) + TestCountersDoNotInterfere (1 — though it has no sub-tests, just a single test function with multiple asserts) = 5 assertions
- expense TestExpenseFormat (3) + TestExpenseIsHigh (6) + TestExpenseApplyDiscount (5) + TestExpenseBump (5) = 19 sub-tests

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
make slides-dev LESSON=09-pointers &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/09-pointers/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/09-pointers/slides/slides.md
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`/`301`/`302` codes.

- [ ] **Step 7: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/09-pointers/slides/slides.md && echo OK
rm -rf dist
```

Expected: two success lines.

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

- `lessons/09-pointers/` contains 12 files in the subpackage layout.
- `make test` passes; `make test-exercises` shows lesson 09's tests passing vacuously.
- `make test-lesson LESSON=09-pointers` runs both sides correctly.
- `make slides-dev LESSON=09-pointers` serves the deck.
- `make slides-build` produces `dist/` containing the lesson.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- 6 commits on the branch.

## What ships next

**Plan M — Lesson 10 (Interfaces).** Same per-lesson pattern. Lesson 10 introduces implicit interface satisfaction, small interfaces (`io.Reader`/`io.Writer`/`error`), `any` (formerly `interface{}`), and type assertions. The main exercise introduces a `Store` interface in the tracker — abstracts persistence; JSON + in-memory implementations. The tracker continues to evolve.
