# Lesson 12: Generics

## What you'll learn

By the end of this lesson you can:

- Write generic functions with type parameters: `func F[T any](x T) T`.
- Choose the right constraint: `any`, `comparable`, or `cmp.Ordered`.
- Let the compiler infer type parameters (and recognise when you must write them explicitly).
- Use multiple type parameters (`Map[T, U any]`) to transform between types.
- Know when **not** to use generics — the Go community's restraint.

## What's different from L11

L11 added rich error handling (`ErrNotFound`, `ParseError`) to the tracker. Lesson 12 is **standalone** — no tracker code, no `expense` package, no cmd binary. Generics are about mechanics, not domain. The Go team's official advice is "use generics sparingly", and forcing them into the tracker would be the exact premature-abstraction anti-pattern the Go team warns about.

The tracker returns in **L13** (Encoding & I/O) — we'll crack open the storage package and build a CSV importer.

## The package layout

```
lessons/12-generics/
├── exercises/
│   ├── warmup/maxg/   ← you implement: Max[T cmp.Ordered]
│   └── slicesx/        ← you implement: Filter[T any], Map[T, U any]
└── solutions/          ← reference implementations
```

In a real library you'd merge `maxg` into `slicesx` — they're sibling helpers. We split them so you implement `Max` first as a self-contained warm-up, then tackle the two-function main package. In real-world Go you'd probably also reach for the stdlib's `slices.Max` (Go 1.21+) rather than rolling your own.

---

## Concept 1 — Type Parameters

A **type parameter** is a placeholder for a type, chosen by the caller (or inferred by the compiler):

```go
func Identity[T any](x T) T {
    return x
}

Identity(42)         // T=int
Identity("hello")    // T=string
Identity([]int{1,2}) // T=[]int
```

The syntax: `[T any]` after the function name. `T` is the parameter name (convention: single capital letter — `T`, `U`, `V`, ...); `any` is the constraint ("any type at all" — we'll constrain it tighter when we need to).

This single declaration replaces what — pre-Go-1.18 — needed either a separate function per type (the lesson-04 approach: `MinMaxInt`, `MinMaxFloat`) or `interface{}` (now `any`) with runtime type assertions. Generics give you the typed code with single-source.

A useful idiom for getting the zero value of a type parameter: `var zero T`. You can't write `T{}` (T might not be a struct), and `*new(T)` works but is ugly. `var zero T` is what the stdlib does.

### Common mistake

Forgetting the brackets at the **declaration site**:

```go
func Max(xs []T) T { ... }  // ❌ "undefined: T"
```

The compiler doesn't know what `T` is. Add `[T cmp.Ordered]` between the function name and the parameter list:

```go
func Max[T cmp.Ordered](xs []T) (T, error) { ... }  // ✓
```

---

## Concept 2 — Constraints

A constraint defines the set of types `T` is allowed to be — and therefore the operations you can perform on a `T` value:

| Constraint     | Operations             | Source        |
|----------------|------------------------|---------------|
| `any`          | Assignment, copy       | builtin       |
| `comparable`   | `==`, `!=`             | builtin       |
| `cmp.Ordered`  | `<`, `>`, `<=`, `>=`   | `cmp` (1.21+) |

If you try to use `>` inside a function with `T any`, the compiler rejects it:

```
invalid operation: x > y (type T does not support comparison)
```

Switch the constraint to `cmp.Ordered` and it compiles — but callers now can only pass slices of int / float / string types. `Max([]bool{})` won't compile because `bool` isn't ordered.

The `cmp` package was added in Go 1.21. Before that, you needed `golang.org/x/exp/constraints`. The course is on Go 1.23, so `cmp.Ordered` is the stdlib path.

### Common mistake

Picking `any` first and getting stuck on operations. The compiler error tells you exactly which constraint you need — read it and lift the constraint to the **smallest** one that supports your ops. Don't over-constrain (`cmp.Ordered` when `comparable` would do); don't under-constrain (`any` when you need `==`).

---

## Concept 3 — Type Inference

Calling generic functions should feel like calling regular functions:

```go
xs := []int{1, 2, 3}
Filter(xs, even)        // T=int, inferred from xs
Map(xs, strconv.Itoa)   // T=int, U=string, both inferred
```

You don't write `Filter[int](xs, even)` unless the compiler can't figure `T` out from the arguments. The compiler walks the argument list and unifies the types.

### When inference fails

Two situations:

1. **`nil` carries no type.** `Max(nil)` can't be inferred — `nil` is a typed-nil for *any* slice type. Write `Max[int](nil)` explicitly. (This is why the warmup's `TestMaxEmpty` writes `Max[int](nil)`.)
2. **Return-type-only generics.** `func Zero[T any]() T` has no argument-side type to look at. Callers **must** write `Zero[int]()`.

