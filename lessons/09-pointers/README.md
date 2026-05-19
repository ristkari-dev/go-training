# Lesson 09: Pointers

## Learning goals

- Understand `&` (address-of) and `*` (dereference); know what a pointer IS and what its zero value is.
- Use pointer receivers (`func (c *T) M()`) when a method needs to mutate the receiver.
- Pick value vs pointer receivers consistently per type, and know the convention for mixing.
- Recognise the common nil-pointer panic patterns and how to guard against them.

## Prerequisites

- Lessons 01-08 (all of Phase 1). Lesson 06 in particular — the "value vs pointer receivers (light)" concept gets the deep dive here.

## What's different about this lesson

**Phase 2 begins.** A few conventions shift starting lesson 09:

- **Subpackages are the default.** `exercises/warmup/counter/` and `exercises/expense/`, mirrored in `solutions/`. No flat `warmup.go`/`main.go` at the top level.
- **No `cmd/` binary in this lesson.** Lesson 09 is about pointer mechanics; the verification surface is `go test`. Later Phase 2 lessons add `cmd/` binaries where they fit (lesson 10's `Store` swap; lesson 13's CSV importer).
- **`Warmup*` symbol prefix retired.** Subpackages provide their own namespace — there's no name collision risk between `counter.Counter` and `expense.Expense`. Warmup types are now named for what they are.

## Concepts

### Pointer basics

Two operators do all the work:

- `&x` — "address of x" → gives you a pointer (type `*T`)
- `*p` — "dereference p" → gives you the value the pointer points at (type `T`)

```go
x := 7
p := &x          // p has type *int; p points at x
fmt.Println(*p)  // 7 — dereferencing p reads x

*p = 42          // dereferencing p as the assignment target writes x
fmt.Println(x)   // 42 — x has been mutated via the pointer
```

Three things to internalise:

- **Types match up.** If `x` is `int`, `&x` is `*int`. If `x` is `Expense`, `&x` is `*Expense`. The `*T` notation is uniform.
- **Reading: `*p`. Writing: `*p = newValue`.** The `*` works both ways.
- **Zero value of any pointer type is `nil`.** Dereferencing nil panics at runtime — Go's most common runtime crash.

Passing a pointer to a function so it can mutate the caller's value:

```go
func increment(p *int) {
	*p++   // *p reads the int; *p = *p + 1 writes it back
}

x := 5
increment(&x)         // pass the address of x
fmt.Println(x)        // 6 — x was mutated
```

This is the building block. Method receivers are a special case of this pattern.

**Common mistake.** Dereferencing a `nil` pointer:

```go
var p *int       // nil
fmt.Println(*p)  // PANIC: runtime error: invalid memory address or nil pointer dereference
```

Same panic shows up in subtler forms — for example, a struct with a pointer field that someone forgot to initialise:

```go
type Box struct{ Contents *int }
b := Box{}                  // b.Contents is nil
fmt.Println(*b.Contents)    // PANIC
```

The fix is always the same: make sure the pointer is non-nil before dereferencing. Either initialise (`b.Contents = new(int)`), or check (`if b.Contents != nil { ... }`).

### Pointer receivers — deep dive

A method receiver is just a special first argument. Two forms:

```go
type Counter struct{ n int }

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
c.Inc()                 // Go takes &c implicitly, passes as the *Counter receiver
fmt.Println(c.Value())  // 1 — mutation stuck
```

Two rules:

- **Use a pointer receiver to mutate the receiver.** No mutation? Use a value receiver.
- **Use a pointer receiver to avoid copying a large struct.** For tiny types (`Counter` with one int), the copy is free; value receivers are fine. For big structs (10+ fields, embedded slices), pointer receivers save the copy.

**Method set rules (the formal version):**

- A type `T` has all its value-receiver methods.
- A type `*T` has all of `T`'s methods PLUS the pointer-receiver methods.
- When the receiver variable is **addressable** (a local variable; a struct field), Go automatically takes `&` for pointer-receiver method calls. In practice: just write `e.M()` regardless.

**Convention: consistency across a type's methods.** If ANY method on a type uses a pointer receiver, make ALL methods on that type use pointer receivers. Mixing is allowed but confuses readers ("does `c.Total()` see updates from `c.Add(...)`?").

> Note: lesson 09's `Expense` deliberately MIXES value and pointer receivers — `Format`/`IsHigh` value, `ApplyDiscount`/`Bump` pointer — as a teaching example. The contrast is the point. Real-world code would pick one and stick with it.

**Common mistake.** Mixing receiver types on the same type for no good reason:

```go
type Cart struct{ items []string }

func (c Cart) Count() int            // value receiver
func (c *Cart) Add(item string)      // pointer receiver
func (c Cart) Total() float64        // value receiver
```

This compiles. But the inconsistency is a smell. Fix: pick one. If you need any pointer receiver, make them all pointer receivers.

### When to use pointers

A common newcomer instinct is "I should use pointers everywhere for performance." That's wrong — Go's compiler is good at avoiding unnecessary copies, and pointers carry their own costs (the cognitive overhead of "is this nil?", an extra heap allocation in some cases).

**Use a pointer when:**

- **You need to mutate the value through a function or method call.** This is the dominant reason.
- **The value is genuinely large** (10+ fields, embedded slices/maps in the hundreds of bytes).
- **You need to express "no value here"** with `nil`. A `*Config` parameter where `nil` means "use defaults" is a common pattern.

**Use a value when:**

- **The type is small** (`int`, `float64`, `string`, a struct with 1-3 small fields). The copy is free; pointers add indirection for no gain.
- **The method is read-only.** Value receivers make this intent explicit.

