# Lesson 08: Phase 1 capstone — expense tracker CLI

## Learning goals

- Apply the `cmd/<binaryname>/main.go` convention to organise binaries within a module.
- Compose focused library packages (`expense`, `summary`, `storage`) under an orchestration `main.go`.
- Write end-to-end integration tests for a CLI using `os/exec` + `t.TempDir()` against an isolated temp JSON file.
- Treat a provided package (`storage`) as a black box — use its API without touching its internals.
- Bring together everything from Phase 1: structs, methods, packages, modules, errors, tests, idiomatic Go.

## Prerequisites

- Lessons 01-07. Lesson 07 in particular — the subpackage layout escalates here.

## What's different about this lesson

Lesson 08 is the biggest lesson of Phase 1 (24 files vs the usual 12-14). The shape:

```
lessons/08-capstone/exercises/
├── warmup/
│   ├── clock/                 (package clock — Now() helper)
│   │   ├── clock.go
│   │   └── clock_test.go
│   └── cmd/
│       └── timestamp/         (package main — tiny demo binary)
│           └── main.go
├── expense/                   (package expense — Expense + methods)
│   ├── expense.go
│   └── expense_test.go
├── summary/                   (package summary — TotalsByCategory + BiggestCategory + Bar)
│   ├── summary.go
│   └── summary_test.go
├── storage/                   (package storage — PROVIDED, treat as black box)
│   └── storage.go
└── cmd/
    └── expenses/              (package main — the capstone binary)
        ├── main.go
        └── main_test.go       (integration tests)
```

The `storage` package is fully implemented and uses `encoding/json` internally — don't modify it. The rest you build.

## Concepts

### The `cmd/` convention

Real-world Go modules with binaries put each binary in its own subdirectory under `cmd/`. The directory name becomes the binary's name. A single module can host multiple binaries this way without polluting the root.

```
yourmodule/
├── go.mod
├── expense/                   library package
├── summary/                   library package
└── cmd/
    ├── expenses/main.go       package main — "expenses" binary
    └── another-tool/main.go   package main — second binary
```

Three things to internalise:

- **`cmd/` is a convention, not a language feature.** Go tooling treats `cmd/foo/main.go` identically to `foo/main.go`. The convention is purely for discoverability: every reader of the module knows "binaries live under `cmd/`."
- **Build/run by directory:**

  ```bash
  go build -o ./bin/expenses ./cmd/expenses
  go run ./cmd/expenses summary
  go install ./cmd/expenses    # installs to $GOPATH/bin (or $GOBIN)
  ```

- **The binary name is the directory name.** `go build ./cmd/expenses` produces a binary called `expenses`.

**Common mistake.** Forgetting that `cmd/expenses/` is a *directory*, not a file:

```bash
go run cmd/expenses.go        # wrong — no such file
```

Or:

```go
package main                   // in a file named cmd/expenses.go at the module root
```

Always: `cmd/<binaryname>/main.go`. The `main.go` lives **inside** the per-binary directory.

### Composing multiple packages into a CLI

A real CLI's `main.go` is *orchestration code* — it parses CLI arguments and delegates to focused subpackages. The "deps go one way" rule keeps the architecture sane.

Lesson 08's dependency graph:

```
cmd/expenses/main.go  ─┬─→  expense    (Expense + Format + IsHigh)
                       ├─→  summary    (TotalsByCategory + BiggestCategory + Bar)
                       └─→  storage    (LoadExpenses + SaveExpenses)
                              └─→  expense   (storage knows about Expense too)
```

Three rules:

- **`main.go` is the only `package main`.** Everything else is a library — testable in isolation, reusable across binaries.
- **Dependencies flow one direction.** `cmd/expenses` depends on `expense`/`summary`/`storage`; none of those depend back. The libraries don't know `cmd/expenses` exists.
- **`storage` depends on `expense`** (because it loads/saves `[]expense.Expense`). That's fine — small "data-type at the bottom" packages get imported by many others.

