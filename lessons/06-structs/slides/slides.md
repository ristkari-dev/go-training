<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">06</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Structs and methods</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Define your own types with <code>struct</code>, attach behaviour with methods, prefer value receivers (the Phase 1 default) and understand when pointer receivers come in, plus a light touch of embedding and exported-vs-unexported fields.</p>
</div>
</div>
</div>

---

## What we'll cover

- `type T struct { ... }` — struct definitions, literal forms, zero-value structs.
- Methods with value receivers: `func (t T) M() ...` — calling, method set.
- Pointer receivers (light): `func (t *T) M() ...` — when to mutate, full treatment in lesson 09.
- Embedding + exported/unexported fields — light, full treatment in lesson 07.

---

## Concept 1: Struct definition and literals

### Motivation

Structs are how you bundle related data into a single value. Three named fields beat three parallel slices any day — `Expense{Date, Amount, Category}` carries its own structure, while `(amounts []float64, categories []string)` is just two slices the programmer has to keep in sync. Phase 1 has been edging toward this — lesson 05's `TotalsByCategory` took parallel slices; lesson 06's takes a `[]Expense`.

---

### The basics

```go
package main

import "fmt"

// Type declaration — once per type, at package level.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func main() {
	// 1. Named-field literal (the only style you should write).
	a := Expense{
		Date:     "2026-05-12",
		Amount:   4.50,
		Category: "coffee",
	}

	// 2. Positional literal — fields in declaration order.
	// Avoid this in real code — it breaks the moment someone adds a field.
	b := Expense{"2026-05-12", 12, "lunch"}

	// 3. Partial named literal — omitted fields get their zero value.
	c := Expense{Date: "2026-05-12"}     // Amount: 0, Category: ""

	// 4. Zero-value struct — every field at its type's zero value.
	var d Expense                          // {Date: "", Amount: 0, Category: ""}

	fmt.Printf("%+v\n", a)  // {Date:2026-05-12 Amount:4.5 Category:coffee}
	fmt.Printf("%+v\n", b)
	fmt.Printf("%+v\n", c)
	fmt.Printf("%+v\n", d)

	// Field access.
	fmt.Println(a.Amount, a.Category)

	// Field mutation (when the variable is addressable).
	a.Amount = 5.00
	fmt.Println(a.Amount)
}
```

Three idioms:

- **Always use named-field literals.** `Expense{Date: ..., Amount: ..., Category: ...}` survives field reordering and additions. Positional literals look concise but break under maintenance.
- **Every field starts at its zero value.** Same rule as lesson 02 for plain variables — string `""`, numeric `0`, bool `false`.
- **`%+v` prints field names** in `fmt.Printf` — useful for debugging.

---

### A worked example

The lesson's main exercise — `Expense` with all three fields exported, ready for tests to construct values directly:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func main() {
	es := []Expense{
		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
		{Date: "2026-05-12", Amount: 75, Category: "rent"},
	}
	for _, e := range es {
		fmt.Printf("%s  €%.2f  %s\n", e.Date, e.Amount, e.Category)
	}
}
```

A slice of structs is the bread-and-butter "rows of data" shape in Go. We'll come back to this in concept 2 once we add methods.

---

### Common mistake

Positional literals + adding a field = silent bug:

```go
// Original code.
type Expense struct {
	Date   string
	Amount float64
}

e := Expense{"2026-05-12", 4.50}  // works
```

Later, someone adds `Category` to the struct:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string   // new field
}

e := Expense{"2026-05-12", 4.50}  // ERROR: too few values in struct literal
```

Now every positional literal in the codebase fails to compile. With named literals, this isn't an issue — the new field just gets its zero value.

```go
e := Expense{Date: "2026-05-12", Amount: 4.50}   // Category: "" implicitly
```

Use named-field literals from the start.

---

### Recap

- `type T struct { Field Type; ... }` defines a struct type at package level.
- Use named-field literals: `T{Field: value, ...}`. Avoid positional literals.
- Every field gets its zero value when omitted.
- `%+v` in `Printf` prints field names — handy for debugging.

