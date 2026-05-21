# Plan O — Lesson 12 (Generics) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 12 of Phase 2 — students learn Go generics through a small standalone `slices`-style helper library. Five concepts: type parameters (`func F[T any](x T) T`), constraints (`any` / `comparable` / `cmp.Ordered`), type inference (when the compiler infers `T` from args and when it can't), multiple type parameters (`Map[T, U]`), and Go's culture of restraint around generics ("if a function is called once with one type, don't bother"). The warm-up implements a generic `Max[T cmp.Ordered](xs []T) (T, error)`. The main exercise implements `Filter[T any]` + `Map[T, U any]`. **Standalone** — no tracker integration; no `expense` subpackage; no cmd binary. The Phase 2 spec is explicit that forcing generics into the tracker is the premature-abstraction anti-pattern Go itself warns against.

**Architecture:** Same per-lesson pattern as Plans D-N. Six tasks. Skeleton tests in warmup + main (Phase 2 default). Heavy-explanatory slide deck with **five** concept blocks (type parameters → constraints → type inference → multiple type parameters → when NOT to use generics).

**Tech Stack:** Go 1.23 stdlib only (`cmp` for `cmp.Ordered`, `errors`, `fmt`, `strconv`, `strings`, `testing`). Reveal.js 5.1.0.

---

## Scope

After Plan O: lesson 12 is complete; `make test` green; both subpackages exercised through their public APIs in solution tests; slides build and `12-generics` lands in the index.

### Design decisions (2 user-approved + 7 plan-recommended)

**User-approved via brainstorming:**

1. **Two subpackages: `warmup/maxg/` + `slicesx/`.** Preserves the Phase 2 warmup/main split established by L09-L11 (skeleton tests in both halves; warmup is a self-contained mini-exercise; main package is what students "build"). Tradeoff: the eventual `slicesx` package only has 2 functions instead of 3, which slightly weakens the "small reusable lib" framing — addressed in the README/slides with "you'd merge maxg into slicesx in real life, but we kept them separate here so you implement Max first as a single-function warm-up."

2. **Five concepts in the slide deck** (type parameters → constraints → type inference → multiple type parameters → when NOT to use generics). Type inference gets its own slot rather than being a sub-bullet because the Common-mistake material is genuinely useful (students sometimes write redundant explicit `[int]` or call functions where inference fails). The 5-concept rhythm is a one-off bump from L09/L10/L11's 4 concepts — justified by generics having that one extra fundamentally distinct sub-topic.

**Plan-recommended:**

3. **Standalone — no `expense` subpackage, no cmd binary.** Per Phase 2 spec. The "Filter on []Expense" motivation is shown as an illustrative slide block only (no real import). Keeps the lesson tightly focused on generics mechanics.

4. **`Max` returns plain `errors.New("maxg: empty slice")` — no sentinel.** L11 just taught sentinels; reusing the pattern here would only be motivated by callers needing to programmatically distinguish "empty" from other failures (there are no other failures). Adding a sentinel adds noise without pedagogical payoff. Lesson 12 is about generics, not errors.

5. **Warmup package named `maxg`** (terse, "g" for generic, avoids the Go 1.21+ `max` builtin). `slicesx` for the main package (mirrors community convention for stdlib-adjacent helpers like `golang.org/x/exp/slices` → `slicesx`).

6. **`cmp.Ordered` (stdlib Go 1.21+) — not `constraints.Ordered`.** No `golang.org/x/exp/constraints` third-party. Course is on Go 1.23.

7. **Pre-allocate output slices.** `Filter` uses `make([]T, 0, len(xs))` (worst-case zero re-allocs); `Map` uses `make([]U, len(xs))` (exact size known). Demonstrates Go's allocation-aware idiom in passing — students saw `make` in L05.

8. **`Map[T, U any]` — both type params unconstrained.** No `comparable` or `Ordered`. The lesson teaches that **most** generic functions need only `any` (constraint-free). Students see `Ordered` once in `Max` and understand it's used only when comparisons matter.

9. **Test surface: ~12 solution sub-tests.** `Max`: 4 (positive, negative, single, empty→error); `Filter`: 4 (even ints, by-length strings, empty, no matches); `Map`: 4 (int→string via `strconv.Itoa`, string→string via `strings.ToUpper`, string→int via `len`, empty). Skeleton tests reference all needed identifiers via `_ = ...` so imports stay satisfied but bodies don't call the panicking stubs (vacuous pass).

---

## Plans F-N lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion (`.golangci.yml` line 24) covers nested subpackages — no config changes needed.
3. `→` arrow consistency.
4. Common-mistake content in README (5 per lesson, one per concept).
5. Slides + README written inline by controller (Plan K/L/M/N pattern).
6. `gofmt -w .` and `go vet ./...` mentions.
7. Skeleton tests pass vacuously (never call the panicking stubs).
8. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).
9. No `internal/` (formal introduction in L15).

---

## File Structure

After Plan O (12 files total — same shape as L09; smaller than L10/L11 because no expense + no cmd):

```
lessons/12-generics/
├── README.md                                       (Task 5)
├── slides/
│   ├── index.html, slides.md, assets/.gitkeep      (Task 1 + Task 4)
├── exercises/
│   ├── warmup/maxg/
│   │   ├── maxg.go                                 (Task 2)
│   │   └── maxg_test.go                            (Task 2 — SKELETON)
│   └── slicesx/
│       ├── slicesx.go                              (Task 3)
│       └── slicesx_test.go                         (Task 3 — SKELETON)
└── solutions/
    └── (mirrored structure)
```

---

## Conventions

- **Branch:** `feature/plan-o-lesson-12-generics`
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold + restructure

Same dance as Plans J/K/L/M/N.

