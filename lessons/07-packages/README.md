# Lesson 07: Packages and modules

## Learning goals

- Split your code into packages — one package per directory, files share a `package X` declaration.
- Use Go's full module-rooted import paths to bring in your own subpackages.
- Apply the **formal** exported-vs-unexported rule: capitalised first letter → public, lower-case → private to the package. Same rule everywhere.
- Tour Go's daily-use tooling: `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`.

## Prerequisites

- Lessons 01-06. In particular, lesson 06's `Expense` + `Format`/`IsHigh` is what we're refactoring into a subpackage here.

## What's different about this lesson's layout

This is the first lesson where `exercises/` and `solutions/` are NOT flat. Each side contains a top-level `main.go` driver plus two subpackages:

```
lessons/07-packages/exercises/
├── main.go                    # package main — imports greet + expense
├── greet/
│   ├── greet.go               # package greet
│   └── greet_test.go
└── expense/
    ├── expense.go             # package expense
    └── expense_test.go
```

There's no `main_test.go` at the top level — `main.go` is a demo binary; verify it by running `go run ./lessons/07-packages/exercises`. Tests live in the subpackages.

## Concepts

### Splitting code into packages

Up to now we've put everything into the `exercises` (or `solutions`) package. Real Go programs split related code into focused **packages**, where each package is a single directory holding files that share a `package X` declaration.

Three rules:

1. **One package per directory.** Every `.go` file in a directory must declare the same package.
2. **Package name = directory name (usually).** A directory called `greet/` contains files that declare `package greet`.
3. **The `package` line is the very first non-comment line** of every `.go` file.

```go
// File: lessons/07-packages/exercises/greet/greet.go
package greet

func Greet(name string) string {
	return "Hello, " + name + "!"
}
```

```go
// File: lessons/07-packages/exercises/expense/expense.go
package expense

type Expense struct {
	Date     string
	Amount   float64
	Category string
}
```

To run the binary: `go run ./lessons/07-packages/exercises`.
To run all tests under the lesson: `go test ./lessons/07-packages/exercises/...` (the `...` means "this directory and all subdirectories").

**Common mistake.** Putting two different `package` declarations in the same directory:

```go
// File: greet/greet.go
package greet

// File: greet/hello.go
package hello   // wrong: must match the rest of the directory
```

Go complains: `found packages greet (greet.go) and hello (hello.go) in greet`. Fix: every `.go` file in a directory must share the same package declaration.

### Imports and module paths

Stdlib imports use short names:

```go
import "fmt"
import "strings"
```

Your own subpackages use the **full module-rooted path** (always — Go does not have relative imports):

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