---

## Concept 2: Methods (value receivers)

### Motivation

A method is a function attached to a type. Behavioural code that operates on a value of type T gets attached to T as a method, so you can call `t.Method()` instead of `Method(t)`. Same logic, better organisation.

---

### The basics

A method declaration looks like a function with one extra bit — the **receiver** before the method name:

```go
package main

import "fmt"

type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Method on Expense. The (e Expense) is the receiver — it makes this function
// a method of type Expense.
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// Another method on the same type.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

func main() {
	e := Expense{Date: "2026-05-12", Amount: 75, Category: "rent"}
	fmt.Println(e.Format())   // 2026-05-12  €75.00   rent
	fmt.Println(e.IsHigh())   // true
}
```

Three things:

- **Receiver syntax: `(e Expense)`** — like an extra parameter that comes before the method name. By convention, the receiver variable is one short letter or two — `e Expense`, not `expense Expense`.
- **Call site: `e.Format()`** — looks just like calling a method in any OO language. Under the hood, this is `Format(e)` with `e` passed as the receiver.
- **Methods live on the same package as the type.** You can only define methods on types declared in your own package — you can't add methods to `int` or `string`.

---

### A worked example

The main exercise — slice of Expense, format each, total by category:

```go
package main

import "fmt"

func main() {
	es := []Expense{
		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
		{Date: "2026-05-12", Amount: 75, Category: "rent"},
	}
	for _, e := range es {
		marker := ""
		if e.IsHigh() {
			marker = "  (!)"
		}
		fmt.Println(e.Format() + marker)
	}
}
```

`Format` and `IsHigh` are methods, so we call them on each element naturally. Same logic, much better organised than a top-level `FormatExpense(date, amount, category)` function with positional arguments.

---

### Common mistake

Forgetting the receiver type — writing a plain function and trying to call it as a method:

```go
// Wrong — this is just a function, not a method.
func Format(e Expense) string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func main() {
	e := Expense{...}
	fmt.Println(e.Format())   // error: e.Format undefined (type Expense has no field or method Format)
}
```

The fix is the receiver — turn `func Format(e Expense)` into `func (e Expense) Format()`. The change is small but it's what makes the function a method.

---

### Recap

- Methods are functions with a receiver — `func (e Expense) Format() string`.
- Call sites look like OO: `e.Format()`.
- You can only define methods on types declared in your own package.
- Convention: receiver names are short (`e Expense`, not `expense Expense`).

---

## Concept 3: Pointer receivers (light)

### Motivation

Value receivers (`(t T)`) get a *copy* of the receiver. That's fine for read-only methods, but if you want to mutate the value, you need a **pointer receiver** (`(t *T)`). Pointer receivers also avoid copying large structs. Pointers are introduced properly in lesson 09; here we just show the syntax and the rule, so you recognise it when you see it.

---

### The basics

```go
package main

import "fmt"

type Counter struct {
	n int
}

// Value receiver — copies the Counter. Mutations don't affect the caller.
func (c Counter) IncrementByValue() {
	c.n++   // mutates the local copy only
}

// Pointer receiver — receives a pointer to the Counter. Mutations affect the caller.
func (c *Counter) IncrementByPointer() {
	c.n++   // mutates the original
}

func main() {
	c := Counter{n: 0}
	c.IncrementByValue()
	fmt.Println(c.n)   // 0 — IncrementByValue had no effect

	c.IncrementByPointer()
	fmt.Println(c.n)   // 1 — IncrementByPointer mutated the original
}
```

Two rules of thumb (lesson 09 explains *why*):

- **Use a value receiver when the method just reads.** `Format()` and `IsHigh()` in our `Expense` are value receivers — they don't change the struct. Most of Phase 1's methods are value receivers.
- **Use a pointer receiver when the method mutates the receiver** — `IncrementByPointer` modifies `c.n`. Or when the struct is large and copying it is expensive.

Tiny stylistic rule: be consistent. If *one* method on a type needs a pointer receiver, the convention is to make *all* methods on that type pointer receivers. You don't have to follow this in Phase 1 since we're using value receivers, but you'll see it in real code.

