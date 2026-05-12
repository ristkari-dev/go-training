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
