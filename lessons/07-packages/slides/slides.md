<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">07</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations</div>
<h1>Packages and modules</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Split your code into packages, import them via Go's full module-rooted paths, formalise the exported-vs-unexported rule, and tour the daily-use Go tools (<code>gofmt</code>, <code>go vet</code>, <code>go doc</code>, <code>go list</code>, <code>go mod tidy</code>).</p>
</div>
</div>
</div>

---

## What we'll cover

- **Packages** — the `package X` declaration, the file → package → directory mapping, why split code at all.
- **Imports and module paths** — `import "..."`, the long `github.com/...` paths, what `go.mod` does.
- **Exported vs unexported (formal)** — capitalised → public, lower-cased → private. Same rule everywhere.
- **Tooling tour** — `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`. One slide each.

---

## Concept 1: Splitting code into packages

### Motivation

Up to now we've put everything into the `exercises` (or `solutions`) package. That works for tiny lessons but doesn't scale — real Go programs split related code into focused **packages**, where each package is a single directory holding files that share a `package X` declaration. Lesson 07 takes the `Expense` type from lesson 06 and promotes it into its own `expense` package; the warm-up does the same with a tiny `greet` package.

---

### The basics

Three rules to internalise:

1. **One package per directory.** Every `.go` file in a directory must declare the same package.
2. **Package name = directory name (usually).** A directory called `greet/` contains files declaring `package greet`. There are exceptions (`main` packages can live in any directory) but the convention is uniform.
3. **The `package` line is the very first non-comment line in the file.**

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

func (e Expense) Format() string { /* ... */ }
```

```go
// File: lessons/07-packages/exercises/main.go
package main

import (
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
)

func main() {
	fmt.Println(greet.Greet("Aki"))
	e := expense.Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}
	fmt.Println(e.Format())
}
```

Three files, three packages, three directories. The `main` package lives at the root of the lesson; `greet` and `expense` are subpackages one level down. The relationship between directory layout and package name is direct: the path from the module root maps onto how you import it.

---

### A worked example

Lesson 07's full `exercises/` shape:

```
lessons/07-packages/exercises/
├── main.go                    # package main
├── greet/
│   ├── greet.go               # package greet
│   └── greet_test.go          # package greet
└── expense/
    ├── expense.go             # package expense
    └── expense_test.go        # package expense
```

`main.go` imports `greet` and `expense`. The two subpackages don't import each other — they're independent. Tests for each subpackage live alongside their code.

To run the binary: `go run ./lessons/07-packages/exercises`.
To run all tests in the lesson: `go test ./lessons/07-packages/exercises/...` (the `...` means "this directory and all subdirectories").

---

### Common mistake

Putting two different `package` declarations in the same directory:

```go
// File: greet/greet.go
package greet

// File: greet/hello.go
package hello   // wrong: must match the rest of the directory
```

Go complains:

```
found packages greet (greet.go) and hello (hello.go) in greet
```

Fix: make every `.go` file in `greet/` start with `package greet`. One directory, one package — no exceptions.

---

### Recap

- One package per directory; every file in a directory shares its `package X` declaration.
- Package name conventionally matches the directory name.
- A `main` package is a binary entry point; non-main packages are libraries.

---

## Concept 2: Imports and module paths

### Motivation

To use code from another package, you `import` it. Go's import paths are module-rooted strings — they look long, but that's because they're unambiguous: every package in the world has exactly one canonical import path, and Go's tooling never has to guess which "utils" package you meant.

---

### The basics

Import the standard library by short name:

```go
import "fmt"
import "strings"
import "errors"
```

Import your own subpackages by full module path:

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

Group multiple imports in parentheses (`gofmt` enforces this):

```go
import (
	"fmt"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"
)
```

The blank line between stdlib imports and module-internal imports is convention — `gofmt` tolerates either form, but `goimports` enforces the split.

After import, refer to exported names through the package qualifier:

```go
greet.Greet("Aki")             // last segment of import path is the qualifier
expense.Expense{Date: "..."}
expense.TotalsByCategory(es)
```

The qualifier is the **last segment of the import path**, which (by convention) matches the package's `package X` declaration. So `import ".../greet"` gives you `greet.Greet` access.

---

### A worked example

What `go.mod` does — Go's module file at the repo root says:

```
module github.com/ristkari-dev/go-training

go 1.23
```

The `module` line declares the **import path prefix** for everything in this module. So when you write `import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet"`, Go takes everything after the module-path prefix (`lessons/07-packages/exercises/greet`) as a directory path relative to the repo root, and finds the matching package there.

That's why imports look long but unambiguous: the module path is a globally unique prefix (typically a Git URL), and the rest is the repo's directory layout.

