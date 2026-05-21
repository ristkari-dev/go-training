# Lesson 11: Errors

## Learning goals

- Declare **sentinel errors** with `var ErrXxx = errors.New(...)` for the "callers need to discriminate" case.
- **Wrap errors with `%w`** to add context without losing the underlying error.
- Use `errors.Is(err, sentinel)` to walk the wrap chain for a sentinel; `errors.As(err, &typedVar)` for a typed error.
- Define **custom error types** via struct + `Error() string` + (optional) `Unwrap() error`.
- Apply the error-design philosophy: bare wrap if the caller doesn't discriminate; sentinel if it discriminates by identity; custom type if it needs structured data.

## Prerequisites

- Lessons 01-10. Lesson 10 in particular — the `store.Store` interface and JSONStore are what this lesson upgrades.

## What's different about this lesson

This lesson **changes the behaviour** of JSONStore from lesson 10:

- **L10**: `JSONStore.Load()` silently returned `([], nil)` when the file didn't exist (friendly first-time use).
- **L11**: `JSONStore.Load()` returns a wrapped `store.ErrNotFound` instead. The friendliness moves to the caller — `cmd/expenses` has a new `loadOrEmpty` helper that uses `errors.Is(err, store.ErrNotFound)` to handle first-time use explicitly.

Why? Pedagogy: this is the perfect place to demonstrate `errors.Is` on a sentinel students already understand. And design: pushing the "is missing file an error?" decision out of JSONStore and into the caller makes the contract honest. JSONStore says "the file isn't there"; the caller decides what that means.

## Concepts

### Sentinel errors

A sentinel error is a package-level error variable callers can compare against to discriminate one kind of failure from another.

```go
var ErrNotFound = errors.New("not found")

func lookup(id string) (string, error) {
	if id == "" {
		return "", ErrNotFound
	}
	return "value-for-" + id, nil
}

err := lookup("")
if errors.Is(err, ErrNotFound) {
	// handled gracefully
}
```

Three things to internalise:

- **`errors.New("...")` returns a `error` value.** That value is unique — comparing the same variable against itself works.
- **Sentinels are package-level `var`s**, not constants. Convention: `ErrXxx` exported, `errXxx` package-private.
- **Use sentinels when the caller needs to discriminate.** If the caller just bubbles the error up, no sentinel is needed — a plain wrapped error is enough.

**Common mistake.** Comparing wrapped errors with `==`:

```go
// WRONG — fails if the error was wrapped at any point.
if err == ErrNotFound {
	// ...
}
```

`fmt.Errorf("...: %w", err)` produces an error that *wraps* ErrNotFound. `err == ErrNotFound` returns false because the wrapped value is a different concrete type. Use `errors.Is`:

```go
// RIGHT — walks the chain.
if errors.Is(err, ErrNotFound) {
	// ...
}
```

### Wrapping with `%w`

`fmt.Errorf("...: %w", err)` adds context without losing the underlying value. `errors.Is` and `errors.As` can still walk through.

```go
err := errors.New("not found")
wrapped := fmt.Errorf("loading %s: %w", path, err)
fmt.Println(wrapped)                  // loading /home/aki/.expenses.json: not found
fmt.Println(errors.Is(wrapped, err))  // true — inner error still findable
```

`%v` is the bypass version — formats as text but doesn't preserve the chain:

```go
// %v — discards the wrap chain.
wrapped := fmt.Errorf("loading %s: %v", path, err)
errors.Is(wrapped, err)  // false
```

