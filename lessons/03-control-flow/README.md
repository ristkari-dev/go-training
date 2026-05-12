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