- [ ] **Step 1:** `make new-lesson NAME=12-generics`
- [ ] **Step 2:** Delete 8 unwanted flat scaffolder files:

```bash
rm lessons/12-generics/exercises/warmup.go
rm lessons/12-generics/exercises/warmup_test.go
rm lessons/12-generics/exercises/main.go
rm lessons/12-generics/exercises/main_test.go
rm lessons/12-generics/solutions/warmup.go
rm lessons/12-generics/solutions/warmup_test.go
rm lessons/12-generics/solutions/main.go
rm lessons/12-generics/solutions/main_test.go
```

- [ ] **Step 3:** Verify 4-file scaffolded tree (README + slides only):

```bash
find lessons/12-generics -type f | sort
# Expected:
# lessons/12-generics/README.md
# lessons/12-generics/slides/assets/.gitkeep
# lessons/12-generics/slides/index.html
# lessons/12-generics/slides/slides.md
```

- [ ] **Step 4:** Commit:

```bash
git add lessons/12-generics/
git commit -m "feat(lessons): scaffold lesson 12-generics with empty subpackage layout"
```

---

## Task 2: Author the warm-up — `maxg` subpackage

`Max[T cmp.Ordered](xs []T) (T, error)` — the generic version of L04's WarmupMinMax. Single function, single constraint, gets students familiar with the type-parameter syntax before they tackle multi-parameter generics in Task 3.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/12-generics/{exercises,solutions}/warmup/maxg`

- [ ] **Step 2:** Create `lessons/12-generics/exercises/warmup/maxg/maxg.go`:

```go
// Package maxg is the lesson 12 warm-up: a generic Max function.
//
// In lesson 04 (functions) you wrote separate Min/Max helpers for ints
// and floats — code duplication forced by Go's lack of generics back
// then. Go 1.18 added type parameters; Go 1.21 added the cmp.Ordered
// constraint to the standard library. Now we can write Max ONCE and use
// it on any ordered type (int, float64, string, time.Duration, ...).
//
// This is also a tiny demo of error handling for the empty-slice case
// (introduced in L11): an empty input returns the zero value of T plus
// a non-nil error.
package maxg

import "cmp"

// Max returns the largest element of xs.
//
// Type parameter T is constrained by cmp.Ordered — the set of types
// that support <, <=, >, >= (all integer types, all float types, strings).
// You CANNOT call Max on []bool or []struct{...} without an Ordered
// constraint — the compiler rejects it at the call site.
//
// On empty input, returns (zero T, non-nil error). On non-empty input,
// returns (largest element, nil).
//
// Type inference: callers usually write Max(xs) — the compiler infers T
// from the element type of xs. The explicit form Max[int](xs) is legal
// but unnecessary.
//
// Hint:
//   1. if len(xs) == 0 → return zero, errors.New("maxg: empty slice")
//      where `var zero T` gives the zero value for the inferred T.
//   2. m := xs[0]
//   3. for _, x := range xs[1:] → if x > m → m = x
//   4. return m, nil
func Max[T cmp.Ordered](xs []T) (T, error) {
	_ = cmp.Less[int]
	panic("TODO: handle empty, then linear scan; return largest element")
}
```

- [ ] **Step 3:** Create `lessons/12-generics/exercises/warmup/maxg/maxg_test.go` (SKELETON):

```go
package maxg

import (
	"testing"
)

// TestMaxInts is a SKELETON. Cover at least:
//   - A slice of positive ints (e.g., {1, 5, 3, 2} → 5)
//   - A slice of negative ints (e.g., {-1, -5, -3} → -1)
//   - A single-element slice (e.g., {42} → 42)
func TestMaxInts(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want int
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := Max(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestMaxStrings is a SKELETON. Demonstrate that Max works on []string
// too — same call, different element type (compiler infers T=string).
func TestMaxStrings(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		// TODO: at least 2 cases.
		// Hint: strings compare lexicographically. "banana" > "apple".
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got, err := Max(tc.in); assert.
			_ = tc
		})
	}
}

// TestMaxEmpty is a SKELETON. Assert that Max(nil) and Max([]int{}) both
// return a non-nil error (and a zero value, but the zero matters less —
// the error is the signal).
func TestMaxEmpty(t *testing.T) {
	// TODO:
	//   _, err := Max([]int{})
	//   if err == nil → t.Fatal("expected error for empty slice")
	//   _, err = Max[int](nil)  // explicit type param needed; nil has no element type
	//   if err == nil → t.Fatal("expected error for nil slice")
}
```

- [ ] **Step 4:** Create `lessons/12-generics/solutions/warmup/maxg/maxg.go`:

```go
// Package maxg is the lesson 12 warm-up reference implementation.
package maxg

import (
	"cmp"
	"errors"
)

// Max returns the largest element of xs (cmp.Ordered constraint).
func Max[T cmp.Ordered](xs []T) (T, error) {
	var zero T
	if len(xs) == 0 {
		return zero, errors.New("maxg: empty slice")
	}
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m, nil
}
```

- [ ] **Step 5:** Create `lessons/12-generics/solutions/warmup/maxg/maxg_test.go`:

```go
package maxg

import (
	"testing"
)

func TestMaxInts(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want int
	}{
		{"positive", []int{1, 5, 3, 2}, 5},
		{"negative", []int{-1, -5, -3}, -1},
		{"mixed", []int{-2, 0, 7, -100, 3}, 7},
		{"single", []int{42}, 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.in)
			if err != nil {
				t.Fatalf("Max(%v) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Max(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestMaxStrings(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"fruits", []string{"apple", "banana", "cherry"}, "cherry"},
		{"single", []string{"only"}, "only"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.in)
			if err != nil {
				t.Fatalf("Max(%v) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Max(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMaxEmpty(t *testing.T) {
	if _, err := Max([]int{}); err == nil {
		t.Fatal("Max([]int{}): expected error, got nil")
	}
	// nil slice — type inference can't see []int from nil alone, so we
	// have to be explicit. (This is also one of the "Common mistake"
	// slide examples in concept 3 / type inference.)
	if _, err := Max[int](nil); err == nil {
		t.Fatal("Max[int](nil): expected error, got nil")
	}
}
```

