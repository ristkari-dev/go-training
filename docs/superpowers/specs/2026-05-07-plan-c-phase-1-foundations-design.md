# Plan C — Phase 1 (Foundations) Design

**Status:** Approved (brainstorming complete, awaiting implementation plan)
**Date:** 2026-05-07
**Owner:** Aki Ristkari
**Master design:** [`2026-05-05-go-course-design.md`](2026-05-05-go-course-design.md)

## Summary

Plan C delivers Phase 1 of the Go training course: lessons 1-8, taking absolute beginners from "what is `go run`" to a small but real multi-package CLI. Each lesson uses a focused per-lesson example chosen to foreshadow the Phase 1 capstone (a personal expense tracker). Lessons 1-7 use a flat exercise layout; lesson 7 introduces packages and lesson 8's capstone uses subfolders. Heavy-explanatory slides serve both live lectures and self-study. Plan C also ships the small platform changes Phase 1 needs: an updated lesson scaffolder, an updated slides skeleton, and a provided `storage` package for the capstone.

## Audience and pacing

- Phase 1 students: software engineering students with no prior production language experience.
- Each lesson: ~90 minutes live + ~60-90 minutes self-study.
- 8 lessons total. Roughly half a semester of weekly sessions.

## Core decisions (locked in during brainstorming)

| Decision | Choice |
|---|---|
| Cohesion across lessons | Per-lesson examples + capstone in lesson 8 |
| Exercise format | Warm-up (5-10 min) + main (30-45 min) per lesson |
| Slide style | Heavy-explanatory (textbook-flavoured) |
| Capstone | Personal expense tracker CLI with JSON persistence |
| Persistence approach | Provided `storage` package (students treat as black box) |
| Layout — lessons 1-6 | Flat (`exercises/warmup.go` + `exercises/main.go` in same package) |
| Layout — lesson 7 | Subfolders introduced (the act of creating them IS the exercise) |
| Layout — lesson 8 | Subfolders + `cmd/` (capstone) |
| Tests as spec, students start writing tests | Lesson 4 |
| Tooling thread first appearance | `gofmt` mention lesson 1; `go vet` lesson 4; full tour lesson 7 |
| Imports through Phase 1 | Stdlib only |

## Curriculum

Each lesson below has the same shape: concept(s), per-lesson example that foreshadows the capstone, warm-up exercise, main exercise.

### Lesson 1 — Hello, Go

- **Concepts:** install + verify, `go run`, `go mod init`, `main` function, importing `fmt`, basic `Println` / `Printf` formatting.
- **Per-lesson example:** A program that prints `Hello! You spent €23.50 on coffee today.` Hardcoded values; the student edits strings/numbers and re-runs.
- **Warm-up:** Print `"Hello, Go!"`; then read a name from input and print `"Hello, <name>!"`.
- **Main:** Extend the example to print three expenses with `Printf` formatting (decimals, alignment).
- **New imports introduced:** `fmt`.

### Lesson 2 — Variables, types, operators

- **Concepts:** typed declarations (`var x int = …`), short-form (`x := …`), zero values, basic numeric/string types, type conversions, constants, `iota`.
- **Per-lesson example:** Declare typed variables for one expense (`date string`, `amount float64`, `category string`); compute the per-day average and a 14% tip with arithmetic operators.
- **Warm-up:** Declare and print variables of every basic type with their zero values; convert between them (`int → float64`, `string → []byte`, etc.).
- **Main:** Given three hardcoded amounts, compute total / average / max-min using arithmetic and print a small summary table.
- **New imports introduced:** none (stays in `fmt`).

### Lesson 3 — Control flow

- **Concepts:** `if` / `else if` / `else`, `for` (the only loop, in its three forms), `switch` (with a `switch v.(type)` teaser as an out-of-scope reference), early returns.
- **Per-lesson example:** Categorise an amount: under €10 → "snack", €10-50 → "regular", over €50 → "splurge". Read amounts in a loop with `fmt.Scanln`.
- **Warm-up:** Classify an integer as positive/negative/zero with `if`; FizzBuzz for 1-20 with `for`.
- **Main:** Read N amounts in a `for` loop, classify each, count per category, print a summary.
- **New imports introduced:** `os`, `strconv`.

### Lesson 4 — Functions & first tests

