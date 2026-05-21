<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">12</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Generics</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn Go generics — type parameters with <code>func F[T any](x T) T</code>, constraints (<code>any</code> / <code>comparable</code> / <code>cmp.Ordered</code>), type inference at the call site, multiple type parameters (<code>Map[T, U]</code>), and Go's culture of restraint: "if a function is called once with one type, don't bother."</p>
</div>
</div>
</div>

---

## What we'll cover

- **Type parameters** — `func F[T any](x T) T`; the algorithm written once, used with any type.
- **Constraints** — `any`, `comparable`, `cmp.Ordered`; the operations each unlocks.
- **Type inference** — the compiler figures out `T` from the args; when it can't.
- **Multiple type parameters** — `Map[T, U]`; input and output types are independent.
- **When NOT to use generics** — Go's restraint; "if a function is called once with one type, don't bother."

---

## Concept 1: Type parameters

### Motivation

In **lesson 4** we wrote two min/max helpers — one for `int`, one for `float64`:

```go
func MinMaxInt(xs []int) (int, int) { ... }
func MinMaxFloat(xs []float64) (float64, float64) { ... }
```

The bodies are **identical** except for the types. Until Go 1.18, this was the only option. Generics let us write the algorithm **once** and use it with any element type.

---

### The basics

```go
package main

import "fmt"

func Identity[T any](x T) T {
	return x
}

func main() {
	fmt.Println(Identity(42))          // T=int   — 42
	fmt.Println(Identity("hello"))     // T=string — hello
	fmt.Println(Identity([]int{1, 2})) // T=[]int  — [1 2]
}
```

Three things to internalise:

- **`[T any]` after the function name** declares a type parameter. `T` is the placeholder; `any` is the constraint ("any type at all").
- **The caller picks (or the compiler infers) which concrete type to use.** One function, many callers, no duplication.
- **The compiler stencils per type.** Internally, the compiler generates specialised code for each type instantiation — but you write the algorithm once.

---

### A worked example

Lesson 12's `maxg.Max` — a single generic function replacing L04's pair of typed helpers:

```go
package maxg

import (
	"cmp"
	"errors"
)

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

Used:

```go
Max([]int{3, 1, 5, 2})       // 5, nil  (T=int — inferred)
Max([]string{"b", "a", "c"}) // "c", nil (T=string — inferred)
```

`cmp.Ordered` is the constraint that allows `>` — covered in concept 2. The `var zero T` idiom gives you the zero value for the inferred type parameter.

---

### Common mistake

Forgetting the brackets at the **declaration site**:

```go
// the wrong way
func Max(xs []T) (T, error) { ... }  // ❌ undefined: T
```

The compiler doesn't know what `T` is. Fix: add `[T cmp.Ordered]` between the function name and the parameter list.

```go
func Max[T cmp.Ordered](xs []T) (T, error) { ... }  // ✓
```

The brackets must come **before** the regular parameter list — they're part of the function's signature.

---

### Recap

- `func F[T any](x T) T` — `[T any]` declares the type parameter.
- The caller picks the concrete type (or the compiler infers it).
- The algorithm is written **once**; the compiler stencils per type.
- `var zero T` gives the zero value of a type parameter.

---

## Concept 2: Constraints

### Motivation

If `T` is "any type at all" (`any`), the only operations you can do on a `T` value are the ones that **every** type supports: assignment, passing as an argument, copying. You can't write `x + y` or `x > y` — the compiler doesn't know if those operators are defined for the chosen `T`.

To unlock more operations, **constrain** `T` to a smaller set of types.

---

### The basics

Three constraints cover almost everything:

| Constraint     | Operations unlocked              | Source        |
|----------------|----------------------------------|---------------|
| `any`          | Assignment, args, copy (nothing else) | builtin       |
| `comparable`   | `==`, `!=`                       | builtin       |
| `cmp.Ordered`  | `<`, `<=`, `>`, `>=` (plus `==`, `!=`) | `cmp` (1.21+) |

```go
func Identity[T any](x T) T          { return x }         // no ops needed
func IsEqual[T comparable](a, b T) bool { return a == b } // uses ==
func Max[T cmp.Ordered](xs []T) T    { ... }              // uses >
```

The `cmp` package was added in Go 1.21 — before that you needed `golang.org/x/exp/constraints`. We're on Go 1.23, so `cmp.Ordered` is in the stdlib.

---

### A worked example

`Max` requires `>`. With `T any`, the compiler rejects:

```go
func Max[T any](xs []T) T {
	m := xs[0]
	for _, x := range xs[1:] {
		if x > m { // ❌ invalid operation: x > m (type T does not support comparison)
			m = x
		}
	}
	return m
}
```

Change `any` to `cmp.Ordered` and it compiles — but now callers can only pass slices of int / float / string types:

```go
func Max[T cmp.Ordered](xs []T) T { ... } // ✓

