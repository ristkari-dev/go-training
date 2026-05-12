<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">04</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Functions &amp; first tests</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Write functions with multiple return values, named returns, and variadic arguments; use <code>defer</code> for cleanup; return and handle <code>error</code>s with <code>errors.New</code>; and write your first table tests with <code>testing.T</code>.</p>
</div>
</div>
</div>

---

## What we'll cover

- Function shape: multi-return values, named returns, variadic arguments (`xs ...int`).
- `defer` — running cleanup when a function returns, regardless of how it returns.
- The `error` return convention: returning `(value, error)`, the `if err != nil` shape, `errors.New`.
- Testing: `*_test.go` files, `func TestXxx(t *testing.T)`, table tests with `[]struct{}`, `t.Errorf` vs `t.Fatalf`.

---

## Concept 1: Function shape

### Motivation

You've used functions since lesson 01. Now we look at the parts of a function declaration that you haven't met yet: multiple return values, named returns, and variadic arguments. Multi-return is everywhere in Go's standard library — it's how `error` gets returned alongside a result.

---

### The basics

```go
package main

import "fmt"

// Multi-return — two values.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// Named returns — variables declared in the signature.
// The bare `return` at the end returns whatever's in q and r.
func divmodNamed(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

// Variadic — accepts zero or more ints.
// Inside the function, xs is a []int.
func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

func main() {
	q, r := divmod(17, 5)
	fmt.Println(q, r)         // 3 2

	q2, r2 := divmodNamed(17, 5)
	fmt.Println(q2, r2)       // 3 2

	fmt.Println(sum())             // 0
	fmt.Println(sum(1, 2, 3))      // 6
	fmt.Println(sum(1, 2, 3, 4))   // 10

	// Passing a slice to a variadic — note the trailing ... .
	xs := []int{10, 20, 30}
	fmt.Println(sum(xs...))         // 60
}
```

Three things:

- **Multi-return** uses parentheses around the return types. The caller destructures with `a, b := f()`.
- **Named returns** declare the variables in the signature. The bare `return` at the end is shorthand for "return all named returns." Use named returns to document what each return means; lesson 06 (structs) shows when they're worth the extra ceremony.
- **Variadic** — `xs ...T` is `xs []T` inside the function. Pass an existing slice with the `slice...` syntax (note the trailing `...`).

---

### A worked example

The expense theme — `WarmupMinMax(xs ...int) (int, int, error)` from your warm-up exercise:

```go
package main

import (
	"errors"
	"fmt"
)

var errEmpty = errors.New("WarmupMinMax: requires at least one value")

func WarmupMinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmpty
	}
	min, max := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, nil
}

func main() {
	lo, hi, err := WarmupMinMax(3, 1, 4, 1, 5, 9, 2, 6)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("min:", lo, "max:", hi)   // min: 1 max: 9
}
```

Three things show up in one function: variadic input, three-value multi-return, and the `error` return for the bad case. We'll dig into errors in concept 3.

---

### Common mistake

Forgetting the trailing `...` when passing a slice to a variadic:

```go
package main

import "fmt"

func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

func main() {
	nums := []int{1, 2, 3}
	fmt.Println(sum(nums))   // wrong: passes []int as a single arg
}
```

Go complains:

```
cannot use nums (variable of type []int) as int value in argument to sum
```

The fix is `sum(nums...)` — the three dots spread the slice into individual arguments. Easy to forget.

---

### Recap

- Multi-return: `func f() (T, U) { return … }`; caller does `a, b := f()`.
- Named returns: declare result variables in the signature; bare `return` returns them all.
- Variadic: `xs ...T` arrives as `[]T`. Spread an existing slice with `slice...`.

---

## Concept 2: `defer`

### Motivation

When a function returns, sometimes you want to run cleanup code regardless of *how* it returned — early return, normal return, even a panic. `defer` is Go's way to schedule that cleanup at the function's exit point. The classic use is closing files and unlocking mutexes; we'll meet both for real in later lessons.

---

### The basics

`defer expr` schedules `expr` to run when the surrounding function returns. Deferred calls run in **LIFO order** (last deferred, first to run):

```go
package main

import "fmt"

func main() {
	defer fmt.Println("world")  // scheduled — runs LAST
	defer fmt.Println("middle") // scheduled — runs before "world"
	fmt.Println("hello")        // runs first (no defer)
}
```

Output:

```
hello
middle
world
```

The deferred expression's *arguments* are evaluated at the `defer` line, but the *call* happens at function exit. That distinction trips people up sometimes (see the common mistake below).

---

### A worked example

A small file-close pattern you'll write hundreds of times once we get to I/O (lesson 13):

```go
// Roughly what real Go code looks like — we cover os.Open in lesson 13.
func loadAndPrint(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()  // runs when loadAndPrint returns, regardless of how

	// ... read from f ...
	return nil
}
```