- **Concepts:** multi-return, named returns, variadic, `defer`. `*_test.go` files. Table tests via the `[]struct{}` pattern. The canonical `if err != nil { return …, err }` shape.
- **Per-lesson example:** Extract `Categorise(amount float64) string` and `FormatExpense(date string, amount float64, cat string) string` from lesson 3's example. Write tests.
- **Warm-up:** Implement `Add(a, b int) int` and `MinMax(xs ...int) (int, int, error)`, write tests for both.
- **Main:** Implement and test `Categorise` and `FormatExpense` using table tests.
- **New imports introduced:** `errors`, `testing` (in test files).
- **Tooling thread:** `go vet ./...` introduced.

### Lesson 5 — Composite types I: arrays, slices, maps

- **Concepts:** array vs slice, slice internals (len/cap), `append`, `range`, slice-of-slice idioms, maps (zero value, lookup, `_, ok := m[k]` exists check, deletion).
- **Per-lesson example:** A `[]float64` of expense amounts and a parallel `[]string` of categories. Compute sum/average via `range`. Build a `map[string]float64` of category totals.
- **Warm-up:** Implement `Sum(xs []int) int`, `Max(xs []int) int`, `Unique(xs []string) []string` with tests.
- **Main:** Given parallel slices `(amounts, categories)`, return a `map[string]float64` of totals per category. Reject empty input with a clear error.
- **New imports introduced:** none (parallel-slice-only design avoids `sort` for now).

### Lesson 6 — Composite types II: structs & methods

- **Concepts:** struct definition, struct literal forms, value vs pointer receivers (favouring value receivers for now), struct embedding (light), exported vs unexported fields (light — full treatment lesson 7).
- **Per-lesson example:** `type Expense struct { Date string; Amount float64; Category string }`. Methods: `(e Expense) Format() string`, `(e Expense) IsHigh() bool` (over €50). Refactor lesson 5's parallel-slice logic to take `[]Expense`.
- **Warm-up:** Define `Point{X, Y float64}` and `DistanceFromOrigin() float64`; tests.
- **Main:** Define `Expense`, implement `Format` and `IsHigh`, write `TotalsByCategory(es []Expense) map[string]float64`. Tests for everything.
- **New imports introduced:** none.

### Lesson 7 — Packages & modules

- **Concepts:** splitting code into packages, exported vs unexported (formal), `go.mod` revisited, imports, brief tour of relevant std-lib packages (`os`, `bufio`, `strings`, `strconv`), `gofmt`, `go vet`, `go doc`.
- **Per-lesson example:** Refactor lesson 6's `Expense` + methods code into a proper `expense` package. The main file imports it.
- **Warm-up:** Take a small `Greet*` helper from earlier, split it into a `greet` subpackage and import it.
- **Main:** Refactor lesson 6's `Expense` code into `exercises/expense/` subfolder + `exercises/main.go` that imports it.
- **Layout note:** This is where students first encounter subfolders — the act of creating them is the exercise.
- **New imports introduced:** `bufio`, `strings`.

### Lesson 8 — Phase 1 capstone — expense tracker CLI

- **Concepts:** integrating multiple packages, `cmd/` convention, using a provided package as a black-box dependency, end-to-end CLI design.
- **Per-lesson example:** The capstone itself.
- **Warm-up:** A small "wire two packages and a main" exercise — a tiny `clock` package providing `Now()` plus a `cmd/timestamp` that imports it.
- **Main:** The capstone — `add`/`list`/`summary` subcommands, JSON-backed file persistence via the provided `storage` package, a text bar chart in the summary output. See "Capstone shape" section below.
- **Layout note:** Subfolders + `cmd/`.
- **New imports introduced:** `encoding/json` (only inside the provided `storage` package — students treat it as opaque).

## Capstone shape (lesson 8)

### What the user sees

```
$ expenses add 2026-05-07 4.50 coffee
added: 2026-05-07  €4.50  coffee

$ expenses add 2026-05-07 12.00 lunch
added: 2026-05-07  €12.00  lunch

$ expenses list
2026-05-06  €23.50  groceries
2026-05-07  €4.50   coffee
2026-05-07  €12.00  lunch

$ expenses summary
3 expenses, total €40.00
by category:
  coffee     €4.50   ▌
  groceries  €23.50  ███████▌
  lunch      €12.00  ████
biggest category: groceries (€23.50)
```