Max([]int{1, 2, 3})       // ✓ int is Ordered
Max([]string{"a", "b"})   // ✓ string is Ordered
Max([]bool{true, false})  // ❌ bool isn't Ordered — won't compile
```

The constraint is enforced at the **call site**: the compiler verifies that the type-argument satisfies the constraint.

---

### Common mistake

Reaching for `any` first and getting stuck on operations. The compiler error tells you exactly which constraint you need:

```
invalid operation: x == y (incomparable types in type set)
  → switch to comparable

invalid operation: x > y (type T does not support comparison)
  → switch to cmp.Ordered
```

Read the error and lift the constraint to the **smallest** one that supports your ops. Don't over-constrain (`cmp.Ordered` when `comparable` would do); don't under-constrain (`any` when you need `==`).

---

### Recap

- `any` is the loosest constraint — no ops beyond assignment.
- `comparable` adds `==` and `!=`.
- `cmp.Ordered` (Go 1.21+ stdlib) adds `<`, `>`, `<=`, `>=`.
- The compiler enforces constraints at the call site.

---

## Concept 3: Type inference

### Motivation

Generic code shouldn't be **noisier** at the call site than typed code. Imagine if every caller had to write:

```go
Filter[int](xs, even)
Map[int, string](xs, strconv.Itoa)
```

That's worse than the duplicated typed versions we started with. Go's **type inference** lets the compiler figure out `T` (and `U`) from the arguments — so callers just write `Filter(xs, even)` and `Map(xs, strconv.Itoa)`.

---

### The basics

The compiler walks the argument list and unifies the types:

```go
xs := []int{1, 2, 3}
even := func(x int) bool { return x%2 == 0 }

Filter(xs, even)
//     ↑    ↑
//   []int  func(int) bool
//     ↓
//   T = int   ← inferred
```

If at least one argument fixes `T`, inference succeeds. If no argument constrains `T`, you **must** write it explicitly.

```go
Filter[int](xs, even)  // explicit — works but verbose
Filter(xs, even)       // implicit — idiomatic
```

---

### A worked example

Three cases — first two work; the third forces you to be explicit:

```go
xs := []int{1, 2, 3}

// (1) Element type fixes T
Filter(xs, even)                  // T inferred = int ✓

// (2) Two parameters, both inferred
Map(xs, strconv.Itoa)             // T from xs, U from strconv.Itoa's return ✓
                                  // T=int, U=string

// (3) nil has no element type
Max(nil)                          // ❌ cannot infer T — nil is typed-nil for any slice
Max[int](nil)                     // ✓ explicit
```

`nil` is a typed-nil for **any** slice type, so the compiler can't pick. The explicit `[int]` fixes it. This is why the warmup's `TestMaxEmpty` writes `Max[int](nil)`.

---

### Common mistake

**Return-type-only generics** — where `T` appears only in the return type — can't be inferred:

```go
// the wrong way (well, it works, but callers always have to be explicit)
func Zero[T any]() T {
	var z T
	return z
}

x := Zero()       // ❌ cannot infer T
x := Zero[int]()  // ✓ T=int, x is int 0
```

The compiler has nothing to look at — no arguments with types — so it gives up. You **must** write `Zero[int]()`. Other languages call these "phantom type parameters". They show up rarely in Go.

The opposite mistake: writing the explicit form when inference would work. `Filter[int](xs, even)` is verbose; `Filter(xs, even)` is the idiomatic call.

---

### Recap

- The compiler infers type parameters from argument types.
- If no argument fixes `T`, you must write it explicitly: `Func[Type](args)`.
- `nil` carries no type → can't be used for inference alone.
- Idiomatic Go: write the explicit form **only** when inference fails.

---

## Concept 4: Multiple type parameters

### Motivation

`Filter` has one type parameter — the element type stays the same after filtering. `Map` is different: it **transforms** the type. Input is `[]T`; output is `[]U`. The two types are independent.

This is where multiple type parameters pay off.

---

### The basics

Declare multiple type parameters in one bracket pair:

```go
func F[T, U any](x T) U { ... }
```

The compiler infers each independently from its source. For `Map`, `T` comes from `xs` and `U` from `f`'s return type:

```go
func Map[T, U any](xs []T, f func(T) U) []U { ... }

Map([]int{1, 2}, strconv.Itoa)
//    ↑              ↑
//   []int           func(int) string
//     ↓             ↓        ↓
//   T = int        T = int   U = string
```

You can have more than two: `func H[A, B, C any](...)`. In practice, two is the most common; three is rare; four is a code smell.

---

### A worked example

The three canonical `Map` shapes:

```go
// T → U (different types)
Map([]int{1, 2, 3}, strconv.Itoa)            // []string{"1","2","3"}

// T → T (same type)
Map([]string{"hi", "ok"}, strings.ToUpper)   // []string{"HI","OK"}