Two important things:

- `defer f.Close()` is on the very next line after the error check — that's the conventional shape. You won't forget to close the file because Go guarantees it.
- If `loadAndPrint` returns early with an error after the defer, `f.Close()` still runs. If `loadAndPrint` returns normally, `f.Close()` still runs. If `loadAndPrint` panics, `f.Close()` still runs.

---

### Common mistake

Forgetting that argument evaluation happens at the `defer` line, not at function exit:

```go
package main

import "fmt"

func main() {
	i := 1
	defer fmt.Println("deferred i =", i)  // i evaluated to 1 here
	i = 2
	fmt.Println("current i =", i)
}
```

Output:

```
current i = 2
deferred i = 1
```

The defer captured `i`'s value at the defer statement, not at the function's exit. If you want the latter behaviour, wrap the call in a function literal: `defer func() { fmt.Println("deferred i =", i) }()`.

---

### Recap

- `defer expr` schedules `expr` for the function's exit.
- Multiple defers run in LIFO order.
- Arguments are evaluated at the `defer` line; the call itself runs later.
- Most common use: file close, mutex unlock — you'll meet both in later lessons.

---

## Concept 3: Errors

### Motivation

In Go, functions that can fail return an `error` as their last result. Callers check it with `if err != nil { … }`. There are no exceptions, no try/catch — error handling is just another value you check. That's verbose but predictable; you always know exactly where errors can come from.

---

### The basics

The shape — function returns `(result, error)`, caller checks the error:

```go
package main

import (
	"errors"
	"fmt"
)

// Sentinel error declared as a package-level var.
var errEmpty = errors.New("WarmupMinMax: requires at least one value")

func WarmupMinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmpty
	}
	// ... normal-path implementation ...
	return xs[0], xs[0], nil
}

func main() {
	lo, hi, err := WarmupMinMax()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(lo, hi)
}
```

Output:

```
error: WarmupMinMax: requires at least one value
```

Two patterns to learn here:

1. **`errors.New("message")`** — the simplest way to make an error. Stash it in a package-level `var` if you want callers (or your own code) to compare against it. We call these "sentinel errors."
2. **`if err != nil { return ..., err }`** — the canonical guard clause. Echoes lesson 03's early-return pattern; you'll write this thousands of times.

`error` itself is a built-in type with one method: `Error() string`. That's all there is to it. (We'll see what makes it special — and how to build your own error types — in lesson 10 when interfaces arrive. For now, treat `error` as "a value that knows how to describe itself.")

---

### A worked example

Combining a multi-return function with the canonical error-check shape:

```go
package main

import (
	"errors"
	"fmt"
)

var errDivByZero = errors.New("division by zero")

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errDivByZero
	}
	return a / b, nil
}

func main() {
	q, err := divide(10, 3)
	if err != nil {
		fmt.Println("first error:", err)
		return
	}
	fmt.Println("10 / 3 =", q)

	q2, err := divide(10, 0)
	if err != nil {
		fmt.Println("second error:", err)
		return
	}
	fmt.Println("10 / 0 =", q2)   // never reached
}
```

Output:

```
10 / 3 = 3
second error: division by zero
```

Notice the early return after each error check. No `else` — the rest of the function only runs when there's no error.

---

### Common mistake

Ignoring the error by assigning it to `_`:

```go
q, _ := divide(10, 0)    // squashes the error
fmt.Println(q)           // prints 0 — but there was an error you ignored
```

Go's compiler doesn't flag this (`_` means "I deliberately don't want this"), but `golangci-lint`'s `errcheck` does. Almost always you want to check the error and handle it. If you really do want to ignore it, leave a comment explaining why — your future self will thank you.

---

### Recap

- Functions that can fail return `(result, error)`. The error is always the last return.
- `errors.New("msg")` creates a simple error. Stash sentinels in package-level `var`s if you want comparisons.
- `if err != nil { return ..., err }` is *the* Go shape for error propagation.
- Don't silently swallow errors with `_` unless you have a clear reason.

---

## Concept 4: Testing with `testing.T`

### Motivation

Go ships with a testing framework in the standard library — no third-party libraries required. Tests live in `*_test.go` files alongside your code. The `go test` command finds and runs them. Three patterns will carry you through Phase 1: the basic `func TestXxx(t *testing.T)` shape, the table-test pattern with `[]struct{}`, and the difference between `t.Errorf` (report and continue) and `t.Fatalf` (report and stop).

---

### The basics

A minimal test:

```go
package mypkg

import "testing"

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d, want %d", got, want)
	}
}
```

Three rules:

