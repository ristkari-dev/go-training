# Lesson 10: Interfaces

## Learning goals

- Understand **implicit interface satisfaction** — a type with the right method set IS the interface, no declaration needed.
- Recognise small interfaces (`io.Reader`, `io.Writer`, `error`, `fmt.Stringer`) and why they're the idiomatic shape in Go.
- Use `any` and type assertions (`v, ok := x.(T)`) when you genuinely need to handle "a value of unknown type."
- Apply the rule of thumb: **accept interfaces, return structs**.

## Prerequisites

- Lessons 01-09. Lesson 09 in particular — pointer receivers come up here when constructors return `*JSONStore` vs `Store` for the interface variable.

## What's different about this lesson

Lesson 10 adds a `cmd/expenses` binary back (lesson 09 had no binary). The CLI is a refactor of L08's expense tracker, now using a `store.Store` interface and a new `-store=mem|json` flag. The lesson's pedagogical payoff: **the same `cmdAdd` / `cmdList` / `cmdSummary` code works for either store** — that's interface polymorphism in action.

> A note on `-store=mem`: the in-memory store loses its data when the process exits. `add` then `list` in separate process runs won't show anything for `-store=mem`. That's by design — MemoryStore is mainly useful for tests. For real persistence use `-store=json` (the default).

## Concepts

### Implicit satisfaction

Go's most distinctive feature in one phrase: **no `implements` keyword**. A type satisfies an interface if it has the right method set. The check happens at the call site, not at the type declaration.

```go
type Greeter interface {
	Greet(name string) string
}

// English has a Greet method, so it IS a Greeter — no declaration needed.
type English struct{}

func (English) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Finnish ALSO has a Greet method. English and Finnish have no declared relationship.
type Finnish struct{}

func (Finnish) Greet(name string) string {
	return fmt.Sprintf("Hei, %s!", name)
}

// A function that accepts a Greeter — doesn't care about the concrete type.
greet := func(g Greeter) {
	fmt.Println(g.Greet("World"))
}

greet(English{})   // Hello, World!
greet(Finnish{})   // Hei, World!

// A slice of Greeters holds either.
greeters := []Greeter{English{}, Finnish{}}
for _, g := range greeters {
	fmt.Println(g.Greet("Aki"))
}
```

Three things to internalise:

- **Satisfaction is automatic.** `English` doesn't say "implements Greeter" anywhere; Go figures it out from the method set.
- **Two unrelated types can satisfy the same interface.** No shared ancestry needed.
- **`fmt.Stringer` is the canonical example.** Any type with `String() string` automatically pretty-prints via `fmt.Println` — that's how you customise formatting for your own types.

**Compile-time check idiom.** When you want the compiler to confirm a type satisfies an interface (defensive code; documentation):

```go
var _ Greeter = English{}
```

This compiles only if `English` satisfies `Greeter`. Many stdlib packages have these at the top of files where the relationship matters.

**Common mistake.** Trying to *declare* the relationship explicitly:

```go
// WRONG — Go has no syntax for this.
type English struct{} implements Greeter
```

In Go, satisfaction is purely structural. The check is automatic. Use the `var _ I = T{}` idiom if you want a compile-time guarantee.

### Small interfaces (`io.Reader`/`io.Writer`/`error`)

Go's stdlib is built around tiny interfaces. The pattern is deliberate: **small interfaces compose; big interfaces don't**.

The canonical 1-method interfaces:

```go
// io.Reader — read into p, return how many bytes you read.
type Reader interface {
	Read(p []byte) (n int, err error)
}

// io.Writer — write bytes from p, return how many you wrote.
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
// io.ReadWriter requires both Read and Write.
type ReadWriter interface {
	Reader
	Writer
}
```

This lesson's `Store` interface is two methods — still small:

```go
type Store interface {
	Load() ([]expense.Expense, error)
	Save(es []expense.Expense) error
}
```

Two implementations — `JSONStore` (file-backed) and `MemoryStore` (in-memory, mostly for tests). The cmd/expenses binary takes a `store.Store` interface variable; it doesn't care which is behind it.

**Common mistake.** Designing a single big interface upfront:

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

