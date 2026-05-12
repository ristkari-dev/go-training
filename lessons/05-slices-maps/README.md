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