// T → U (different types again)
Map([]string{"hi", "world"}, func(s string) int { return len(s) })
//                                              → []int{2, 5}
```

The function passed as `f` determines `U`. The compiler stitches it all together — no manual type-juggling.

Implementation (from `slicesx/slicesx.go`):

```go
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}
```

Pre-allocates the output to exactly `len(xs)` — we know the output size in advance, so we can skip the `append`-and-grow dance.

---

### Common mistake

Assuming `U` is derivable from `T` alone. It isn't — `U` is whatever the **function** `f` returns. If you write:

```go
// the wrong assumption
Map([]int{1, 2}, func(x int) any { return x })  // U = any
```

…you get back `[]any`, not `[]int`. The compiler trusts the function signature. If you want `[]int` out, the function's return type must be `int`.

---

### Recap

- `func F[T, U any](...)` declares two type parameters.
- Each is inferred from its own argument source.
- The output type of `Map` is whatever the transform function returns.
- More than 2 type params is rare; more than 3 is a code smell.

---

## Concept 5: When NOT to use generics

### Motivation

Generics are powerful. They're also **easy to overuse**. Go's standard library uses them sparingly — many obvious "generic" places (the `map` and `slice` builtins) were intentionally kept as language features rather than user-written generics. The Go team's release notes for 1.18 said:

> We expect generics to be used relatively sparingly. Most Go programs will continue to be written without them.

Three years later, that prediction has mostly held.

---

### The basics

Three costs of generics:

1. **Cognitive complexity.** Readers parse `[T cmp.Ordered]` machinery before reaching the logic. Worth it for `Max`; wasteful for a one-off helper.
2. **API surface.** Generic functions leak constraint complexity to callers. A constraint chain involving custom interface unions for a one-off helper is overkill.
3. **Compile time / binary size.** The compiler stencils per type instantiation. Rarely a problem; matters in performance-critical or embedded contexts.

Together: **prefer concrete types until generics clearly earn their keep.**

---

### A worked example

A "bad" generic:

```go
func PrintTwice[T any](x T) {
	fmt.Println(x)
	fmt.Println(x)
}
```

What's `T any` buying us? `fmt.Println` accepts `any` already. Just write:

```go
func PrintTwice(x any) {
	fmt.Println(x)
	fmt.Println(x)
}
```

Identical behaviour. No generics. Less to read. The `[T any]` was decorative — a "generics tax" with no offsetting payoff.

Contrast with `Max[T cmp.Ordered]` — that one **needs** the type parameter to (a) return a typed value (not `any`) and (b) constrain to ordered types so `>` works. Generics earn their keep there.

---

### Common mistake

Reaching for generics **first** when refactoring duplicated typed code. Better triage:

1. **Called with N≥3 distinct types, growth likely?** → Generics make sense.
2. **Called with 2-3 fixed types, no growth?** → Keep them separate, or write an `any` + type switch.
3. **Called with 1 type, and the duplication just *looks* generic-ish?** → Don't bother.

The Go community's reflex: **prefer concrete types until generics are clearly justified**.

```go
// the wrong way — generic for no reason
func ParseInts[T int | int64](s string) (T, error) { ... }
// Only ever called with int. Just write `func ParseInt(s string) (int, error)`.
```

---

### Recap

- Generics aren't free — cognitive cost, API surface, occasionally compile/binary cost.
- "If a function is called once with one type, don't bother."
- Lead with concrete types; reach for generics when 3+ types and growth argue for it.
- The stdlib (`slices`, `maps`, `cmp`) uses generics with restraint — emulate that.

---

## Practice

### Warm-up

Implement `Max[T cmp.Ordered](xs []T) (T, error)` — the generic version of L04's typed min/max helpers. Empty slice → zero + non-nil error.

```bash
cd lessons/12-generics/exercises/warmup/maxg
go test -v
```

---

### Main

Implement `Filter[T any]` and `Map[T, U any]` in the `slicesx` subpackage. Cover the test cases (filter evens, map int→string via `strconv.Itoa`, map string→string via `strings.ToUpper`).

```bash
cd lessons/12-generics/exercises/slicesx
go test -v
```

Note:
The warmup and main live in separate subpackages so you implement `Max` first as a self-contained mini-exercise, then tackle the two-function main package. In a real library these would be merged into one `slicesx` — and you'd probably also reach for the stdlib's `slices.Max` (Go 1.21+) rather than rolling your own.

---

## Closing thought

The Go team's release notes for 1.18 (the first version with generics) included this:

> We expect generics to be used relatively sparingly. Most Go programs will continue to be written without them.

Three years later (Go 1.21 added `cmp.Ordered`; `slices` and `maps` packages went stable in 1.21), the prediction has mostly held. Generics power a small set of high-leverage stdlib packages (`slices`, `maps`, `cmp`) and otherwise stay out of the way.

Use them when they earn their keep. Skip them when they don't.

---

## What we learned

- Type parameters: `func F[T any](x T) T` — the algorithm once, used with any type.
- Constraints: `any` / `comparable` / `cmp.Ordered` — choose the smallest that allows your ops.
- Type inference: idiomatic Go relies on it; only write `[T]` when the compiler can't pick.
- Multiple type parameters: `Map[T, U]` — input and output types independent.
- Restraint: "if a function is called once with one type, don't bother."

---

## Up next

Lesson 13 — **Encoding & I/O**. We crack open the `storage` package (a black box since L08), formally meet `io.Reader` and `io.Writer`, and build a CSV-to-expenses importer. The tracker is back.
