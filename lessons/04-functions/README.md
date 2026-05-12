# Lesson 04: Functions & first tests

## Learning goals

- Write functions with multiple return values, named returns, and variadic arguments.
- Use `defer` to schedule cleanup at function exit.
- Return and handle `error`s with the canonical `if err != nil { return …, err }` shape.
- Write your first tests using `testing.T`, the table-test pattern, and `t.Errorf` vs `t.Fatalf`.
- Get into the habit of running `go vet ./...` alongside `go test`.

## Prerequisites

- Lesson 01 — `go run`, `package main`, `fmt`.
- Lesson 02 — variable forms, types, arithmetic.
- Lesson 03 — `if`/`else`, `for`, `switch`, early returns.

## Concepts

### Function shape: multi-return, named returns, variadic

You've used functions since lesson 01. Three additional shapes round out the toolkit:

```go
// Multi-return — two values.
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// Named returns — variables declared in the signature.
// The bare `return` returns whatever's in q and r.
func divmodNamed(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

// Variadic — accepts zero or more ints.
// Inside the function, xs is []int.
func sum(xs ...int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}
```

Calling them:

```go
q, r := divmod(17, 5)       // 3 2
q2, r2 := divmodNamed(17, 5) // 3 2

sum()                        // 0
sum(1, 2, 3)                 // 6
sum(1, 2, 3, 4)              // 10

xs := []int{10, 20, 30}
sum(xs...)                   // 60 — note the trailing ...
```

Three points:

- **Multi-return** is everywhere in the stdlib — most notably for `(result, error)` pairs (concept 3).
- **Named returns** declare result variables in the signature. The bare `return` returns them all. Useful when the return values are non-obvious from the types alone.
- **Variadic** — `xs ...T` is `xs []T` inside the function. Spread an existing slice with `slice...`.

**Common mistake.** Forgetting the trailing `...` when passing a slice to a variadic:

```go
nums := []int{1, 2, 3}
fmt.Println(sum(nums))    // compile error: cannot use nums as int
```

Fix: `sum(nums...)`.

### `defer`

`defer expr` schedules `expr` to run when the surrounding function returns. Cleanup that needs to happen regardless of return path (early return, normal return, panic) goes here.

```go
func main() {
	defer fmt.Println("world")  // runs last
	defer fmt.Println("middle") // runs second
	fmt.Println("hello")        // runs first
}
// Output:
// hello
// middle
// world
```

Deferred calls run in **LIFO order** (last deferred, first to run).

The most common use is paired with resource acquisition — open a file, immediately defer the close:

```go
// You'll write this exact shape hundreds of times once we hit lesson 13's I/O.
f, err := os.Open(path)
if err != nil {
	return err
}
defer f.Close()
// ... use f ...
```

`f.Close()` will run when the surrounding function returns, no matter how. You can't forget.

**Common mistake.** Argument evaluation happens at the `defer` line, not at function exit:

```go
i := 1
defer fmt.Println("deferred i =", i)  // i evaluated to 1 right now
i = 2
fmt.Println("current i =", i)
```

Output:

```
current i = 2
deferred i = 1
```

If you want late binding, wrap the call in a function literal: `defer func() { fmt.Println(i) }()`.

### Errors

Go functions that can fail return an `error` as their last result. Callers check it with `if err != nil`. There's no `try`/`catch`; error handling is just another value to inspect.

```go
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
		fmt.Println("error:", err)
		return
	}
	fmt.Println("10 / 3 =", q)
}
```

Two patterns to internalise:

- **`errors.New("message")`** — the simplest way to make an error. Stash recurring ones in package-level `var`s as "sentinels" (e.g. `var errDivByZero = errors.New(...)`). Compare using `errors.Is(err, sentinelVar)` — your warm-up test uses this shape. (`errors.Is` is plain equality for unwrapped sentinels today; lesson 11 explains why it's the future-proof spelling once errors get wrapped.)
- **`if err != nil { return ..., err }`** — *the* canonical guard. You'll write it thousands of times.

`error` itself is a built-in type with one method: `Error() string`. (What makes that work is interfaces — lesson 10.)

**Common mistake.** Silently ignoring errors with `_`:

```go
q, _ := divide(10, 0)
fmt.Println(q)           // prints 0 — but there was an error
```

`golangci-lint`'s `errcheck` flags it. If you really do want to discard the error, leave a comment explaining why.

### Testing with `testing.T`

Tests live in `*_test.go` files alongside your code. `go test ./...` finds and runs them.

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

1. Filename ends in `_test.go`.
2. Function name starts with `Test`, takes `*testing.T`.
3. Report failures with `t.Errorf` (record and continue) or `t.Fatalf` (record and stop).

The **table-test pattern** scales the basic shape to many cases:

```go
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

`t.Run(name, func)` makes each case a separately-named sub-test. Sub-tests are individually addressable (`go test -run TestAdd/positive`), and one failing case doesn't block the rest.

`t.Errorf` vs `t.Fatalf`:

- **`t.Errorf`** — record the failure, keep running. Use when the rest of the test still gives useful information.
- **`t.Fatalf`** — record the failure, stop the test function. Use when continuing would crash or produce noise.

**Common mistake.** Using `t.Errorf` when the next line would crash:

```go
result, err := mightFail()
if err != nil {
	t.Errorf("unexpected error: %v", err)   // wrong: keeps going
}
result.DoSomething()                          // crashes if result is nil
```

Fix: `t.Fatalf` when continuing would crash.

## Exercise: warm-up

Open `exercises/warmup.go`. Two functions:

- `WarmupAdd(a, b int) int` — return `a + b`.
- `WarmupMinMax(xs ...int) (int, int, error)` — return the smallest and largest, plus `nil`. For empty input, return `(0, 0, errEmptyWarmupMinMax)` (the sentinel is already declared at the top of the file).

Tests are pre-written. Implement until they pass.

## Exercise: main

**This is the first lesson where you write tests yourself.** Open `exercises/main_test.go` — it ships as a *skeleton*:

```go
func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		// TODO: add at least 5 cases here...
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call Categorise(tc.amount), compare with tc.want...
			_ = tc
		})
	}
}
```

Your job is to fill in:

1. **The `cases` slice** — at least 5 entries for `TestCategorise`, at least 3 for `TestFormatExpense`. Cover the snack/regular/splurge branches and the boundary amounts (9.99, 10, 50, 50.01) for Categorise; for FormatExpense, build the expected output strings by hand using the format string in the doc comment.
2. **The `t.Run` body** — call the function, compare against `tc.want`, report a failure with `t.Errorf` if they differ. Replace the `_ = tc` line — it's just there so the empty skeleton compiles.

Look at `exercises/warmup_test.go` (and `solutions/main_test.go` if you want to compare to the reference) for the shape.

The two implementations in `exercises/main.go`:

- `Categorise(amount float64) string` — return `"snack"` (`amount < 10`), `"regular"` (`10 <= amount <= 50`), or `"splurge"` (`amount > 50`). Same as lesson 03.
- `FormatExpense(date string, amount float64, cat string) string` — return `"YYYY-MM-DD  €AMOUNT  category"` using `fmt.Sprintf` with `"%s  €%-7.2f %s"`. Note the spacing in the expected outputs in the doc comment.

## How to run

```bash
cd lessons/04-functions/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main
```

This lesson introduces two new habits — get into both of them now and you won't have to come back:

```bash
gofmt -w .       # reformat your code to canonical Go style
go vet ./...     # find subtle bugs go test won't catch
```

`gofmt` is non-negotiable (CI enforces it). `go vet` is run by CI too; running it locally catches the surprises before you commit. From this lesson onwards, every lesson README mentions it.

Once both exercises pass, compare your code (especially your tests) against `solutions/`.

## Going further

### Read

- [A Tour of Go — Functions](https://go.dev/tour/basics/4) — multi-return values and named returns with playgrounds.
- [Effective Go — Defer](https://go.dev/doc/effective_go#defer) — short explanation of `defer`'s semantics and why Go's compiler likes resource pairs.
- [Go blog — Error handling and Go](https://go.dev/blog/error-handling-and-go) — the canonical introduction to the `error` type and the `if err != nil` shape (skip the `errors.Is`/`errors.As` bits for now — lesson 11).
- [`testing` package docs](https://pkg.go.dev/testing) — the stdlib reference. You'll meet `t.Helper`, `t.TempDir`, and subtests-beyond-table-tests in lesson 15.

### Try

- **A third return for divmod.** Add `divmodChecked(a, b int) (q, r int, err error)` that returns `errDivByZero` when `b == 0`. Write the test for it.
- **Run a single sub-test.** Pick one of your `TestCategorise/...` sub-tests and run only that one: `go test -run "TestCategorise/upper-snack-edge" -v`. Useful when you want to focus on one failing case.
- **`go vet` finds a real bug.** Try writing `fmt.Printf("%d", "hello")` in a small program and run `go vet ./...` — it'll catch the type mismatch before you run the program. Format-string bugs are the most common thing `go vet` finds.
