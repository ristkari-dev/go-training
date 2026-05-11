## Lesson 02

# Variables, types, operators

Learning goal: declare typed variables in Go's three forms, recognise basic types and their zero values, do arithmetic with operators, and convert between numeric types.

---

## What we'll cover

- Three ways to declare a variable: `var x int = 5`, `var x = 5`, `x := 5`.
- Basic types — `int`, `float64`, `string`, `bool` — and their zero values.
- Arithmetic operators (`+`, `-`, `*`, `/`, `%`) and a tip-calculation example.
- Explicit type conversions (`float64(x)`, `int(y)`) and why Go is strict.
- Constants with `const`.

---

## Concept 1: Variables and constants

### Motivation

Variables hold values. We need a way to introduce a name, optionally give it a type, and either assign a starting value or rely on Go's default. Go has three syntactic forms — they're equivalent, but you'll see them in different places, so it's worth knowing all three.

---

### The basics

Variables hold values that can be declared in three forms. Constants declare values that never change. For sequential constants, `iota` provides an auto-incrementing counter inside a `const ( ... )` block.

--

### Code

```go
package main

import "fmt"

func main() {
	var x int = 5      // explicit: type and initial value
	var y = 5          // type inferred from the literal (int)
	z := 5             // short form: only inside functions

	const Pi = 3.14    // constant: value never changes

	fmt.Println(x, y, z, Pi)
}
```

The three variable forms are equivalent in this example. `:=` is the most common in practice — it's shorter and Go infers the type for you. The explicit `var x int = 5` form is more common at package level (outside functions) where `:=` isn't allowed.

`const` declares a value that the compiler refuses to let you change after the declaration.

For sequential constants — like enum values — use `iota`, a counter that starts at 0 inside a `const ( ... )` block:

```go
const (
	Cents = iota  // 0
	Euros         // 1
	Dollars       // 2
)
```

Each line implicitly takes the previous line's expression, so listing names alone gives you 0, 1, 2.

---

### A worked example

The expense theme — declare an expense as three typed variables.

--

### Code

```go
package main

import "fmt"

func main() {
	date := "2026-05-08"
	amount := 4.50
	category := "coffee"

	fmt.Printf("%s  €%.2f  %s\n", date, amount, category)
}
```

Three short-form declarations, three different types (`string`, `float64`, `string`), all inferred from their literal initialisers. The output is the same as lesson 01's `FormatExpense` — except now the values are real variables you can compute with.

---

### Common mistake

Trying to use `:=` outside a function.

--

### Code (broken)

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

---

### Recap

- Three variable forms: `var x int = 5` (explicit), `var x = 5` (inferred), `x := 5` (short).
- `:=` only works inside functions.
- `const` declares a compile-time-fixed value.
- `iota` is a counter inside a `const ( … )` block; useful for enum-like sequences.

---

## Concept 2: Basic types and zero values

### Motivation

Every variable in Go has a type. The type determines what values the variable can hold and what operations you can do with it. Even more important: every variable always has a value, even before you assign one — Go fills in the type's "zero value" so there's no such thing as an "uninitialised variable" you can read.

---

### The basics

The four most common types are `int`, `float64`, `string`, and `bool`. Every variable starts at its type's zero value if you don't assign one.

--

### Code

```go
package main

import "fmt"

func main() {
	var i int       // 0
	var f float64   // 0.0
	var s string    // ""
	var b bool      // false

	fmt.Println(i, f, s, b)
}
```

Output:

```
0 0  false
```

The four most common types you'll meet first:

- `int` — whole numbers. Size depends on the platform (usually 64-bit). Zero value: `0`.
- `float64` — 64-bit floating-point numbers. Zero value: `0.0`.
- `string` — immutable sequence of bytes. Zero value: `""` (empty).
- `bool` — `true` or `false`. Zero value: `false`.

---

### A worked example

You can declare and assign in one go, or split them. Both produce the same final state.

--

### Code

```go
package main

import "fmt"

func main() {
	// declare-and-assign
	count := 3
	average := 13.33
	currency := "€"
	settled := true

	// or: declare-then-assign (verbose)
	var anotherCount int
	anotherCount = 3

	fmt.Println(count, average, currency, settled)
	fmt.Println("declared then assigned:", anotherCount)
}
```

The split form is occasionally useful when the assignment depends on a condition you'll write in the next few lines (lesson 03 territory).

---

### Common mistake

Assuming Go has `nil` for numeric types.

--

### Code (broken)

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

`nil` exists in Go but only for pointers, slices, maps, channels, functions, and interfaces — not for basic types. An uninitialised `int` is `0`, not `nil`. The same goes for `float64` (0.0), `string` (""), and `bool` (false).

---

### Recap

- Four basic types you'll use first: `int`, `float64`, `string`, `bool`.
- Every variable has a value — the zero value if you didn't assign one.
- Zero values: `0`, `0.0`, `""`, `false`. No nil for basic types.

---

## Concept 3: Operators and arithmetic

### Motivation

You've seen variables. Now we want to compute new values from them. Arithmetic operators (`+`, `-`, `*`, `/`, `%`) work on numbers; comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) work on values of the same type and produce booleans; logical operators (`&&`, `||`, `!`) combine booleans.

---

### The basics

Go's five arithmetic operators cover addition, subtraction, multiplication, division, and remainder. Integer division drops the fractional part — convert to float when you need fractions.

--

### Code

```go
package main

import "fmt"

func main() {
	a := 10
	b := 3

	fmt.Println(a + b)   // 13
	fmt.Println(a - b)   // 7
	fmt.Println(a * b)   // 30
	fmt.Println(a / b)   // 3   (integer division — fractional part dropped)
	fmt.Println(a % b)   // 1   (remainder)

	fmt.Println(a == b)  // false
	fmt.Println(a > b)   // true
}
```

