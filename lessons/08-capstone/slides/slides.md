<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">08</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 1 — Foundations · Capstone</div>
<h1>Expense tracker CLI</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Compose everything from Phase 1 into a real CLI: three subcommands (<code>add</code>, <code>list</code>, <code>summary</code>) backed by JSON file persistence, a text bar chart in the summary, and integration tests that drive the binary via <code>os/exec</code> against a temp file.</p>
</div>
</div>
</div>

---

## What we'll cover

- The **`cmd/` convention** — `package main` lives at `cmd/<binaryname>/main.go`; the directory name becomes the binary name.
- **Composing multiple packages into a CLI** — the orchestration `main.go` calls into focused subpackages (`expense`, `summary`, `storage`).
- **End-to-end integration testing** — `os/exec` + `t.TempDir()` to drive the binary against an isolated temp JSON file.

---

## Concept 1: The `cmd/` convention

### Motivation

Lesson 07 introduced packages and subpackages. Real-world Go modules with binaries put each binary in its own subdirectory under `cmd/` — the directory name becomes the binary's name. A single module can host multiple binaries this way without polluting the root.

---

### The basics

```
yourmodule/
├── go.mod                   (module github.com/you/yourmodule)
├── expense/
│   └── expense.go           (package expense — library code)
├── summary/
│   └── summary.go           (package summary — library code)
└── cmd/
    ├── expenses/
    │   └── main.go          (package main — the "expenses" binary)
    └── another-tool/
        └── main.go          (package main — a second binary)
```

Three things to internalise:

- **`cmd/` is a convention, not a Go language feature.** Nothing requires it; the Go tooling treats `cmd/foo/main.go` identically to `foo/main.go`. The convention is purely about discoverability — every reader of the module knows "the binaries live under `cmd/`."
- **Build and run by directory:**

  ```bash
  go build -o ./bin/expenses ./cmd/expenses
  go run ./cmd/expenses summary
  go install ./cmd/expenses     # installs to $GOPATH/bin
  ```

- **The binary name is the directory name.** `go build ./cmd/expenses` produces a binary called `expenses`. Renaming via `-o` works, but follow the convention if you can.

---

### A worked example

Lesson 08's full `exercises/` shape (collapsing test files for clarity):

```
lessons/08-capstone/exercises/
├── warmup/
│   ├── clock/clock.go        package clock — library
│   └── cmd/
│       └── timestamp/main.go  package main — tiny binary
├── expense/expense.go         package expense — library
├── summary/summary.go         package summary — library
├── storage/storage.go         package storage — library (PROVIDED)
└── cmd/
    └── expenses/main.go       package main — the capstone binary
```

Two binaries — the tiny `cmd/timestamp` (warmup) and the main `cmd/expenses` (capstone) — siblings under `cmd/`. Each library subpackage exposes its own focused API; the binaries compose them.

To run both:

```bash
go run ./lessons/08-capstone/solutions/warmup/cmd/timestamp
go run ./lessons/08-capstone/solutions/cmd/expenses summary
```

---

### Common mistake

Forgetting that `cmd/expenses/` is a *directory*, not a file. New Go programmers sometimes try:

```bash
go run cmd/expenses.go        # wrong — there's no file by that name
```

Or:

```go
package main                   // in a file named cmd/expenses.go at the module root
```

`cmd/` is a directory. The actual `main.go` file lives **inside** the per-binary subdirectory. Always:

```
cmd/<binaryname>/main.go      ✓
cmd/<binaryname>.go            ✗
```

---

### Recap

- `cmd/<binaryname>/main.go` — the convention for binary entry points.
- Multiple binaries are siblings under `cmd/`.
- The binary name is the directory name (build/run by directory, not by file).
- Convention, not language feature — but follow it.

---

## Concept 2: Composing multiple packages into a CLI

### Motivation

A real CLI's `main.go` is *orchestration code* — it parses CLI arguments and delegates the actual work to focused subpackages. The "deps go one way" rule keeps the architecture sane: the orchestration depends on the libraries, the libraries don't depend on the orchestration.

---

### The basics

Lesson 08's dependency graph:

```
cmd/expenses/main.go  ─┬─→  expense    (Expense + Format + IsHigh)
                       ├─→  summary    (TotalsByCategory + BiggestCategory + Bar)
                       └─→  storage    (LoadExpenses + SaveExpenses)
                              └─→  expense   (storage knows about Expense too)
```

Three rules:

- **`main.go` is the only `package main`.** Everything else is a library — testable in isolation, reusable across binaries.
- **Dependencies flow one direction.** `cmd/expenses` depends on `expense`/`summary`/`storage`; none of those depend back on `cmd/expenses`. The libraries don't even know `cmd/expenses` exists.
- **`storage` depends on `expense`** (because it loads/saves `[]expense.Expense`). That's fine — small "data type at the bottom" packages get imported by many others. `expense` doesn't depend on anything inside the lesson.

---

### A worked example

