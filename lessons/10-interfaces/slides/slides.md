<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">10</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Interfaces</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Use Go's implicit interface satisfaction to write code that depends on behaviour rather than on concrete types; recognise small interfaces (<code>io.Reader</code>, <code>io.Writer</code>, <code>error</code>) as the idiomatic shape; use <code>any</code> and type assertions when you must; and follow the "accept interfaces, return structs" rule.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Implicit satisfaction** — no `implements` keyword; any type with the method set is automatically an instance.
- **Small interfaces** — `io.Reader` / `io.Writer` / `error` are 1-method interfaces; small interfaces compose, big ones don't.
- **`any` and type assertions** — `any` is `interface{}`; `v, ok := x.(T)` extracts a concrete type when you need it.
- **Interface design** — "accept interfaces, return structs"; empty interface as a code-smell signal.

---

## Concept 1: Implicit satisfaction

### Motivation

Go's most distinctive feature in one phrase: **no `implements` keyword**. A type satisfies an interface if it has the right methods. The compiler checks this at the call site, not at the type declaration. The consequence: you can write a function that takes an interface, then pass in any type that happens to have the right methods — including types you didn't write.

---

### The basics

```go
package main

import "fmt"

// Greeter is a one-method interface.
type Greeter interface {
	Greet(name string) string
}

// English has a Greet method, so it IS a Greeter — no declaration needed.
type English struct{}

func (English) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Finnish ALSO has a Greet method, so it's ALSO a Greeter.
// English and Finnish have no declared relationship.
type Finnish struct{}

func (Finnish) Greet(name string) string {
	return fmt.Sprintf("Hei, %s!", name)
}

func main() {
	// A function that accepts a Greeter — doesn't care about the concrete type.
	greet := func(g Greeter) {
		fmt.Println(g.Greet("World"))
	}

	greet(English{})  // Hello, World!
	greet(Finnish{})  // Hei, World!

	// A slice of Greeters can hold values of any concrete type that satisfies it.
	greeters := []Greeter{English{}, Finnish{}}
	for _, g := range greeters {
		fmt.Println(g.Greet("Aki"))
	}
}
```

Three things to internalise:

- **Satisfaction is automatic.** `English` doesn't say "implements Greeter" anywhere — Go figures it out from the method set.
- **Two unrelated types can satisfy the same interface.** `English` and `Finnish` are separate structs; no shared ancestry; both are `Greeter` values.
- **`fmt.Stringer` is the canonical example.** `interface{ String() string }`. Any type that has a `String() string` method automatically gets pretty-printed when passed to `fmt.Println`. This is how you customise formatting for your own types.

---

### A worked example

The `Greeter` slice in the warm-up is the simplest possible polymorphism demo. Real-world Go uses it constantly:

```go
// stdlib: io.Reader is an interface, *os.File implements it, so does
// strings.Reader, bytes.Buffer, http.Request.Body...
func readAll(r io.Reader) ([]byte, error) {
	// works for any of the above — no type-switch needed
}

// Caller picks the source:
readAll(os.Stdin)
readAll(strings.NewReader("hello"))
readAll(httpResponse.Body)
```

You'll see lesson 13 (encoding & I/O) lean on this idea heavily.

---

### Common mistake

Trying to *declare* that a type implements an interface (like Java's `implements` or Rust's `impl Trait for`):

```go
// WRONG — Go has no syntax for this.
type English struct{} implements Greeter

// Or trying to "register" the relationship explicitly somewhere — also wrong.
```

In Go, satisfaction is purely structural. The check happens when you try to use the type as the interface (assign to an interface variable, pass to a function that accepts the interface). If the method set matches, it works. If not, you get a compile error.

To force a *compile-time* check that your type satisfies an interface — a useful sanity check — use the assertion idiom:

```go
// Compile fails if English doesn't satisfy Greeter.
var _ Greeter = English{}
```

This declares a blank variable of type `Greeter`, initialised with an `English` value. It compiles only if `English` satisfies `Greeter`. Many stdlib packages use this trick at the top of files where the relationship matters.

---

### Recap

- Implicit satisfaction: any type with the method set IS the interface.
- No `implements` keyword. No declared relationship.
- `var _ I = T{}` is the compile-time check idiom.
- `fmt.Stringer` is the canonical example you've already been using indirectly.

---

## Concept 2: Small interfaces

### Motivation