```bash
$ go list -m
github.com/ristkari-dev/go-training

$ go list ./lessons/07-packages/exercises/...
github.com/ristkari-dev/go-training/lessons/07-packages/exercises
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense
github.com/ristkari-dev/go-training/lessons/07-packages/exercises/greet
```

---

### Common mistake

Confusing relative paths (which Go does NOT support) with module paths:

```go
// Wrong — Go doesn't have relative imports.
import "./expense"
import "../shared/util"
```

Go complains:

```
"./expense" is relative, but relative import paths are not supported in module mode
```

Fix: always use the full module-rooted path, even for code in your own repo:

```go
import "github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense"
```

This sometimes feels verbose for a one-line change, but it's the same rule everywhere — there are no special cases for "stuff in the same module."

---

### Recap

- Imports are full module-rooted strings — even for your own subpackages.
- The package qualifier is the last segment of the import path: `import ".../greet"` → `greet.Greet`.
- `go.mod`'s `module` line declares the import-path prefix for everything in the module.
- No relative imports. Ever.

---

## Concept 3: Exported vs unexported (formal)

### Motivation

Lesson 06 introduced this rule lightly. Now we formalise it: **a name is exported if its first letter is uppercase, unexported otherwise**. The rule is uniform — fields, methods, functions, types, constants, variables, even `init` functions all follow it. Exported = visible from other packages; unexported = visible only within the declaring package.

---

### The basics

```go
package billing

// Exported — callers from other packages can use it.
type Invoice struct {
	Number int       // exported
	Amount float64   // exported
	notes  string    // unexported — field is package-private
}

// Exported method on Invoice.
func (i Invoice) Display() string {
	return fmt.Sprintf("Invoice #%d: €%.2f", i.Number, i.Amount)
}

// Unexported function — internal helper.
func formatNotes(s string) string {
	return strings.TrimSpace(s)
}

// Exported — the constructor is the API; the field is hidden.
func NewInvoice(number int, amount float64, notes string) Invoice {
	return Invoice{
		Number: number,
		Amount: amount,
		notes:  formatNotes(notes),
	}
}
```

From another package:

```go
import "github.com/example/myapp/billing"

inv := billing.NewInvoice(1, 12.50, "  groceries  ")  // works
inv.Display()                                          // works — exported method
inv.Number = 2                                         // works — exported field
inv.notes = "..."                                      // ERROR: unexported field
```

The capitalisation rule applies to:

- **Functions** — `Greet` exported, `greet` unexported (collides with the package name here, but that's fine).
- **Types** — `Expense` exported, `expense` unexported.
- **Methods** — `Format()` exported, `format()` unexported.
- **Fields** — `Date string` exported, `date string` unexported.
- **Constants and variables** — `MaxRetries = 3` exported, `maxRetries = 3` unexported.

This is the **only** access-control mechanism in Go. There's no `public` / `private` / `protected` keyword.

---

### A worked example

Compare two API designs for the `expense` package:

```go
// API design A — fields exported, no constructor.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Caller does:
//   e := expense.Expense{Date: "...", Amount: 4.50, Category: "coffee"}
```

```go
// API design B — fields unexported, constructor required.
type Expense struct {
	date     string
	amount   float64
	category string
}

func NewExpense(date string, amount float64, category string) (Expense, error) {
	// ... validate, normalise, etc ...
	return Expense{date: date, amount: amount, category: category}, nil
}

// Caller does:
//   e, err := expense.NewExpense("...", 4.50, "coffee")
```

Design A is what the lesson uses — fields exported, no validation, anyone can construct. Simple and direct; the caller has no guarantee that the fields are well-formed.

Design B is what real-world Go often does for types with invariants — the package controls construction, validates inputs, and exposes the data through methods. The fields stay hidden; the package's API is what the caller sees.

Phase 1 sticks with design A because we need our tests to construct values via struct literals. Production code often goes with B.

---

### Common mistake

Trying to access an unexported field from another package and getting a confusing-looking error:

```go
// In package mypkg:
type Counter struct {
	count int       // unexported
}

// In another package:
c := mypkg.Counter{count: 5}
```

Go says:

```
unknown field count in struct literal of type mypkg.Counter
```

The error says "unknown field" — but the field exists; it's just invisible from outside `mypkg`. Capitalise it (`Count`), or expose it through a method (`func (c Counter) Count() int { return c.count }`).

---

### Recap

- Capitalised first letter → exported (visible from other packages).
- Lower-case first letter → unexported (package-private).
- Same rule for fields, methods, functions, types, constants, variables.
- Go has no other access modifiers.

---

## Concept 4: Tooling tour

### Motivation

Go ships with a small set of tools you'll run every day. We've already met `go run`, `go build`, `go test`, `gofmt`, and `go vet`. Lesson 07 rounds out the daily-use set with `go doc`, `go list`, and `go mod tidy`, and reinforces the formatters/linters as part of every commit.

---

### The basics

**`gofmt`** — Go's canonical formatter. There's exactly one correct style and `gofmt` enforces it.

```bash
gofmt -l lessons/07-packages/    # list files that would be reformatted
gofmt -w lessons/07-packages/    # rewrite them in place
```

CI runs `gofmt -l` as a check. If the output is non-empty, the build fails.

**`go vet`** — static analyser that catches subtle bugs `go build` won't:

```bash
go vet ./...
```

Catches: wrong format verbs in `Printf` (`%d` with a string), shadowed variables, unreachable code, unused locks, missing struct field tags, and more. Run it locally; CI does too.

**`go doc`** — read the documentation for a package or symbol from the command line:

```bash
go doc fmt                        # the entire fmt package
go doc fmt.Sprintf                # one function
go doc github.com/ristkari-dev/go-training/lessons/07-packages/exercises/expense.Expense
```

The doc comments you write on exported names are what `go doc` shows. `go doc -all` includes unexported.

**`go list`** — enumerate packages or modules:

```bash
go list ./...                              # all packages in this module
go list ./lessons/07-packages/...          # packages under one directory
go list -m                                 # the current module's path
```

Useful in scripts (e.g., the project's `make test` does `go list ./... | grep -v '/exercises$'`).

**`go mod tidy`** — synchronise `go.mod` and `go.sum` with the imports actually used:

```bash
go mod tidy
```

After adding or removing imports, run this. It removes unused module dependencies and adds missing ones. CI in larger projects often runs `go mod tidy && git diff --exit-code go.mod` to catch un-tidied PRs.

---

### A worked example

A workflow you'll do dozens of times a day:

```bash
# Edit your code...
$ vim lessons/07-packages/exercises/expense/expense.go

# Format it.
$ gofmt -w lessons/07-packages/exercises/expense/

# Check static analysis.
$ go vet ./lessons/07-packages/...

# Run the tests.
$ go test ./lessons/07-packages/...

# All green — commit.
$ git add ...
$ git commit -m "..."
```

Make these four steps muscle memory. CI runs them all; if you run them locally first, you'll never push a broken build.

---

### Common mistake

Skipping `gofmt` and seeing a CI failure on a one-character indentation difference. The fix is to run `gofmt -w .` before every commit, and ideally to wire your editor to format on save.

A second common one: editing imports manually and getting "unused import" or "imported and not used" errors. `goimports` (a sibling of `gofmt`) auto-adjusts the import block; install it with:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

then run `goimports -w .` to add missing imports and remove unused ones.

---

### Recap

- `gofmt` — canonical formatter. CI enforces it.
- `go vet` — static analyser. Run alongside `go test`.
- `go doc` — package/symbol docs from the command line.
- `go list` — enumerate packages or the module path.
- `go mod tidy` — sync `go.mod` with what's actually imported.

---

## Practice

### Warm-up

In `exercises/greet/`:

- `greet/greet.go` — implement `Greet(name string) string` returning `"Hello, <name>!"`.
- `greet/greet_test.go` — fill in the skeleton `TestGreet` (cases + assertion body).

```bash
cd lessons/07-packages/exercises
go test ./greet/... -v
```

---

### Main

In `exercises/expense/`:

- `expense/expense.go` — implement `Expense.Format()`, `Expense.IsHigh()`, and `TotalsByCategory(es []Expense)`. Same logic as lesson 06; same byte-output for `Format`.
- `expense/expense_test.go` — fill in three skeleton tests.

The top-level `main.go` is already wired up — it imports both subpackages and prints a small report. Once the subpackages work, `go run ./lessons/07-packages/exercises` produces the expected output (see the doc comment on `main.go`).

```bash
cd lessons/07-packages/exercises
go test ./...    # warm-up + main
go run .         # the binary
```

Note:
For live: do the import-path tour live — type one of the long `import "github.com/..."` paths on the projector, then split it: "module path / repo subdirectory / package name". The realisation that "the import path is just the directory layout from the module root" is the lesson's lightbulb. Demo `go list ./lessons/07-packages/...` to show the discovered packages, and `go doc` on `expense.Expense` to show the doc comments rendering.

---

## What we learned

- One package per directory; package name conventionally matches the directory.
- Imports are full module-rooted paths (`github.com/.../subdir/pkg`) — no relative imports.
- The package qualifier is the last segment of the import path.
- Capitalised first letter → exported; lower-case → unexported. Uniform rule across fields, methods, functions, types, constants, variables.
- Daily Go tools: `gofmt`, `go vet`, `go doc`, `go list`, `go mod tidy`. Run all of them.

---

## Up next

Lesson 08 — Phase 1 capstone. The expense tracker CLI: `add`, `list`, `summary` subcommands, JSON-backed persistence (via a provided `storage` package), `cmd/` convention, end-to-end CLI design.
