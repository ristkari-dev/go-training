<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">03</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Control flow</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Branch on values with <code>if</code>/<code>else</code>, prefer early returns to deep nesting, write the three forms of <code>for</code> loop Go provides, and pick between <code>if</code>-chains and <code>switch</code> for multi-way branches.</p>
</div>
</div>
</div>

---

## What we'll cover

- `if` / `else if` / `else` and Go's comparison and logical operators.
- Early returns / guard clauses — why deeply nested code is a smell.
- `for` in three forms: infinite, condition-only, and C-style. Plus a peek at `for ... range`.
- `switch`, both tagless (`switch { case … }`) and tagged (`switch x { case … }`). Brief note on `switch v.(type)` (Phase 2).

---

## Concept 1: `if`, `else`, and comparison

### Motivation

Programs need to behave differently depending on the values they're working with. Go's `if` is the basic branching construct — it takes a boolean expression and a body, and optionally chains with `else if` and `else`. Unlike many languages, Go does not put parentheses around the condition.

---

### The basics

```go
package main

import "fmt"

func main() {
	x := 7

	if x > 0 {
		fmt.Println("positive")
	} else if x < 0 {
		fmt.Println("negative")
	} else {
		fmt.Println("zero")
	}
}
```

Three things to notice:

- **No parentheses around the condition.** `if x > 0` not `if (x > 0)`. The braces are required, however — even for a one-line body.
- **`else if` and `else` chain on the same line as the closing brace** of the previous block. `gofmt` enforces this.
- **The condition must be a `bool`.** Go won't accept `if x { … }` for a non-bool `x`.

Comparison operators: `==`, `!=`, `<`, `>`, `<=`, `>=`. Logical operators: `&&` (and), `||` (or), `!` (not). Both groups produce `bool`.

---

### A worked example

The expense theme — a first version of the `Categorise` function with an `if`/`else if`/`else` chain:

```go
package main

import "fmt"

// Categorise — first version, using an if/else if/else chain.
func Categorise(amount float64) string {
	if amount < 10 {
		return "snack"
	} else if amount <= 50 {
		return "regular"
	}
	return "splurge"
}

func main() {
	fmt.Println(Categorise(4.50))   // snack
	fmt.Println(Categorise(12.00))  // regular
	fmt.Println(Categorise(75.00))  // splurge
}
```

We'll come back to this in concept 4 and refactor it to a tagless `switch`. Same behaviour, slightly cleaner shape.

---

### Common mistake

Forgetting that the condition must be `bool` and trying C-style truthy ints:

```go
package main

func main() {
	x := 7
	if x {           // wrong: Go won't coerce int to bool
		// ...
	}
}
```

Go complains:

```
non-boolean condition in if statement
```

Write `if x != 0 { … }` or `if x > 0 { … }` — be explicit about what "truthy" means.

---

### Recap

- `if` / `else if` / `else` — conditions are bare booleans, no parentheses, braces required.
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`. Logical: `&&`, `||`, `!`.
- Go does not coerce non-booleans to `bool` — be explicit.

---

## Concept 2: Early returns and guard clauses

### Motivation

Deeply nested `if` statements are hard to read. Go culture leans hard toward returning early — handle the bad / unusual cases first, then let the happy path flow through the bottom of the function with no indentation. The pattern is sometimes called "guard clauses."

---

### The basics

A function with one input check, written two ways:

```go
// Nested style — discouraged.
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

// Early-return style — preferred.
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

Both functions compute the same thing. The second one is easier to read because each guard clause sits on the left margin and you can scan straight down to the happy path at the bottom.

This is *the* idiomatic Go shape for any function that has prerequisites — it gets used heavily once we introduce error handling in lesson 04 (`if err != nil { return err }`).

---

### A worked example

A small validator that rejects empty input or weird amounts:

```go
package main

import "fmt"

func describe(category string, amount float64) string {
	if category == "" {
		return "(no category)"
	}
	if amount < 0 {
		return "(negative amount)"
	}
	if amount == 0 {
		return fmt.Sprintf("%s: free", category)
	}
	return fmt.Sprintf("%s: €%.2f", category, amount)
}

func main() {
	fmt.Println(describe("coffee", 4.50))   // coffee: €4.50
	fmt.Println(describe("", 12))           // (no category)
	fmt.Println(describe("coffee", -1))     // (negative amount)
	fmt.Println(describe("free-sample", 0)) // free-sample: free
}
```