Go's stdlib is built around tiny interfaces — `io.Reader` has one method, `io.Writer` has one method, `error` has one method, `fmt.Stringer` has one method. The pattern is deliberate: **small interfaces compose; big interfaces don't**. A 1-method interface is satisfied by almost any type that does roughly the right thing. A 20-method interface is satisfied by almost nothing.

---

### The basics

The canonical small interfaces:

```go
// io.Reader — read some bytes into p, return how many you read.
type Reader interface {
	Read(p []byte) (n int, err error)
}

// io.Writer — write some bytes from p, return how many you wrote.
type Writer interface {
	Write(p []byte) (n int, err error)
}

// error — describe yourself as a string.
type error interface {
	Error() string
}

// fmt.Stringer — print yourself.
type Stringer interface {
	String() string
}
```

Each is one method. Each is satisfied by dozens of stdlib types. Each composes with the others via embedding:

```go
// io.ReadWriter is the union — anything that satisfies BOTH Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}
```

The composition rule: an interface that embeds other interfaces requires all the embedded methods. *Files* satisfy `io.ReadWriter` because `*os.File` has both `Read` and `Write` methods.

---

### A worked example

Lesson 10's `Store` interface — two methods, the smallest API that covers "save and load":

```go
package store

import "github.com/ristkari-dev/go-training/lessons/10-interfaces/solutions/expense"

type Store interface {
	Load() ([]expense.Expense, error)
	Save(es []expense.Expense) error
}
```

Two implementations — `JSONStore` (file-backed) and `MemoryStore` (in-memory, mostly for tests). The cmd/expenses binary takes a `store.Store` interface variable and doesn't care which is behind it:

```go
func cmdAdd(s store.Store, args []string) error {
	es, err := s.Load()       // s could be *JSONStore or *MemoryStore
	if err != nil { return err }
	// ... mutate es ...
	return s.Save(es)
}
```

Same code paths through `cmdAdd`, `cmdList`, `cmdSummary` — the implementation is picked at startup from a flag. That's the architectural payoff.

---

### Common mistake

Designing a single big interface upfront, hoping it'll cover everything:

```go
// Smell — a "Service" with 20 methods.
type ExpenseService interface {
	Load() ([]Expense, error)
	Save([]Expense) error
	Add(Expense) error
	Update(int, Expense) error
	Delete(int) error
	Find(string) ([]Expense, error)
	Sort(string) error
	Filter(func(Expense) bool) []Expense
	Export(string) error
	Import(string) error
	// ... 10 more ...
}
```

Now every implementation has to implement all 20 — or use a "stub returns ErrNotImplemented" anti-pattern. The fix: split into multiple small interfaces, each focused on one job:

```go
type Loader  interface { Load() ([]Expense, error) }
type Saver   interface { Save([]Expense) error }
type Adder   interface { Add(Expense) error }
type Finder  interface { Find(string) ([]Expense, error) }
```

Functions take the interfaces they need; implementations satisfy the ones they support. Go stdlib does this everywhere (`io.Reader`, `io.Writer`, `io.Closer`, `io.ReadWriter`, ...). Interface segregation, except Go gets there naturally because writing tiny interfaces is the default.

---

### Recap

- Small interfaces compose; big interfaces don't.
- Canonical 1-method interfaces: `io.Reader`, `io.Writer`, `error`, `fmt.Stringer`.
- Embed interfaces to build larger ones from smaller pieces.
- Don't design a 20-method "Service" interface upfront.

---

## Concept 3: `any` and type assertions

### Motivation

Sometimes you need to handle "a value of unknown type." Go 1.18 added `any` as an alias for `interface{}` — "any value satisfies it because it has no methods to check." When you have an `any` and need to use its concrete type, you reach for a **type assertion**.

---

### The basics

`any` is just `interface{}`:

```go
// These are equivalent.
var x interface{}
var x any

// any can hold anything.
var x any = 42
x = "hello"
x = []int{1, 2, 3}
x = func() {}
```

You can't do much with an `any` directly — it has no methods. To use it, you have to assert what type it actually is:

```go
var x any = "hello"

// The "panicking" form — panics if x isn't a string.
s := x.(string)
fmt.Println(s)  // hello

// The "comma, ok" form — sets ok=false instead of panicking.
s, ok := x.(string)
if ok {
	fmt.Println(s)  // hello
}

// What if x is something else?
var y any = 42
s, ok := y.(string)
fmt.Println(s, ok)  // "", false  (zero value of string + ok=false)
```