> Note on the empty-slice test: students often forget to test the nil-slice case separately. In Go, `len(nil) == 0`, so the same branch handles both — but the type-inference angle is pedagogically interesting (you can't write `Max(nil)`; the compiler has no T to infer).

> Note on `var zero T`: this is the idiomatic Go way to obtain the zero value of a type parameter. `T{}` doesn't compile (T might not be a struct); `*new(T)` works but is ugly. `var zero T` is what the stdlib does (see `cmp.Or` source).

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/12-generics/
go test ./lessons/12-generics/exercises/warmup/maxg/... -v 2>&1 | tail -10
go test ./lessons/12-generics/solutions/warmup/maxg/... -v 2>&1 | tail -20
make test
golangci-lint run ./...
go vet ./...

git add lessons/12-generics/exercises/warmup/ lessons/12-generics/solutions/warmup/
git commit -m "feat(lesson-12): warmup — maxg (generic Max[T cmp.Ordered])"
```

Expected: gofmt empty; exercises pass vacuously; solutions TestMaxInts (4 sub-tests) + TestMaxStrings (2 sub-tests) + TestMaxEmpty pass; make test green; lint 0 issues; vet clean.

---

## Task 3: Main — `slicesx` (Filter + Map)

The main package: `Filter[T any]` (one type parameter) and `Map[T, U any]` (two type parameters). Both unconstrained — students see that `any` is the most common case.

**Files (4 total — 2 per side):**

- [ ] **Step 1:** Create directories.

```bash
mkdir -p lessons/12-generics/{exercises,solutions}/slicesx
```

- [ ] **Step 2:** Create `lessons/12-generics/exercises/slicesx/slicesx.go`:

```go
// Package slicesx is a small slices-style helper library. It demonstrates
// Go generics: type parameters, unconstrained type parameters (any),
// multiple type parameters, and type inference at the call site.
//
// In real code you'd use the stdlib slices package (Go 1.21+) — it has
// Filter (via custom code: slices.DeleteFunc with inversion) and there's
// an "x/exp/slices.Map" community variant. We reimplement them here to
// teach the generics machinery from scratch.
//
// Both functions are unconstrained (T any, U any). The lesson teaches
// that MOST generic functions need only `any` — only operations that
// need <, ==, etc. require richer constraints (see maxg.Max).
package slicesx

// Filter returns a new slice containing the elements of xs for which
// pred returns true. The output preserves input order.
//
// The output slice is pre-allocated with capacity len(xs) — best case
// (no matches) wastes len(xs) ints; worst case (all match) does zero
// re-allocations. Reasonable tradeoff for unknown match rate.
//
// Examples:
//   Filter([]int{1, 2, 3, 4}, func(x int) bool { return x%2 == 0 }) → [2, 4]
//   Filter([]string{"a", "bb", "ccc"}, func(s string) bool { return len(s) > 1 }) → ["bb", "ccc"]
//
// Type inference: callers write Filter(xs, pred) — the compiler infers T
// from the element type of xs.
//
// Hint:
//   1. out := make([]T, 0, len(xs))
//   2. for _, x := range xs → if pred(x) → out = append(out, x)
//   3. return out
func Filter[T any](xs []T, pred func(T) bool) []T {
	panic("TODO: pre-allocate with cap len(xs); append elements where pred is true")
}

// Map returns a new slice where each element is the result of applying
// f to the corresponding input element. The output type U is
// (potentially) different from the input type T.
//
// This is where multiple type parameters earn their keep: the function
// has one input type and one output type, and they can be different.
//
// Examples:
//   Map([]int{1, 2, 3}, strconv.Itoa)              → ["1", "2", "3"]   // T=int, U=string
//   Map([]string{"hi", "world"}, strings.ToUpper)  → ["HI", "WORLD"]   // T=U=string
//   Map([]string{"hi", "world"}, len)              → [2, 5]             // T=string, U=int
//
// Output is pre-allocated to exactly len(xs) — we know the output size
// in advance.
//
// Type inference: callers write Map(xs, f) — both T and U are inferred
// from the function signature.
//
// Hint:
//   1. out := make([]U, len(xs))
//   2. for i, x := range xs → out[i] = f(x)
//   3. return out
func Map[T, U any](xs []T, f func(T) U) []U {
	panic("TODO: pre-allocate with size len(xs); transform each element")
}
```

- [ ] **Step 3:** Create `lessons/12-generics/exercises/slicesx/slicesx_test.go` (SKELETON):

```go
package slicesx

import (
	"strconv"
	"strings"
	"testing"
)

// TestFilterInts is a SKELETON. Cover at least:
//   - Filter even numbers from a mix of positives/negatives
//   - Filter with a predicate that matches nothing → empty (or nil)
//   - Filter empty input → empty
func TestFilterInts(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	cases := []struct {
		name string
		in   []int
		pred func(int) bool
		want []int
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Filter(tc.in, tc.pred); compare to tc.want.
			_ = tc
			_ = even
		})
	}
}

// TestFilterStrings is a SKELETON. Filter strings by length (e.g., keep
// only strings of length > 1). Demonstrates that the same Filter works
// with any element type — type inference picks T=string.
func TestFilterStrings(t *testing.T) {
	longerThan1 := func(s string) bool { return len(s) > 1 }
	// TODO: write one assertion. E.g.:
	//   got := Filter([]string{"a", "bb", "ccc"}, longerThan1)
	//   want := []string{"bb", "ccc"}
	//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
	_ = longerThan1
}