1. **Filename ends in `_test.go`.** `go build` ignores these; `go test` runs them.
2. **Function name starts with `Test`** and takes `*testing.T`. Any other signature is invisible to `go test`.
3. **Report failures with `t.Errorf`/`t.Fatalf`.** The message you pass becomes the test failure output — make it useful (input + got + want).

Run them with `go test ./...` from anywhere in the module, or `go test ./path/to/package` to scope.

---

### A worked example — the table test pattern

When you have many cases to check, write one test that iterates over a slice of structs:

```go
package mypkg

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
```

What `t.Run(name, func)` buys you:

- Each case is reported as a separate sub-test (`TestAdd/positive`, `TestAdd/zero`).
- One case failing doesn't stop the others — the loop keeps going.
- You can run a single case with `go test -run TestAdd/positive`.

This is the shape used in your warm-up tests AND the shape you'll fill in for the main exercise.

`t.Errorf` vs `t.Fatalf`:

- **`t.Errorf("...")`** — record the failure and **keep running** the current test function. Use it for assertions where checking the rest still gives useful information.
- **`t.Fatalf("...")`** — record the failure and **stop** the current test function. Use it when continuing would crash (e.g., dereferencing a nil pointer) or produce noise.

---

### A note on `go vet`

Alongside `go test`, get into the habit of running `go vet ./...`. It's a sibling stdlib tool that catches subtle bugs — wrong format verbs in `Printf`, shadowed variables, unreachable code, missing struct field tags. CI runs it, and from this lesson on, the lesson README's "How to run" mentions it explicitly.

```bash
go vet ./...
```

---

### Common mistake

Using `t.Errorf` when you should use `t.Fatalf`:

```go
result, err := mightFail()
if err != nil {
	t.Errorf("unexpected error: %v", err)   // wrong: keeps going
}
result.DoSomething()                          // crashes if result is nil
```

Fix: `t.Fatalf` when the next line would crash, `t.Errorf` when it wouldn't:

```go
result, err := mightFail()
if err != nil {
	t.Fatalf("unexpected error: %v", err)   // stops the test cleanly
}
result.DoSomething()
```

---

### Recap

- Tests live in `*_test.go` files; `go test ./...` runs them.
- `func TestXxx(t *testing.T)` is the only test signature `go test` recognises.
- Table tests: a `[]struct{}` of cases, a `for ... t.Run` loop. Sub-tests are individually addressable and one failure doesn't block the rest.
- `t.Errorf` → record and continue. `t.Fatalf` → record and stop.
- `go vet ./...` catches the bugs `go test` won't.

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `WarmupAdd(a, b int) int` — the simplest function in the course. Return `a + b`.
- `WarmupMinMax(xs ...int) (int, int, error)` — variadic input, multi-return, error for empty input. Return `(0, 0, errEmptyWarmupMinMax)` when `len(xs) == 0`; otherwise return the smallest, the largest, and `nil`.

Tests are pre-written.

```bash
cd lessons/04-functions/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `Categorise(amount float64) string` — same signature and behaviour as lesson 03's. Use whichever shape (if-chain or tagless switch) you prefer.
- `FormatExpense(date string, amount float64, cat string) string` — return `"YYYY-MM-DD  €AMOUNT  category"` using `fmt.Sprintf` with `%s  €%-7.2f %s`.

For the **first time, you write the tests yourself** — `exercises/main_test.go` ships as a skeleton with the struct shape and the loop scaffold; you fill in the case rows and the assertion body. The README walks through the test shape step by step. Cases to include:

- For Categorise — at least 5 cases including the boundary values (9.99, 10, 50, 50.01).
- For FormatExpense — at least 3 cases. Build the expected output strings by hand using the format.

```bash
cd lessons/04-functions/exercises
go test -v
```

Note:
For live: emphasise the test-authoring jump. Walk through writing one Categorise case live on the projector — name the case, fill in the input, type out the expected string, run the test, watch it fail with a useful message, fix the implementation, watch it pass. That round-trip is the single most valuable habit Phase 1 hands students.

---

## What we learned

- Multi-return values are everywhere in Go (especially `(result, error)`).
- Variadic args (`xs ...T`) arrive as `[]T`. Spread a slice with `slice...`.
- Named returns are sugar for documenting result variables.
- `defer` schedules cleanup at function exit; arguments evaluate at the defer line, the call runs at return time.
- `error` is Go's failure convention. `errors.New` makes a sentinel; `if err != nil { return ..., err }` is the canonical propagation shape.
- Tests live in `*_test.go`; table tests with `[]struct{}` and `t.Run` are *the* pattern for parameterised testing. `t.Errorf` vs `t.Fatalf` is the recovery-vs-stop distinction.
- `go vet ./...` — run it alongside `go test`.

---

## Up next

Lesson 05 — Composite types I: arrays, slices, maps (`[]T`, `len`/`cap`, `append`, `range`, `map[K]V`).