**Use the comma-ok form by default.** The panic form is for cases where the type is guaranteed by surrounding logic (e.g. just after a type switch on `x.(type)`).

A `switch v := x.(type)` is the "I have a list of types I expect" pattern:

```go
func describe(x any) string {
	switch v := x.(type) {
	case string:
		return "string: " + v
	case int:
		return fmt.Sprintf("int: %d", v)
	case fmt.Stringer:
		return "stringer: " + v.String()
	default:
		return fmt.Sprintf("unknown: %T", x)
	}
}
```

---

### A worked example

A small helper that handles `fmt.Stringer` specially:

```go
func describe(x any) string {
	if s, ok := x.(fmt.Stringer); ok {
		return "stringer: " + s.String()
	}
	return fmt.Sprintf("value: %v", x)
}

describe(time.Now())           // stringer: 2026-05-20 14:30:00...  (Stringer)
describe(42)                    // value: 42
describe([]int{1, 2, 3})        // value: [1 2 3]
```

This is "interface check at runtime." Useful in library code (logging frameworks, serialisers) where you want to give some types special treatment. Avoid it in application code — usually you can rephrase the problem to use static types.

---

### Common mistake

Using `any` everywhere instead of writing focused interfaces:

```go
// Smell — `any` parameter type.
func handle(x any) {
	switch v := x.(type) {
	case Greeter:
		v.Greet("World")
	case fmt.Stringer:
		v.String()
	}
}
```

If the caller could only sensibly pass a `Greeter` or a `Stringer`, write two separate functions or define a new interface that covers both. `any` says "I don't know what I'm getting" — which usually means the code is doing two unrelated jobs.

A second one: assuming `any` is cheap. Each value stored in an `any` carries a runtime type tag (the type info) plus a pointer to the data. Compared to a typed parameter, an `any` is slower at runtime and harder to reason about. Use it sparingly.

---

### Recap

- `any` = `interface{}`. Use the alias since Go 1.18.
- Type assertions: `v, ok := x.(T)` is the safe form; `v := x.(T)` panics on mismatch.
- `switch v := x.(type)` is the multi-branch form.
- Don't use `any` to avoid writing types. Most "any" code can be a focused interface.

---

## Concept 4: Interface design — accept interfaces, return structs

### Motivation

Two complementary rules of thumb make Go interface design tractable:

1. **Accept interfaces** — functions that operate on values should take interface parameters when possible. The caller gets to pick the concrete type.
2. **Return structs** — constructors and factories should return concrete types. The caller can then decide whether to assign to an interface variable or use the type's full API.

Together they keep dependencies flowing in the right direction: producers of values commit to types; consumers of values accept interfaces.

---

### The basics

```go
// Constructor — returns the concrete type *JSONStore.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

// Function that uses the value — accepts the Store INTERFACE.
func process(s Store) error {
	es, err := s.Load()
	if err != nil { return err }
	// ...
	return s.Save(es)
}
```

The caller:

```go
js := NewJSONStore("/tmp/x.json")   // js is *JSONStore (concrete)
process(js)                          // implicitly converted to Store at the call site
```

Two benefits:

- **The caller gets the full API on the concrete type** (e.g. `js.Path` is a JSONStore-specific field that wouldn't be visible through `Store`).
- **`process` works with any future implementation** — `MemoryStore`, `RedisStore`, `MockStore` in tests — without changing.

---

### A worked example

Lesson 10's structure follows this exactly:

```go
// In store/store.go — constructors return concrete *JSONStore and *MemoryStore.
func NewJSONStore(path string) *JSONStore { ... }
func NewMemoryStore() *MemoryStore { ... }

// In cmd/expenses/main.go — buildStore returns the interface.
func buildStore(kind, path string) (store.Store, error) {
	switch kind {
	case "mem":   return store.NewMemoryStore(), nil
	case "json":  return store.NewJSONStore(path), nil
	}
	return nil, fmt.Errorf("unsupported")
}

// cmdAdd/cmdList/cmdSummary all take store.Store — they don't care which.
func cmdAdd(s store.Store, args []string) error { ... }
```

`buildStore`'s declared return type is `store.Store` *intentionally* — at that point we want the interface, because the caller (`cmdAdd` etc) operates through the interface. Compare with `NewJSONStore` which returns `*JSONStore` — anyone who wants JSONStore-specific behaviour gets it.

---

### Common mistake

Taking `*Interface` as a parameter (pointer to interface):

```go
// WRONG — almost always.
func process(s *Store) error {
	return (*s).Save(...)
}
```

Interfaces are already reference-like internally (an interface value is a pair of (type-tag, data-pointer)). Adding `*` on top is wasted indirection. The correct shape:

```go
func process(s Store) error {
	return s.Save(...)
}
```

If you genuinely want to mutate which value the interface points at (rare), use a named field on a struct rather than a `*Interface` parameter. The Go FAQ has a specific entry for this question.

A second common mistake: returning an interface from a constructor:

```go
// SMELL — returning the interface from a constructor.
func NewJSONStore(path string) Store {
	return &JSONStore{Path: path}
}
```

The caller has lost access to JSONStore-specific behaviour (e.g. the `Path` field). The caller will sometimes wrap with a type assertion to get it back, which is ugly. Return the concrete type; let the caller decide to upcast to the interface.

---

### Recap

- Accept interfaces — your function works with any current and future impl.
- Return structs — the caller keeps full access to type-specific API.
- Don't take `*Interface` parameters — interfaces are already reference-like.
- Don't return interfaces from constructors — return the concrete type.

---

## Practice

### Warm-up

In `exercises/warmup/greet/`:

- `greet.go` — implement `English.Greet(name)` returning `"Hello, <name>!"` and `Finnish.Greet(name)` returning `"Hei, <name>!"`. Both are one-line `fmt.Sprintf`.
- `greet_test.go` — fill in the skeleton: a `[]Greeter` table with both English and Finnish entries, asserting on output strings; plus `TestGreeterSlice` showing `[]Greeter{English{}, Finnish{}}` works as a polymorphic collection.

```bash
cd lessons/10-interfaces/exercises
go test ./warmup/greet/... -v
```

---

### Main

In `exercises/`:

- `expense/expense.go` — already complete (verbatim L09 forward-port). Don't modify.
- `store/store.go` — implement four methods:
  - `(j *JSONStore) Load() ([]expense.Expense, error)` — read the file; if missing return `([]expense.Expense{}, nil)`.
  - `(j *JSONStore) Save(es []expense.Expense) error` — pretty-print JSON; create parent dir; write file.
  - `(m *MemoryStore) Load() ([]expense.Expense, error)` — return a defensive copy of `m.items`.
  - `(m *MemoryStore) Save(es []expense.Expense) error` — replace `m.items` with a defensive copy of `es`.
- `store/store_test.go` — fill in `runStoreContract` to verify the three behaviours (empty Load → non-nil empty + nil err; Save then Load roundtrips; Save replaces). Then wire `TestJSONStore` and `TestMemoryStore` to call `runStoreContract` with the appropriate store.
- `cmd/expenses/main.go` — already complete (working code; study how it depends on `store.Store`). Don't modify.

Once the store implementations are in, the binary works end-to-end:

```bash
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP
go run ./cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
go run ./cmd/expenses -store=json -file=$TMP list
go run ./cmd/expenses -store=json -file=$TMP summary
rm $TMP
```

> Note on `-store=mem`: `MemoryStore` doesn't persist across invocations. `add` then `list` in separate process runs won't show anything — by design. MemoryStore is mainly useful for tests.

```bash
cd lessons/10-interfaces/exercises
go test ./...
```

Note:
For live: open `cmd/expenses/main.go` on the projector and trace one subcommand path — show how the same `cmdAdd` body works for both store flavours via the interface. The "interface variable holds either implementation" moment lands here. Then run the binary with `-store=mem` and `-store=json` back to back to make the polymorphism concrete.

---

## What we learned

- Implicit satisfaction: any type with the method set IS the interface. No `implements` keyword.
- Small interfaces compose; big interfaces don't. `io.Reader`/`io.Writer`/`error` are the templates.
- `any` is `interface{}`. Type assertions (`v, ok := x.(T)`) extract concrete types; use the comma-ok form by default.
- Accept interfaces, return structs. Functions accept the smallest interface they need; constructors return the concrete type.
- Don't take `*Interface` parameters; don't return interfaces from constructors.

---

## Up next

Lesson 11 — Errors. Sentinel errors (`var ErrNotFound = errors.New(...)`), wrapping with `%w`, `errors.Is`/`errors.As`, custom error types. The tracker's storage gets richer error context (path + line number for malformed JSON).