### Common mistake

Writing the explicit form when inference would work:

```go
Filter[int](xs, even)   // ❌ verbose
Filter(xs, even)        // ✓ idiomatic
```

Lean on inference; reach for the explicit form only when the compiler makes you.

---

## Concept 4 — Multiple Type Parameters

When the input type and output type are different — like `Map` transforming `[]int` to `[]string` — you need two type parameters:

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

Each parameter is inferred independently: `T` from `xs`'s element type, `U` from `f`'s return type. The function passed as `f` determines `U` — the compiler trusts the signature.

You can have more than two type parameters (`func H[A, B, C any](...)`). In practice, two is the most common; three is rare; four is a code smell.

### Common mistake

Assuming `U` follows from `T`. It doesn't — `U` is whatever the function `f` returns. If you write `func(x int) any`, you get back `[]any`. The compiler trusts the function signature; if you want `[]int` out, the function's return type must be `int`.

---

## Concept 5 — When NOT to Use Generics

The Go team's release notes for 1.18 said:

> We expect generics to be used relatively sparingly. Most Go programs will continue to be written without them.

Three costs:

1. **Cognitive complexity.** Readers parse `[T cmp.Ordered]` machinery before reaching the logic.
2. **API surface.** Generic functions leak constraint complexity to callers.
3. **Compile time / binary size.** Rarely a problem; can matter in embedded or performance-critical code.

Rule of thumb: **if a function is called once with one type, don't bother**. Generics earn their keep when 3+ types are likely now or later. Until then, prefer concrete types.

### Common mistake

Reaching for generics **first** when refactoring duplicated typed code. Better order:

1. Called with N≥3 distinct types, growth likely? → Generics.
2. Called with 2-3 fixed types, no growth? → Keep separate or use `any` + type switch.
3. Called with 1 type? → Don't bother.

The Go community's reflex: **prefer concrete types until generics are clearly justified**.

---

## Exercise: warm-up — `maxg`

Implement `Max[T cmp.Ordered](xs []T) (T, error)` in `exercises/warmup/maxg/maxg.go`. Empty slice returns the zero value of `T` plus a non-nil error; non-empty returns the largest element plus `nil`. Tests cover positive ints, negative ints, strings (lexicographic compare), and the empty-slice error case.

**Time:** 5-10 minutes.

## Exercise: main — `slicesx`

Implement `Filter[T any]` and `Map[T, U any]` in `exercises/slicesx/slicesx.go`. Both pre-allocate their output slices (Filter with `cap(0, len(xs))`; Map with exact `len(xs)`). Tests cover element filtering on `[]int` and `[]string`, plus `Map` from `int→string`, `string→string`, and `string→int`.

**Time:** 15-25 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...           # auto-format
go vet ./...             # catch shadowed vars, wrong format verbs, etc.
go test ./...            # run the suite
```

The compiler will already catch most generics mistakes (constraint mismatches, inference failures) at compile time — generics in Go are designed to fail early. Trust the compiler; read its errors carefully.

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/12-generics/exercises
go test ./...

# Or run a specific subpackage
go test ./warmup/maxg -v
go test ./slicesx -v

# Reference solution
go test ./lessons/12-generics/solutions/... -v
```

## Going further

### Read

- **Go blog — "Why Generics" (2019)** — Ian Lance Taylor's pre-1.18 motivation and constraints walkthrough: <https://go.dev/blog/why-generics>
- **Go 1.18 release notes — generics section** — what shipped: <https://go.dev/doc/go1.18#generics>
- **`slices` package docs** — see how the stdlib actually uses generics: <https://pkg.go.dev/slices>

### Try

- **Reduce.** Add `Reduce[T, U any](xs []T, init U, f func(U, T) U) U` to your `slicesx` — a left-fold. Test with sum-of-ints (`func(acc, x int) int { return acc + x }`) and string-concat (`func(acc, s string) string { return acc + s }`).
- **Zip.** Add `Zip[T, U any](xs []T, ys []U) []Pair[T, U]` where `type Pair[T, U any] struct { First T; Second U }`. Decide what to do if `len(xs) != len(ys)` — truncate, error, or pad?
