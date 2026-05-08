# Lesson 02: Variables, types, operators

## Learning goals

- Declare variables in Go's three forms (`var x int = 5`, `var x = 5`, `x := 5`) and pick the right one for the context.
- Recognise the four basic types you'll see most often (`int`, `float64`, `string`, `bool`) and their zero values.
- Use arithmetic operators (`+`, `-`, `*`, `/`, `%`) and understand integer-vs-float division.
- Convert between numeric types explicitly with `T(value)` syntax.
- Declare values that never change with `const`.

## Prerequisites

- Lesson 01 (Hello, Go) — you should be comfortable running a Go file with `go run` and using `fmt.Println` / `fmt.Printf`.

## Concepts

### Variables and constants

Variables hold values. Go has three equivalent ways to declare one:

```go
var x int = 5      // explicit: type and initial value
var y = 5          // type inferred from the literal (int)
z := 5             // short form: only allowed inside functions
```

All three produce an `int` variable holding `5`. The short form `z := 5` is the most common in practice — it's compact, and Go figures out the type from the expression on the right. The explicit `var x int = 5` form is essential at package level (outside any function), where `:=` isn't allowed.

For values that should never change after declaration, use `const`:

```go
const Pi = 3.14
const TipRate = 0.14
```

The compiler will reject any attempt to assign a new value to a constant.

**Common mistake.** Trying to use `:=` outside a function:

```go
package main

import "fmt"

z := 5  // wrong: := only allowed inside functions

func main() {
	fmt.Println(z)
}
```

Go complains:

```
syntax error: non-declaration statement outside function body
```

At package level, you must use `var` (or `const`). Inside a function, both forms work.

### Basic types and zero values

Every variable in Go has a type, and the type determines what values it can hold and what operations are allowed. The four types you'll meet first:

- `int` — whole numbers. Size is platform-dependent (almost always 64-bit on modern machines). Zero value: `0`.
- `float64` — 64-bit floating-point numbers. Zero value: `0.0`.
- `string` — an immutable sequence of bytes (typically interpreted as UTF-8 text). Zero value: `""` (empty string).
- `bool` — `true` or `false`. Zero value: `false`.

Even more important than the type list: every variable always has a value. If you declare `var x int` without assigning, `x` is `0`. There's no such thing in Go as reading an uninitialised variable — the language guarantees that wouldn't happen.

```go
var i int       // 0
var f float64   // 0.0
var s string    // ""
var b bool      // false
```

**Common mistake.** Assuming Go has `nil` for numeric types:

```go
var x int
if x == nil {  // wrong: int is never nil
	// unreachable
}
```

Go complains:

```
cannot convert nil to type int
```

`nil` exists in Go but only for pointers, slices, maps, channels, functions, and interfaces — not for basic types. An uninitialised `int` is `0`, not `nil`.

### Operators and arithmetic

Arithmetic operators do what you'd expect:

```go
a := 10
b := 3

a + b   // 13
a - b   // 7
a * b   // 30
a / b   // 3   (integer division — fractional part dropped)
a % b   // 1   (remainder)
```

Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) work on values of the same type and produce a `bool`. Logical operators (`&&`, `||`, `!`) combine booleans.

The expense theme — a 14% tip:

```go
const TipRate = 0.14

amount := 40.00
tip := amount * TipRate         // 5.60
totalWithTip := amount + tip    // 45.60

fmt.Printf("Tip (14%%): €%.2f\n", tip)
```

Notice `%%` in the format string — that's how you print a literal `%` with `Printf`.

**Common mistake.** Integer division silently drops the fractional part:

```go
total := 7
count := 2
avg := total / count   // 3, not 3.5
```

If you wanted `3.5`, at least one operand needs to be a float:

```go
avg := float64(total) / float64(count)   // 3.5
```

This trips up everyone the first time. The compiler doesn't warn — it just gives the integer answer.

### Type conversions