The shape of `cmd/expenses/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/storage"
	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/summary"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point. It returns an error instead of calling
// os.Exit so integration tests can drive it without process teardown.
func run(args []string) error {
	// Parse -file=path option.
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}
	switch args[0] {
	case "add":
		return cmdAdd(path, args[1:])
	case "list":
		return cmdList(path)
	case "summary":
		return cmdSummary(path)
	}
	return fmt.Errorf("unknown subcommand %q", args[0])
}

func cmdAdd(path string, args []string) error {
	es, err := storage.LoadExpenses(path)            // storage call
	if err != nil { return err }
	e := expense.Expense{Date: args[0], /*...*/}     // expense type
	es = append(es, e)
	if err := storage.SaveExpenses(path, es); err != nil { // storage call
		return err
	}
	fmt.Println("added:", e.Format())                 // method call
	return nil
}
```

`cmdAdd` is ~10 lines of orchestration: load → mutate → save → print. The actual logic (what an `Expense` is, how to format it, how JSON gets to disk) lives in the libraries — `main.go` just wires them together.

The same pattern repeats for `cmdList` (load → print each) and `cmdSummary` (load → totals → bar chart → biggest).

---

### Common mistake

Putting business logic inside `main.go` instead of a library subpackage. Example: writing the bar-chart logic directly in `cmdSummary` instead of `summary.Bar`:

```go
// Wrong — bar chart embedded in cmd/expenses/main.go
func cmdSummary(path string) error {
	// ...
	for k, v := range totals {
		units := v / maxTotal * 8                 // logic in main.go
		full := int(units)
		bar := strings.Repeat("█", full)
		// ... 20 more lines ...
	}
}
```

Two problems: (1) the logic isn't unit-testable in isolation (every test has to drive `main`); (2) it can't be reused from another binary. The fix is to put the logic in `summary.Bar` and have `main.go` just call it:

```go
fmt.Printf("  %s: %s\n", k, summary.Bar(v, maxTotal, 8))
```

`main.go` shrinks; the logic lives in a tested, reusable package.

---

### Recap

- `main.go` is orchestration — parse args, call libraries, format output. Keep the logic in the libraries.
- One-way deps: `cmd → libraries`, never the other direction.
- Each library is independently testable and reusable.

---

## Concept 3: End-to-end integration testing

### Motivation

Unit tests on `summary.TotalsByCategory` or `expense.Format` are great, but they don't catch wiring bugs — e.g., passing `args[1]` where you meant `args[2]`, or forgetting to call `storage.SaveExpenses` after appending. **Integration tests** drive the actual binary against a real (temp) file, verifying the wiring works end-to-end.

---

### The basics

The pattern: build the binary once per test, run it with subcommand + args, assert on stdout / stderr / exit code / the file it touched.

```go
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type runFn func(args ...string) (stdout, stderr string, err error)

func buildAndRun(t *testing.T) runFn {
	t.Helper()
	binDir := t.TempDir()                  // 1. isolated temp dir for the binary
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

func TestAddSingle(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	stdout, _, err := run("-file="+tmpFile, "add", "2026-05-18", "4.50", "coffee")
	if err != nil { t.Fatalf("add failed: %v", err) }
	if !strings.Contains(stdout, "added: 2026-05-18  €4.50    coffee") {
		t.Errorf("stdout missing add confirmation: %q", stdout)
	}
	if _, err := os.ReadFile(tmpFile); err != nil {
		t.Errorf("expected JSON file at %s: %v", tmpFile, err)
	}
}
```

Four parts:

- **`t.TempDir()`** — Go's testing package gives every test its own temp directory, auto-cleaned at test end. No leftover files, no test pollution.
- **`exec.Command("go", "build", "-o", binPath, ".")`** — build the binary into the temp dir. The `.` means "the package in the current directory" (which is `cmd/expenses` because the test file lives there).
- **`exec.Command(binPath, args...)`** — run the binary with subcommand args. `os/exec` lets you capture stdout/stderr separately.
- **Assertions on stdout** — `strings.Contains` is fine for "does the output mention X?" checks. Don't assert on whitespace-sensitive bar chart output — terminal rendering differs.

---

### A worked example

Test the full add → list → summary flow:

```go
func TestAddThenSummary(t *testing.T) {
	run := buildAndRun(t)
	tmpFile := filepath.Join(t.TempDir(), "expenses.json")

	// Set up: add three expenses.
	for _, args := range [][]string{
		{"add", "2026-05-18", "4.50", "coffee"},
		{"add", "2026-05-18", "12.00", "lunch"},
		{"add", "2026-05-18", "23.50", "groceries"},
	} {
		if _, stderr, err := run(append([]string{"-file=" + tmpFile}, args...)...); err != nil {
			t.Fatalf("setup add: %v\n%s", err, stderr)
		}
	}

	// Run summary and check key lines.
	stdout, _, err := run("-file="+tmpFile, "summary")
	if err != nil { t.Fatalf("summary: %v", err) }
	want := []string{
		"3 expenses, total €40.00",
		"biggest category: groceries (€23.50)",
	}
	for _, w := range want {
		if !strings.Contains(stdout, w) {
			t.Errorf("summary output missing %q\nfull stdout:\n%s", w, stdout)
		}
	}
}
```