Group imports in parentheses (`gofmt`'s convention; `goimports` enforces a blank line between stdlib and module-internal):

```go
import (
	"fmt"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
)
```

Use the imported package via its qualifier — the **last segment of the import path**:

```go
greet.Greet("Aki")
expense.Expense{Date: "..."}
expense.TotalsByCategory(es)
```

Why so verbose? Because `go.mod` at the repo root declares the module's import-path prefix:

```
module github.com/ristkari-dev/go-training

go 1.23
```

Anything imported under `github.com/ristkari-dev/go-training/...` is interpreted as a directory path relative to the repo root. Every package in the world has exactly one canonical import path, and Go's tooling never has to guess which one you meant.

```bash
$ go list ./lessons/07-packages/exercises/...
github.com/ristkari-dev/go-training/lessons/07-packages/exercises
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet
```

**Common mistake.** Trying relative imports:

```go
import "./expense"      // ERROR
import "../shared"      // ERROR
```

Go complains: `"./expense" is relative, but relative import paths are not supported in module mode`. Fix: use the full module-rooted path. Always.

### Exported vs unexported (formal)

Lesson 06 introduced this. Now we formalise: **a name is exported if its first letter is uppercase, unexported otherwise**. The rule is uniform across fields, methods, functions, types, constants, variables — Go has no other access modifiers.

```go
package billing

type Invoice struct {
	Number int       // exported — accessible from other packages
	Amount float64   // exported
	notes  string    // unexported — package-private
}

func (i Invoice) Display() string {     // exported method
	return fmt.Sprintf("...", i.notes)  // can read notes from within the package
}

func formatNotes(s string) string {     // unexported function — internal helper
	return strings.TrimSpace(s)
}
```

From another package:

```go
inv := billing.Invoice{Number: 1, Amount: 12.50}   // works
inv.notes = "..."                                   // ERROR: unexported field
```

Two API patterns to recognise:

```go
// Pattern A — fields exported, no constructor.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}
// Caller: e := expense.Expense{Date: "...", Amount: 4.50, Category: "coffee"}
```

```go
// Pattern B — fields unexported, constructor required.
type Expense struct {
	date, amount, category interface{}   // hidden
}
func NewExpense(date string, amount float64, cat string) (Expense, error) { /* ... */ }
// Caller: e, err := expense.NewExpense("...", 4.50, "coffee")
```

Pattern A is what we use this lesson — direct, simple, and our tests need to construct via struct literals. Pattern B is what production code often does for types with invariants — the constructor validates and the caller can't break the type's rules.

**Common mistake.** Lowercase = invisible from other packages, even though the field exists:

```go
// In package mypkg:
type Counter struct { count int }

// In another package:
c := mypkg.Counter{count: 5}
```

Go says: `unknown field count in struct literal of type mypkg.Counter` — even though `count` clearly exists. Fix: capitalise it (`Count`) or expose it through a method (`func (c Counter) Count() int { return c.count }`).

### Tooling tour

Five commands you'll run every day. We've met some already.

```bash
gofmt -l ./...              # list files needing reformat (CI fails if non-empty)
gofmt -w ./...              # rewrite in canonical format

go vet ./...                # static analysis — wrong Printf verbs, shadow vars, etc.

go doc fmt                  # docs for the fmt package
go doc fmt.Sprintf          # docs for one symbol
go doc -all fmt             # include unexported

go list ./...               # enumerate packages
go list -m                  # the module's path

go mod tidy                 # sync go.mod with imports actually used
```

The four-step rhythm before every commit:

```bash
gofmt -w ./...    # format
go vet ./...      # static check
go test ./...     # tests
git commit ...
```

Make it muscle memory. CI runs all four; running them locally first means you never push a broken build.

**Common mistake.** Skipping `gofmt` and seeing CI fail on indentation. Fix: wire your editor to format on save (`gofmt -w` on every save), or remember to run it before every commit.

A second one: editing imports manually and getting "unused import" or "imported and not used" errors. Install `goimports` (a sibling of `gofmt`):

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

Then `goimports -w .` adds missing imports and removes unused ones in one shot.

## Exercise: warm-up

In `exercises/greet/greet.go`:

- `func Greet(name string) string` — return `"Hello, <name>!"`. One-liner with `fmt.Sprintf`.

Then fill in the skeleton `TestGreet` in `exercises/greet/greet_test.go`. At least 4 cases: a normal name, the literal `"World"`, an empty name, and one with non-ASCII characters.

## Exercise: main

Three things in `exercises/expense/expense.go`:

- `func (e Expense) Format() string` — same as lesson 06's `Format`, byte-for-byte. Use `fmt.Sprintf("%s  €%-7.2f %s", ...)`.
- `func (e Expense) IsHigh() bool` — `e.Amount > 50`. The boundary value 50 is **not** high.
- `func TotalsByCategory(es []Expense) map[string]float64` — accumulate totals per category. Empty/nil input returns a non-nil empty map.

Fill in the three skeleton tests in `exercises/expense/expense_test.go`. Same case suggestions as lesson 06.

The top-level `main.go` is already wired — it imports both subpackages and runs them. Once the implementations are in, `go run ./lessons/07-packages/exercises` should produce:

```
Hello, Aki!
================
2026-05-15  €4.50    coffee
2026-05-15  €12.00   lunch
2026-05-15  €75.00   rent  (high)
2026-05-15  €9.99    coffee
----------------
totals: coffee=14.49 lunch=12.00 rent=75.00
```

> A note on `make test-lesson LESSON=07-packages`: lesson 07's exercise tests ship with empty `cases` slices, so the make output may show `ok` (vacuous pass) before you add cases. Don't be fooled — run `go test -v` and look for sub-tests; if there are none, you haven't added cases yet.

## How to run

```bash
cd lessons/07-packages/exercises
go test ./greet/... -v        # warm-up
go test ./expense/... -v      # main
go test ./...                  # both
go run .                       # the binary (panics until you implement the stubs)
```

Daily habits, now formalised:

```bash
gofmt -w ./...    # format
go vet ./...      # static analysis
```

Both run by CI. Add `goimports -w .` to your editor's save hook for the smoothest workflow.

## Going further

### Read

- [Effective Go — Names](https://go.dev/doc/effective_go#names) — package and identifier naming.
- [Go blog — Organizing Go code](https://go.dev/blog/organizing-go-code) — the canonical guide to package organisation.
- [Go module reference — Module paths](https://go.dev/ref/mod#module-path) — exhaustive reference for what makes a valid module path.

### Try

- **Add a third subpackage.** Create `exercises/greet/farewell/farewell.go` (yes, nested under greet — Go allows it) with a `Farewell(name string)` function. Import it from `main.go`. Note how the import path now reflects the deeper nesting.
- **Constructor pattern.** Convert `Expense` to Pattern B from the slides — make all fields unexported, write `NewExpense(date, amount, category)` as the only construction path. Update the test and `main.go` to use it. Observe what breaks (struct-literal construction is no longer valid from outside the `expense` package).
- **`go doc` on your code.** Run `go doc github.com/ristkari-dev/go-training/lessons/07-packages/solutions/expense.Expense` and read the output — it's your own doc comments rendered.