Three guards, then the happy path. No `else` anywhere — each guard ends with `return`, so the rest of the function can assume the prior guards passed.

---

### Common mistake

The "useless else" — keeping an `else` after a guard that already returns:

```go
// Style smell — flagged by some linters.
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	} else {
		return fmt.Sprintf("€%.2f", amount)
	}
}
```

The `else` adds no information — if `amount < 0` returned, control already left the function. Drop the `else`:

```go
func describe(amount float64) string {
	if amount < 0 {
		return "(negative)"
	}
	return fmt.Sprintf("€%.2f", amount)
}
```

Reads identically; one less indentation level.

---

### Recap

- Prefer early returns / guard clauses over deeply nested `if`.
- After a guard with `return`, drop the redundant `else`.
- This pattern is *the* shape Go's stdlib uses everywhere — get comfortable with it now; lesson 04's error handling leans on it heavily.

---

## Concept 3: `for` — Go's only loop

### Motivation

Go has *one* loop keyword: `for`. There's no `while`, no `do…while`, no `loop`. The single `for` keyword takes three syntactic forms which between them cover every iteration pattern other languages spread across multiple keywords.

---

### The basics

The three forms:

```go
package main

import "fmt"

func main() {
	// 1. Infinite for — like `while true` in other languages.
	count := 0
	for {
		if count >= 3 {
			break
		}
		fmt.Println("infinite", count)
		count++
	}

	// 2. Condition-only for — like a classical `while`.
	x := 1
	for x < 10 {
		x *= 2
	}
	fmt.Println("doubled to", x) // 16

	// 3. C-style for — init, condition, post.
	for i := 0; i < 3; i++ {
		fmt.Println("c-style", i)
	}
}
```

Each form is just `for` with a different number of clauses. The C-style form is the one you'll write most often. `break` exits the loop; `continue` skips to the next iteration.

---

### A worked example

Walking a slice with `for ... range` — a fourth form, not technically a "shape" but worth showing now since we'll use it in the main exercise:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12.00, 75.00}
	for i, a := range amounts {
		fmt.Printf("%d: €%.2f\n", i, a)
	}

	// Index-only — drop the value with _.
	for i := range amounts {
		fmt.Println(i)
	}

	// Value-only — drop the index with _.
	for _, a := range amounts {
		fmt.Println(a)
	}
}
```

`range` works on slices, arrays, strings, maps, and channels. We're using it on a slice here. We cover slices properly in lesson 05; for now, treat `[]float64` as "a sequence of float64s" you can `range` over.

---

### Common mistake

C-style for with the wrong loop bound — off by one:

```go
package main

import "fmt"

func main() {
	amounts := []float64{4.50, 12.00, 75.00}
	for i := 0; i <= len(amounts); i++ {  // wrong: <= overruns
		fmt.Println(amounts[i])
	}
}
```

Go panics at runtime:

```
panic: runtime error: index out of range [3] with length 3
```

Slices and arrays are zero-indexed: valid indices are `0` to `len-1`. The loop condition should be `i < len(amounts)`, not `<=`. (Better still: use `for i, a := range amounts` and let the compiler handle the bounds.)

---

### Recap

- One keyword (`for`), three syntactic forms: infinite, condition-only, C-style.
- `break` / `continue` work as you'd expect.
- `for ... range` walks a collection — convenient and bounds-safe.
- Always check loop bounds: indices run `0` to `len-1`.

---

## Concept 4: `switch`

### Motivation

A long `if`/`else if` chain comparing one value against many possibilities is a "switch in disguise." Go's `switch` makes that intent explicit and reads more cleanly. Go's `switch` is also slightly more powerful than C/Java's: it can switch on a tag (`switch x { case 1: ... }`) or stay tagless (`switch { case x > 0: ... }`).

Cases do **not** fall through by default — the opposite of C. You don't need `break`.

---

### The basics

Tagged switch — branch on the value of an expression:

```go
package main

import "fmt"

func main() {
	day := "Tue"
	switch day {
	case "Mon", "Tue", "Wed", "Thu", "Fri":
		fmt.Println("weekday")
	case "Sat", "Sun":
		fmt.Println("weekend")
	default:
		fmt.Println("unknown day:", day)
	}
}
```

Tagless switch — each case is an arbitrary boolean expression:

```go
package main