Lesson 08's `cmdAdd` is ~10 lines of orchestration: load → mutate → save → print. The actual logic lives in the libraries; `main.go` just wires them together.

```go
func cmdAdd(path string, args []string) error {
	es, err := storage.LoadExpenses(path)
	if err != nil { return err }
	e := expense.Expense{Date: args[0], /*...*/}
	es = append(es, e)
	if err := storage.SaveExpenses(path, es); err != nil { return err }
	fmt.Println("added:", e.Format())
	return nil
}
```

The same load → operate → save shape repeats in `cmdList` and `cmdSummary`.

**Common mistake.** Putting business logic inside `main.go` instead of a library. Example: writing the bar-chart math directly in `cmdSummary` instead of `summary.Bar`. Two problems: (1) the logic isn't unit-testable in isolation; (2) it can't be reused. Fix: put the logic in `summary.Bar` and have `main.go` just call it.

```go
fmt.Printf("  %s: %s\n", k, summary.Bar(v, maxTotal, 8))
```

`main.go` shrinks; the logic lives in a tested, reusable package.

### End-to-end integration testing

Unit tests on individual functions don't catch wiring bugs — passing `args[1]` where you meant `args[2]`, or forgetting to call `storage.SaveExpenses` after appending. Integration tests drive the actual binary against a real (temp) file, verifying the wiring works.

The pattern — build the binary once per test, run it with args, assert on stdout / stderr / file contents:

```go
type runFn func(args ...string) (stdout, stderr string, err error)

func buildAndRun(t *testing.T) runFn {
	t.Helper()
	binDir := t.TempDir()                  // isolated temp dir for the binary
	binPath := filepath.Join(binDir, "expenses")
	if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return func(args ...string) (string, string, error) {
		cmd := exec.Command(binPath, args...)
		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
}
```

Four parts:

- **`t.TempDir()`** — Go's testing package gives every test its own temp directory, auto-cleaned at test end. No leftover files, no test pollution.
- **`exec.Command("go", "build", "-o", binPath, ".")`** — build the binary into the temp dir.
- **`exec.Command(binPath, args...)`** — run the binary with subcommand args.
- **Assertions on stdout** — `strings.Contains` is fine for "does the output mention X?". Don't assert on whitespace-sensitive bar-chart output.

**Common mistake.** Asserting on full whitespace-sensitive output:

```go
// Wrong — Unicode block chars + variable-width spacing
if stdout != "3 expenses, total €40.00\nby category:\n  coffee     €4.50    █▌\n..." {
	t.Errorf("got %q", stdout)
}
```

Terminal/font rendering can change which Unicode characters print; the spacing in `Bar` output is whitespace-sensitive. Fix: assert that specific *substrings* appear. The integration test's job is "does the CLI wire the right things together?" — not "is the bar chart pixel-perfect?". Bar-chart correctness is `summary.Bar`'s unit-test responsibility.

## Exercise: warm-up

In `exercises/warmup/`:

- `clock/clock.go` — implement `Now() string` returning the current UTC time formatted as `"2006-01-02 15:04:05"` via `time.Now().UTC().Format(...)`.
- `clock/clock_test.go` — fill in the skeleton (assert the FORMAT only — length 19, separators at the right positions; not the actual time).
- `cmd/timestamp/main.go` — already wired; imports `clock`, calls `clock.Now()`, prints. Verify by running it.

```bash
cd lessons/08-capstone/exercises
go test ./warmup/clock/... -v
go run ./warmup/cmd/timestamp
```

## Exercise: main

Three packages to implement + one orchestration binary:

- **`expense/expense.go`** — `Expense` struct + `Format()` + `IsHigh()`. Same as lesson 07; you can copy from your work there. Note: the struct has JSON tags (`` `json:"date"` `` etc) so the storage package's `encoding/json` calls map fields to lowercase JSON keys.
- **`summary/summary.go`** — three pure functions:
  - `TotalsByCategory(es []expense.Expense) map[string]float64` — same as lesson 06/07.
  - `BiggestCategory(totals map[string]float64) (name string, total float64)` — biggest by total; alphabetical tie-break; `("", 0)` for empty.
  - `Bar(value, max float64, width int) string` — Unicode block characters proportional to `value/max`. `█` (full) and `▌` (half). Returns `""` for empty/zero inputs.