Now every implementation has to implement all 20 — or use a "stub returns ErrNotImplemented" anti-pattern. Fix: split into multiple small interfaces, each focused on one job. Functions take the interfaces they need; implementations satisfy the ones they support.

### `any` and type assertions

`any` is just `interface{}` (alias added in Go 1.18). When you have an `any` and need to use its concrete type, you reach for a **type assertion**.

```go
var x any = "hello"

// The "panicking" form — panics if x isn't a string.
s := x.(string)
fmt.Println(s)   // hello

// The "comma, ok" form — safe; sets ok=false instead of panicking.
s, ok := x.(string)
if ok {
	fmt.Println(s)
}

// What if x is something else?
var y any = 42
s, ok = y.(string)
fmt.Println(s, ok)   // "", false
```

**Use the comma-ok form by default.** The panic form is for cases where the type is guaranteed by surrounding logic.

A type switch is the multi-branch form:

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

**Common mistake.** Using `any` everywhere instead of writing focused interfaces:

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

If the caller could only sensibly pass a Greeter or a Stringer, write two functions or define a new interface that covers both. `any` says "I don't know what I'm getting" — which usually means the code is doing two unrelated jobs.

A second one: `any` isn't free at runtime. Each value stored in an `any` carries a runtime type tag plus a pointer. Compared to a typed parameter, an `any` is slower and harder to reason about. Use it sparingly.

### Interface design — accept interfaces, return structs

Two complementary rules of thumb:

1. **Accept interfaces** — functions that operate on values should take interface parameters when possible. The caller gets to pick the concrete type.
2. **Return structs** — constructors and factories should return concrete types. The caller can then decide whether to assign to an interface variable or use the type's full API.

```go
// Constructor — returns the concrete type *JSONStore.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

// Function that uses the value — accepts the Store INTERFACE.
func process(s Store) error {
	es, err := s.Load()
	// ...
	return s.Save(es)
}
```

Caller:

```go
js := NewJSONStore("/tmp/x.json")   // *JSONStore (concrete)
process(js)                          // implicitly converted to Store
```

Two benefits:

- **Caller gets the full API on the concrete type** (e.g. `js.Path` field).
- **`process` works with any future implementation** (`MemoryStore`, `MockStore` for tests, etc.) — no change needed.

This lesson's `buildStore` function in `cmd/expenses/main.go` *intentionally* returns the interface (`store.Store`), because at that point the caller wants polymorphism — `cmdAdd` etc. only see the interface methods.

**Common mistake.** Taking `*Interface` as a parameter:

```go
// WRONG — almost always.
func process(s *Store) error {
	return (*s).Save(...)
}
```

Interfaces are already reference-like internally. Adding `*` on top is wasted indirection. Use the plain interface type:

```go
func process(s Store) error {
	return s.Save(...)
}
```

A related mistake: returning an interface from a constructor. The caller loses access to type-specific behaviour. Return the concrete type; let the caller decide.

## Exercise: warm-up

In `exercises/warmup/greet/`:

- `greet.go` — implement `English.Greet(name)` returning `"Hello, <name>!"` and `Finnish.Greet(name)` returning `"Hei, <name>!"`. Both are one-line `fmt.Sprintf`.
- `greet_test.go` — fill in the skeleton:
  - `TestGreeters` — a table with both English and Finnish entries, asserting exact greeting strings (at least 4 cases including the empty-name case).
  - `TestGreeterSlice` — declare `[]Greeter{English{}, Finnish{}}`, loop through asserting each `Greet("World")` returns a non-empty string.

```bash
cd lessons/10-interfaces/exercises
go test ./warmup/greet/... -v
```

## Exercise: main

In `exercises/`:

- `expense/expense.go` — **already complete** (verbatim L09 forward-port). Don't modify.
- `store/store.go` — implement four methods:
  - `(j *JSONStore) Load() ([]expense.Expense, error)` — read the file at `j.Path`. If `errors.Is(err, fs.ErrNotExist)`, return `([]expense.Expense{}, nil)`. Otherwise `json.Unmarshal` into a slice and return it.
  - `(j *JSONStore) Save(es []expense.Expense) error` — `json.MarshalIndent(es, "", "  ")`; create the parent directory with `os.MkdirAll`; write with `os.WriteFile`.
  - `(m *MemoryStore) Load() ([]expense.Expense, error)` — return a defensive copy of `m.items`: `append([]expense.Expense{}, m.items...)`.
  - `(m *MemoryStore) Save(es []expense.Expense) error` — replace `m.items` with a defensive copy of `es`. Return nil.