import "fmt"

func main() {
	amount := 42.0
	switch {
	case amount < 10:
		fmt.Println("snack")
	case amount <= 50:
		fmt.Println("regular")
	default:
		fmt.Println("splurge")
	}
}
```

Two things to remember:

- **No fall-through by default.** Once a case body runs, the switch is done. (You can opt in with `fallthrough`, but it's rare.)
- **Multiple values per case** — comma-separated, as in the weekday example.

---

### A worked example

The `Categorise` refactor: take concept 1's if-else version and turn it into a tagless switch. Same behaviour, less ceremony:

```go
// Categorise — second version, tagless switch.
func Categorise(amount float64) string {
	switch {
	case amount < 10:
		return "snack"
	case amount <= 50:
		return "regular"
	default:
		return "splurge"
	}
}
```

This is the version you'll implement in `exercises/main.go`. It also pairs nicely with `Tally`, where the inner switch dispatches on the string the outer `Categorise` returns:

```go
for _, a := range amounts {
	switch Categorise(a) {
	case "snack":
		snack++
	case "regular":
		regular++
	case "splurge":
		splurge++
	}
}
```

---

### Common mistake

Adding a `break` because muscle memory — and being surprised it does nothing useful:

```go
switch amount {
case 0:
	fmt.Println("zero")
	break  // unnecessary — switch doesn't fall through anyway
case 1:
	fmt.Println("one")
}
```

The `break` doesn't break out of any enclosing loop — it just exits the (already exiting) switch. `gofmt` won't remove it; `golangci-lint`'s `staticcheck` will flag it.

If you genuinely want fall-through from one case into the next, use the `fallthrough` keyword on the line *before* the next case — but you almost never want this. The default no-fallthrough behaviour is what makes `switch` safe in Go.

---

### A note for later

Go has a third form — `switch v.(type)` — for runtime type checks. It needs interfaces, which we cover in lesson 10. You won't write or read one in Phase 1.

---

### Recap

- Tagged `switch x { case … }` and tagless `switch { case bool-expr }`.
- No fall-through by default (the safe default).
- Multiple values per case with commas.
- Use `switch` over an `if`/`else if` chain when you're branching on the same value/condition.

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `WarmupClassify(n int) string` — return `"positive"` / `"negative"` / `"zero"` using `if`/`else if`/`else`.
- `WarmupFizzBuzz(n int) []string` — build the FizzBuzz sequence for `1..n` with a `for` loop and an `if`/`else if` (or `switch`) chain inside. Return an empty (non-nil) slice when `n < 1`.

```bash
cd lessons/03-control-flow/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `Categorise(amount float64) string` — return `"snack"` / `"regular"` / `"splurge"` by amount. Use a tagless `switch`.
- `Tally(amounts []float64) (snack, regular, splurge int)` — walk the slice with `for ... range` and count per category.
- `runMain()` — already provided. It demos the interactive `fmt.Scanln` loop and is *not* covered by `go test`. Try it from the `solutions/` package once your code compiles:

```bash
printf "3\n4.50\n12\n75\n" | go run ./lessons/03-control-flow/solutions
```

```bash
cd lessons/03-control-flow/exercises
go test -v
```

Note:
For live: walk through the if-vs-switch refactor in concept 4 on the projector — same code shape, side by side. The "no fall-through" point lands well as a contrast with C/Java if students have that background. Demo `runMain` interactively at the end if there's time; the off-by-one slide also pays for itself when a student inevitably writes `i <= len(s)`.

---

## What we learned

- `if`, `else if`, `else` — bare booleans (no parentheses), braces required, no implicit truthiness.
- Early returns / guard clauses keep code flat — use them; drop redundant `else`s.
- `for` is Go's only loop. Three forms cover everything; `for ... range` walks collections.
- `switch` — tagged or tagless, no fall-through by default; clearer than long if-chains for multi-way branches.
- A tagless `switch { case bool-expr }` is just an if-chain in nicer clothes — pick whichever reads better in context.

---

## Up next

Lesson 04 — Functions & first tests (multi-return values, the `error` type, table tests, `t.Errorf` vs `t.Fatalf`).