Three subcommands — `add`, `list`, `summary` — driven by `os.Args`. Persistence reads/writes a JSON file (default `~/.expenses.json`, overridable via `-file=path`).

### Package layout

```
lessons/08-capstone/exercises/
├── cmd/
│   └── expenses/
│       └── main.go              # parses os.Args, dispatches subcommand
├── expense/
│   ├── expense.go               # Expense struct + Format() / IsHigh()
│   │                            # (carried forward from lesson 7)
│   └── expense_test.go
├── summary/
│   ├── summary.go               # TotalsByCategory, BiggestCategory, Bar
│   └── summary_test.go
├── storage/                     # PROVIDED — students don't write or modify
│   └── storage.go               # LoadExpenses(path), SaveExpenses(path, []expense.Expense)
└── warmup/
    ├── clock/
    │   └── clock.go
    └── cmd/
        └── timestamp/
            └── main.go
```

`solutions/` mirrors the same layout exactly.

### What the student writes

1. **`expense/`** — copied forward from their lesson 7 work. Lesson 8's tests are slightly stricter (e.g., `Format` must produce a specific column-aligned form).
2. **`summary/`** — three pure functions (no I/O):
   - `TotalsByCategory(es []expense.Expense) map[string]float64`
   - `BiggestCategory(totals map[string]float64) (name string, total float64)`
   - `Bar(value, max float64, width int) string` — produces a string like `"████▌"`.
3. **`cmd/expenses/main.go`** — orchestration: parse `os.Args`, branch on subcommand, call `storage.LoadExpenses` / `storage.SaveExpenses` and the `summary` functions.

### What's provided

- **`storage/storage.go`** — fully implemented (~50 lines). `LoadExpenses(path string) ([]expense.Expense, error)` returns an empty slice + nil error when the file doesn't exist; `SaveExpenses(path string, es []expense.Expense) error` writes pretty-printed JSON via `json.MarshalIndent`. Top-of-file comment: "We'll see how this works in Phase 2 (lesson 13). Treat it as a black box for now."
- **A starter `expense/expense.go`** — empty stubs ready for the student to copy in their lesson 7 code.
- **The full failing test suite** — `expense_test.go`, `summary_test.go`, plus a handful of `cmd/expenses/main_test.go` integration tests that drive the binary via `os/exec` against a temp JSON file.

### Capstone non-scope

- Edit / delete an expense (only `add` / `list` / `summary`).
- Date-range filtering from the CLI.
- Sorting options.
- Colour output.
- Configuration files.
- A "current month" default for `summary`.
- Locking around the JSON file (single-process assumption).
- `flag` package usage — `os.Args` is enough for three subcommands.
- Interfaces.
- Errors deeper than `if err != nil { return err }`.

These appear in lesson 8's "Going further" section as suggestions for self-driven extension.

## Lesson anatomy

### Folder layout (lessons 1-6)

```
lessons/NN-name/
├── README.md
├── slides/
│   ├── index.html
│   ├── slides.md
│   └── assets/
├── exercises/
│   ├── warmup.go
│   ├── warmup_test.go
│   ├── main.go
│   └── main_test.go
└── solutions/
    ├── warmup.go
    ├── warmup_test.go
    ├── main.go
    └── main_test.go
```

Both `warmup.go` and `main.go` live in the same Go package (`exercises` or `solutions`). The warm-up's exported names use a `Warmup*` prefix to avoid collisions with the main exercise's identifiers (e.g., `WarmupAdd`, `WarmupGreet`); the main exercise uses unprefixed names (`Categorise`, `Expense`, `Format`).

### Folder layout (lesson 7)

Subfolders enter for the first time. Starter intentionally has stubs the student must fill in:

```
lessons/07-packages/exercises/
├── main.go                  # imports the expense subpackage
├── main_test.go
├── expense/
│   ├── expense.go           # stub: package declaration + TODO
│   └── expense_test.go
└── warmup/
    ├── greet/
    │   └── greet.go
    └── cmd/
        └── greet/
            └── main.go
```

### Folder layout (lesson 8 — capstone)

Already shown in the "Capstone shape" section above.

### `README.md` structure (every lesson)

