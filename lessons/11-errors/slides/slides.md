<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">11</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 2 — Idiomatic Go</div>
<h1>Errors</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Move from "<code>if err != nil { return err }</code>" to mature error handling: sentinel errors with <code>errors.New</code>, wrapping with <code>fmt.Errorf(&quot;...: %w&quot;, err)</code>, inspecting wrapped chains with <code>errors.Is</code> and <code>errors.As</code>, and custom error types via a struct with <code>Error()</code> and <code>Unwrap()</code> methods.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Sentinel errors** — `var ErrXxx = errors.New(...)` for "callers might want to discriminate."
- **Wrapping with `%w`** — `fmt.Errorf("...: %w", err)` adds context without losing the inner error.
- **`errors.Is` and `errors.As`** — inspect wrapped chains for a sentinel or a typed error.
- **Custom error types** — when a sentinel isn't enough; struct + `Error() string` + `Unwrap() error`.

---

## Concept 1: Sentinel errors

### Motivation

Phase 1 ended with `if err != nil { return err }` as the universal error idiom. That's still right 80% of the time. The remaining 20% — where the caller needs to DISCRIMINATE based on what went wrong — needs more machinery. Sentinel errors are the simplest tool: declare a named error value, return it when the specific condition occurs, callers compare against it.

---

### The basics

```go
package main

import (
	"errors"
	"fmt"
)

// Declare a sentinel — a package-level error variable.
// Convention: name starts with `Err` (exported) or `err` (unexported).
var ErrNotFound = errors.New("not found")

func lookup(id string) (string, error) {
	if id == "" {
		return "", ErrNotFound
	}
	return "value-for-" + id, nil
}

func main() {
	_, err := lookup("")
	if err == ErrNotFound {
		fmt.Println("handled gracefully")
	}
}
```

Three things to internalise:

- **`errors.New("...")` returns a value of type `error`.** That value is unique — comparing it with `==` against the same variable works.
- **Sentinels are package-level `var`s**, not constants (Go's `const` doesn't support `error` values). Convention: capital `ErrXxx` for exported, lowercase `errXxx` for package-private.
- **Use sentinels when the caller needs to discriminate.** If the caller just bubbles the error up, no sentinel is needed — a plain wrapped error is enough.

---

### A worked example

Lesson 11's `store.ErrNotFound` — JSONStore returns it when its backing file doesn't exist:

```go
// store/store.go
var ErrNotFound = errors.New("store: not found")

func (j *JSONStore) Load() ([]expense.Expense, error) {
	data, err := os.ReadFile(j.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
	}
	// ...
}
```

The caller — cmd/expenses — uses it to discriminate first-time use from a real error:

```go
func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}
```

Without `ErrNotFound`, the CLI either has to print a confusing error on first run, or silently swallow ALL load errors. The sentinel lets the caller make the right call.

---

### Common mistake

Comparing wrapped errors with `==` instead of `errors.Is`:

```go
// WRONG — fails if the error was wrapped at any point.
if err == ErrNotFound {
	// ...
}
```

`fmt.Errorf("loading %s: %w", path, ErrNotFound)` produces an error that *wraps* ErrNotFound — but `err == ErrNotFound` returns false because the wrapped value is a different concrete type. Use `errors.Is`:

```go
// RIGHT — walks the wrap chain looking for ErrNotFound.
if errors.Is(err, ErrNotFound) {
	// ...
}
```

`errors.Is` is the safe default. `==` works only for bare (unwrapped) sentinels — and once you start wrapping (which you should), it stops working.

---

### Recap

- `var ErrXxx = errors.New("...")` declares a sentinel.
- Use sentinels when callers need to discriminate based on the kind of error.
- Don't use `==` to compare — use `errors.Is` (which handles wrapped chains).
- If callers just bubble the error, a plain wrapped error is enough; no sentinel needed.

---

## Concept 2: Wrapping with `%w`

### Motivation

A bare error like `errors.New("not found")` tells you what; an error like `loading /home/aki/.expenses.json: not found` tells you what AND where. Wrapping adds context to an error without losing the underlying value. `errors.Is`/`errors.As` can still find the inner error.

---

### The basics

```go
import "fmt"

err := errors.New("not found")
wrapped := fmt.Errorf("loading %s: %w", path, err)
fmt.Println(wrapped)              // loading /home/aki/.expenses.json: not found
fmt.Println(errors.Is(wrapped, err))   // true — the inner error is still findable
```

The `%w` verb is special — only valid inside `fmt.Errorf` — it tells fmt "this argument is an error; preserve the wrap chain." The `%v` verb works too but DOESN'T preserve the chain:

