# Lesson 06: Structs and methods

## Learning goals

- Define your own types with `type T struct { ... }` and construct values with named-field literals.
- Attach behaviour to a type via methods (`func (t T) M() ...`); know when to use a value receiver.
- Recognise pointer receivers (`func (t *T) M() ...`) and the value-vs-pointer distinction — Phase 1 prefers value receivers; lesson 09 covers pointers fully.
- See struct embedding (`type B struct { A; ... }`) for composition, and exported vs unexported names for visibility — both get the full treatment in lesson 07.

## Prerequisites

- Lessons 01-05. In particular, lesson 05's `TotalsByCategory(amounts, categories)` is refactored here into `TotalsByCategory(es []Expense)` — go back and skim lesson 05 if the map-accumulation idiom isn't fresh.

## Concepts

### Struct definition and literals

A struct bundles related fields into one named type:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}
```

Four ways to build a value:

```go
a := Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"}    // named-field
b := Expense{"2026-05-12", 12, "lunch"}                                 // positional — avoid
c := Expense{Date: "2026-05-12"}                                        // partial — others zero
var d Expense                                                           // zero-value
```

Three idioms:

- **Always use named-field literals.** Positional literals look concise but break the moment someone adds a field.
- **Every field starts at its zero value.** `""`, `0`, `false`, `nil` (for slices/maps/pointers/interfaces).
- **`%+v`** in `Printf` prints field names — useful for debugging.

**Common mistake.** Positional literal + new field = compile error:

```go
type Expense struct {
	Date   string
	Amount float64
}

e := Expense{"2026-05-12", 4.50}     // works

// Later, someone adds Category:
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

e := Expense{"2026-05-12", 4.50}     // ERROR: too few values
```

Fix: use named-field literals from the start. `Expense{Date: "...", Amount: ...}` survives field additions — the new field just gets its zero value.

### Methods (value receivers)

A method is a function with a **receiver** before the name:

```go
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func (e Expense) IsHigh() bool {
	return e.Amount > 50
}
```

Call sites look like methods in any OO language:

```go
e := Expense{Date: "2026-05-12", Amount: 75, Category: "rent"}
e.Format()   // "2026-05-12  €75.00   rent"
e.IsHigh()   // true
```

Three things:

- **Receiver syntax** `(e Expense)` is like an extra parameter that comes before the method name. Convention: receiver names are short (`e`, not `expense`).
- **`e.Format()` is shorthand for `Format(e)`** under the hood. The dot is just sugar.
- **Methods live in the same package as the type.** You can't add methods to `int`, `string`, or anything from another package — only to types you declared yourself.

**Common mistake.** Writing a function and trying to call it as a method:

```go
func Format(e Expense) string { ... }   // plain function, not a method

e := Expense{...}
e.Format()                                // error: undefined
```

Fix: add the receiver. Change `func Format(e Expense)` to `func (e Expense) Format()`.

### Pointer receivers (light)

Value receivers `(t T)` get a *copy* of the receiver. Mutations don't reach the caller:

```go
type Counter struct {
	n int
}

func (c Counter) IncrementByValue() {
	c.n++   // mutates the local copy only
}

c := Counter{n: 0}
c.IncrementByValue()
fmt.Println(c.n)   // 0 — IncrementByValue had no effect
```

Pointer receivers `(t *T)` get a pointer to the value; mutations stick:

```go
func (c *Counter) IncrementByPointer() {
	c.n++
}

c.IncrementByPointer()
fmt.Println(c.n)   // 1 — mutation stuck
```

Two rules of thumb (lesson 09 explains *why*):

- **Value receiver when the method just reads.** Most of Phase 1's methods are value receivers.
- **Pointer receiver when the method mutates the receiver** — `IncrementByPointer` above. Or when the struct is large and copying would be expensive.

Convention: if one method on a type needs a pointer receiver, make *all* methods on that type pointer receivers. Be consistent.

Phase 1 leans on a different pattern altogether: **return a new value**, immutability by default.

```go
// Instead of mutating, return a new Expense.
func (e Expense) Discounted(rate float64) Expense {
	e.Amount *= (1 - rate)
	return e
}

// Caller reassigns:
e = e.Discounted(0.10)
```

This sidesteps the pointer-receiver question entirely and is often more readable. Use it when you can.

**Common mistake.** A value-receiver method trying to mutate the struct:

```go
func (e Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)   // mutates the copy, not the caller's value
}

e := Expense{Amount: 100}
e.ApplyDiscount(0.10)
fmt.Println(e.Amount)   // 100 — discount didn't stick
```

Two fixes: `*Expense` pointer receiver, or return a new value (`Discounted` above).

### Embedding + exported/unexported (light)

**Embedding** is Go's "type B has-a type A" shape, with field promotion:

```go
type Address struct {
	City    string
	Country string
}

type Customer struct {
	Address          // embedded — no field name
	Name string
}