---

### A worked example

You won't write a pointer receiver in this lesson — the main exercise stays with value receivers throughout. Here's a preview of when you'd reach for one:

```go
type Cart struct {
	Items []string
}

// Pointer receiver: Add() needs to mutate Cart.Items.
func (c *Cart) Add(item string) {
	c.Items = append(c.Items, item)
}

func main() {
	cart := Cart{}
	cart.Add("apple")
	cart.Add("banana")
	fmt.Println(cart.Items)   // [apple banana]
}
```

If `Add` had a value receiver, the appends would happen on a local copy and `cart.Items` would stay empty.

---

### Common mistake

A value-receiver method "trying" to mutate the struct:

```go
type Expense struct {
	Amount float64
}

// Doesn't work — value receiver mutates the copy.
func (e Expense) Discount(rate float64) {
	e.Amount *= (1 - rate)
}

func main() {
	e := Expense{Amount: 100}
	e.Discount(0.10)
	fmt.Println(e.Amount)   // 100 — discount didn't stick
}
```

Two fixes:

- **Pointer receiver:** `func (e *Expense) Discount(rate float64)`. The method now mutates the original.
- **Return a new value:** `func (e Expense) Discounted(rate float64) Expense` returning the discounted version. Caller writes `e = e.Discounted(0.10)`. This is actually the more idiomatic shape — immutability by default, mutate-by-assignment.

Phase 1 leans on the "return a new value" pattern. Pointer receivers come into their own in lesson 09.

---

### Recap

- Value receiver `(t T)` — gets a copy. Read-only.
- Pointer receiver `(t *T)` — gets a pointer. Can mutate.
- Phase 1 prefers value receivers; lesson 09 covers pointers properly.
- The "return a new value" alternative often reads cleaner than a mutating method.

---

## Concept 4: Embedding + exported/unexported (light)

### Motivation

Two short Go-isms to round out the lesson. **Embedding** is Go's answer to "I want type B to include all of type A's fields and methods" — close to inheritance but actually composition. **Exported vs unexported** is Go's encapsulation rule: capitalised names are public, lower-cased names are private to the package. Both get fuller treatment in lesson 07.

---

### The basics

**Embedding** — declare one struct type inside another with no field name:

```go
package main

import "fmt"

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

func main() {
	c := Customer{
		Address: Address{City: "Helsinki", Country: "Finland"},
		Name:    "Aki",
	}
	fmt.Println(c.Name)                // Aki
	fmt.Println(c.City)                // Helsinki — promoted from Address
	fmt.Println(c.Address.City)        // Helsinki — explicit path also works
	fmt.Println(c.Format())            // Helsinki, Finland — method promoted too
}
```

Two things to notice:

- `c.City` works because the field is *promoted* from the embedded `Address` — there's no separate `City` field on `Customer`.
- `c.Format()` works similarly — `Address`'s method is also promoted.

This isn't inheritance — it's composition. Under the hood, `Customer` *has-a* `Address`; the dot notation just hides the indirection when there's no ambiguity. Lesson 07 shows how this maps to "compose a thin orchestrator on top of focused sub-types."

**Exported vs unexported** — pure case-of-first-letter:

```go
type Expense struct {
	Date     string   // exported (capital D)
	Amount   float64  // exported
	Category string   // exported
	notes    string   // unexported — accessible only within this package
}

func (e Expense) AddNote(note string) Expense {   // exported method
	e.notes = note   // assign within the package; outside callers can't read this
	return e
}

func (e Expense) note() string {   // unexported method
	return e.notes
}
```

From another package, `e.Date` is accessible, `e.notes` is not. `e.AddNote(...)` is accessible, `e.note()` is not. The rule is uniform across fields, methods, functions, types, constants, and variables: capitalised first letter → exported, otherwise → package-private.

We've been declaring all our exercise/solution fields as exported (capitalised) because the tests construct values via struct literals — they need to set every field. Lesson 07 shows how to design APIs that hide implementation details behind unexported fields.

---

### A worked example