- `store/store_test.go` — fill in `runStoreContract(t, s)` to verify three behaviours:
  1. `Load()` on an empty store returns a non-nil empty slice and nil error.
  2. After `Save([e1, e2, e3])`, `Load()` returns those three in order (use `reflect.DeepEqual`).
  3. After a second `Save([e4])`, `Load()` returns just `[e4]` — Save REPLACES.

  Then wire `TestJSONStore` (using `t.TempDir()` for the file path) and `TestMemoryStore` to call `runStoreContract`. **The point of writing the contract once and feeding both impls to it: that's how you test interface implementations idiomatically in Go.**
- `cmd/expenses/main.go` — **already complete** (working code; don't modify). Study how it depends on `store.Store` — the same `cmdAdd`/`cmdList`/`cmdSummary` bodies work for both JSONStore and MemoryStore.

Once the store implementations are in:

```bash
cd lessons/10-interfaces/exercises

# Tests
go test ./...

# Binary with JSON store
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP
go run ./cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
go run ./cmd/expenses -store=json -file=$TMP add 2026-05-20 12.00 lunch
go run ./cmd/expenses -store=json -file=$TMP list
go run ./cmd/expenses -store=json -file=$TMP summary
rm $TMP

# Binary with memory store (single-invocation only)
go run ./cmd/expenses -store=mem add 2026-05-20 4.50 coffee
```

> A note on `make test-lesson LESSON=10-interfaces`: lesson 10's exercise tests ship with empty cases, so the make output may show `ok` (vacuous pass) before you add cases. Run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/10-interfaces/exercises
go test ./warmup/greet/... -v
go test ./store/... -v
go test ./...

# Run the binary
go run ./cmd/expenses [-store=mem|json] [-file=path] <subcommand> [args]
```

Daily habits:

```bash
gofmt -w ./...
go vet ./...
```

CI enforces both.

## Going further

### Read

- [Effective Go — Interfaces](https://go.dev/doc/effective_go#interfaces) — short canonical reference. Especially the "Generality" section on small interfaces.
- [Go FAQ — When should I use a pointer to an interface?](https://go.dev/doc/faq#pointer_to_interface) — useful short read; the answer is almost never.
- [Go blog — Errors are values](https://go.dev/blog/errors-are-values) — lesson 11 preview; argues that `error` being an interface is what enables wrapping.
- [`io` package docs](https://pkg.go.dev/io) — the canonical 1-method interfaces (`Reader`, `Writer`, `Closer`) and their composed forms (`ReadCloser`, `ReadWriteCloser`, etc.). Read the package docs in full once.

### Try

- **A third Store implementation.** Add `type AppendOnlyStore struct { ... }` that wraps a `JSONStore` but disallows `Save` if the new slice is shorter than the existing one (i.e. you can only add, never delete). Plug it into `cmd/expenses` by adding `-store=append`.
- **A `Reader`-accepting CSV importer.** Write `ImportCSV(r io.Reader) ([]expense.Expense, error)` that parses CSV rows (date,amount,category) into expenses. Test it with `strings.NewReader("2026-05-20,4.50,coffee\n...")` — no file I/O needed thanks to `io.Reader`.
- **A `fmt.Stringer` method on Expense.** Add `func (e Expense) String() string` returning the same as `Format()`. Pass an Expense to `fmt.Println(e)` directly — it'll auto-pretty-print. Then add the compile-time check `var _ fmt.Stringer = Expense{}` at the top of expense.go.
- **A type switch on Store.** Write a debug helper `func describe(s Store) string` that uses a `switch v := s.(type)` to print which concrete type is behind the interface (`*JSONStore` with its `.Path` field; `*MemoryStore` with its item count). Useful for logging.