// TestMapIntToString is a SKELETON. Demonstrates Map with T=int, U=string.
// Use strconv.Itoa as the transform.
func TestMapIntToString(t *testing.T) {
	// TODO:
	//   got := Map([]int{1, 2, 3}, strconv.Itoa)
	//   want := []string{"1", "2", "3"}
	//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
	_ = strconv.Itoa
}

// TestMapStringToString is a SKELETON. Demonstrates Map with T=U=string.
// Use strings.ToUpper as the transform.
func TestMapStringToString(t *testing.T) {
	// TODO:
	//   got := Map([]string{"hi", "world"}, strings.ToUpper)
	//   want := []string{"HI", "WORLD"}
	_ = strings.ToUpper
}
```

- [ ] **Step 4:** Create `lessons/12-generics/solutions/slicesx/slicesx.go`:

```go
// Package slicesx is the lesson 12 main package reference implementation.
package slicesx

// Filter returns a new slice containing the elements of xs for which pred
// returns true. Preserves input order. Pre-allocates with cap len(xs).
func Filter[T any](xs []T, pred func(T) bool) []T {
	out := make([]T, 0, len(xs))
	for _, x := range xs {
		if pred(x) {
			out = append(out, x)
		}
	}
	return out
}

// Map returns a new slice where each element is f applied to the
// corresponding input element. Pre-allocates with exact size len(xs).
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}
```

- [ ] **Step 5:** Create `lessons/12-generics/solutions/slicesx/slicesx_test.go`:

```go
package slicesx

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestFilterInts(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	gt100 := func(x int) bool { return x > 100 }
	cases := []struct {
		name string
		in   []int
		pred func(int) bool
		want []int
	}{
		{"evens", []int{1, 2, 3, 4, 5, 6}, even, []int{2, 4, 6}},
		{"no-matches", []int{1, 2, 3}, gt100, []int{}},
		{"empty-input", []int{}, even, []int{}},
		{"all-match", []int{2, 4, 6}, even, []int{2, 4, 6}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.in, tc.pred)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Filter(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestFilterStrings(t *testing.T) {
	longerThan1 := func(s string) bool { return len(s) > 1 }
	got := Filter([]string{"a", "bb", "ccc"}, longerThan1)
	want := []string{"bb", "ccc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter strings = %v, want %v", got, want)
	}
}

func TestMapIntToString(t *testing.T) {
	got := Map([]int{1, 2, 3}, strconv.Itoa)
	want := []string{"1", "2", "3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map ints→strings = %v, want %v", got, want)
	}
}

func TestMapStringToString(t *testing.T) {
	got := Map([]string{"hi", "world"}, strings.ToUpper)
	want := []string{"HI", "WORLD"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map ToUpper = %v, want %v", got, want)
	}
}

func TestMapStringToInt(t *testing.T) {
	got := Map([]string{"hi", "world", ""}, func(s string) int { return len(s) })
	want := []int{2, 5, 0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map len = %v, want %v", got, want)
	}
}

func TestMapEmpty(t *testing.T) {
	got := Map([]int{}, strconv.Itoa)
	if len(got) != 0 {
		t.Errorf("Map([]) = %v, want empty", got)
	}
}
```

> Note on the `"no-matches"` test asserting `[]int{}` not `nil`: `Filter` returns `make([]T, 0, len(xs))` (always a non-nil empty slice). If students accidentally `var out []T` instead, the test catches it because `reflect.DeepEqual([]int{}, nil) == false`. Good safety net.

> Note on `strconv.Itoa`: a function value with signature `func(int) string` — passable directly as `f` in `Map`. Demonstrates that Go's first-class functions compose naturally with generics.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/12-generics/
go test ./lessons/12-generics/exercises/slicesx/... -v 2>&1 | tail -10
go test ./lessons/12-generics/solutions/slicesx/... -v 2>&1 | tail -30
make test
golangci-lint run ./...
go vet ./...

git add lessons/12-generics/exercises/slicesx/ lessons/12-generics/solutions/slicesx/
git commit -m "feat(lesson-12): main — slicesx (generic Filter + Map)"
```

Expected: gofmt empty; exercises pass vacuously; solutions all 5 tests + sub-tests pass; make test green; lint 0 issues; vet clean.

---

## Task 4: Slide deck — 5 concepts

Heavy-explanatory pattern: Motivation → Basics → Worked example → Common mistake → Recap per concept. Plus intro and recap slides.

**File:** `lessons/12-generics/slides/slides.md`

The deck has ~30 slides:

- 1 cover
- 1 intro / lesson roadmap
- ~5 slides × 5 concepts = ~25 concept slides
- 1 wrap-up / what's next
- 1 closing

**Concept order and structure:**

1. **Type Parameters** (Motivation: code duplication of L04's typed helpers / Basics: `func F[T any](x T) T` syntax / Worked: rewriting MinMaxInt+MinMaxFloat as a single Max[T cmp.Ordered] / Common-mistake: forgetting to pass the type-parameter brackets at declaration site — `func F(x T)` doesn't work, T isn't defined / Recap)
2. **Constraints** (Motivation: why can't I write `func Max[T any](xs []T) T` and use `>`? / Basics: `any` (no ops), `comparable` (==, !=), `cmp.Ordered` (<,>,<=,>=) / Worked: `Max[T cmp.Ordered]` compiles because `>` is allowed under Ordered / Common-mistake: trying to use `==` on a `T any` (compiler rejects; switch to `comparable`) / Recap)
3. **Type Inference** (Motivation: nobody wants to write `Filter[int](xs, pred)` every time / Basics: compiler infers T from argument types / Worked: `Filter(xs, even)` works because `xs []int` → T=int / Common-mistake: writing `Max[int](nil)` works but `Max(nil)` fails — type inference needs concrete element types, nil has none. ALSO: return-type-only generics like `func Zero[T any]() T` can't be inferred; caller must write `Zero[int]()` / Recap)
4. **Multiple Type Parameters** (Motivation: Map needs input type AND output type — they're independent / Basics: `func F[T, U any](...)` syntax / Worked: `Map([]int{1,2,3}, strconv.Itoa)` → T=int, U=string, both inferred / Common-mistake: assuming U is derivable from T alone (it isn't — the function signature `func(T) U` fixes the relationship) / Recap)
5. **When NOT to Use Generics** (Motivation: "if a function is called once with one type, don't bother" / Basics: the cognitive cost (readers parse `[T any]` machinery just to learn it's a one-shot helper); the compile-time cost (rare, but stencilling generic functions across many type instantiations bloats the binary in extreme cases); the API-design cost (generic APIs leak constraint complexity to callers) / Worked: a "bad" example — `func PrintTwice[T any](x T) { fmt.Println(x); fmt.Println(x) }` — generic for no reason; just take `any` / Common-mistake: reaching for generics first when refactoring duplicated code, instead of asking "does this need to scale to N types or just 2-3 known ones?" / Recap)

After concept 5, a "Closing thought" slide: the Go community's attitude. Quote from Russ Cox / the Go team's release notes ("we expect generics to be used sparingly").

- [ ] **Step 1:** Author the deck. Open with the cover slide (`# Lesson 12 — Generics` + subtitle + date), then a roadmap slide listing the 5 concepts.

- [ ] **Step 2:** Author concept 1 (Type Parameters):

```markdown
---

## Concept 1 / Type Parameters

### Motivation

In **lesson 4** we wrote two min/max helpers — one for `int`, one for
`float64`:

```go
func MinMaxInt(xs []int) (int, int) { ... }
func MinMaxFloat(xs []float64) (float64, float64) { ... }
```

The bodies are **identical** except for the types. Until Go 1.18, this
was the only option. Generics let us write the algorithm **once**.

---

### Basics

A **type parameter** is a placeholder for a type, chosen by the caller
(or inferred by the compiler):

```go
func Identity[T any](x T) T {
    return x
}

Identity(42)         // T=int
Identity("hello")    // T=string
Identity([]int{1,2}) // T=[]int
```

Syntax: `[T any]` after the function name. `any` is the constraint —
"any type at all". `T` is the parameter name (convention: single capital
letter for the first; `T, U, V, ...` for more).

---

### Worked example

A single generic Max replaces L04's pair of typed helpers:

```go
func Max[T cmp.Ordered](xs []T) (T, error) {
    var zero T
    if len(xs) == 0 {
        return zero, errors.New("maxg: empty slice")
    }
    m := xs[0]
    for _, x := range xs[1:] {
        if x > m { m = x }
    }
    return m, nil
}

Max([]int{3, 1, 5, 2})       // 5, nil  (T=int)
Max([]string{"b", "a", "c"}) // "c", nil (T=string)
```

`cmp.Ordered` is the constraint that allows `>` — covered next.

---

### Common mistake

Forgetting the type-parameter brackets at the **declaration site**:

```go
func Max(xs []T) (T, error) { ... }  // ❌ T is undefined
//        ↑↑↑
// "undefined: T" — you forgot [T cmp.Ordered]
```

Compiler error: ``undeclared name: T``. The fix is to add `[T cmp.Ordered]`
between the function name and the parameter list.

---

### Recap

- `func F[T any](x T) T` — `[T any]` declares the type parameter.
- The caller picks (or the compiler infers) which concrete type to use.
- The algorithm is written **once**; the compiler stencils it per type.
```

- [ ] **Step 3:** Author concept 2 (Constraints):

```markdown
---

## Concept 2 / Constraints

### Motivation

If `T` is "any type at all", the only operations you can do on a `T`
value are the ones that **every** type supports: assignment, passing as
an argument, comparison via `reflect.DeepEqual`. You can't write
`x + y` or `x > y` — the compiler doesn't know if `+` and `>` are
defined for the chosen `T`.

To unlock more operations, **constrain** `T` to a smaller set of types.

---

### Basics

Three constraints cover almost everything:

| Constraint | Operations unlocked | Source |
|---|---|---|
| `any` | None beyond assignment, function args, copying | builtin |
| `comparable` | `==`, `!=` | builtin |
| `cmp.Ordered` | `<`, `<=`, `>`, `>=`, plus `==`, `!=` | `cmp` (Go 1.21+) |

Examples:

```go
func IsEqual[T comparable](a, b T) bool { return a == b }
func Max[T cmp.Ordered](xs []T) (T, error) { ... }  // uses >
func Identity[T any](x T) T { return x }            // no ops needed
```

---

### Worked example

`Max` requires `>`. With `T any`, the compiler rejects:

```go
func Max[T any](xs []T) T {
    m := xs[0]
    for _, x := range xs[1:] {
        if x > m { m = x }  // ❌ invalid operation: x > m (type T does not support comparison)
    }
    return m
}
```

Change `any` to `cmp.Ordered` and it compiles — but now callers can only
pass slices of int / float / string types. `Max([]bool{})` won't
compile (bool isn't ordered).

---

### Common mistake

Reaching for `any` first and getting stuck. If you need ANY operation
beyond assignment, you need a constraint. The Go playground will tell
you exactly which:

```
invalid operation: x == y (incomparable types in type set)
  → switch to comparable
invalid operation: x > y (type T does not support comparison)
  → switch to cmp.Ordered
```

---

### Recap

- `any` is the loosest constraint — no ops beyond assignment.
- `comparable` adds `==` and `!=`.
- `cmp.Ordered` adds `<`, `>`, `<=`, `>=`.
- Custom constraints (interface unions) exist for advanced cases — we
  won't cover them; the stdlib's three constraints handle the 95%.
```

- [ ] **Step 4:** Author concept 3 (Type Inference):

```markdown
---

## Concept 3 / Type Inference

### Motivation

Generic code shouldn't be **noisier** at the call site than typed code.
Imagine if every caller had to write:

```go
Filter[int](xs, even)
Map[int, string](xs, strconv.Itoa)
```

That's worse than the duplicated typed versions we started with. Go's
**type inference** lets the compiler figure out `T` (and `U`) from the
arguments — so callers write `Filter(xs, even)` and `Map(xs, strconv.Itoa)`.

---

### Basics

The compiler walks the argument list and unifies the types:

```go
Filter(xs, even)
//     ↑    ↑
//   []int  func(int) bool
//     ↓
//   T = int   ← inferred
```

If at least one argument fixes `T`, the compiler succeeds. If no
argument constrains `T`, you must write it explicitly.

---

### Worked example

Three cases:

```go
xs := []int{1, 2, 3}

Filter(xs, even)              // T inferred from xs ✓
Map(xs, strconv.Itoa)         // T from xs, U from strconv.Itoa's return ✓
Max[int](nil)                 // nil has no element type — must write [int]
```

The third case: `nil` is a typed-nil for **any** slice type. The
compiler can't pick. The explicit `[int]` fixes it.

---

### Common mistake

**Return-type-only generics** can't be inferred:

```go
func Zero[T any]() T {
    var z T
    return z
}

x := Zero()       // ❌ cannot infer T
x := Zero[int]()  // ✓ T=int, x is int 0
```

The compiler has nothing to look at — no arguments with types — so it
gives up. You **must** write `Zero[int]()`. (Other languages call these
"phantom type parameters". They show up rarely in Go.)

The other common mistake: writing the explicit form when inference
would work. `Filter[int](xs, even)` is verbose; `Filter(xs, even)` is
the idiomatic call.

---

### Recap

- The compiler infers type parameters from argument types.
- If no argument fixes `T`, you must write it explicitly: `Func[Type](args)`.
- `nil` carries no type → can't be used for inference alone.
- Idiomatic Go: write the explicit form **only** when inference fails.
```

- [ ] **Step 5:** Author concept 4 (Multiple Type Parameters):

```markdown
---

## Concept 4 / Multiple Type Parameters

### Motivation

`Filter` has one type parameter — the element type stays the same after
filtering. `Map` is different: it **transforms** the type. Input is
`[]T`; output is `[]U`. The two types are independent.

This is where multiple type parameters pay off.

---

### Basics

Declare multiple type parameters in one bracket pair:

```go
func F[T, U any](x T) U { ... }
```

The compiler infers each independently from arguments. For `Map`, `T`
comes from `xs` and `U` from `f`'s return type:

```go
func Map[T, U any](xs []T, f func(T) U) []U { ... }

Map([]int{1, 2}, strconv.Itoa)
//    ↑              ↑
//   []int           func(int) string
//     ↓             ↓        ↓
//   T = int        T = int   U = string
```

---

### Worked example

The three canonical Map shapes:

```go
// T → U (different types)
Map([]int{1, 2, 3}, strconv.Itoa)            // []string{"1","2","3"}

// T → T (same type)
Map([]string{"hi", "ok"}, strings.ToUpper)   // []string{"HI","OK"}

// T → U (different types again)
Map([]string{"hi", "world"}, func(s string) int { return len(s) })
                                              // []int{2, 5}
```

The function passed as `f` determines `U`. The compiler stitches it all
together.

---

### Common mistake

Assuming `U` is derivable from `T` alone. It isn't — `U` is whatever
the **function** `f` returns. If you write:

```go
Map([]int{1, 2}, func(x int) any { return x })  // U = any
```

…you get back `[]any`, not `[]int`. The compiler trusts your function
signature.

---

### Recap

- `func F[T, U any](...)` declares two type parameters.
- Each is inferred from its own argument source.
- The output type of `Map` is whatever the transform function returns.
```

- [ ] **Step 6:** Author concept 5 (When NOT to use generics):

```markdown
---

## Concept 5 / When NOT to Use Generics

### Motivation

Generics are powerful. They're also **easy to overuse**. Go's standard
library uses them sparingly — many obvious "generic" places (e.g., the
`map` and `slice` builtins) were intentionally kept as language features
rather than user-written generics. The Go team's official guidance:

> Use generics for code that operates on multiple types in the same way.
> If your function works on one type, write it for that type.

---

### Basics

Three costs of generics:

1. **Cognitive complexity.** Readers parse `[T cmp.Ordered]` machinery
   before reaching the logic. Worth it for `Max`; wasteful for a one-off
   helper.
2. **API surface.** Generic functions leak constraint complexity to
   callers. `Max[T cmp.Ordered]` documents "only ordered types" — fine.
   A constraint chain involving custom interface unions for a one-time
   helper is overkill.
3. **Compile time / binary size.** The compiler generates code per
   distinct type instantiation. Rarely a problem; matters in
   performance-critical or embedded contexts.

---

### Worked example

A "bad" generic:

```go
func PrintTwice[T any](x T) {
    fmt.Println(x)
    fmt.Println(x)
}
```

What's `T any` buying us? `fmt.Println` accepts `any` already. Just
write:

```go
func PrintTwice(x any) {
    fmt.Println(x)
    fmt.Println(x)
}
```

Identical behaviour. No generics. Less to read.

---

### Common mistake

Reaching for generics **first** when refactoring duplicated typed code.
Better order:

1. Is the duplicated code called with **N (≥3) distinct types**, where
   N is likely to grow? Generics make sense.
2. Is it called with 2-3 fixed types, unlikely to grow? Keep them
   separate, or write a tiny `any` + type switch.
3. Is it called with 1 type, and the duplication is just because it
   *looks* generic-ish? Don't bother.

The Go community's reflex: **prefer concrete types until generics are
clearly justified**.

---

### Recap

- Generics aren't free — cognitive cost, API surface, occasionally
  compile/runtime cost.
- "If a function is called once with one type, don't bother."
- Lead with concrete types; reach for generics when 3+ types and growth
  argue for it.
```

- [ ] **Step 7:** Add a closing thought slide:

```markdown
---

## Closing Thought

The Go team's release notes for 1.18 (the first version with generics)
included this:

> We expect generics to be used relatively sparingly. Most Go programs
> will continue to be written without them.

Three years later (Go 1.21 added `cmp.Ordered`; `slices` and `maps`
packages went stable in 1.21), the prediction has mostly held. Generics
power a small set of high-leverage stdlib packages (`slices`, `maps`,
`cmp`) and otherwise stay out of the way.

Use them when they earn their keep. Skip them when they don't.

---

## What's Next

Lesson 13: **Encoding & I/O**. We crack open the `storage` package
(which has been a black box since L08), formally meet `io.Reader` /
`io.Writer`, and build a CSV-to-expenses importer.

The tracker is back.
```

- [ ] **Step 8:** Verify slides build:

```bash
make slides-build
grep -q "12-generics" dist/index.html && echo "✓ 12-generics in index"
rm -rf dist
```

- [ ] **Step 9:** Commit:

```bash
git add lessons/12-generics/slides/
git commit -m "feat(lesson-12): slides — Generics (5 concepts)"
```

---

## Task 5: README

Mirrors slide concepts (5 sections); one Common-mistake paragraph per concept; gofmt/vet daily-habits block; "What's different" section.

**File:** `lessons/12-generics/README.md`

- [ ] **Step 1:** Author the README. Template:

```markdown
# Lesson 12 — Generics

## What you'll learn

By the end of this lesson you can:

- Write generic functions with type parameters: `func F[T any](x T) T`.
- Choose the right constraint: `any`, `comparable`, or `cmp.Ordered`.
- Let the compiler infer type parameters (and recognise when you must
  write them explicitly).
- Use multiple type parameters (`Map[T, U any]`) to transform between
  types.
- Know when **not** to use generics — the Go community's restraint.

## What's different from L11

L11 added rich error handling (`ErrNotFound`, `ParseError`) to the
tracker. Lesson 12 is **standalone** — no tracker code. Generics are
about mechanics, not domain. The Go team's official advice is "use
generics sparingly", and forcing them into the tracker would be the
exact premature-abstraction anti-pattern the Go team warns about.

The tracker returns in **L13** (Encoding & I/O) — we'll crack open the
storage package and build a CSV importer.

## The package layout

```
lessons/12-generics/
├── exercises/
│   ├── warmup/maxg/   ← you implement: Max[T cmp.Ordered]
│   └── slicesx/        ← you implement: Filter[T any], Map[T, U any]
└── solutions/          ← reference implementations
```

In a real library you'd merge `maxg` into `slicesx` — they're sibling
helpers. We split them so you implement `Max` first as a self-contained
warm-up, then tackle the two-function main package.

---

## Concept 1 — Type Parameters

A type parameter is a placeholder for a type, chosen by the caller (or
inferred by the compiler):

```go
func Identity[T any](x T) T {
    return x
}

Identity(42)         // T=int
Identity("hello")    // T=string
```

The syntax: `[T any]` after the function name. `T` is the parameter
name (convention: single capital letter); `any` is the constraint
("any type at all" — we'll constrain it tighter when we need to).

This single declaration replaces what — pre-Go-1.18 — needed a separate
function per type, or `interface{}` (now `any`) with runtime type
assertions.

### Common mistake

Forgetting the brackets at the **declaration site**:

```go
func Max(xs []T) T { ... }  // ❌ "undefined: T"
```

The compiler doesn't know what `T` is. Add `[T cmp.Ordered]` between
the function name and the parameter list.

---

## Concept 2 — Constraints

A constraint defines the set of types `T` is allowed to be — and
therefore the operations you can perform on a `T` value:

| Constraint     | Operations             | Source        |
|----------------|------------------------|---------------|
| `any`          | Assignment only        | builtin       |
| `comparable`   | `==`, `!=`             | builtin       |
| `cmp.Ordered`  | `<`, `>`, `<=`, `>=`   | `cmp` (1.21+) |

If you try to use `>` inside a function with `T any`, the compiler
rejects it:

```
invalid operation: x > y (type T does not support comparison)
```

Switch the constraint to `cmp.Ordered` and it compiles — but callers
now can only pass slices of int / float / string types.

### Common mistake

Picking `any` first and getting stuck. The compiler error tells you
exactly which constraint you need — read it and lift the constraint to
the smallest one that supports your ops.

---

## Concept 3 — Type Inference

Calling generic functions should feel like calling regular functions:

```go
xs := []int{1, 2, 3}
Filter(xs, even)        // T=int, inferred from xs
Map(xs, strconv.Itoa)   // T=int, U=string, both inferred
```

You don't write `Filter[int](xs, even)` unless the compiler can't
figure `T` out from the arguments.

### When inference fails

Two situations:

1. **`nil` carries no type.** `Max(nil)` can't be inferred — `nil` is a
   typed-nil for *any* slice type. Write `Max[int](nil)` explicitly.
2. **Return-type-only generics.** `func Zero[T any]() T` has no
   argument-side type to look at. Callers **must** write `Zero[int]()`.

### Common mistake

Writing the explicit form when inference would work:

```go
Filter[int](xs, even)   // ❌ verbose
Filter(xs, even)        // ✓ idiomatic
```

Lean on inference; reach for the explicit form only when the compiler
makes you.

---

## Concept 4 — Multiple Type Parameters

When the input type and output type are different — like `Map`
transforming `[]int` to `[]string` — you need two type parameters:

```go
func Map[T, U any](xs []T, f func(T) U) []U {
    out := make([]U, len(xs))
    for i, x := range xs {
        out[i] = f(x)
    }
    return out
}

Map([]int{1, 2, 3}, strconv.Itoa)  // []string{"1","2","3"}
```

Each parameter is inferred independently: `T` from `xs`'s element type,
`U` from `f`'s return type.

### Common mistake

Assuming `U` follows from `T`. It doesn't — `U` is whatever the
function `f` returns. If you write `func(x int) any`, you get back
`[]any`. The compiler trusts the function signature.

---

## Concept 5 — When NOT to Use Generics

The Go team's release notes for 1.18 said:

> We expect generics to be used relatively sparingly. Most Go programs
> will continue to be written without them.

Three costs:

1. **Cognitive complexity.** Readers parse `[T cmp.Ordered]` machinery
   before reaching the logic.
2. **API surface.** Generic functions leak constraint complexity to
   callers.
3. **Compile time / binary size.** Rarely a problem; can matter in
   embedded or performance-critical code.

Rule of thumb: **if a function is called once with one type, don't
bother**. Generics earn their keep when 3+ types are likely now or
later. Until then, prefer concrete types.

### Common mistake

Reaching for generics **first** when refactoring duplicated typed code.
Better order:

1. Called with N≥3 distinct types, growth likely? → Generics.
2. Called with 2-3 fixed types, no growth? → Keep separate or use
   `any` + type switch.
3. Called with 1 type? → Don't bother.

---

## Daily habits

After every change:

```bash
gofmt -w ./...           # auto-format
go vet ./...             # catch shadowed vars, wrong format verbs, etc.
go test ./...            # run the suite
```

The compiler will already catch generics mistakes (constraint
mismatches, inference failures) at compile time — generics in Go are
designed to fail early.

---

## Run the tests

```bash
# Skeleton (passes vacuously until you implement)
go test ./lessons/12-generics/exercises/...

# Reference solution
go test ./lessons/12-generics/solutions/...
```
```

- [ ] **Step 2:** Commit:

```bash
git add lessons/12-generics/README.md
git commit -m "docs(lesson-12): README — Generics self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1:** Full test sweep.

```bash
make test
```

Expected: all lessons (01-12) + tools pass.

- [ ] **Step 2:** Solution-only verbose test for lesson 12.

```bash
go test -v ./lessons/12-generics/solutions/...
```

Expected:
- `maxg.TestMaxInts` — 4 sub-tests PASS
- `maxg.TestMaxStrings` — 2 sub-tests PASS
- `maxg.TestMaxEmpty` PASS
- `slicesx.TestFilterInts` — 4 sub-tests PASS
- `slicesx.TestFilterStrings` PASS
- `slicesx.TestMapIntToString` PASS
- `slicesx.TestMapStringToString` PASS
- `slicesx.TestMapStringToInt` PASS
- `slicesx.TestMapEmpty` PASS

- [ ] **Step 3:** Exercise tests pass vacuously.

```bash
go test ./lessons/12-generics/exercises/...
```

Expected: ok (no sub-tests run — case slices are empty).

- [ ] **Step 4:** Static analysis clean.

```bash
go vet ./...
golangci-lint run ./...
gofmt -l lessons/12-generics/
```

Expected: zero output from gofmt; 0 issues from lint; vet clean.

- [ ] **Step 5:** Slides build smoke.

```bash
make slides-build
grep -q "12-generics" dist/index.html
rm -rf dist
```

Expected: `built dist`; grep finds; cleanup succeeds.

- [ ] **Step 6:** (No binary smoke test — L12 has no cmd binary by design.)

- [ ] **Step 7:** Final code review subagent → push branch → create PR.

```bash
git push -u origin feature/plan-o-lesson-12-generics
gh pr create --title "feat(lesson-12): Generics — type params, constraints, inference, restraint" --body "..."
```

---

## Verification

After all 6 tasks:

```bash
make test                                            # green across all lessons
go test -v ./lessons/12-generics/solutions/... 2>&1 | grep -E "PASS|FAIL"   # 5 tests, ~12 sub-tests, all PASS
go vet ./...                                         # clean
golangci-lint run ./...                              # 0 issues
gofmt -l lessons/12-generics/                        # empty
make slides-build && grep -q "12-generics" dist/index.html && rm -rf dist  # green
```

## Critical file paths

To be created:

- `lessons/12-generics/` (directory)
- `lessons/12-generics/README.md`
- `lessons/12-generics/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/12-generics/exercises/warmup/maxg/{maxg.go, maxg_test.go}`
- `lessons/12-generics/exercises/slicesx/{slicesx.go, slicesx_test.go}`
- `lessons/12-generics/solutions/...` (mirrored)

To be referenced (not modified):

- `docs/superpowers/specs/2026-05-18-phase-2-idiomatic-go-design.md` (Lesson 12 section — locked the standalone shape)
- `.golangci.yml` (no changes; existing exclusion `lessons/.*/exercises/` covers nested paths)
- `lessons/11-errors/...` (lesson 11 stays untouched; L12 introduces no carry-forward)