Go is strict about types. You can't add an `int` to a `float64`, or compare a `string` to a `[]byte`, without an explicit conversion. This feels strict at first but prevents whole categories of bugs that other languages let through.

The syntax is `T(value)` — read it as "cast to T":

```go
count := 3            // int
average := 13.33      // float64

asFloat := float64(count)    // 3.0
asInt := int(average)        // 13 (fractional part truncated)
```

`float64(count)` converts the int to float64 with no loss. `int(average)` converts the float64 to int by *truncating toward zero* — `int(13.33)` is `13`, `int(-2.9)` is `-2`. There's no rounding.

A cents-to-euros example:

```go
amountCents := 450
amountEuros := float64(amountCents) / 100.0   // 4.5
```

Without the conversion, `amountCents / 100` would be integer division (`4`, not `4.5`).

**Common mistake.** Truncation, not rounding, when converting float64 to int:

```go
amount := 3.9
asInt := int(amount)
fmt.Println(asInt)   // 3, not 4
```

`int(3.9)` is `3`, not `4`. Go truncates toward zero — it doesn't round to the nearest integer. If you need rounding, you'd use `math.Round` (lesson 14 territory) — but for simple display, `%.2f` formatting is usually what you actually want.

## Exercise: warm-up

Open `exercises/warmup.go`. Two functions to implement:

- `WarmupZeroValues() (int, float64, string, bool)` — return the zero values of these four types. The simplest way is to declare four variables with `var` and return them.
- `WarmupConvert(intVal int, floatVal float64) (asFloat float64, asInt int)` — return `float64(intVal)` and `int(floatVal)` (the float-to-int conversion truncates toward zero).

The tests cover positive, zero, negative, and large values.

## Exercise: main

Open `exercises/main.go`. The `TipRate` constant is already declared at the top of the file — use it in your `Summary`.

- `Total(a, b, c float64) float64` — return `a + b + c`.
- `Average(a, b, c float64) float64` — return `(a + b + c) / 3.0`. Note the `3.0` (not `3`) to keep float64 arithmetic.
- `Summary(a, b, c float64) string` — return a multi-line summary using `Total`, `Average`, and `TipRate`. Use `fmt.Sprintf` with `%.2f` for each amount and `\n` between lines. The expected format is:

```
3 expenses
Total: €40.00
Average: €13.33
Tip (14%): €5.60
Total with tip: €45.60
```

Notice `(14%)` in the output — that's `%%` in your format string (escape for `%`).

## How to run

```bash
cd lessons/02-variables/exercises
go test -run Warmup -v   # warm-up only
go test -v                # everything
```

After you finish (or while iterating), run `gofmt -w .` from the lesson folder to keep your code in canonical Go style. Make this a habit.

Once both exercises pass, take a look at `solutions/` to compare your code with the reference.

## Going further

### Read

- [A Tour of Go — Variables](https://go.dev/tour/basics/8) — the official tour's variable section, with playgrounds.
- [Effective Go — Constants](https://go.dev/doc/effective_go#constants) — short explanation of typed and untyped constants.
- [The Go specification — Numeric types](https://go.dev/ref/spec#Numeric_types) — exhaustive reference for Go's numeric types (skim it; you don't need to memorise it).

### Try

- **`iota` in action.** Add a constant block with `iota` to `solutions/main.go`: `const (Cents = iota; Euros; Dollars)` gives Cents=0, Euros=1, Dollars=2. Try it in a small program. No reference solution provided — `iota` is an "extra credit" topic for lesson 02.
- **Format width specifiers.** Go's `Printf` supports width specifiers: `%8.2f` pads the float to width 8, right-aligned. Modify your `Summary` to right-align the amounts in a clean column. Different from the spec — but it's a useful display trick worth knowing.
- **Integer overflow.** Try `var x int8 = 127; x = x + 1` and print `x`. The result is surprising. Look up "integer overflow" if you want to understand why.