`strings.Contains` on a couple of stable lines is enough — if any of the assertions fail, the wiring is wrong somewhere.

---

### Common mistake

Asserting on whitespace-sensitive output:

```go
// Wrong — Unicode block chars + variable-width spacing
if stdout != "3 expenses, total €40.00\nby category:\n  coffee     €4.50    █▌\n  ..." {
	t.Errorf("got %q", stdout)
}
```

Two problems: (1) terminal/font rendering can change which Unicode characters print; (2) the spacing in `summary.Bar` output is whitespace-sensitive (`%-7.2f` width). Tests that match the full output are fragile.

Fix: assert that specific *substrings* appear. The integration test's job is "does the CLI wire the right things together?" — not "is the bar chart pixel-perfect?". The bar-chart correctness is `summary.Bar`'s unit-test responsibility.

---

### Recap

- Integration tests build the binary once, run it many times against a temp file.
- `t.TempDir()` gives each test isolated, auto-cleaned scratch space.
- `os/exec` captures stdout/stderr; check key substrings, not full output.
- Use integration tests to catch wiring bugs that unit tests miss; use unit tests for the deep correctness of individual functions.

---

## Phase 1 capstone — what you've built

After lesson 08 you have a working CLI:

```bash
$ expenses add 2026-05-18 4.50 coffee
added: 2026-05-18  €4.50    coffee

$ expenses add 2026-05-18 12.00 lunch
added: 2026-05-18  €12.00   lunch

$ expenses add 2026-05-18 23.50 groceries
added: 2026-05-18  €23.50   groceries

$ expenses list
2026-05-18  €4.50    coffee
2026-05-18  €12.00   lunch
2026-05-18  €23.50   groceries

$ expenses summary
3 expenses, total €40.00
by category:
  coffee     €4.50    █▌
  groceries  €23.50   ████████
  lunch      €12.00   ████
biggest category: groceries (€23.50)
```

JSON-backed (`~/.expenses.json` by default), three subcommands, integration-tested. About 200 lines of code total, organised into four packages.

---

## Practice

### Warm-up

In `exercises/warmup/`:

- `clock/clock.go` — implement `Now() string` returning the current UTC time formatted as `"2006-01-02 15:04:05"` via `time.Now().UTC().Format(...)`.
- `clock/clock_test.go` — fill in the skeleton (assert the format only, not the actual time).
- `cmd/timestamp/main.go` — already wired; just imports `clock` and prints.

```bash
cd lessons/08-capstone/exercises
go test ./warmup/clock/... -v
go run ./warmup/cmd/timestamp
```

---

### Main

Three subpackages to implement + one orchestration binary:

- `expense/expense.go` — `Expense` struct + `Format()` + `IsHigh()`. Same as lesson 07; you can copy.
- `summary/summary.go` — `TotalsByCategory(es []expense.Expense)`, `BiggestCategory(totals)`, `Bar(value, max, width)`.
- `storage/storage.go` — **already provided**; treat as a black box.
- `cmd/expenses/main.go` — fill in `cmdAdd`, `cmdList`, `cmdSummary` (the parsing/dispatch scaffolding is provided).

Skeleton tests in `expense/`, `summary/`, AND `cmd/expenses/main_test.go` (integration tests).

```bash
cd lessons/08-capstone/exercises
go test ./...                    # all four test files
go run ./cmd/expenses -file=/tmp/x.json add 2026-05-18 4.50 coffee
go run ./cmd/expenses -file=/tmp/x.json list
go run ./cmd/expenses -file=/tmp/x.json summary
```

Note:
For live: walk through the layout on the projector — show the directory tree first, then walk the dependency graph (cmd/expenses depends on expense + summary + storage; storage depends on expense). The "one-way deps" point lands best when students see the tree side-by-side with the import graph. Demo the `os/exec` integration test pattern — it's the moment students realise "I can test a CLI like any other Go code."

---

## What we learned

- The `cmd/<binaryname>/main.go` convention — binaries as siblings under `cmd/`.
- Composing focused libraries into a CLI orchestration in `main.go`; one-way dependencies.
- `os/exec` + `t.TempDir()` integration testing — drives the binary against an isolated temp file.
- The full Phase 1 toolkit applied: structs (Expense), methods (Format/IsHigh), packages (the four subpackages), errors (`if err != nil { return err }` everywhere in `cmd/expenses`), tests (unit + integration), `gofmt` + `go vet`, and a provided package (storage) treated as a black box.

---

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

You can read, write, and test idiomatic Go for small programs. Phase 2 (lessons 09-15) goes deeper: pointers, interfaces, error design, generics, std-lib literacy.

---

## Up next

Lesson 09 — Pointers. The "&" and "*" you've been carefully avoiding all of Phase 1. (Phase 2 begins.)