func (a Address) Format() string {
	return fmt.Sprintf("%s, %s", a.City, a.Country)
}

c := Customer{
	Address: Address{City: "Helsinki", Country: "Finland"},
	Name:    "Aki",
}
c.City                // "Helsinki" — promoted from Address
c.Address.City        // also "Helsinki" — explicit path
c.Format()            // "Helsinki, Finland" — method also promoted
```

`Address` fields and methods are *promoted* onto `Customer` — `c.City` looks identical to a real field on `Customer`. It isn't inheritance; it's composition with a dot-notation shortcut.

**Exported vs unexported** is pure case-of-first-letter:

```go
type Expense struct {
	Date     string   // exported (capital D) — visible from other packages
	Amount   float64  // exported
	Category string   // exported
	notes    string   // unexported (lower-case n) — package-private
}
```

The rule applies to everything: fields, methods, functions, types, constants, variables. Capitalised → public; lower-case → private to the declaring package.

We've been declaring all exercise/solution fields as exported because the tests construct values with struct literals. Lesson 07 shows how to design APIs that hide internals behind unexported names.

**Common mistake.** Lowercase = unexported = invisible to other packages:

```go
// In package mypkg:
type Counter struct {
	count int
}

// In another package:
c := mypkg.Counter{count: 5}   // ERROR: cannot refer to unexported field
```

Fix: capitalise the field (`Count`), or expose it through a method (`func (c Counter) Count() int { return c.count }`).

## Exercise: warm-up

Two things in `exercises/warmup.go`:

- `type WarmupPoint struct { X, Y float64 }` — already defined.
- `func (p WarmupPoint) DistanceFromOrigin() float64` — implement. Return `math.Sqrt(p.X*p.X + p.Y*p.Y)`.

Then fill in the skeleton `TestWarmupPointDistanceFromOrigin` in `exercises/warmup_test.go`. Floats need a tolerance — use `math.Abs(got - want) < 1e-9` (import `math` if you want) or compute the absolute difference inline.

## Exercise: main

Three things in `exercises/main.go`:

- `type Expense struct { Date string; Amount float64; Category string }` — already defined.
- `func (e Expense) Format() string` — use `fmt.Sprintf` with `"%s  €%-7.2f %s"`. Same byte-for-byte output as lesson 04's `FormatExpense`. The doc comment has the three expected output strings.
- `func (e Expense) IsHigh() bool` — return `e.Amount > 50`. Note: 50 itself is **not** high.
- `func TotalsByCategory(es []Expense) map[string]float64` — walk the slice with `range`, accumulate into a `map[string]float64`. Empty/nil input must return a non-nil empty map.

Fill in the three skeleton tests in `exercises/main_test.go`:

- `TestExpenseFormat` — at least 3 cases. The `%-7.2f` width matters — build the expected strings carefully.
- `TestExpenseIsHigh` — at least 4 cases, including the boundary value 50 (not high) and just over 50 (high).
- `TestTotalsByCategory` — at least 4 cases. Cover distinct categories, repeated category (totals accumulate), single-element slice, and empty input (non-nil empty map).

> A note on `make test-lesson LESSON=06-structs`: lesson 06's exercise tests now ship with empty `cases` slices, so the make output may show `ok` (vacuous pass) before you add cases. Don't be fooled — run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/06-structs/exercises
go test -run Warmup -v   # warm-up only
go test -v                # warm-up + main
```

Keep the lesson 04 habits:

```bash
gofmt -w .       # reformat to canonical Go style
go vet ./...     # subtle bug catcher
```

CI enforces both.

## Going further

### Read

- [Effective Go — Structs](https://go.dev/doc/effective_go#composite_literals) — short canonical reference.
- [Effective Go — Embedding](https://go.dev/doc/effective_go#embedding) — the "composition not inheritance" case study.
- [Go FAQ — Why doesn't Go have inheritance?](https://go.dev/doc/faq#inheritance) — the design rationale; useful counterpoint if you're coming from an OO background.

### Try

- **A pointer-receiver variant.** Rewrite `Expense.IsHigh()` with a pointer receiver and observe that nothing changes for the caller (because `IsHigh` doesn't mutate). Then write a mutating method — `Expense.Bump(amount float64)` that adds to `Amount` — and observe that you need a pointer receiver for it to stick.
- **Embedding in action.** Add a `Timestamped` type with a `CreatedAt string` field, then define `type TimestampedExpense struct { Expense; Timestamped }` and verify that `te.Format()` and `te.CreatedAt` both work without explicit field paths.
- **An unexported field.** Add a private `note string` field to `Expense` and an `AddNote(string)` method (value receiver — returns a new `Expense`). Run `go test` from `solutions/` — the test should still pass because it doesn't touch the unexported field. Then try to access `e.note` from a scratch file in `exercises/` — it works, because both files are in the same package. Move that scratch file to a different package (a `cmd/play/main.go` would do it) and watch the compiler reject it.