`+`, `-`, `*`, `/`, `%` are the five basic arithmetic operators. `%` is the remainder operator, only valid for integers.

Note `a / b` is `3`, not `3.333…`. When both operands are integers, Go does *integer division* — the fractional part is dropped. To get a float result, convert at least one operand: `float64(a) / float64(b)` would give `3.333…`.

---

### A worked example

The expense theme — compute a 14% tip.

--

### Code

```go
package main

import "fmt"

func main() {
	const TipRate = 0.14

	amount := 40.00
	tip := amount * TipRate
	totalWithTip := amount + tip

	fmt.Printf("Subtotal: €%.2f\n", amount)
	fmt.Printf("Tip (14%%): €%.2f\n", tip)
	fmt.Printf("Total: €%.2f\n", totalWithTip)
}
```

Output:

```
Subtotal: €40.00
Tip (14%): €5.60
Total: €45.60
```

Notice `%%` in the format string — that's how you print a literal `%` with `Printf`.

---

### Common mistake

Integer division silently drops the fractional part.

--

### Code (broken)

```go
package main

import "fmt"

func main() {
	total := 7
	count := 2
	avg := total / count
	fmt.Println(avg)  // 3, not 3.5
}
```

If you wanted `3.5`, both operands need to be floats:

```go
avg := float64(total) / float64(count)  // 3.5
```

This bites everyone the first time. The compiler doesn't warn — it just gives the integer answer.

---

### Recap

- Arithmetic: `+`, `-`, `*`, `/`, `%`. `%` is remainder; only on integers.
- Integer division drops the fractional part — convert to float for fractional results.
- Comparison (`==`, `<`, …) yields a `bool`.
- Print a literal `%` in `Printf` with `%%`.

---

## Concept 4: Type conversions

### Motivation

Go is strict about types. You can't add an `int` to a `float64`, or compare a `string` to a `[]byte`, without an explicit conversion. This feels strict at first but it prevents whole categories of bugs that other languages let through. The syntax for conversion is `T(value)` — read it as "cast to T."

---

### The basics

Type conversions are explicit in Go. Use `float64(x)` or `int(y)` to convert between numeric types — the compiler won't do it for you.

--

### Code

```go
package main

import "fmt"

func main() {
	count := 3            // int
	average := 13.33      // float64

	// average + count  // this would fail to compile

	asFloat := float64(count)
	asInt := int(average)

	fmt.Println(asFloat, asInt)  // 3 13
}
```

`float64(count)` converts the int to float64 (no information lost). `int(average)` converts the float64 to int — but the fractional part is *truncated toward zero*, so `int(13.33)` is `13`, not `14`. There's no rounding.

---

### A worked example

The expense theme — say cents are stored as integers, but we want to display them as euros.

--

### Code

```go
package main

import "fmt"

func main() {
	amountCents := 450  // €4.50 in cents
	amountEuros := float64(amountCents) / 100.0
	fmt.Printf("€%.2f\n", amountEuros)
}
```

Output:

```
€4.50
```

Without the conversion, `amountCents / 100` would be integer division (`4`, not `4.5`). The `float64(amountCents)` promotes one side; the literal `100.0` is already a float64, so the division produces a float64.

---

### Common mistake

Truncation, not rounding, when converting float64 to int.

--

### Code (broken)

```go
package main

import "fmt"

func main() {
	amount := 3.9
	asInt := int(amount)
	fmt.Println(asInt)  // 3, not 4
}
```

`int(3.9)` is `3`. `int(-2.9)` is `-2` — truncates *toward zero*, not toward negative infinity. If you need rounding, you'd use `math.Round` (lesson 14 territory) — but for simple display purposes, `%.2f` formatting is usually what you actually want.

---

### Recap

- Conversion syntax: `T(value)`.
- Numeric conversions are explicit — Go won't promote types implicitly.
- `int(floatValue)` truncates toward zero — no rounding.
- Use conversions to escape integer division when you want fractional results.

---

## Practice

### Warm-up

Two small functions in `exercises/warmup.go`:

- `WarmupZeroValues()` returns the zero values of `int`, `float64`, `string`, `bool`.
- `WarmupConvert(intVal, floatVal)` returns `float64(intVal)` and `int(floatVal)`.

```bash
cd lessons/02-variables/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `TipRate` is already defined as a `const`.
- `Total(a, b, c)` and `Average(a, b, c)` are simple arithmetic.
- `Summary(a, b, c)` uses Total, Average, and TipRate to produce a multi-line string with `fmt.Sprintf` and `\n`.

```bash
cd lessons/02-variables/exercises
go test -v
```

Note:
For live: walk through the integer-division gotcha on the projector (`7/2 = 3`, not `3.5`). Also the `int(3.9) = 3` truncation. Both produce surprised faces and lock the lesson in. Skip the worked example in concept 2's "split assignment" form if running short — the exercises don't require it.

---

## What we learned

- Three variable forms: `var x int = 5`, `var x = 5`, `x := 5`. Use `:=` inside functions; `var` at package level.
- Basic types: `int`, `float64`, `string`, `bool`. Zero values: `0`, `0.0`, `""`, `false`.
- Arithmetic: `+`, `-`, `*`, `/`, `%`. Integer division drops fractions — convert to float to get fractional results.
- Conversions are explicit (`float64(x)`, `int(y)`); `int(float)` truncates toward zero.
- `const` declares values that never change; `iota` is a counter for sequences.
- Run `gofmt -w .` to keep your code formatted.

---

## Up next

Lesson 03 — Control flow (`if`, `for`, `switch`).