```go
// %v — just formats as text; loses the wrap chain.
wrapped := fmt.Errorf("loading %s: %v", path, err)
fmt.Println(errors.Is(wrapped, err))   // false — wrap chain broken
```

Use `%w` whenever you want callers to be able to `errors.Is`/`errors.As` through the wrap. Use `%v` only when you genuinely want to discard the inner error (e.g. the inner error contains secrets you don't want to leak).

---

### A worked example

The lesson's JSONStore.Load wraps both error cases:

```go
func (j *JSONStore) Load() ([]expense.Expense, error) {
	data, err := os.ReadFile(j.Path)
	if errors.Is(err, fs.ErrNotExist) {
		// Wrap our sentinel with path context.
		return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
	}
	if err != nil {
		// Wrap the OS error with path context too.
		return nil, fmt.Errorf("loading %s: %w", j.Path, err)
	}
	// ...
}
```

Either way the caller sees a useful message:

```
loading /home/aki/.expenses.json: not found
loading /home/aki/.expenses.json: permission denied
```

And either way `errors.Is(err, ErrNotFound)` finds the sentinel when it's there.

---

### Common mistake

Wrapping with `%v` instead of `%w`:

```go
// WRONG — %v formats but doesn't preserve the chain.
return fmt.Errorf("loading %s: %v", path, err)
```

Now the caller can't tell apart "file missing" from "permission denied" — `errors.Is(err, ErrNotFound)` returns false even when the underlying error was ErrNotFound. Always use `%w` unless you have a specific reason not to.

A second one: wrapping the same error twice with different messages, accidentally losing context:

```go
err := errors.New("not found")
e1 := fmt.Errorf("step 1: %w", err)
e2 := fmt.Errorf("step 2: %v", e1)   // wrong: %v drops the chain
```

`errors.Is(e2, err)` is false. Fix: use `%w` in the second wrap too. The wrap chain is a stack — every layer either extends it (`%w`) or terminates it (`%v`).

---

### Recap

- `fmt.Errorf("...: %w", err)` wraps. `%v` formats only (loses the chain).
- Wrap to add context: the file path, the operation name, the row number.
- The chain is walked by `errors.Is` and `errors.As`.
- Use `%v` only when you intentionally want to discard the inner error.

---

## Concept 3: `errors.Is` and `errors.As`

### Motivation

Once you start wrapping, you need to walk the chain to find specific errors. `errors.Is` answers "is the sentinel X anywhere in the chain?". `errors.As` answers "is there a typed error of type T in the chain, and if so, give me a typed handle to it?".

---

### The basics

```go
import "errors"

var ErrNotFound = errors.New("not found")

func find(id string) error {
	return fmt.Errorf("find(%s): %w", id, ErrNotFound)
}

err := find("abc")

// errors.Is — sentinel check. Walks the chain.
if errors.Is(err, ErrNotFound) {
	fmt.Println("not found, that's OK")
}
```

`errors.As` is the typed counterpart — extracts a specific error type from the chain:

```go
type ParseError struct {
	Line int
	Path string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: parse error", e.Path, e.Line)
}

func loadFile(path string) error {
	return &ParseError{Line: 5, Path: path}
}

err := loadFile("config.json")

var pe *ParseError
if errors.As(err, &pe) {
	fmt.Printf("parse failed at line %d of %s\n", pe.Line, pe.Path)
}
```

The first argument to `errors.As` is the error chain; the second is a **pointer to a variable of the target type**. If a matching error is found, the variable is set; `As` returns true.

Two rules:

- **`errors.Is` for sentinels** (values you compare against — `ErrNotFound`, `io.EOF`).
- **`errors.As` for types** (structs you want typed access to — `*ParseError`, `*os.PathError`).

---

### A worked example

Lesson 11's cmd/expenses uses both:

```go
// In main(): branch on error type for a richer message.
if err := run(os.Args[1:]); err != nil {
	var pe *store.ParseError
	if errors.As(err, &pe) {
		fmt.Fprintf(os.Stderr, "error: could not parse %s at line %d: %v\n",
			pe.Path, pe.Line, pe.Cause)
	} else {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	os.Exit(1)
}

// In loadOrEmpty(): use errors.Is for sentinel detection.
func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}
```

`errors.As` gives main() access to ParseError's fields — `pe.Path`, `pe.Line`, `pe.Cause` — for a much better message than a bare `err.Error()`.

---

### Common mistake

Passing the wrong second argument to `errors.As`:

```go
// WRONG — second arg must be a POINTER to the target type.
var pe *ParseError
if errors.As(err, pe) {   // missing &
	// ...
}

// WRONG — second arg must be a *pointer-to-the-typed-error*, not the type itself.
if errors.As(err, &ParseError{}) {   // & of a value, not the variable
	// ...
}
```

Both compile but neither does what you want. The right shape is:

```go
var pe *ParseError
if errors.As(err, &pe) {
	// pe is now non-nil and points at the ParseError in the chain
}
```

The second arg is `**T` (pointer to a pointer-to-the-typed-error). The library uses reflection to extract and assign. The compiler doesn't catch this — Go vet does, sometimes.

A second one: forgetting that `errors.As` mutates its second argument:

```go
var pe *ParseError
errors.As(err, &pe)
// pe is now set OR still nil — check the return value of As before using pe.

if pe.Line > 0 { // ← will panic if pe is nil (As returned false)
```

Always guard with `if errors.As(err, &pe) { ... }`.

---

### Recap

- `errors.Is(err, sentinel)` — walks the chain looking for `sentinel`.
- `errors.As(err, &typedVar)` — walks the chain looking for an error of type `T`, sets `typedVar` if found.
- The second arg of `errors.As` is `**T`. Always.
- Always check the return value before using the extracted variable.

---

## Concept 4: Custom error types

### Motivation

Sentinels say "this thing happened" — same value every time. Sometimes you need to say "this thing happened AND here's some structured info about it" — a line number, a file path, an HTTP status code. That's a custom error type: a struct with an `Error() string` method.

---

### The basics

A custom error type is just a struct that implements `error`:

```go
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

// Error makes *ParseError satisfy the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse %s:%d: %v", e.Path, e.Line, e.Cause)
}

// Unwrap returns the underlying error so errors.Is/errors.As can chase
// through the chain.
func (e *ParseError) Unwrap() error {
	return e.Cause
}
```

Two methods, two responsibilities:

- **`Error() string`** — formats the error for human output. Convention: lowercase, no trailing newline, no period.
- **`Unwrap() error`** — returns the underlying error. Lets `errors.Is`/`errors.As` walk through. Optional; only define if you wrap another error.

Use a custom type when:

- Callers need typed access to fields (line number, status code, retry-after).
- The error carries enough structured data that a string formatting wouldn't capture it.

Otherwise: a sentinel + wrap is simpler.

---

### A worked example

Lesson 11's `store.ParseError`:

```go
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("store parse %s:%d: %v", e.Path, e.Line, e.Cause)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}
```

Caller in main():

```go
var pe *store.ParseError
if errors.As(err, &pe) {
	fmt.Fprintf(os.Stderr, "error: could not parse %s at line %d: %v\n",
		pe.Path, pe.Line, pe.Cause)
}
```

Without the custom type, main() would have to parse the error message string to extract path and line — fragile and ugly. With the type, the data is right there.

### Error-design philosophy

Three questions to ask when deciding the error shape:

1. **Does the caller need to discriminate this error from other errors?**
   - No → just wrap with `fmt.Errorf("...: %w", err)`. Move on.
   - Yes → continue.

2. **Does the caller need structured data (fields), or just identity (this-vs-that)?**
   - Just identity → sentinel `var ErrXxx = errors.New(...)`.
   - Structured data → custom type with the fields.

3. **Does the error wrap something else?**
   - Yes → add `Unwrap() error` method (for custom types) OR use `%w` (for `fmt.Errorf`).
   - No → no Unwrap needed.

That's the whole design space. Apply it case-by-case; don't try to design the whole error taxonomy upfront.

---

### Common mistake

Defining `Error()` on a value receiver when the struct has pointer-receiver methods elsewhere:

```go
type ParseError struct{ Line int }

func (e ParseError) Error() string { ... }   // value receiver
func (e *ParseError) Wrap(err error) { ... }   // pointer receiver

err := &ParseError{Line: 5}
fmt.Println(err.Error())          // works — fine
fmt.Println(errors.Is(err, ...))  // also works — but the As story gets weird
```

Mixing value and pointer receivers across the methods of the same type is a smell (lesson 09's consistency rule). For error types specifically: define Error on a pointer receiver. The error interface satisfaction works either way; pointer receiver is the convention.

A second mistake: returning a non-nil `*ParseError` typed value as `error`:

```go
func loadFile(path string) error {
	var pe *ParseError  // nil pointer
	// ... oops, never assigned ...
	return pe   // ← THIS IS NOT NIL when returned as error
}

err := loadFile(...)
if err != nil {
	// This branch runs! Even though pe was nil!
}
```

This is the famous "typed nil" gotcha. `(*ParseError)(nil)` boxed into an `error` interface is non-nil because the interface has a type tag even when the value is nil. Fix: return `nil` directly when there's no error:

```go
func loadFile(path string) error {
	pe, err := tryLoad(path)
	if err != nil {
		return pe   // pe is non-nil here
	}
	return nil    // not (*ParseError)(nil)
}
```

`go vet` catches some forms of this; not all.

---

### Recap

- Custom error type = struct + `Error() string` + (optional) `Unwrap() error`.
- Use when callers need typed access to fields.
- Define methods on pointer receivers (`*T`) by convention.
- Watch out for the typed-nil gotcha: return `nil` literally, not a nil pointer of your error type.

---

## Practice

### Warm-up

In `exercises/warmup/parseage/`:

- `parseage.go` — implement `ParseAge(s string) (int, error)`. Call `strconv.Atoi`; on error, wrap with `fmt.Errorf("parseage: invalid age %q: %w", s, err)` so the message has context AND the underlying `strconv.ErrSyntax` is still findable.
- `parseage_test.go` — fill in the skeleton:
  - `TestParseAgeValid` — at least 3 happy-path cases.
  - `TestParseAgeInvalidWraps` — at least 3 failing cases. Assert THREE things per case: error non-nil, message contains the input string, `errors.Is(err, strconv.ErrSyntax)` is true.

```bash
cd lessons/11-errors/exercises
go test ./warmup/parseage/... -v
```

---

### Main

In `exercises/`:

- `expense/expense.go` — **already complete** (verbatim L10 forward-port). Don't modify.
- `store/store.go` — implement four things:
  - `(e *ParseError) Error() string` — `fmt.Sprintf("store parse %s:%d: %v", e.Path, e.Line, e.Cause)`.
  - `(e *ParseError) Unwrap() error` — return `e.Cause`.
  - `(j *JSONStore) Load()` — read the file; if missing return wrapped `ErrNotFound`; if other I/O error wrap with `%w`; if unmarshal fails return `&ParseError{Path: j.Path, Line: lineFromJSONErr(data, err), Cause: err}`; otherwise return the parsed slice.
  - `(j *JSONStore) Save()` — same as L10 (marshal with indent, MkdirAll, WriteFile).
- `store/store_test.go` — fill in five test functions:
  - `runStoreContract` — 2 contract checks (Save then Load roundtrips; Save replaces). Drops L10's "empty Load → empty + nil" because JSONStore now returns ErrNotFound for empty.
  - `TestJSONStore` — temp dir + `runStoreContract`.
  - `TestMemoryStore` — `runStoreContract` on `NewMemoryStore()`.
  - `TestJSONStoreLoadReturnsErrNotFound` — assert `errors.Is(err, ErrNotFound)` for a missing file; message contains the path.
  - `TestJSONStoreLoadReturnsParseError` — write `"not json"` to a temp file; assert `errors.As(err, &pe)`; assert `pe.Path`/`pe.Line > 0`/`pe.Cause` are all set.
  - `TestMemoryStoreLoadNeverReturnsErrNotFound` — sanity check that MemoryStore.Load doesn't error.
- `cmd/expenses/main.go` — **already complete** (working code; study how it uses `errors.As` in main() and `errors.Is` in loadOrEmpty). Don't modify.

```bash
cd lessons/11-errors/exercises
go test ./...

# Smoke tests
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP

# First-time use (file missing — ErrNotFound path; CLI handles it gracefully)
go run ./cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
go run ./cmd/expenses -store=json -file=$TMP list

# Malformed JSON (ParseError path; CLI gives a rich message)
echo "not json" > $TMP
go run ./cmd/expenses -store=json -file=$TMP list   # should exit 1 with parse error

rm $TMP
```

Note:
For live: demo the first-time path first (no setup; just run `add` with a fresh -file path). The "no error message" outcome is the lesson's first payoff — sentinels enable graceful handling. Then `echo "not json" > $TMP` and run `list` to demo the ParseError path. The "error: could not parse /tmp/.../...: 1 at line 1: ..." message is the second payoff — custom types enable rich messages.

---

## What we learned

- Sentinel errors (`var ErrXxx = errors.New(...)`) — for "callers need to discriminate." Compare with `errors.Is`, never `==`.
- Wrapping with `%w` — `fmt.Errorf("...: %w", err)` adds context; preserves the chain.
- `errors.Is(err, sentinel)` walks the chain for a sentinel. `errors.As(err, &typedVar)` walks the chain for a typed error.
- Custom error types — struct + `Error()` + (optional) `Unwrap()`. Use when callers need typed access to fields.
- Watch the typed-nil gotcha: return literal `nil`, not a nil pointer of your error type.

---

## Up next

Lesson 12 — Generics (standalone). Type parameters, constraints (`cmp.Ordered`), when NOT to use generics. Standalone example: a small `slices`-style helper library (Filter/Map/Max). The tracker doesn't evolve this lesson — forcing generics into the tracker would be the premature-abstraction anti-pattern. New import: `cmp`.