Use `%w` whenever you want callers to be able to inspect the underlying error. Use `%v` only when you want to discard the inner error intentionally (rare; e.g. the inner contains secrets you can't leak).

This lesson's JSONStore wraps both error cases for context:

```go
if errors.Is(err, fs.ErrNotExist) {
	return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
}
if err != nil {
	return nil, fmt.Errorf("loading %s: %w", j.Path, err)
}
```

Either way the caller sees `loading /home/aki/.expenses.json: not found` (or `: permission denied`), AND can still `errors.Is(err, ErrNotFound)` to discriminate.

**Common mistake.** Wrapping with `%v` instead of `%w`:

```go
// WRONG — breaks the chain.
return fmt.Errorf("loading %s: %v", path, err)
```

The caller can no longer use `errors.Is` to detect specific underlying errors. Always use `%w`.

A second one: wrapping twice but with mixed verbs:

```go
e1 := fmt.Errorf("step 1: %w", err)
e2 := fmt.Errorf("step 2: %v", e1)   // %v drops the chain
// errors.Is(e2, err) is false
```

Every layer either extends the chain (`%w`) or terminates it (`%v`). Be consistent — typically all `%w` unless you have a reason.

### `errors.Is` and `errors.As`

Once you start wrapping, you walk the chain to find specific errors.

**`errors.Is`** — sentinel check; walks the chain:

```go
if errors.Is(err, ErrNotFound) {
	// ...
}
```

**`errors.As`** — typed extraction; walks the chain looking for an error of a specific type:

```go
type ParseError struct {
	Path string
	Line int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: parse error", e.Path, e.Line)
}

err := loadFile("config.json")

var pe *ParseError
if errors.As(err, &pe) {
	fmt.Printf("parse failed at line %d of %s\n", pe.Line, pe.Path)
}
```

Two rules:

- **`errors.Is` for sentinels** (values you compare against — `ErrNotFound`, `io.EOF`).
- **`errors.As` for types** (structs you want typed access to — `*ParseError`, `*os.PathError`).

This lesson's `cmd/expenses` uses both — `errors.As` in `main()` for ParseError; `errors.Is` in `loadOrEmpty` for ErrNotFound.

**Common mistake.** Wrong second argument to `errors.As`:

```go
// WRONG — must be a POINTER to the target type.
var pe *ParseError
if errors.As(err, pe) {   // missing &
	// ...
}
```

The right shape:

```go
var pe *ParseError
if errors.As(err, &pe) {   // & of the variable; library mutates via reflection
	// pe is now non-nil if ParseError was in the chain
}
```

A second one: forgetting to check the return value of `errors.As` before using the extracted variable:

```go
var pe *ParseError
errors.As(err, &pe)
if pe.Line > 0 { // panics if pe is nil because As returned false
```

Always guard: `if errors.As(err, &pe) { ... }`.

### Custom error types

When sentinels aren't enough — when the caller needs structured data (line number, path, status code) — define a custom error type.

```go
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse %s:%d: %v", e.Path, e.Line, e.Cause)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}
```

Two methods, two responsibilities:

- **`Error() string`** — formats for human output. Convention: lowercase, no trailing newline, no period.
- **`Unwrap() error`** — returns the underlying error. Lets `errors.Is`/`errors.As` walk through. Optional; only define if you wrap another error.

Use a custom type when:

- Callers need typed access to fields (line number, status code, retry-after).
- The error carries enough structured data that string formatting wouldn't capture it.

Otherwise: a sentinel + wrap is simpler.

**Error-design philosophy** — three questions to ask:

1. **Does the caller need to discriminate this error from other errors?**
   - No → just wrap with `fmt.Errorf("...: %w", err)`.
   - Yes → continue.

2. **Does the caller need structured data (fields), or just identity?**
   - Just identity → sentinel.
   - Structured data → custom type.

3. **Does the error wrap something else?**
   - Yes → `%w` (for `fmt.Errorf`) or `Unwrap()` method (for custom types).
   - No → no Unwrap needed.

That's the whole design space. Apply case-by-case.

**Common mistake.** The typed-nil gotcha — returning a nil pointer of your error type as `error`:

```go
func loadFile(path string) error {
	var pe *ParseError  // nil pointer
	// ... oops, never assigned ...
	return pe   // ← NOT NIL when returned as error
}

err := loadFile(...)
if err != nil {   // ← TRUE, even though pe was nil!
	// ...
}
```

`(*ParseError)(nil)` boxed into an `error` interface is non-nil because the interface has a type tag. Fix: return literal `nil`:

```go
func loadFile(path string) error {
	pe, err := tryLoad(path)
	if err != nil {
		return pe   // pe is non-nil here
	}
	return nil   // not (*ParseError)(nil)
}
```

`go vet` catches some forms of this; not all.

## Exercise: warm-up

In `exercises/warmup/parseage/`:

- `parseage.go` — implement `ParseAge(s string) (int, error)`. Call `strconv.Atoi`; on error, wrap with `fmt.Errorf("parseage: invalid age %q: %w", s, err)`.
- `parseage_test.go` — fill in two skeleton tests:
  - `TestParseAgeValid` — at least 3 happy cases (positive, zero, negative).
  - `TestParseAgeInvalidWraps` — at least 3 failing cases. Each asserts THREE things: error non-nil, message contains the input string, `errors.Is(err, strconv.ErrSyntax)` is true.

```bash
cd lessons/11-errors/exercises
go test ./warmup/parseage/... -v
```

## Exercise: main

In `exercises/`:

- `expense/expense.go` — **already complete** (verbatim L10 forward-port). Don't modify.
- `store/store.go` — implement:
  - `(e *ParseError) Error() string` — `fmt.Sprintf("store parse %s:%d: %v", e.Path, e.Line, e.Cause)`.
  - `(e *ParseError) Unwrap() error` — return `e.Cause`.
  - `(j *JSONStore) Load()` — see the doc-comment hint. Read file; if missing return wrapped `ErrNotFound`; if other I/O wrap with `%w`; if unmarshal fails return `&ParseError{Path, Line, Cause}`; otherwise return the slice.
  - `(j *JSONStore) Save()` — same as L10 (marshal with indent, MkdirAll, WriteFile).
- `store/store_test.go` — fill in:
  - `runStoreContract` — 2 contract checks (Save+Load roundtrips; Save replaces). The L10 "empty Load → empty + nil" check is gone — covered by the new TestJSONStoreLoadReturnsErrNotFound test below.
  - `TestJSONStore`, `TestMemoryStore` — temp dir + call `runStoreContract`.
  - `TestJSONStoreLoadReturnsErrNotFound` — non-existent path; assert `errors.Is(err, ErrNotFound)` and message contains the path.
  - `TestJSONStoreLoadReturnsParseError` — write `"not json"` to a temp file; assert `errors.As(err, &pe)`; check `pe.Path`, `pe.Line > 0`, `pe.Cause`.
  - `TestMemoryStoreLoadNeverReturnsErrNotFound` — sanity check.
- `cmd/expenses/main.go` — **already complete** (working code). Study how `main()` uses `errors.As` for ParseError and `loadOrEmpty` uses `errors.Is` for ErrNotFound. Don't modify.

```bash
cd lessons/11-errors/exercises
go test ./...
```

Then exercise the binary's two new error paths:

```bash
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP

# First-time use — ErrNotFound path; CLI handles it gracefully via loadOrEmpty
go run ./cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
go run ./cmd/expenses -store=json -file=$TMP list

# Malformed JSON — ParseError path; CLI gives a rich error message via errors.As
echo "not json" > $TMP
go run ./cmd/expenses -store=json -file=$TMP list  # exits 1 with a parse error

rm $TMP
```

Expected behaviour:
- First-time `add` prints `added: ...` (no error message — `loadOrEmpty` ate the ErrNotFound).
- `list` prints the one expense.
- After writing `"not json"`, `list` prints `error: could not parse /tmp/expenses-XXXXXX.json at line 1: invalid character 'o' in literal null (expecting 'u')` and exits 1.

> A note on `make test-lesson LESSON=11-errors`: lesson 11's exercise tests ship with empty cases / placeholder bodies, so the make output may show `ok` (vacuous pass) before you add cases. Run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/11-errors/exercises
go test ./warmup/parseage/... -v
go test ./store/... -v
go test ./...

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

- [Go blog — Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — the canonical introduction to `%w`, `errors.Is`, `errors.As`. Worth reading in full.
- [Go blog — Errors are values](https://go.dev/blog/errors-are-values) — Rob Pike on why errors as plain values beats exceptions.
- [Effective Go — Errors](https://go.dev/doc/effective_go#errors) — short canonical reference.
- [`errors` package docs](https://pkg.go.dev/errors) — `Is`, `As`, `Unwrap`, `Join`. Skim `Join` for awareness (multi-error handling — useful but not covered in this lesson).

### Try

- **An ErrInvalidExpense sentinel.** Add validation to `cmdAdd` — reject empty date, negative amount, empty category — by returning a wrapped sentinel `store.ErrInvalidExpense`. Update main() to recognise it and print a specific message.
- **A FieldError custom type.** Generalise the validation above: a `FieldError struct { Field, Value, Reason string }` for "this specific field is wrong." Wire it through cmdAdd's validation. Use `errors.As` in main() to print a nicely-formatted message.
- **Multiple errors from one operation.** Read `errors.Join` (Go 1.20+). Update `cmdAdd` to return ALL validation errors at once instead of just the first one. Update main() to print each on its own line.
- **A test for the typed-nil gotcha.** Write a function that intentionally returns a `*ParseError(nil)` as `error`. Write a test that catches the gotcha. Use `go vet` (which sometimes catches this) and verify it reports the issue.