A quick worked example from this lesson — `Expense` is small (~24 bytes) and `Format`/`IsHigh` are read-only, so those stay value receivers. `ApplyDiscount` and `Bump` need to mutate `e.Amount`, so they're pointer receivers. The decision was driven by mutation, not by size.

**Escape analysis (high-level intuition).** Go's compiler decides whether each variable lives on the **stack** (fast) or the **heap** (garbage-collected). Taking `&x` doesn't automatically put `x` on the heap — the compiler can keep it on the stack if it proves the pointer doesn't outlive the function.

You almost never need to think about this. The rule of thumb: **don't use pointers for performance unless a benchmark shows a real difference.** Lesson 22 covers profiling.

**Common mistake.** Using pointers as the default "just in case":

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

Pointers add cognitive overhead. For small read-only data, the value form is simpler.

A related anti-pattern: `*[]T` (pointer to slice) — slices are already reference-like (the header is a struct with a pointer + length + capacity). You almost never need `*[]T`. Use `[]T`.

### Common nil-pointer pitfalls

`nil` pointer dereferences are Go's most common runtime panic. Three flavours:

**1. Uninitialised pointer.**

```go
var p *int       // nil
fmt.Println(*p)  // PANIC
```

Fix: initialise, or check.

**2. Nil pointer field in a struct.**

```go
type Box struct{ Contents *int }
b := Box{}                  // b.Contents is nil
fmt.Println(*b.Contents)    // PANIC
```

Fix: don't ship structs with required pointer fields unless your constructor enforces them. Or check before dereferencing.

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

This actually WORKS in Go — `nil.Length()` doesn't panic as long as the method handles `l == nil` explicitly. But it's also a common foot-gun: forget the nil-guard and the next deref panics.

**Common mistake.** Returning a nil pointer alongside a nil error — caller can't tell apart "successfully found nothing" vs "failed":

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

## Exercise: warm-up

Two methods to implement in `exercises/warmup/counter/counter.go`:

- `(c *Counter) Inc()` — pointer receiver. One-liner: `c.n++`.
- `(c Counter) Value() int` — value receiver. One-liner: `return c.n`.

Then fill in the skeleton tests in `counter_test.go`:

- `TestCounter` — table-style: declare a `var c Counter`, call `Inc()` `incCount` times, assert `c.Value() == want`. At least 4 cases.
- `TestCountersDoNotInterfere` — two separate Counters, increment one a few times, increment the other once, assert they have independent counts.

```bash
cd lessons/09-pointers/exercises
go test ./warmup/counter/... -v
```

## Exercise: main

Four methods in `exercises/expense/expense.go`:

- `(e Expense) Format() string` — same as lesson 04/06/07/08. Use `fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)`.
- `(e Expense) IsHigh() bool` — `e.Amount > 50`. Boundary 50 is NOT high.
- `(e *Expense) ApplyDiscount(rate float64)` — `e.Amount *= (1 - rate)`. Pointer receiver — the change sticks on the caller.
- `(e *Expense) Bump(amount float64)` — `e.Amount += amount`. Pointer receiver.

Fill in the four skeleton tests:

- `TestExpenseFormat` — at least 3 cases (small/medium/large amounts).
- `TestExpenseIsHigh` — at least 4 cases incl. 49.99, 50 (false), 50.01 (true).
- `TestExpenseApplyDiscount` — at least 4 cases. The key assertion: after `e.ApplyDiscount(rate)`, `e.Amount` is the discounted value. **This is the lesson's payoff** — the mutation stuck.
- `TestExpenseBump` — at least 4 cases. Positive, negative (refund), zero, on-zero-start.

> A note on `make test-lesson LESSON=09-pointers`: lesson 09's exercise tests ship with empty cases, so the make output may show `ok` (vacuous pass) before you add cases. Run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/09-pointers/exercises
go test ./warmup/counter/... -v   # warm-up
go test ./expense/... -v           # main
go test ./...                       # both
```

Keep the daily habits:

```bash
gofmt -w ./...    # format
go vet ./...      # static analysis
```

CI enforces both.

## Going further

### Read

- [A Tour of Go — Pointers](https://go.dev/tour/moretypes/1) — the official tour's pointer page.
- [Effective Go — Pointers vs Values](https://go.dev/doc/effective_go#pointers_vs_values) — short canonical reference for the receiver rules.
- [Go FAQ — When should I use a pointer to an interface?](https://go.dev/doc/faq#pointer_to_interface) — useful prep for lesson 10 (interfaces). Short answer: almost never.

### Try

- **Write a value-receiver `ApplyDiscount` and watch it fail.** Rewrite `func (e *Expense) ApplyDiscount(rate float64)` as `func (e Expense) ApplyDiscount(rate float64)`. Re-run the tests — they fail because the mutation no longer sticks. This is the lesson's core payoff in a single experiment.
- **Add `(e *Expense) String() string` returning the same as Format.** Go will automatically use `String()` when you pass `*Expense` (or `Expense`) to `fmt.Println`. Bonus: read up on the `fmt.Stringer` interface — it's a teaser for lesson 10.
- **A method on a nil receiver.** Add `(c *Counter) IsZero() bool` returning `c == nil || c.n == 0`. Call it on a nil `*Counter` — observe that it works because of the explicit nil check. Then remove the nil check and watch it panic.
- **Pointer to slice.** Try `func (e *Expense) MoreCategories(cats ...string)` that appends to a `*[]string` field. Convince yourself it's needlessly complex compared to a value-receiver method that returns a new slice.
