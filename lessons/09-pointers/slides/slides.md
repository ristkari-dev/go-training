<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">09</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Pointers</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Understand <code>&amp;</code> (address-of) and <code>*</code> (dereference); use pointer receivers when methods need to mutate; pick value vs pointer receivers consistently; and avoid the most common runtime panic in Go — the nil-pointer dereference.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Pointer basics** — `&` and `*`, what a pointer is, value vs reference semantics, nil pointers.
- **Pointer receivers — deep dive** — `(c *T) M()` vs `(c T) M()`; method set rules; consistency across methods.
- **When to use pointers** — mutation (the main reason), large structs (the other reason), the "don't reach by default" rule.
- **Common nil-pointer pitfalls** — uninitialised pointer dereferences, nil pointer fields, methods on nil receivers.

---

## Concept 1: Pointer basics

### Motivation

Up to lesson 8, every variable held its value directly. Now we meet the other half: a **pointer** holds the *address* of a value. Pointers let you share a value across function/method boundaries — so a callee can mutate the caller's data — and avoid copying large structs around. Lesson 06 mentioned pointer receivers briefly; lesson 09 goes deep.

---

### The basics

Two operators do all the work:

- `&x` — "address of x" → gives you a pointer (type `*T`)
- `*p` — "dereference p" → gives you the value the pointer points at (type `T`)

```go
package main

import "fmt"

func main() {
	x := 7
	p := &x          // p has type *int; p points at x
	fmt.Println(*p)  // 7 — dereferencing p reads x

	*p = 42          // dereferencing p as the assignment target writes x
	fmt.Println(x)   // 42 — x has been mutated via the pointer
}
```

Three things to internalise:

- **Types match up.** If `x` is `int`, `&x` is `*int`. If `x` is `Expense`, `&x` is `*Expense`. The `*T` notation is uniform.
- **Reading a pointer: `*p`.** Writing through a pointer: `*p = newValue`. The `*` works both ways.
- **The zero value of any pointer type is `nil`.** A `var p *int` is `nil` — pointing nowhere. Dereferencing it (`*p`) panics at runtime.

---

### A worked example

Passing a pointer to a function so it can mutate the caller's value:

```go
package main

import "fmt"

// Plain function — the caller passes a *int, the function mutates the
// caller's int through the pointer.
func increment(p *int) {
	*p++   // *p reads the int; *p = *p + 1 writes it back
}

func main() {
	x := 5
	increment(&x)         // pass the address of x
	fmt.Println(x)        // 6 — x was mutated

	var p *int            // nil pointer
	// increment(p)       // would panic: nil pointer dereference inside increment
	_ = p
}
```

This is the building block. Method receivers (concept 2) are a special case of this pattern — when you call `e.Bump(5)` on a value `e`, Go silently takes `&e` and passes it to the pointer-receiver method.

---

### Common mistake

Dereferencing a `nil` pointer — Go's most common runtime panic:

```go
package main

import "fmt"

func main() {
	var p *int       // nil
	fmt.Println(*p)  // PANIC: runtime error: invalid memory address or nil pointer dereference
}
```

Same panic shows up in subtler forms — for example, a struct with a pointer field that someone forgot to initialise:

```go
type Box struct{ Contents *int }
b := Box{}            // b.Contents is nil
fmt.Println(*b.Contents)  // PANIC
```

The fix is always the same: make sure the pointer is non-nil before you dereference. Either initialise (`b.Contents = new(int)`), or check (`if b.Contents != nil { fmt.Println(*b.Contents) }`).

---

### Recap

- `&x` produces a pointer; `*p` dereferences one.
- Type `*T` is "pointer to T"; zero value is `nil`.
- Pass a pointer to let a callee mutate the caller's value.
- Dereferencing `nil` panics — most common runtime panic in Go.

---

## Concept 2: Pointer receivers — deep dive

### Motivation

Lesson 06 said "value receivers copy; pointer receivers share" and left it there. Lesson 09 looks at the consequence: which methods mutate, what the method set rules are, and why consistency matters.

---

### The basics

A method receiver is just a special first argument. Two forms:

```go
type Counter struct {
	n int
}

// Value receiver — `c` is a COPY of the caller's Counter.
// Mutations to c stay inside this function.
func (c Counter) Value() int {
	return c.n
}

// Pointer receiver — `c` is a POINTER to the caller's Counter.
// Mutations through c (i.e. c.n++) affect the caller.
func (c *Counter) Inc() {
	c.n++
}
```

Call sites look identical:

```go
var c Counter
c.Inc()           // Go takes &c implicitly, passes as the *Counter receiver
fmt.Println(c.Value())  // 1 — mutation stuck
```

Two rules:

- **Use a pointer receiver to mutate the receiver.** No mutation? Use a value receiver.
- **Use a pointer receiver to avoid copying a large struct.** For tiny types (`Counter` with one int, `Point` with two floats), the copy is free; value receivers are fine. For big structs (10+ fields, embedded slices), pointer receivers save the copy.

---

### A worked example

Lesson 09's main exercise — `Expense` with two pointer-receiver methods:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func (e Expense) Format() string     // value receiver — reads, doesn't mutate
func (e Expense) IsHigh() bool        // value receiver — reads
func (e *Expense) ApplyDiscount(rate float64)  // pointer receiver — mutates
func (e *Expense) Bump(amount float64)         // pointer receiver — mutates

// Calling them:
e := Expense{Amount: 100}
fmt.Println(e.Format())   // works — value-receiver method
e.ApplyDiscount(0.10)     // works — Go takes &e implicitly
fmt.Println(e.Amount)     // 90 — mutation stuck
```

Notice the call sites are uniform — `e.M()` works whether `M` is a value-receiver or pointer-receiver method. The DIFFERENCE is whether mutations to `e` survive after the call.

### Method set rules (the formal version)

- A type `T` has all its value-receiver methods.
- A type `*T` has all of `T`'s methods PLUS the pointer-receiver methods.
- When the receiver variable is **addressable** (a local variable; a struct field), Go automatically takes `&` for pointer-receiver method calls.

In practice: just write `e.M()` regardless. The exception is when the variable is NOT addressable — e.g. a map value (`m["key"].Bump(5)` doesn't compile because map values aren't addressable). Rare enough to defer until you hit it.

---

### Common mistake

Mixing receiver types on the same type for no good reason:

```go
type Cart struct{ items []string }

func (c Cart) Count() int            // value receiver
func (c *Cart) Add(item string)      // pointer receiver
func (c Cart) Total() float64        // value receiver
```

This compiles. But the inconsistency confuses readers: do I get a copy or a reference? Does `c.Total()` see updates from `c.Add(...)` made elsewhere?

**The convention:** if ANY method on a type uses a pointer receiver, make ALL methods on that type use pointer receivers. Be consistent across the type's whole method set.

(Our lesson 09 `Expense` deliberately mixes them, AS A TEACHING EXAMPLE — the contrast between `Format`/`IsHigh` (value) and `ApplyDiscount`/`Bump` (pointer) is the lesson. Real-world code would pick one and stick with it. The README notes this.)

---

### Recap

- `(c T) M()` — value receiver, copies. Read-only by convention.
- `(c *T) M()` — pointer receiver, shares. Use for mutation or large structs.
- Call sites are identical — `e.M()` works either way.
- Be consistent across a type's methods: if any method needs a pointer receiver, all should.

---

## Concept 3: When to use pointers

### Motivation

A common newcomer instinct is "I should use pointers everywhere for performance." That's wrong — Go's compiler is good at avoiding unnecessary copies, and pointers carry their own costs (the cognitive overhead of "is this nil?", an extra heap allocation in some cases). The actual rules are simple.

---

### The basics

Use a pointer when:

- **You need to mutate the value through a function or method call.** This is the dominant reason.
- **The value is genuinely large** (10+ fields, embedded slices/maps in the hundreds of bytes). Copying becomes wasteful.
- **You need to express "no value here"** with `nil`. A `*Config` parameter where `nil` means "use defaults" is a common pattern (though optional struct fields often communicate the same idea more clearly).

Use a value when:

- **The type is small** (`int`, `float64`, `string`, a struct with 1-3 small fields). The copy is free; pointers add indirection for no gain.
- **The method is read-only.** Value receivers make this intent explicit.

---

### A worked example

Phase 1's `Expense` deliberately stayed all-value-receiver:

```go
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