Our `Expense` type is the simplest possible case — three exported fields, no embedding. Even so, the capitalisation matters: rename `Date` to `date` in `lessons/06-structs/solutions/main.go` and `solutions/main_test.go` would stop compiling, because `Expense{Date: "..."}` in the test would refer to a non-existent field.

A toy embedding example (you won't write this in the exercise but worth seeing):

```go
type Timestamped struct {
	CreatedAt string
}

type TimestampedExpense struct {
	Expense        // embed Expense
	Timestamped    // embed Timestamped
}

func main() {
	te := TimestampedExpense{
		Expense:     Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
		Timestamped: Timestamped{CreatedAt: "2026-05-12T08:30:00Z"},
	}
	fmt.Println(te.Format(), te.CreatedAt)   // both promoted from embedded types
}
```

---

### Common mistake

Forgetting that lowercase = unexported, and being surprised when an external test can't see the field:

```go
// In package mypkg:
type Counter struct {
	count int   // unexported
}

// In another package's test:
c := mypkg.Counter{count: 5}   // ERROR: cannot refer to unexported field
```

Fix: either capitalise the field (`Count`), or expose it through an exported method (`func (c Counter) Count() int { return c.count }`).

The whole-package convention to follow: if a field's value is part of the type's public behaviour, export it. If it's an implementation detail, keep it unexported. Lesson 07 explores this more rigorously.

---

### Recap

- **Embedding:** `type Outer struct { Inner; ... }` (no field name) — `Inner`'s fields and methods are promoted onto `Outer`. Composition, not inheritance.
- **Exported vs unexported:** capitalised first letter → visible to other packages, lowercased → package-private. Applies to fields, methods, functions, types, etc.
- Both topics get fuller treatment in lesson 07 (packages and modules).

---

## Practice

### Warm-up

In `exercises/warmup.go`:

- `type WarmupPoint struct { X, Y float64 }` — already defined.
- `func (p WarmupPoint) DistanceFromOrigin() float64` — implement. Return `math.Sqrt(p.X*p.X + p.Y*p.Y)`.

The skeleton test is in `exercises/warmup_test.go`. Floats need a tolerance — use `math.Abs(got - want) < 1e-9` (and import `math`), or compute the absolute difference inline.

```bash
cd lessons/06-structs/exercises
go test -run Warmup -v
```

---

### Main

In `exercises/main.go`:

- `type Expense struct { Date string; Amount float64; Category string }` — already defined.
- `func (e Expense) Format() string` — return `"%s  €%-7.2f %s"` using `fmt.Sprintf`. Same as lesson 04's `FormatExpense`, but as a method.
- `func (e Expense) IsHigh() bool` — return `e.Amount > 50`. Boundary value 50 is NOT high.
- `func TotalsByCategory(es []Expense) map[string]float64` — walk `es` with `range`, accumulate into a `map[string]float64`. Empty/nil input returns a non-nil empty map.

Skeleton tests in `exercises/main_test.go` — fill in cases for all three.

```bash
cd lessons/06-structs/exercises
go test -v
```

Note:
For live: walk through the field-promotion magic of embedding on the projector — define Timestamped + TimestampedExpense, range over a `[]TimestampedExpense` calling `.Format()` and `.CreatedAt` to show both promoted. The "promoted = composition, not inheritance" point lands well as a contrast for students with OO backgrounds.

---

## What we learned

- `type T struct { ... }` defines a struct; always use named-field literals.
- Methods attach behaviour to a type via the receiver: `func (e Expense) Format()`.
- Value receivers (`(t T)`) get a copy; pointer receivers (`(t *T)`) can mutate. Phase 1 prefers value receivers + "return a new value" patterns; lesson 09 covers pointers properly.
- Embedding (`type B struct { A; ... }`) promotes `A`'s fields and methods onto `B` — composition, not inheritance.
- Exported (capitalised) vs unexported (lowercased) names. Both fields and methods.

---

## Up next

Lesson 07 — Packages and modules (splitting code into packages, formal exported/unexported, `gofmt`/`go vet`/`go doc` tour).