1. **Learning goals** — 3-5 bullets.
2. **Prerequisites** — links to earlier lessons.
3. **Concepts** — full prose mirroring the deck. For Phase 1, this is several paragraphs per concept, not a recap.
4. **Exercise: warm-up** — what to build, expected `go test` output.
5. **Exercise: main** — same shape, larger.
6. **How to run** — exact `go test` commands for both warm-up and main.
7. **Going further** — 2-3 links + 1-2 stretch problems (no solutions provided for stretch).

### Slide structure (heavy-explanatory pattern)

Each lesson's `slides.md` follows a repeating per-concept pattern:

1. **Title slide** — lesson number, title, one-line learning goal.
2. **What we'll cover** — bullet list of concepts.
3. **For each concept** (typically 3-5 per lesson):
   - **Motivation** — why this concept exists.
   - **The basics** — minimal code introducing the concept.
   - **A worked example** — substantive code using the concept in context.
   - **Common mistake** — what NOT to do, with the bug the compiler/runtime would surface.
   - **Recap** — bullet list of takeaways.
4. **Practice** — points at exercises folder; names the warm-up and main.
5. **What we learned** — phase-end recap across all concepts.
6. **Up next** — pointer to the next lesson.

Estimated 25-35 slides per lesson. Speaker notes (`Note:` blocks) carry "what to actually say in the live lecture" — the slide prose is the self-study text.

## Platform updates (Plan C tasks 1-3, before lessons are authored)

### Update the lesson scaffolder template

`tools/new-lesson/template/` currently produces a single-exercise layout. Add the warm-up files alongside the existing main files:

```
tools/new-lesson/template/
├── README.md.tmpl                         # updated: warm-up + main exercise sections
├── slides/
│   ├── index.html.tmpl                    # unchanged
│   ├── slides.md.tmpl                     # rewritten: heavy-explanatory skeleton
│   └── assets/.gitkeep                    # unchanged
├── exercises/
│   ├── warmup.go.tmpl                     # NEW
│   ├── warmup_test.go.tmpl                # NEW
│   ├── main.go.tmpl                       # unchanged
│   └── main_test.go.tmpl                  # unchanged
└── solutions/
    ├── warmup.go.tmpl                     # NEW
    ├── warmup_test.go.tmpl                # NEW
    ├── main.go.tmpl                       # unchanged
    └── main_test.go.tmpl                  # unchanged
```

The warm-up template uses a `WarmupGreet`-prefixed name; the main uses `Greet`. The naming convention is documented in `CONTRIBUTING.md`.

The `slides.md.tmpl` rewrite pre-stocks the heavy-explanatory pattern — three skeletal "Concept N" blocks with Motivation/Basics/Worked example/Common mistake/Recap subsections — so authors fill in content rather than designing structure.

The scaffolder Go code (`tools/new-lesson/main.go`) doesn't change — it walks any `.tmpl` files it finds.

The scaffolder also updates the existing scaffolder tests to reflect the new template files (`TestScaffoldCreatesExpectedTree` gets four extra paths in its `want` list).

### The provided `storage` package (lesson 8 only)

```
lessons/08-capstone/exercises/storage/storage.go
lessons/08-capstone/solutions/storage/storage.go
```

Byte-identical files. ~50 lines. Top-of-file comment explaining "treat as black box; covered in lesson 13."

`storage` imports the lesson's own `expense` package via the full module path:

```go
import "github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
```

This long import is itself a small teaching moment about module paths; the slides briefly show it.

### Documentation updates

`CONTRIBUTING.md` (top-level) gets a new section "Phase 1 conventions":

- Warm-up + main exercise pattern with the `Warmup*` naming convention.
- Heavy-explanatory slide pattern (per-concept Motivation/Basics/etc.).
- "Tests are the spec" extended: students start writing tests at lesson 4.
- Import-limited-to-stdlib rule for Phase 1.
- Pointer to this design doc and the Plan C document.

### What's NOT changing

- `Makefile` — `go test ./...` already covers warm-up files; no new targets needed.
- `tools/slides-dev` — still serves any lesson's deck regardless of how many concepts.
- `tools/build-index` — still walks `lessons/*/slides/`.
- The deploy pipeline.
- The root `go.mod` — no new dependencies.

## Cross-cutting threads

### Testing thread

- **Lessons 1-3:** Failing tests ship in `*_test.go`. Students don't write tests themselves — they edit the `.go` files until tests pass.
- **Lesson 4:** Testing introduced formally — filename convention, `func TestXxx(t *testing.T)`, `t.Errorf` vs `t.Fatalf`, table tests via `[]struct{}`. Lesson 4's main exercise is the first time students *author* test code.
- **Lessons 5-8:** Exercises ship with skeleton tests (some empty `t.Run` blocks) that students must complete in addition to writing implementation code.