func (e Expense) Format() string      // value — small struct, read-only
func (e Expense) IsHigh() bool         // value — small struct, read-only
```

Lesson 09 adds mutation methods — pointer receivers:

```go
func (e *Expense) ApplyDiscount(rate float64)
func (e *Expense) Bump(amount float64)
```

The decision was driven by mutation, not by size. `Expense` is small (24 bytes-ish — two strings + a float). Copying it is free.

A counter-example — a hypothetical `Report` with thousands of lines:

```go
type Report struct {
	Lines []string  // could be a megabyte
}

func (r *Report) Summarise() string  // pointer — avoid copying the slice header
```

Even though `Summarise` doesn't mutate `r`, we use a pointer receiver because copying `*Report` is cheaper than copying the `Report` struct (whose slice header is 24 bytes — fine — but conventional). Be honest with yourself about whether "large" applies before reaching for `*`.

### Escape analysis (high-level intuition only)

Go's compiler decides whether each variable lives on the **stack** (fast, automatically reclaimed at function return) or the **heap** (slower, garbage-collected). Taking `&x` doesn't automatically put `x` on the heap — the compiler will keep it on the stack if it can prove the pointer doesn't outlive the function. That's **escape analysis**.

You almost never need to think about this. The rule of thumb: **don't use pointers for performance unless a benchmark shows a real difference.** Phase 3 (lesson 22) covers profiling and benchmarks properly; you can defer all "is this faster?" questions until then.

---

### Common mistake

Using pointers as the default "just in case":

```go
// Wrong — gratuitous pointer for a tiny read-only type.
func describe(p *string) string {
	return "got: " + *p
}

// Right — string is small and read-only.
func describe(s string) string {
	return "got: " + s
}
```

Pointers add cognitive overhead. The reader has to think "is this nil? Will the caller's value change behind my back?" For small read-only data, the value form is simpler.

A related anti-pattern: `*[]T` (pointer to slice) — slices are already reference-like under the hood (the header is a struct with a pointer + length + capacity). You almost never need `*[]T`. Use `[]T`.

---

### Recap

- Mutation → pointer receiver.
- Large struct → pointer receiver (with a real benchmark, not a guess).
- Small read-only type → value receiver.
- Don't reach for pointers by default. They add cognitive overhead.

---

## Concept 4: Common nil-pointer pitfalls

### Motivation

`nil` pointer dereferences are Go's most common runtime panic. Every Go programmer hits this at some point — usually multiple times in their first month. The patterns are predictable; learning to recognise them saves hours.

---

### The basics

Three flavours:

**1. Uninitialised pointer.**

```go
var p *int       // nil
fmt.Println(*p)  // PANIC
```

The fix: initialise with `new()` or `&` of something concrete. Or check before dereferencing.

**2. Nil pointer field in a struct.**

```go
type Box struct{ Contents *int }
b := Box{}                  // b.Contents is nil (zero value of *int)
fmt.Println(*b.Contents)    // PANIC
```

The fix: don't ship structs with required pointer fields unless your constructor enforces them. Or check before dereferencing.

**3. Methods on nil receivers.**

```go
type List struct{ next *List; value int }

func (l *List) Length() int {
	if l == nil {
		return 0
	}
	return 1 + l.next.Length()
}

var empty *List  // nil
fmt.Println(empty.Length())  // 0 — explicitly handled
```

This actually WORKS in Go — `nil.Length()` doesn't panic as long as the method handles `l == nil` explicitly. Many stdlib types use this pattern. But it's also a common foot-gun: forget the `if l == nil` and the next deref panics.

---

### A worked example

Lesson 09's `Expense` doesn't actually use pointer fields, so this slide is mostly preventive. But here's the shape you'll see in real code (and in lesson 10's interfaces):

```go
type Storage struct {
	cache *Cache  // pointer — may be nil if caching is disabled
}