- **`storage/storage.go`** — **already provided**; treat as a black box. We'll see how it works in Phase 2 (lesson 13).
- **`cmd/expenses/main.go`** — fill in `cmdAdd`, `cmdList`, `cmdSummary`. The parsing/dispatch scaffolding is provided; you implement the per-subcommand bodies (hints are in the doc comments).

**Plus skeleton tests** in `expense/`, `summary/`, AND `cmd/expenses/main_test.go` (integration tests). Same fill-in-the-cases pattern as lessons 04-07. The integration test scaffolding (`buildAndRun` helper) is provided; you fill in the per-test logic.

> A note on `make test-lesson LESSON=08-capstone`: lesson 08's exercise tests ship with empty cases (and the integration tests have placeholder bodies), so the make output may show `ok` (vacuous pass) before you add cases. Run `go test -v` and look for sub-tests; if there are none, you haven't written assertions yet.

## How to run

```bash
cd lessons/08-capstone/exercises

# Tests (warmup + each subpackage + integration)
go test ./...

# The capstone binary (panics until you implement the stubs)
go run ./cmd/expenses -file=/tmp/x.json add 2026-05-18 4.50 coffee
go run ./cmd/expenses -file=/tmp/x.json add 2026-05-18 12.00 lunch
go run ./cmd/expenses -file=/tmp/x.json list
go run ./cmd/expenses -file=/tmp/x.json summary
```

Once everything's implemented, `go run ... summary` produces something like:

```
3 expenses, total €40.00
by category:
  coffee     €4.50    █▌
  groceries  €23.50   ████████
  lunch      €12.00   ████
biggest category: groceries (€23.50)
```

Daily habits:

```bash
gofmt -w ./...    # format
go vet ./...      # static analysis
```

Both run by CI.

## Phase 1 — complete!

You've come a long way. Across eight lessons:

- **01 Hello** — `go run`, `package main`, `fmt`.
- **02 Variables** — types, zero values, arithmetic, conversions.
- **03 Control flow** — `if`, `for`, `switch`, early returns.
- **04 Functions & first tests** — multi-return, variadic, `defer`, `errors`, table tests.
- **05 Slices and maps** — composite types and idioms.
- **06 Structs and methods** — own types + behaviour.
- **07 Packages and modules** — splitting code, module paths, tooling tour.
- **08 Capstone** — composing it all into a real CLI.

You can read, write, and test idiomatic Go for small programs. Phase 2 (lessons 09-15) goes deeper into the language: pointers, interfaces, error design, generics, std-lib literacy.

## Going further

### Read

- [Effective Go — Names](https://go.dev/doc/effective_go#names) — package and identifier naming, revisited.
- [Go blog — Organizing Go code](https://go.dev/blog/organizing-go-code) — package organisation patterns.
- [Standard project layout](https://github.com/golang-standards/project-layout) — a community convention (note: it's a convention, not an official Go standard).

### Try

- **Edit / delete an expense.** Add `edit INDEX [DATE] [AMOUNT] [CATEGORY]` and `delete INDEX` subcommands. Watch how the dispatch in `run()` grows — does it suggest a refactor? (Phase 2's `flag` package would help.)
- **A `summary -days=N` flag.** Filter the summary to the last N days. You'd need to parse dates as `time.Time` (lesson 14 territory) to do comparisons. Try the string-prefix shortcut first.
- **Sort the list.** Add `list -sort=date` and `list -sort=amount` flags. Uses `sort.Slice` (lesson 14 territory). Stretch.
- **Colour output.** Use ANSI escape codes to colour the bar chart by category. Tiny but satisfying.
- **A web UI.** Use `net/http` (lesson 23) to serve a minimal HTML page that lists the expenses and shows the summary chart. Long-stretch for Phase 4.