Deferred to Phase 2 (lesson 15): subtests with `t.Run` for parameterised cases beyond table tests, `t.Helper`, `t.TempDir`, golden files, fuzzing, benchmarks.

### Tooling thread

- **Lesson 1:** `gofmt -w .` mentioned as a habit.
- **Lesson 4:** `go vet ./...` introduced; appears in every lesson README's "How to run" from this point.
- **Lesson 7:** Formal tour: `gofmt`, `go vet`, `go doc <pkg>`, `go list ./...`, `go mod tidy`. One slide per tool.
- **Lesson 8:** `golangci-lint` mentioned, `make lint` shown. The detailed `.golangci.yml` walk-through waits for Phase 2.

### "Going further" thread

Per-lesson README has a `## Going further` section with two parts:

- **Read** — 1-3 short links (Go blog posts, std-lib docs, occasional book references).
- **Try** — 1-2 stretch problems harder than the main exercise. Self-graded — no solutions in `solutions/`. README explicitly notes this.

### Identifier-naming thread

- Full-word names: `TotalsByCategory`, not `tbc`; `Categorise`, not `cat`.
- Idiomatic short names where Go itself uses them — `i` for loop indices, `r io.Reader`, `err error`. Lesson 4 calls this out explicitly when std-lib argument names appear.
- Test names: `TestFunctionName` for simple tests; subtests within table tests use the input as the subtest name.

### Error-handling thread

- Lessons 1-3: no errors (no failure modes worth handling).
- Lesson 4: introduces `error` as a return type via `MinMax(xs ...int) (int, int, error)` (empty input → error). The canonical `if err != nil { return ..., err }` pattern.
- Lessons 5-8: students return errors, callers check them.

Deferred to Phase 2 (lesson 11): sentinel errors, error wrapping with `%w`, `errors.Is` / `errors.As`, custom error types, error design philosophy.

### Imports thread

- Stdlib only through all of Phase 1.
- New packages introduced lesson by lesson, summarised on a slide near the start of the lesson.

## Out of scope (Phase 1 boundaries)

### Deferred to Phase 2 (lessons 9-15)

- Pointers (lesson 9). Phase 1 uses value receivers exclusively in lesson 6; slices behave "pointer-like" via their header. The pointer concept itself doesn't appear.
- Interfaces (lesson 10). No `interface{}`/`any` in Phase 1 code.
- Errors deeper than `if err != nil { return err }` (lesson 11).
- Generics (lesson 12). All Phase 1 functions are concrete on argument types.
- JSON / encoding deeper than the provided black box (lesson 13).
- Std-lib literacy beyond a tiny core (lesson 14).
- Idiomatic `cmd/`/`internal/` formalisation (lesson 15). Phase 1 uses `cmd/` minimally for the capstone but doesn't formalise the convention.
- Advanced testing — subtests as a structural pattern, `t.Helper`, `t.TempDir`, golden files (lesson 15).

### Deferred to Phase 3 (concurrency)

- Goroutines, channels, `select`, `sync`, `context`. Phase 1 does the synchronous equivalent everywhere.
- Networking, syscalls beyond `os.Open`/`os.ReadFile`.
- Profiling, benchmarking, fuzzing.

### Deferred to Phase 4 (production)

- HTTP servers / clients. gRPC. Configuration / secrets / signal handling. Observability. Containerisation of student code.

### Hard non-goals (not arriving later either, in Phase 1 framing)

- No third-party dependencies — root `go.mod`'s `require` block stays empty through all of Phase 1.
- No web framework. No databases. No GUI.
- No "good practice" lectures detached from code.
- No automated grading beyond `go test`.
- No discussion of generics/interfaces as forward references except a one-line mention if a student asks.

## Open items deferred to implementation planning

- Exact wording for the "treat as black box" comment in `storage/storage.go`.
- Concrete "Going further" link selection per lesson (the reading list).
- Whether the capstone's `cmd/expenses/main_test.go` integration tests use `go run` or `go build` + exec (defaults to `go build` + exec in `t.TempDir()` for speed).
- Slide deck length per lesson — target is 25-35 but actual depends on how much each concept needs.