func (s *Storage) Load(key string) (string, error) {
	if s.cache != nil {                  // guard
		if v, ok := s.cache.Get(key); ok {
			return v, nil
		}
	}
	return s.loadFromDisk(key)
}
```

The guard pattern (`if x != nil { use x }`) shows up wherever optional pointers exist. Get into the habit of asking yourself: "Could this pointer be nil here? What happens if it is?"

---

### Common mistake

Returning a nil pointer alongside a nil error — caller can't tell apart "successfully found nothing" vs "failed":

```go
func Find(id string) (*User, error) {
	u := db.Lookup(id)
	if u == nil {
		return nil, nil   // bad: caller doesn't know if there was an error
	}
	return u, nil
}
```

Two fixes:

```go
// Option A: a sentinel error
var ErrNotFound = errors.New("not found")

func Find(id string) (*User, error) {
	u := db.Lookup(id)
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// Option B: a value type + ok flag
func Find(id string) (User, bool) { ... }
```

Lesson 11 (errors) goes deep on the sentinel pattern. The general rule: **don't return `(nil, nil)` from `(T, error)`-returning functions** — the caller has no way to discriminate, and you'll cause panics downstream.

---

### Recap

- Three nil panic flavours: uninitialised, nil struct field, missing nil-guard on a receiver method.
- The fix is always: initialise, or check.
- Methods on nil receivers CAN work if the method handles `l == nil` explicitly. Don't rely on this without testing.
- Don't return `(nil, nil)` from `(T, error)` functions. Sentinel + error, or value + ok.

---

## Practice

### Warm-up

In `exercises/warmup/counter/`:

- `counter.go` — implement `(c *Counter) Inc()` and `(c Counter) Value() int`. Both are one-liners.
- `counter_test.go` — fill in the skeleton (TestCounter table with at least 4 cases + TestCountersDoNotInterfere asserting two Counters are independent).

```bash
cd lessons/09-pointers/exercises
go test ./warmup/counter/... -v
```

---

### Main

In `exercises/expense/`:

- `expense.go` — implement four methods:
  - `(e Expense) Format() string` — same as lesson 04/06/07/08
  - `(e Expense) IsHigh() bool` — `e.Amount > 50`
  - `(e *Expense) ApplyDiscount(rate float64)` — `e.Amount *= (1 - rate)`
  - `(e *Expense) Bump(amount float64)` — `e.Amount += amount`
- `expense_test.go` — fill in four skeleton tests. The mutation tests (`TestExpenseApplyDiscount`, `TestExpenseBump`) verify that the change STUCK on the caller's Expense — that's the lesson's payoff.

```bash
cd lessons/09-pointers/exercises
go test ./expense/... -v
```

Note:
For live: type out a small example on the projector — `e := Expense{Amount: 100}`; `e.ApplyDiscount(0.10)`; `fmt.Println(e.Amount)`. The "90, not 100" moment is the lesson's "aha." Then do the same with a value receiver (rewrite ApplyDiscount as `func (e Expense)` instead of `func (e *Expense)`) and watch e.Amount stay at 100. That contrast lands the receiver-mechanic forever.

---

## What we learned

- Pointers (`&x`, `*p`): pass a value's address so a callee can mutate it. Type is `*T`; zero value is `nil`; dereferencing `nil` panics.
- Pointer receivers `(c *T) M()` — for mutation or large structs. Value receivers `(c T) M()` — for read-only or small types.
- Call sites are uniform: `e.M()` works regardless of receiver type (assuming the variable is addressable).
- Convention: pick one receiver type per concrete type. Mixing is allowed but confuses readers.
- Don't reach for pointers by default — they add cognitive overhead and aren't a "performance trick."
- Nil-pointer panics: most common Go runtime crash. Initialise or guard.

---

## Up next

Lesson 10 — Interfaces. Implicit satisfaction, `io.Reader`/`io.Writer`, `any`, type assertions. The tracker gets a `Store` interface (with JSON + in-memory implementations) — your first "swap the implementation under the same API" refactor.
